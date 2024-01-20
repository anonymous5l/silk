package silk

import (
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"unsafe"
)

type EncoderOption struct {
	SampleRate            int32
	MaxInternalSampleRate int32
	PacketSize            int32
	BitRate               int32
	PacketLossPercentage  int32
	Complexity            int32
	UseInBandFEC          int32
	UseDTX                int32
}

type Encoder struct {
	opts  *EncoderOption
	state *encoder_state_FIX
}

func NewEncoder(opts *EncoderOption) *Encoder {
	return &Encoder{
		opts:  opts,
		state: initEncoder(),
	}
}

func (e *Encoder) EncodePayload(in []int16, out []byte) (int, error) {
	encControl := e.opts
	psEnc := e.state

	var (
		max_internal_fs_kHz, PacketSize_ms, PacketLoss_perc, UseInBandFEC, UseDTX int32
		ret                                                                       = 0
		nSamplesToBuffer, Complexity, input_10ms, nSamplesFromInput               int32
		TargetRate_bps, API_fs_Hz                                                 int32
		MaxBytesOut                                                               int16
	)
	if ((encControl.SampleRate != 8000) &&
		(encControl.SampleRate != 12000) &&
		(encControl.SampleRate != 16000) &&
		(encControl.SampleRate != 24000) &&
		(encControl.SampleRate != 32000) &&
		(encControl.SampleRate != 44100) &&
		(encControl.SampleRate != 48000)) ||
		((encControl.MaxInternalSampleRate != 8000) &&
			(encControl.MaxInternalSampleRate != 12000) &&
			(encControl.MaxInternalSampleRate != 16000) &&
			(encControl.MaxInternalSampleRate != 24000)) {
		return 0, ErrEncodeFSNotSupported
	}

	API_fs_Hz = encControl.SampleRate
	max_internal_fs_kHz = int32(encControl.MaxInternalSampleRate>>10) + 1
	PacketSize_ms = DIV32(1000*encControl.PacketSize, API_fs_Hz)
	TargetRate_bps = encControl.BitRate
	PacketLoss_perc = encControl.PacketLossPercentage
	UseInBandFEC = encControl.UseInBandFEC
	Complexity = encControl.Complexity
	UseDTX = encControl.UseDTX

	psEnc.sCmn.API_fs_Hz = API_fs_Hz
	psEnc.sCmn.maxInternal_fs_kHz = max_internal_fs_kHz
	psEnc.sCmn.useInBandFEC = UseInBandFEC

	nSamplesIn := int32(len(in))

	input_10ms = DIV32(100*nSamplesIn, API_fs_Hz)
	if input_10ms*API_fs_Hz != 100*nSamplesIn || nSamplesIn < 0 {
		return 0, ErrEncodeInputInvalidNoOfSamples
	}

	TargetRate_bps = LIMIT(TargetRate_bps, MIN_TARGET_RATE_BPS, MAX_TARGET_RATE_BPS)
	if ret = control_encoder_FIX(psEnc, PacketSize_ms, TargetRate_bps,
		PacketLoss_perc, UseDTX, Complexity); ret != 0 {
		return 0, fmt.Errorf("invalid control encoder ret %d", ret)
	}

	if 1000*nSamplesIn > psEnc.sCmn.PacketSize_ms*API_fs_Hz {
		return 0, ErrEncodeInputInvalidNoOfSamples
	}

	samplesIn := mem2Slice(in)

	nBytesOut := int16(len(out))

	if min(API_fs_Hz, 1000*max_internal_fs_kHz) == 24000 &&
		psEnc.sCmn.sSWBdetect.SWB_detected == 0 &&
		psEnc.sCmn.sSWBdetect.WB_detected == 0 {
		detect_SWB_input(&psEnc.sCmn.sSWBdetect, samplesIn, nSamplesIn)
	}

	outData := mem2Slice[byte](out)

	MaxBytesOut = 0
	for {
		nSamplesToBuffer = psEnc.sCmn.frame_length - psEnc.sCmn.inputBufIx
		if API_fs_Hz == SMULBB(1000, psEnc.sCmn.fs_kHz) {
			nSamplesToBuffer = min(nSamplesToBuffer, nSamplesIn)
			nSamplesFromInput = nSamplesToBuffer
			samplesIn.copy(psEnc.sCmn.inputBuf.off(int(psEnc.sCmn.inputBufIx)), int(nSamplesFromInput))
		} else {
			nSamplesToBuffer = min(nSamplesToBuffer, 10*input_10ms*psEnc.sCmn.fs_kHz)
			nSamplesFromInput = DIV32_16(nSamplesToBuffer*API_fs_Hz, int16(psEnc.sCmn.fs_kHz*1000))
			ret += resampler(&psEnc.sCmn.resampler_state,
				psEnc.sCmn.inputBuf.off(int(psEnc.sCmn.inputBufIx)), samplesIn, nSamplesFromInput)
		}
		samplesIn = samplesIn.off(int(nSamplesFromInput))
		nSamplesIn -= nSamplesFromInput
		psEnc.sCmn.inputBufIx += nSamplesToBuffer

		if psEnc.sCmn.inputBufIx >= psEnc.sCmn.frame_length {
			if MaxBytesOut == 0 {
				MaxBytesOut = nBytesOut
				if ret = encode_frame_FIX(psEnc, outData, &MaxBytesOut, psEnc.sCmn.inputBuf); ret != 0 {
					return 0, fmt.Errorf("encode frame failed code %d", ret)
				}
			} else {
				if ret = encode_frame_FIX(psEnc, outData, &nBytesOut, psEnc.sCmn.inputBuf); ret != 0 {
					return 0, fmt.Errorf("encode frame failed code %d", ret)
				}
			}
			psEnc.sCmn.inputBufIx = 0
			psEnc.sCmn.controlled_since_last_payload = 0

			if nSamplesIn == 0 {
				break
			}
		} else {
			break
		}
	}

	nBytesOut = MaxBytesOut
	if psEnc.sCmn.useDTX != 0 && psEnc.sCmn.inDTX != 0 {
		nBytesOut = 0
	}

	return int(nBytesOut), nil
}

func Encode(option *EncoderOption, r io.Reader, w io.Writer) error {
	out := make([]byte, 250*5, 250*5)
	in := make([]int16, option.PacketSize, option.PacketSize)
	encoder := NewEncoder(option)

	dupin := in[0:]
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&dupin))
	sh.Len *= 2
	sh.Cap *= 2

	if _, err := w.Write(magicV3); err != nil {
		return err
	}

	for {
		n, err := r.Read(*(*[]byte)(unsafe.Pointer(sh)))
		if err == io.EOF {
			break
		}

		sizeIn := n / 2
		outSize, err := encoder.EncodePayload(in[:sizeIn], out)
		if err != nil {
			return err
		}

		if err = binary.Write(w, binary.LittleEndian, int16(outSize)); err != nil {
			return err
		}

		if _, err = w.Write(out[:outSize]); err != nil {
			return err
		}
	}
	return nil
}

func initEncoder() *encoder_state_FIX {

	ret := 0
	es := &encoder_state_FIX{}
	if ret += init_encoder_FIX(es); ret != 0 {
		return nil
	}

	return es
}

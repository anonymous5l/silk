package silk

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

type DecodeResult struct {
	SampleRate             int32
	FrameSize              int32
	FramePerPacket         int32
	MoreInterDecoderFrames int32
	InBandFECOffset        int32
	Data                   []int16
}

type Decoder struct {
	sampleRate int32
	state      *decoder_state
}

func initDecoder() *decoder_state {
	struc := &decoder_state{}
	init_decoder(struc)
	return struc
}

func NewDecoder(sampleRate int) *Decoder {
	return &Decoder{
		sampleRate: int32(sampleRate),
		state:      initDecoder(),
	}
}

func (d *Decoder) DecodePayload(in []byte, lostFlag bool) (*DecodeResult, error) {
	var (
		ret                     = 0
		used_bytes, prev_fs_kHz int32
		samplesOut              = alloc[int16](MAX_API_FS_KHZ * FRAME_LENGTH_MS)
		nBytesIn                = int32(len(in))
		nSamplesOut             int16
	)

	psDec := d.state
	decControl := &DecodeResult{
		SampleRate: d.sampleRate,
	}

	if psDec.moreInternalDecoderFrames == 0 {
		psDec.nFramesDecoded = 0
	}

	if psDec.moreInternalDecoderFrames == 0 &&
		!lostFlag &&
		nBytesIn > MAX_ARITHM_BYTES {
		lostFlag = true
		ret = DEC_PAYLOAD_TOO_LARGE
	}

	prev_fs_kHz = psDec.fs_kHz

	inData := mem2Slice[byte](in)

	ret += decode_frame(psDec, samplesOut, &nSamplesOut, inData, nBytesIn,
		lostFlag, &used_bytes)

	if used_bytes != 0 {
		if psDec.nBytesLeft > 0 && psDec.FrameTermination == MORE_FRAMES && psDec.nFramesDecoded < 5 {
			psDec.moreInternalDecoderFrames = 1
		} else {
			psDec.moreInternalDecoderFrames = 0
			psDec.nFramesInPacket = psDec.nFramesDecoded

			if psDec.vadFlag == VOICE_ACTIVITY {
				if psDec.FrameTermination == LAST_FRAME {
					psDec.no_FEC_counter++
					if psDec.no_FEC_counter > NO_LBRR_THRES {
						psDec.inband_FEC_offset = 0
					}
				} else if psDec.FrameTermination == LBRR_VER1 {
					psDec.inband_FEC_offset = 1
					psDec.no_FEC_counter = 0
				} else if psDec.FrameTermination == LBRR_VER2 {
					psDec.inband_FEC_offset = 2
					psDec.no_FEC_counter = 0
				}
			}
		}
	}

	if MAX_API_FS_KHZ*1000 < decControl.SampleRate ||
		8000 > decControl.SampleRate {
		return nil, ErrDecodeInvalidSamplingFrequency
	}

	if psDec.fs_kHz*1000 != decControl.SampleRate {
		samplesOut_tmp := alloc[int16](MAX_API_FS_KHZ * FRAME_LENGTH_MS)

		samplesOut.copy(samplesOut_tmp, int(nSamplesOut))

		if prev_fs_kHz != psDec.fs_kHz || psDec.prev_API_sampleRate != decControl.SampleRate {
			ret = resampler_init(psDec.resampler_state, SMULBB(psDec.fs_kHz, 1000), decControl.SampleRate)
		}

		ret += resampler(psDec.resampler_state, samplesOut, samplesOut_tmp, int32(nSamplesOut))

		nSamplesOut = int16(DIV32(int32(nSamplesOut)*decControl.SampleRate, psDec.fs_kHz*1000))
	}

	psDec.prev_API_sampleRate = decControl.SampleRate

	decControl.FrameSize = decControl.SampleRate / 50
	decControl.FramePerPacket = psDec.nFramesInPacket
	decControl.InBandFECOffset = psDec.inband_FEC_offset
	decControl.MoreInterDecoderFrames = psDec.moreInternalDecoderFrames
	decControl.Data = samplesOut.From()[:nSamplesOut]

	return decControl, nil
}

func Decode(sampleRate int, r io.Reader, w io.Writer) error {
	magic := make([]byte, len(magicV3), len(magicV3))
	if _, err := io.ReadFull(r, magic); err != nil {
		return err
	}
	if bytes.Compare(magic, magicV3) != 0 {
		return errors.New("invalid silk v3 head")
	}

	chunk := make([]byte, MAX_ARITHM_BYTES, MAX_ARITHM_BYTES)
	decoder := NewDecoder(sampleRate)
	payloadSize := uint16(0)
	for {
		if err := binary.Read(r, binary.LittleEndian, &payloadSize); err == io.ErrUnexpectedEOF {
			break
		} else if err != nil {
			return err
		}
		if _, err := io.ReadFull(r, chunk[:payloadSize]); err != nil {
			if err != io.ErrUnexpectedEOF {
				return err
			}
			break
		}
		dr, err := decoder.DecodePayload(chunk[:payloadSize], false)
		if err != nil {
			return err
		}
		if _, err = w.Write(slice2[byte](mem2Slice(dr.Data)).From()); err != nil {
			return err
		}
	}
	return nil
}

func search_for_LBRR(inData *slice[byte], nBytesIn int32, lost_offset int32, LBRRData *slice[byte], nLBRRBytes *int16) {
	sDec := &decoder_state{}
	sDec.init()

	sDecCtrl := &decoder_control{}
	sDecCtrl.init()

	TempQ := alloc[int32](MAX_FRAME_LENGTH)

	if lost_offset < 1 || lost_offset > MAX_LBRR_DELAY {
		*nLBRRBytes = 0
		return
	}

	sDec.nFramesDecoded = 0
	sDec.fs_kHz = 0
	sDec.lossCnt = 0

	range_dec_init(sDec.sRC, inData, nBytesIn)

	for {
		decode_parameters(sDec, sDecCtrl, TempQ, 0)

		if sDec.sRC.error != 0 {
			*nLBRRBytes = 0
			return
		}
		if ((sDec.FrameTermination-1)&lost_offset) != 0 && sDec.FrameTermination > 0 && sDec.nBytesLeft >= 0 {
			*nLBRRBytes = int16(sDec.nBytesLeft)
			inData.off(int(nBytesIn-sDec.nBytesLeft)).copy(LBRRData, int(sDec.nBytesLeft))
			break
		}
		if sDec.nBytesLeft > 0 && sDec.FrameTermination == MORE_FRAMES {
			sDec.nFramesDecoded++
		} else {
			LBRRData = nil
			*nLBRRBytes = 0
			break
		}
	}
}

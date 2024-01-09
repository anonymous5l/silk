package silk

import (
	"encoding/binary"
	"math"
)

type rangeCoderState struct {
	BaseQ32  uint32
	RangeQ16 uint32
	buffer   []byte
	bufferIx int
}

func newRangeCoderState(buffer []byte) (*rangeCoderState, error) {
	state := &rangeCoderState{}
	state.RangeQ16 = 0x0000FFFF
	if len(buffer) > 0 {
		if len(buffer) > MaxArithmBytes {
			return nil, ErrRangeCodeDecodePayloadTooLarge
		}

		state.BaseQ32 = binary.BigEndian.Uint32(buffer)
		state.buffer = buffer
		state.bufferIx = 0
	}
	return state, nil
}

func (sRC *rangeCoderState) decodeSplit(pChild1, pChild2 *int32, p int32, shellTable []uint16) (err error) {
	if p > 0 {
		cdfMiddle := p >> 1
		cdf := shellTable[ShellCodeTableOffsets[p]:]
		if *pChild1, err = sRC.rangeDecoder(cdf, cdfMiddle); err != nil {
			return
		}
		*pChild2 = p - *pChild1
	} else {
		*pChild1 = 0
		*pChild2 = 0
	}
	return
}

func (sRC *rangeCoderState) shellDecoder(pulses0 []int32, pulses4 int32) (err error) {
	pulses3 := make([]int32, 2, 2)
	pulses2 := make([]int32, 4, 4)
	pulses1 := make([]int32, 8, 8)

	if err = sRC.decodeSplit(&pulses3[0], &pulses3[1], pulses4, ShellCodeTable3); err != nil {
		return
	}

	if err = sRC.decodeSplit(&pulses2[0], &pulses2[1], pulses3[0], ShellCodeTable2); err != nil {
		return
	}

	if err = sRC.decodeSplit(&pulses1[0], &pulses1[1], pulses2[0], ShellCodeTable1); err != nil {
		return
	}
	if err = sRC.decodeSplit(&pulses0[0], &pulses0[1], pulses1[0], ShellCodeTable0); err != nil {
		return
	}
	if err = sRC.decodeSplit(&pulses0[2], &pulses0[3], pulses1[1], ShellCodeTable0); err != nil {
		return
	}

	if err = sRC.decodeSplit(&pulses1[2], &pulses1[3], pulses2[1], ShellCodeTable1); err != nil {
		return
	}
	if err = sRC.decodeSplit(&pulses0[4], &pulses0[5], pulses1[2], ShellCodeTable0); err != nil {
		return
	}
	if err = sRC.decodeSplit(&pulses0[6], &pulses0[7], pulses1[3], ShellCodeTable0); err != nil {
		return
	}

	if err = sRC.decodeSplit(&pulses2[2], &pulses2[3], pulses3[1], ShellCodeTable2); err != nil {
		return
	}

	if err = sRC.decodeSplit(&pulses1[4], &pulses1[5], pulses2[2], ShellCodeTable1); err != nil {
		return
	}
	if err = sRC.decodeSplit(&pulses0[8], &pulses0[9], pulses1[4], ShellCodeTable0); err != nil {
		return
	}
	if err = sRC.decodeSplit(&pulses0[10], &pulses0[11], pulses1[5], ShellCodeTable0); err != nil {
		return
	}

	if err = sRC.decodeSplit(&pulses1[6], &pulses1[7], pulses2[3], ShellCodeTable1); err != nil {
		return
	}
	if err = sRC.decodeSplit(&pulses0[12], &pulses0[13], pulses1[6], ShellCodeTable0); err != nil {
		return
	}
	if err = sRC.decodeSplit(&pulses0[14], &pulses0[15], pulses1[7], ShellCodeTable0); err != nil {
		return
	}

	return
}

func (sRC *rangeCoderState) rangeDecoder(prob []uint16, probIndex int32) (int32, error) {
	highQ16, lowQ16 := prob[probIndex], uint16(0)
	rangeQ16, baseQ32 := sRC.RangeQ16, sRC.BaseQ32
	rangeQ32 := uint32(0)
	buffer := sRC.buffer[4:]
	bufferIx := sRC.bufferIx

	baseTmp := u32mul(rangeQ16, uint32(highQ16))
	if baseTmp > baseQ32 {
		for {
			probIndex--
			lowQ16 = prob[probIndex]
			baseTmp = u32mul(rangeQ16, uint32(lowQ16))
			if baseTmp <= baseQ32 {
				break
			}
			highQ16 = lowQ16
			if highQ16 == 0 {
				return 0, ErrRangeCoderCDFOutOfRange
			}
		}
	} else {
		for {
			lowQ16 = highQ16
			probIndex++
			highQ16 = prob[probIndex]
			baseTmp = u32mul(rangeQ16, uint32(highQ16))
			if baseTmp > baseQ32 {
				probIndex--
				break
			}
			if highQ16 == 0xFFFF {
				return 0, ErrRangeCoderCDFOutOfRange
			}
		}
	}

	baseQ32 -= u32mul(rangeQ16, uint32(lowQ16))
	rangeQ32 = u32mul(rangeQ16, uint32(highQ16-lowQ16))

	if rangeQ32&0xFF000000 > 0 {
		rangeQ16 = u32rshift(rangeQ32, 16)
	} else {
		if rangeQ32&0xFFFF0000 > 0 {
			rangeQ16 = u32rshift(rangeQ32, 8)
			if u32rshift(baseQ32, 24) > 0 {
				return 0, ErrRangeCoderNormalizationFailed
			}
		} else {
			rangeQ16 = rangeQ32
			if u32rshift(baseQ32, 16) > 0 {
				return 0, ErrRangeCoderNormalizationFailed
			}
			baseQ32 = baseQ32 << 8
			if bufferIx < len(buffer) {
				baseQ32 |= uint32(buffer[bufferIx])
				bufferIx++
			}
		}

		baseQ32 = u32lshift(baseQ32, 8)
		if bufferIx < len(buffer) {
			baseQ32 |= uint32(buffer[bufferIx])
			bufferIx++
		}
	}

	if rangeQ16 == 0 {
		return 0, ErrRangeCoderZeroIntervalWidth
	}

	sRC.BaseQ32 = baseQ32
	sRC.RangeQ16 = rangeQ16
	sRC.bufferIx = bufferIx

	return probIndex, nil
}

func (sRC *rangeCoderState) rangeDecoderMulti(prob [][]uint16, probStartIx []int32, data []int32, nSymbols int) (err error) {
	for k := 0; k < nSymbols; k++ {
		if data[k], err = sRC.rangeDecoder(prob[k], probStartIx[k]); err != nil {
			return
		}
	}
	return
}

func (sRC *rangeCoderState) rangeCoderGetLength(nBytes *int32) int32 {
	nBits := lshift(int32(sRC.bufferIx), 3) + clz32(int32(sRC.RangeQ16-1)) - 14

	*nBytes = rshift(nBits+7, 3)

	return nBits
}

func (sRC *rangeCoderState) rangeCoderCheckAfterDecoding() error {
	var nBytes int32
	bitsInStream := sRC.rangeCoderGetLength(&nBytes)

	if nBytes-1 >= int32(len(sRC.buffer)) {
		return ErrRangeCoderDecoderCheckFailed
	}

	if bitsInStream&7 > 0 {
		mask := byte(rshift(0xff, bitsInStream&7))
		if (sRC.buffer[nBytes-1] & mask) != mask {
			return ErrRangeCoderDecoderCheckFailed
		}
	}

	return nil
}

func (sRC *rangeCoderState) decodeSigns(q []int32, length, sigtype, QuantOffsetType, RateLevelIndex int32) (err error) {
	cdf := make([]uint16, 3, 3)
	i := smulbb(NRateLevels-1, lshift(sigtype, 1)+QuantOffsetType) + RateLevelIndex
	cdf[1] = SignCDF[i]
	cdf[2] = 65535

	var data int32

	for i = int32(0); i < length; i++ {
		if q[i] > 0 {
			if data, err = sRC.rangeDecoder(cdf, 1); err != nil {
				return
			}
			q[i] *= lshift(data, 1) - 1
		}
	}

	return
}

type decoderControl struct {
	pitchL      []int32
	GainsQ16    []int32
	Seed        int32
	PredCoefQ12 [][]int16
	LTPCoefQ14  []int16
	LTPScaleQ14 int32

	PERIndex         int32
	RateLevelIndex   int32
	QuantOffsetType  int32
	sigType          int32
	NLSFInterpCoefQ2 int32
}

func (d *decoderControl) init() {
	d.pitchL = make([]int32, NBSubFR, NBSubFR)
	d.GainsQ16 = make([]int32, NBSubFR, NBSubFR)
	d.PredCoefQ12 = make([][]int16, 2, 2)
	d.PredCoefQ12[0] = make([]int16, MaxLPCOrder, MaxLPCOrder)
	d.PredCoefQ12[1] = make([]int16, MaxLPCOrder, MaxLPCOrder)
	d.LTPCoefQ14 = make([]int16, LTPOrder*NBSubFR, LTPOrder*NBSubFR)
}

type decoderState struct {
	sRC                       *rangeCoderState
	prevInvGainQ16            int32
	sLTPQ16                   []int32
	sLPCQ14                   []int32
	excQ10                    []int32
	resQ10                    []int32
	outBuf                    []int16
	lagPrev                   int32
	LastGainIndex             int32
	LastGainIndexEnhLayer     int32
	typeOffsetPrev            int32
	HPState                   []int32
	HPA                       []int16
	HPB                       []int16
	fskHz                     int32
	prevAPISampleRate         int32
	frameLength               int32
	subfrLength               int32
	LPCOrder                  int32
	prevNLSFQ15               []int32
	firstFrameAfterReset      int
	nBytesLeft                int32
	nFramesDecoded            int
	nFramesInPacket           int
	moreInternalDecoderFrames int
	FrameTermination          int32
	psNLSFCB                  [2]NLSFCB
	vadFlag                   int32
	noFECCounter              int32
	inBandFECOffset           int32
	sCNG                      *CNG
	lossCnt                   int32
	prevSigType               int32
	sPLC                      *PLC
}

func (s *decoderState) init() {
	s.sLTPQ16 = make([]int32, 2*MaxFrameLength, 2*MaxFrameLength)
	s.sLPCQ14 = make([]int32, MaxFrameLength/NBSubFR+MaxLPCOrder, MaxFrameLength/NBSubFR+MaxLPCOrder)
	s.excQ10 = make([]int32, MaxFrameLength, MaxFrameLength)
	s.resQ10 = make([]int32, MaxFrameLength, MaxFrameLength)
	s.outBuf = make([]int16, 2*MaxFrameLength, 2*MaxFrameLength)
	s.HPState = make([]int32, DecHPOrder, DecHPOrder)
	s.prevNLSFQ15 = make([]int32, MaxLPCOrder, MaxLPCOrder)
	s.sCNG = &CNG{}
	s.sCNG.init()
	s.sPLC = &PLC{}
	s.sPLC.init()
}

func (s *decoderState) CNGReset() {
	NLSFStepQ15 := math.MaxInt16 / (s.LPCOrder + 1)
	NLSFAccQ15 := int32(0)
	s.sCNG.smthNLSFQ15 = make([]int32, MaxLPCOrder)
	for i := int32(0); i < s.LPCOrder; i++ {
		NLSFAccQ15 += NLSFStepQ15
		s.sCNG.smthNLSFQ15[i] = NLSFAccQ15
	}
	s.sCNG.smthGainQ16 = 0
	s.sCNG.randSeed = 0x307880
}

func (s *decoderState) PLCReset() {
	s.sPLC.pitchLQ8 = s.frameLength >> 1
}

type Decoder struct {
	state *decoderState
}

func NewDecoder() (*Decoder, error) {
	var err error

	decoder := &Decoder{}
	if err = decoder.init(); err != nil {
		return nil, err
	}

	return decoder, nil
}

func (d *Decoder) setFs(fskHz int32) error {
	if d.state.fskHz != fskHz {
		d.state.fskHz = fskHz
		d.state.frameLength = smulbb(FrameLengthMS, fskHz)
		d.state.subfrLength = smulbb(FrameLengthMS/NBSubFR, fskHz)

		if d.state.fskHz == 8 {
			d.state.LPCOrder = MinLPCOrder
			d.state.psNLSFCB[0] = NLSFCB010
			d.state.psNLSFCB[1] = NLSFCB110
		} else {
			d.state.LPCOrder = MaxLPCOrder
			d.state.psNLSFCB[0] = NLSFCB016
			d.state.psNLSFCB[1] = NLSFCB116
		}

		memset(d.state.sLPCQ14, 0, MaxLPCOrder)
		memset(d.state.outBuf, 0, MaxFrameLength)
		memset(d.state.prevNLSFQ15, 0, MaxLPCOrder)

		d.state.lagPrev = 100
		d.state.LastGainIndex = 1
		d.state.prevSigType = 0
		d.state.firstFrameAfterReset = 1

		switch fskHz {
		case 24:
			d.state.HPA = DecodeAHP24
			d.state.HPB = DecodeBHP24
		case 16:
			d.state.HPA = DecodeAHP16
			d.state.HPB = DecodeBHP16
		case 12:
			d.state.HPA = DecodeAHP12
			d.state.HPB = DecodeBHP12
		case 8:
			d.state.HPA = DecodeAHP8
			d.state.HPB = DecodeBHP8
		default:
			return ErrUnsupportedSamplingRate
		}
	}

	assert(d.state.frameLength > 0 && d.state.frameLength <= MaxFrameLength)
	return nil
}

func (d *Decoder) init() error {
	d.state = &decoderState{}
	d.state.init()

	if err := d.setFs(24); err != nil {
		return err
	}

	d.state.firstFrameAfterReset = 1
	d.state.prevInvGainQ16 = 0x10000

	d.state.CNGReset()
	d.state.PLCReset()

	return nil
}

func (d *Decoder) decodeParameters(psDecCtrl *decoderControl, q []int32, fullDecoding bool) (err error) {
	var psNLSFCB *NLSFCB

	fskHzDec, Ix := int32(0), int32(0)
	gainsIndices := make([]int32, NBSubFR, NBSubFR)
	NLSFIndices := make([]int32, NLSFMSVQMaxCBStages, NLSFMSVQMaxCBStages)
	pNLSFQ15 := make([]int32, MaxLPCOrder, MaxLPCOrder)
	pNLSF0Q15 := make([]int32, MaxLPCOrder, MaxLPCOrder)
	Ixs := make([]int32, NBSubFR, NBSubFR)

	psRC := d.state.sRC

	if d.state.nFramesDecoded == 0 {
		if Ix, err = psRC.rangeDecoder(SamplingRatesCDF, SamplingRatesOffset); err != nil {
			return
		}

		if Ix < 0 || Ix > 3 {
			return ErrRangeCoderIllegalSamplingRate
		}

		fskHzDec = SamplingRatesTable[Ix]

		if err = d.setFs(fskHzDec); err != nil {
			return
		}
	}

	if d.state.nFramesDecoded == 0 {
		if Ix, err = psRC.rangeDecoder(TypeOffsetCDF, TypeOffsetCDFOffset); err != nil {
			return
		}
	} else {
		if Ix, err = psRC.rangeDecoder(TypeOffsetJointCDF[d.state.typeOffsetPrev],
			TypeOffsetCDFOffset); err != nil {
			return
		}
	}

	psDecCtrl.sigType = rshift(Ix, 1)
	psDecCtrl.QuantOffsetType = Ix & 1
	d.state.typeOffsetPrev = Ix

	if d.state.nFramesDecoded == 0 {
		if gainsIndices[0], err = psRC.rangeDecoder(GainCDF[psDecCtrl.sigType], GainCDFOffset); err != nil {
			return
		}
	} else {
		if gainsIndices[0], err = psRC.rangeDecoder(DeltaGainCDF, DeltaGainCDFOffset); err != nil {
			return
		}
	}

	for i := 1; i < NBSubFR; i++ {
		if gainsIndices[i], err = psRC.rangeDecoder(DeltaGainCDF, DeltaGainCDFOffset); err != nil {
			return
		}
	}

	gainsDequant(psDecCtrl.GainsQ16, gainsIndices, &d.state.LastGainIndex, d.state.nFramesDecoded)

	psNLSFCB = &d.state.psNLSFCB[psDecCtrl.sigType]

	if err = psRC.rangeDecoderMulti(psNLSFCB.StartPtr, psNLSFCB.MiddleIx,
		NLSFIndices, int(psNLSFCB.nStages)); err != nil {
		return
	}

	psNLSFCB.MSVQDecode(pNLSFQ15, NLSFIndices, d.state.LPCOrder)

	if psDecCtrl.NLSFInterpCoefQ2, err = psRC.rangeDecoder(
		NLSFInterpolationFactorCDF, NLSFInterpolationFactorOffset); err != nil {
		return
	}

	if d.state.firstFrameAfterReset == 1 {
		psDecCtrl.NLSFInterpCoefQ2 = 4
	}

	if fullDecoding {
		_NLSF2AStable(psDecCtrl.PredCoefQ12[1], pNLSFQ15, d.state.LPCOrder)

		if psDecCtrl.NLSFInterpCoefQ2 < 4 {
			for i := int32(0); i < d.state.LPCOrder; i++ {
				pNLSF0Q15[i] = d.state.prevNLSFQ15[i] +
					rshift(mul(psDecCtrl.NLSFInterpCoefQ2, pNLSFQ15[i]-d.state.prevNLSFQ15[i]), 2)
			}
			_NLSF2AStable(psDecCtrl.PredCoefQ12[0], pNLSF0Q15, d.state.LPCOrder)
		} else {
			for i := int32(0); i < d.state.LPCOrder; i++ {
				psDecCtrl.PredCoefQ12[0][i] = psDecCtrl.PredCoefQ12[1][i]
			}
		}
	}

	for i := int32(0); i < d.state.LPCOrder; i++ {
		d.state.prevNLSFQ15[i] = pNLSFQ15[i]
	}

	if d.state.lossCnt > 0 {
		bwexpander(psDecCtrl.PredCoefQ12[0], d.state.LPCOrder, BWEAfterLossQ16)
		bwexpander(psDecCtrl.PredCoefQ12[1], d.state.LPCOrder, BWEAfterLossQ16)
	}

	var cbkPtrQ14 []int16

	if psDecCtrl.sigType == SIGTypeVoiced {
		switch d.state.fskHz {
		case 8:
			if Ixs[0], err = psRC.rangeDecoder(PitchLagNBCDF, PitchLagNBCDFOffset); err != nil {
				return
			}
		case 12:
			if Ixs[0], err = psRC.rangeDecoder(PitchLagMBCDF, PitchLagMBCDFOffset); err != nil {
				return
			}
		case 16:
			if Ixs[0], err = psRC.rangeDecoder(PitchLagWBCDF, PitchLagWBCDFOffset); err != nil {
				return
			}
		default:
			if Ixs[0], err = psRC.rangeDecoder(PitchLagSWBCDF, PitchLagSWBCDFOffset); err != nil {
				return
			}
		}

		if d.state.fskHz == 8 {
			if Ixs[1], err = psRC.rangeDecoder(PitchContourNBCDF, PitchContourNBCDFOffset); err != nil {
				return
			}
		} else {
			if Ixs[1], err = psRC.rangeDecoder(PitchContourCDF, PitchContourCDFOffset); err != nil {
				return
			}
		}

		decodePitch(Ixs[0], Ixs[1], psDecCtrl.pitchL, d.state.fskHz)

		if psDecCtrl.PERIndex, err = psRC.rangeDecoder(LTPPerIndexCDF, LTPPerIndexCDFOffset); err != nil {
			return
		}

		cbkPtrQ14 = LTPVqPtrsQ14[psDecCtrl.PERIndex]

		for k := int32(0); k < NBSubFR; k++ {
			if Ix, err = psRC.rangeDecoder(
				LTPGainCDFPtrs[psDecCtrl.PERIndex], LTPGainCDFOffsets[psDecCtrl.PERIndex]); err != nil {
				return
			}

			for i := int32(0); i < LTPOrder; i++ {
				psDecCtrl.LTPCoefQ14[k*LTPOrder+i] = cbkPtrQ14[Ix*LTPOrder+i]
			}
		}

		if Ix, err = psRC.rangeDecoder(LTPScaleCDF, LTPScaleOffset); err != nil {
			return
		}
		psDecCtrl.LTPScaleQ14 = int32(LTPScalesTableQ14[Ix])
	} else {
		assert(psDecCtrl.sigType == SIGTypeUnvoiced)

		memset(psDecCtrl.pitchL, 0, NBSubFR)
		memset(psDecCtrl.LTPCoefQ14, 0, LTPOrder)
		psDecCtrl.PERIndex = 0
		psDecCtrl.LTPScaleQ14 = 0
	}

	if Ix, err = psRC.rangeDecoder(SeedCDF, SeedOffset); err != nil {
		return
	}
	psDecCtrl.Seed = Ix

	if err = d.decodePulses(psDecCtrl, q, d.state.frameLength); err != nil {
		return
	}

	if d.state.vadFlag, err = psRC.rangeDecoder(VadFlagCDF, VadFlagOffset); err != nil {
		return
	}

	if d.state.FrameTermination, err = psRC.rangeDecoder(FrameTerminationCDF, FrameTerminationOffset); err != nil {
		return
	}

	var nBytesUsed int32

	psRC.rangeCoderGetLength(&nBytesUsed)
	d.state.nBytesLeft = int32(len(psRC.buffer)) - nBytesUsed
	if d.state.nBytesLeft < 0 {
		return ErrRangeCoderReadBeyondBuffer
	}

	if d.state.nBytesLeft == 0 {
		if err = psRC.rangeCoderCheckAfterDecoding(); err != nil {
			return
		}
	}

	return
}

func (d *Decoder) decodePulses(psDecCtrl *decoderControl, q []int32, frameLength int32) (err error) {
	psRC := d.state.sRC

	if psDecCtrl.RateLevelIndex, err = psRC.rangeDecoder(RateLevelsCDF[psDecCtrl.sigType], RateLevelsCDFOffset); err != nil {
		return
	}

	nLshifts, sumPulses := make([]int32, MaxNBShellBlocks), make([]int32, MaxNBShellBlocks)

	iter := frameLength / ShellCodecFrameLength

	cdfPtr := PulsesPerBlockCDF[psDecCtrl.RateLevelIndex]
	for i := int32(0); i < iter; i++ {
		nLshifts[i] = 0
		if sumPulses[i], err = psRC.rangeDecoder(cdfPtr, PulsesPerBlockCDFOffset); err != nil {
			return
		}

		for sumPulses[i] == MaxPulses {
			nLshifts[i]++
			if sumPulses[i], err = psRC.rangeDecoder(
				PulsesPerBlockCDF[NRateLevels-1], PulsesPerBlockCDFOffset); err != nil {
				return
			}
		}
	}

	for i := int32(0); i < iter; i++ {
		if sumPulses[i] > 0 {
			if err = psRC.shellDecoder(q[smulbb(i, ShellCodecFrameLength):], sumPulses[i]); err != nil {
				return
			}
		} else {
			memset(q[smulbb(i, ShellCodecFrameLength):], 0, ShellCodecFrameLength)
		}
	}

	var (
		nLS, bit, absQ int32
		pulsesPtr      []int32
	)

	for i := int32(0); i < iter; i++ {
		if nLshifts[i] > 0 {
			nLS = nLshifts[i]
			pulsesPtr = q[smulbb(i, ShellCodecFrameLength):]

			for k := 0; k < ShellCodecFrameLength; k++ {
				absQ = pulsesPtr[k]
				for j := int32(0); j < nLS; j++ {
					absQ = lshift(absQ, 1)
					if bit, err = psRC.rangeDecoder(LSBCDF, 1); err != nil {
						return
					}
					absQ += bit
				}
				pulsesPtr[k] = absQ
			}
		}
	}

	return psRC.decodeSigns(q, frameLength, psDecCtrl.sigType, psDecCtrl.QuantOffsetType, psDecCtrl.RateLevelIndex)
}

func (d *Decoder) decodeCore(psDecCtrl *decoderControl, xq []int16, q []int32) (err error) {
	assert(d.state.prevInvGainQ16 != 0)

	NLSFInterpolationFlag := int32(0)

	offsetQ10 := int32(QuantizationOffsetsQ10[psDecCtrl.sigType][psDecCtrl.QuantOffsetType])

	if psDecCtrl.NLSFInterpCoefQ2 < (1 << 2) {
		NLSFInterpolationFlag = 1
	} else {
		NLSFInterpolationFlag = 0
	}

	var dither int32

	randSeed := psDecCtrl.Seed
	for i := int32(0); i < d.state.frameLength; i++ {
		randSeed = rand(randSeed)
		dither = rshift(randSeed, 31)

		d.state.excQ10[i] = lshift(q[i], 10) + offsetQ10
		d.state.excQ10[i] = (d.state.excQ10[i] ^ dither) - dither

		randSeed += q[i]
	}

	pexcQ10 := d.state.excQ10
	presQ10 := d.state.resQ10
	pxq := d.state.outBuf[d.state.frameLength:]
	sLTPBufIdx := d.state.frameLength

	var (
		BQ14, AQ12                      []int16
		lag, startIdx, sigType          int32
		invGainQ16, GainQ16, gainAdjQ16 int32
		invGainQ32                      int32
		LTPPredQ14                      int32
		predLagPtr                      []int32
	)

	FiltState := make([]int32, MaxLPCOrder, MaxLPCOrder)
	sLTP := make([]int16, MaxFrameLength, MaxFrameLength)
	vecQ10 := make([]int32, MaxFrameLength/NBSubFR)
	AQ12Tmp := make([]int16, MaxLPCOrder)

	for k := int32(0); k < NBSubFR; k++ {
		AQ12 = psDecCtrl.PredCoefQ12[k>>1]

		for c := int32(0); c < d.state.LPCOrder; c++ {
			AQ12Tmp[c] = AQ12[c]
		}

		BQ14 = psDecCtrl.LTPCoefQ14[k*LTPOrder:]
		GainQ16 = psDecCtrl.GainsQ16[k]
		sigType = psDecCtrl.sigType

		invGainQ16 = inverse32varQ(i32max(GainQ16, 1), 32)
		invGainQ16 = i32min(invGainQ16, math.MaxInt16)

		gainAdjQ16 = 1 << 16
		if invGainQ16 != d.state.prevInvGainQ16 {
			gainAdjQ16 = div32varQ(invGainQ16, d.state.prevInvGainQ16, 16)
		}

		if d.state.lossCnt > 0 && d.state.prevSigType == SIGTypeVoiced &&
			psDecCtrl.sigType == SIGTypeUnvoiced && k < (NBSubFR>>1) {
			memset(BQ14, 0, LTPOrder)
			BQ14[LTPOrder/2] = 1 << 12
			sigType = SIGTypeVoiced
			psDecCtrl.pitchL[k] = d.state.lagPrev
		}

		if sigType == SIGTypeVoiced {
			lag = psDecCtrl.pitchL[k]

			if (k & (3 - lshift(NLSFInterpolationFlag, 1))) == 0 {
				startIdx = d.state.frameLength - lag - d.state.LPCOrder - LTPOrder/2
				assert(startIdx >= 0)
				assert(startIdx <= d.state.frameLength-d.state.LPCOrder)

				_MAPrediction(d.state.outBuf[startIdx+k*(d.state.frameLength>>2):],
					AQ12, FiltState, sLTP[startIdx:], d.state.frameLength-startIdx, d.state.LPCOrder)

				invGainQ32 = lshift(invGainQ16, 16)
				if k == 0 {
					invGainQ32 = lshift(smulwb(invGainQ32, psDecCtrl.LTPScaleQ14), 2)
				}

				for i := int32(0); i < lag+(LTPOrder/2); i++ {
					d.state.sLTPQ16[sLTPBufIdx-i-1] = smulwb(invGainQ32, int32(sLTP[d.state.frameLength-i-1]))
				}

			} else {
				if gainAdjQ16 != 1<<16 {
					for i := int32(0); i < lag+LTPOrder/2; i++ {
						d.state.sLTPQ16[sLTPBufIdx-i-1] = smulww(gainAdjQ16, d.state.sLTPQ16[sLTPBufIdx-i-1])
					}
				}
			}
		}

		for i := 0; i < MaxLPCOrder; i++ {
			d.state.sLPCQ14[i] = smulww(gainAdjQ16, d.state.sLPCQ14[i])
		}

		assert(invGainQ16 != 0)

		d.state.prevInvGainQ16 = invGainQ16

		if sigType == SIGTypeVoiced {
			predLagPtr = d.state.sLTPQ16[(sLTPBufIdx-lag+LTPOrder/2)-4:]

			for i := int32(0); i < d.state.subfrLength; i++ {
				LTPPredQ14 = smulwb(predLagPtr[4], int32(BQ14[0]))
				LTPPredQ14 = smlawb(LTPPredQ14, predLagPtr[3], int32(BQ14[1]))
				LTPPredQ14 = smlawb(LTPPredQ14, predLagPtr[2], int32(BQ14[2]))
				LTPPredQ14 = smlawb(LTPPredQ14, predLagPtr[1], int32(BQ14[3]))
				LTPPredQ14 = smlawb(LTPPredQ14, predLagPtr[0], int32(BQ14[4]))

				predLagPtr = predLagPtr[1:]

				presQ10[i] = pexcQ10[i] + rrshift(LTPPredQ14, 4)

				d.state.sLTPQ16[sLTPBufIdx] = lshift(presQ10[i], 6)

				sLTPBufIdx++
			}
		} else {
			memcpy(presQ10, pexcQ10, int(d.state.subfrLength))
		}

		decodeShortTermPrediction(vecQ10, presQ10, d.state.sLPCQ14, AQ12Tmp, d.state.LPCOrder, d.state.subfrLength)

		for i := int32(0); i < d.state.subfrLength; i++ {
			pxq[i] = sat16(rrshift(smulww(vecQ10[i], GainQ16), 10))
		}

		memcpy(d.state.sLPCQ14, d.state.sLPCQ14[d.state.subfrLength:], MaxLPCOrder)

		pexcQ10 = pexcQ10[d.state.subfrLength:]
		presQ10 = presQ10[d.state.subfrLength:]
		pxq = pxq[d.state.subfrLength:]
	}

	memcpy(xq, d.state.outBuf[d.state.frameLength:], int(d.state.frameLength))

	return
}

func (d *Decoder) decodeFrame(pOut []int16, pN *int16, pCode []byte, lost bool) (decBytes int, err error) {
	sDecCtrl := &decoderControl{}
	sDecCtrl.init()

	L, fsKhzOld := d.state.frameLength, int32(0)
	Pulses := make([]int32, MaxFrameLength, MaxFrameLength)

	sDecCtrl.LTPScaleQ14 = 0
	assert(L > 0 && L <= MaxFrameLength)

	if !lost {
		fsKhzOld = d.state.fskHz

		if d.state.nFramesDecoded == 0 {
			if d.state.sRC, err = newRangeCoderState(pCode); err != nil {
				return
			}
		}

		if err = d.decodeParameters(sDecCtrl, Pulses, true); err != nil {
			lost = true

			if err = d.setFs(fsKhzOld); err != nil {
				return
			}

			decBytes = len(d.state.sRC.buffer)
		} else {
			decBytes = len(d.state.sRC.buffer) - int(d.state.nBytesLeft)
			d.state.nFramesDecoded++

			L = d.state.frameLength

			if err = d.decodeCore(sDecCtrl, pOut, Pulses); err != nil {
				return
			}

			d.state.PLC(sDecCtrl, pOut, L, lost)

			d.state.lossCnt = 0
			d.state.prevSigType = sDecCtrl.sigType

			d.state.firstFrameAfterReset = 0
		}
	}

	if lost {
		d.state.PLC(sDecCtrl, pOut, L, lost)
	}

	memcpy(d.state.outBuf, pOut, int(L))

	d.state.PLCGlueFrames(sDecCtrl, pOut, L)

	if err = d.state.CNG(sDecCtrl, pOut, L); err != nil {
		return
	}

	assert(d.state.fskHz == 12 && (L%3) == 0 ||
		d.state.fskHz != 12 && (L%2) == 0)

	biquad(pOut, d.state.HPB, d.state.HPA, d.state.HPState, pOut, L)

	*pN = int16(L)

	d.state.lagPrev = sDecCtrl.pitchL[NBSubFR-1]

	return
}

func (d *Decoder) KHz() int {
	if d.state != nil && d.state.nFramesDecoded > 0 {
		return int(d.state.fskHz) * 1000
	}
	return 0
}

func (d *Decoder) LossCount() int {
	if d.state != nil && d.state.nFramesDecoded > 0 {
		return int(d.state.lossCnt)
	}
	return 0
}

func (d *Decoder) Decode(lost bool, payload []byte) (out []int16, err error) {
	if d.state.moreInternalDecoderFrames == 0 {
		d.state.nFramesDecoded = 0
	}

	if d.state.moreInternalDecoderFrames == 0 &&
		!lost && len(payload) > MaxArithmBytes {
		lost = true
		err = ErrRangeCodeDecodePayloadTooLarge
	}

	pSamplesOutInternal := make([]int16, MaxApiFSKHZ*FrameLengthMS, MaxApiFSKHZ*FrameLengthMS)

	var (
		usedBytes   int
		nSamplesOut int16
	)

	if usedBytes, err = d.decodeFrame(pSamplesOutInternal, &nSamplesOut, payload, lost); err != nil {
		return
	}

	if usedBytes > 0 {
		if d.state.nBytesLeft > 0 && d.state.FrameTermination == MoreFrames && d.state.nFramesDecoded < 5 {
			d.state.moreInternalDecoderFrames = 1
		} else {
			d.state.moreInternalDecoderFrames = 0
			d.state.nFramesInPacket = d.state.nFramesDecoded

			if d.state.vadFlag == VoiceActivity {
				if d.state.FrameTermination == LastFrame {
					d.state.noFECCounter++
					if d.state.noFECCounter > NoLBRRRhres {
						d.state.inBandFECOffset = 0
					}
				} else if d.state.FrameTermination == LBRRVer1 {
					d.state.inBandFECOffset = 1
					d.state.noFECCounter = 0
				} else if d.state.FrameTermination == LBRRVer2 {
					d.state.inBandFECOffset = 2
					d.state.noFECCounter = 0
				}
			}
		}
	}

	out = pSamplesOutInternal[:nSamplesOut]

	return
}

func biquad(in, B, A []int16, S []int32, out []int16, length int32) {
	var k, in16 int32
	var A0Neg, A1Neg, S0, S1, out32, tmp32 int32

	S0, S1 = S[0], S[1]
	A0Neg = int32(-A[0])
	A1Neg = int32(-A[1])

	for k = 0; k < length; k++ {
		in16 = int32(in[k])
		out32 = smlabb(S0, in16, int32(B[0]))

		S0 = smlabb(S1, in16, int32(B[1]))
		S0 += lshift(smulwb(out32, A0Neg), 3)

		S1 = lshift(smulwb(out32, A1Neg), 3)
		S1 = smlabb(S1, in16, int32(B[2]))

		tmp32 = rrshift(out32, 13) + 1
		out[k] = sat16(tmp32)
	}

	S[0], S[1] = S0, S1
}

func decodeShortTermPrediction(vecQ10 []int32, presQ10 []int32, sLPCQ14 []int32,
	AQ12Tmp []int16, LPCOrder int32, subfrLength int32) {

	var atmp, LpcPredQ10 int32

	if LPCOrder == 16 {
		for i := int32(0); i < subfrLength; i++ {
			atmp = ua2i32(AQ12Tmp[0:])
			LpcPredQ10 = smulwb(sLPCQ14[MaxLPCOrder+i-1], atmp)
			LpcPredQ10 = smlawt(LpcPredQ10, sLPCQ14[MaxLPCOrder+i-2], atmp)
			atmp = ua2i32(AQ12Tmp[2:])
			LpcPredQ10 = smlawb(LpcPredQ10, sLPCQ14[MaxLPCOrder+i-3], atmp)
			LpcPredQ10 = smlawt(LpcPredQ10, sLPCQ14[MaxLPCOrder+i-4], atmp)
			atmp = ua2i32(AQ12Tmp[4:])
			LpcPredQ10 = smlawb(LpcPredQ10, sLPCQ14[MaxLPCOrder+i-5], atmp)
			LpcPredQ10 = smlawt(LpcPredQ10, sLPCQ14[MaxLPCOrder+i-6], atmp)
			atmp = ua2i32(AQ12Tmp[6:])
			LpcPredQ10 = smlawb(LpcPredQ10, sLPCQ14[MaxLPCOrder+i-7], atmp)
			LpcPredQ10 = smlawt(LpcPredQ10, sLPCQ14[MaxLPCOrder+i-8], atmp)
			atmp = ua2i32(AQ12Tmp[8:])
			LpcPredQ10 = smlawb(LpcPredQ10, sLPCQ14[MaxLPCOrder+i-9], atmp)
			LpcPredQ10 = smlawt(LpcPredQ10, sLPCQ14[MaxLPCOrder+i-10], atmp)
			atmp = ua2i32(AQ12Tmp[10:])
			LpcPredQ10 = smlawb(LpcPredQ10, sLPCQ14[MaxLPCOrder+i-11], atmp)
			LpcPredQ10 = smlawt(LpcPredQ10, sLPCQ14[MaxLPCOrder+i-12], atmp)
			atmp = ua2i32(AQ12Tmp[12:])
			LpcPredQ10 = smlawb(LpcPredQ10, sLPCQ14[MaxLPCOrder+i-13], atmp)
			LpcPredQ10 = smlawt(LpcPredQ10, sLPCQ14[MaxLPCOrder+i-14], atmp)
			atmp = ua2i32(AQ12Tmp[14:])
			LpcPredQ10 = smlawb(LpcPredQ10, sLPCQ14[MaxLPCOrder+i-15], atmp)
			LpcPredQ10 = smlawt(LpcPredQ10, sLPCQ14[MaxLPCOrder+i-16], atmp)

			vecQ10[i] = presQ10[i] + LpcPredQ10
			sLPCQ14[MaxLPCOrder+i] = lshift(vecQ10[i], 4)
		}
	} else {
		assert(LPCOrder == 10)

		for i := int32(0); i < subfrLength; i++ {
			atmp = ua2i32(AQ12Tmp[0:])
			LpcPredQ10 = smulwb(sLPCQ14[MaxLPCOrder+i-1], atmp)
			LpcPredQ10 = smlawt(LpcPredQ10, sLPCQ14[MaxLPCOrder+i-2], atmp)
			atmp = ua2i32(AQ12Tmp[2:])
			LpcPredQ10 = smlawb(LpcPredQ10, sLPCQ14[MaxLPCOrder+i-3], atmp)
			LpcPredQ10 = smlawt(LpcPredQ10, sLPCQ14[MaxLPCOrder+i-4], atmp)
			atmp = ua2i32(AQ12Tmp[4:])
			LpcPredQ10 = smlawb(LpcPredQ10, sLPCQ14[MaxLPCOrder+i-5], atmp)
			LpcPredQ10 = smlawt(LpcPredQ10, sLPCQ14[MaxLPCOrder+i-6], atmp)
			atmp = ua2i32(AQ12Tmp[6:])
			LpcPredQ10 = smlawb(LpcPredQ10, sLPCQ14[MaxLPCOrder+i-7], atmp)
			LpcPredQ10 = smlawt(LpcPredQ10, sLPCQ14[MaxLPCOrder+i-8], atmp)
			atmp = ua2i32(AQ12Tmp[8:])
			LpcPredQ10 = smlawb(LpcPredQ10, sLPCQ14[MaxLPCOrder+i-9], atmp)
			LpcPredQ10 = smlawt(LpcPredQ10, sLPCQ14[MaxLPCOrder+i-10], atmp)

			vecQ10[i] = presQ10[i] + LpcPredQ10
			sLPCQ14[MaxLPCOrder+i] = lshift(vecQ10[i], 4)
		}
	}

	//for i := int32(0); i < subfrLength; i++ {
	//	LPCPredQ10 := smulwb(sLPCQ14[MaxLPCOrder+i-1], int32(AQ12Tmp[0]))
	//	LPCPredQ10 = smlawb(LPCPredQ10, sLPCQ14[MaxLPCOrder+i-2], int32(AQ12Tmp[1]))
	//	LPCPredQ10 = smlawb(LPCPredQ10, sLPCQ14[MaxLPCOrder+i-3], int32(AQ12Tmp[2]))
	//	LPCPredQ10 = smlawb(LPCPredQ10, sLPCQ14[MaxLPCOrder+i-4], int32(AQ12Tmp[3]))
	//	LPCPredQ10 = smlawb(LPCPredQ10, sLPCQ14[MaxLPCOrder+i-5], int32(AQ12Tmp[4]))
	//	LPCPredQ10 = smlawb(LPCPredQ10, sLPCQ14[MaxLPCOrder+i-6], int32(AQ12Tmp[5]))
	//	LPCPredQ10 = smlawb(LPCPredQ10, sLPCQ14[MaxLPCOrder+i-7], int32(AQ12Tmp[6]))
	//	LPCPredQ10 = smlawb(LPCPredQ10, sLPCQ14[MaxLPCOrder+i-8], int32(AQ12Tmp[7]))
	//	LPCPredQ10 = smlawb(LPCPredQ10, sLPCQ14[MaxLPCOrder+i-9], int32(AQ12Tmp[8]))
	//	LPCPredQ10 = smlawb(LPCPredQ10, sLPCQ14[MaxLPCOrder+i-10], int32(AQ12Tmp[9]))
	//	for j := int32(10); j < LPCOrder; j++ {
	//		LPCPredQ10 = smlawb(LPCPredQ10, sLPCQ14[MaxLPCOrder+i-j-1], int32(AQ12Tmp[j]))
	//	}
	//
	//	vecQ10[i] = presQ10[i] + LPCPredQ10
	//	sLPCQ14[MaxLPCOrder+i] = vecQ10[i] << 4
	//}
}

func decodePitch(lagIndex int32, contourIndex int32, pitchLags []int32, fskHz int32) {
	minLag := smulbb(PitchESTMinLagMS, fskHz)
	lag := minLag + lagIndex
	if fskHz == 8 {
		for i := 0; i < PitchESTNBSubFR; i++ {
			pitchLags[i] = lag + int32(CBLagsStage2[i][contourIndex])
		}
	} else {
		for i := 0; i < PitchESTNBSubFR; i++ {
			pitchLags[i] = lag + int32(CBLagsStage3[i][contourIndex])
		}
	}
}

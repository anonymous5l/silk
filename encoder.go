package silk

import "math"

const (
	LOG2VariableHPMinFREQQ7 = 809
	RadiansConstantQ19      = 1482
)

type _LBRR struct {
	payload [MaxArithmBytes]uint8
	nBytes  int32
	usage   int32
}

type predict struct {
	pitchLPCWinLength int32
	minPitchLag       int32
	maxPitchLag       int32
	prevNLSFqQ15      [MaxLPCOrder]int32
}

type nsqState struct {
	xq             [2 * MaxFrameLength]int16
	sLTPShpQ10     [2 * MaxFrameLength]int32
	sLPCQ14        [MaxFrameLength/NBSubFR + NSQLPCBufLength]int32
	sAR2Q14        [MaxShapeLPCOrder]int32
	sLFARShpQ12    int32
	lagPrev        int32
	sLTPBufIdx     int32
	sLTPShpBufIdx  int32
	randSeed       int32
	prevInvGainQ16 int32
	rewhiteFlag    int32
}

type _LP struct {
	InLPState         [2]int32
	transitionFrameNo int32
	mode              int32
}

func (psLP *_LP) interpolateFilterTaps(BQ28, AQ28 []int32, ind, facQ16 int32) {
	var nb, na int32

	if ind < TransitionIntNum-1 {
		if facQ16 > 0 {
			if facQ16 == int32(sat16(facQ16)) {
				for nb = 0; nb < TransitionNB; nb++ {
					BQ28[nb] = smlawb(
						TransitionLPBQ28[ind][nb],
						TransitionLPBQ28[ind+1][nb]-TransitionLPBQ28[ind][nb],
						facQ16)
				}
				for na = 0; na < TransitionNA; na++ {
					AQ28[na] = smlawb(
						TransitionLPAQ28[ind][na],
						TransitionLPAQ28[ind+1][na]-TransitionLPAQ28[ind][na],
						facQ16)
				}

			} else if facQ16 == (1 << 15) {
				for nb = 0; nb < TransitionNB; nb++ {
					BQ28[nb] = rshift(
						TransitionLPBQ28[ind][nb]+
							TransitionLPBQ28[ind+1][nb],
						1)
				}
				for na = 0; na < TransitionNA; na++ {
					AQ28[na] = rshift(
						TransitionLPAQ28[ind][na]+
							TransitionLPAQ28[ind+1][na],
						1)
				}
			} else {
				assert(((1 << 16) - facQ16) == int32(sat16((1<<16)-facQ16)))
				for nb = 0; nb < TransitionNB; nb++ {
					BQ28[nb] = smlawb(
						TransitionLPBQ28[ind+1][nb],
						TransitionLPBQ28[ind][nb]-
							TransitionLPBQ28[ind+1][nb],
						(1<<16)-facQ16)
				}
				for na = 0; na < TransitionNA; na++ {
					AQ28[na] = smlawb(
						TransitionLPAQ28[ind+1][na],
						TransitionLPAQ28[ind][na]-
							TransitionLPAQ28[ind+1][na],
						(1<<16)-facQ16)
				}
			}
		} else {
			memcpy(BQ28, TransitionLPBQ28[ind], TransitionNB)
			memcpy(AQ28, TransitionLPAQ28[ind], TransitionNA)
		}
	} else {
		memcpy(BQ28, TransitionLPBQ28[TransitionIntNum-1], TransitionNB)
		memcpy(AQ28, TransitionLPAQ28[TransitionIntNum-1], TransitionNA)
	}
}

func (psLP *_LP) VariableCutoff(out, in []int16, frameLength int32) {
	var (
		BQ28   [TransitionNB]int32
		AQ28   [TransitionNA]int32
		facQ16 int32
		ind    int32
	)

	assert(psLP.transitionFrameNo >= 0)
	assert(((psLP.transitionFrameNo < TransitionFramesDown) && (psLP.mode == 0)) ||
		(psLP.transitionFrameNo <= TransitionFramesUp && psLP.mode == 1))

	if psLP.transitionFrameNo > 0 {
		if psLP.mode == 0 {
			if psLP.transitionFrameNo < TransitionFramesDown {
				facQ16 = lshift(psLP.transitionFrameNo, 16-5)
				ind = rshift(facQ16, 16)
				facQ16 -= lshift(ind, 16)

				assert(ind >= 0)
				assert(ind < TransitionIntNum)

				psLP.interpolateFilterTaps(BQ28[:], AQ28[:], ind, facQ16)
				psLP.transitionFrameNo++
			} else {
				assert(psLP.transitionFrameNo == TransitionFramesDown)

				psLP.interpolateFilterTaps(BQ28[:], AQ28[:], TransitionIntNum-1, 0)
			}
		} else {
			assert(psLP.mode == 1)

			if psLP.transitionFrameNo < TransitionFramesUp {
				facQ16 = lshift(TransitionFramesUp-psLP.transitionFrameNo, 16-6)
				ind = rshift(facQ16, 16)
				facQ16 -= lshift(ind, 16)

				assert(ind >= 0)
				assert(ind < TransitionIntNum)

				psLP.interpolateFilterTaps(BQ28[:], AQ28[:], ind, facQ16)
				psLP.transitionFrameNo++
			} else {
				assert(psLP.transitionFrameNo == TransitionFramesUp)
				psLP.interpolateFilterTaps(BQ28[:], AQ28[:], 0, 0)
			}
		}
	}

	if psLP.transitionFrameNo > 0 {
		biquadAlt(in, BQ28[:], AQ28[:], psLP.InLPState[:], out, frameLength)
	} else {
		memcpy(out, in, int(frameLength))
	}
}

type EncodeOption struct {
	SampleRate           int32
	MaxSampleRate        int32
	BitRate              int32
	PacketSize           int32
	PacketLossPercentage int32
	UseInBandFEC         bool
	UseDTX               bool
	Complexity           int32
}

func (opt *EncodeOption) init() {
	if opt.SampleRate == 0 {
		opt.SampleRate = KHz24000
	}
	if opt.MaxSampleRate == 0 {
		opt.MaxSampleRate = KHz24000
	}
	if opt.SampleRate < opt.MaxSampleRate {
		opt.MaxSampleRate = opt.SampleRate
	}
	if opt.BitRate == 0 {
		opt.BitRate = SWB2WBBitrateBPS
	}
	if opt.Complexity == 0 {
		opt.Complexity = 2
	}
	opt.BitRate = limit(opt.BitRate, MinTargetRateBPS, MaxTargetRateBPS)
}

type encoderControl struct {
	lagIndex             int32
	contourIndex         int32
	PERIndex             int32
	LTPIndex             [NBSubFR]int32
	NLSFIndices          [NLSFMSVQMaxCBStages]int32
	NLSFInterpCoefQ2     int32
	GainsIndices         [NBSubFR]int32
	Seed                 int32
	LTPScaleIndex        int32
	RateLevelIndex       int32
	QuantOffsetType      int32
	sigType              int32
	pitchL               [NBSubFR]int32
	LBRRUsage            int32
	GainsQ16             [NBSubFR]int32
	PredCoefQ12          [2][MaxLPCOrder]int16
	LTPCoefQ14           [LTPOrder * NBSubFR]int16
	LTPScaleQ14          int32
	AR1Q13               [NBSubFR * MaxShapeLPCOrder]int16
	AR2Q13               [NBSubFR * MaxShapeLPCOrder]int16
	LFShpQ14             [NBSubFR]int32
	GainsPreQ14          [NBSubFR]int32
	HarmBoostQ14         [NBSubFR]int32
	TiltQ14              [NBSubFR]int32
	HarmShapeGainQ14     [NBSubFR]int32
	LambdaQ10            int32
	inputQualityQ14      int32
	codingQualityQ14     int32
	pitchFreqLowHz       int32
	currentSNRdBQ7       int32
	sparsenessQ8         int32
	predGainQ16          int32
	LTPredCodGainQ7      int32
	inputQualityBandsQ15 [VADNBands]int32
	inputTiltQ15         int32
	ResNrg               [NBSubFR]int32
	ResNrgQ              [NBSubFR]int32
}

type Encoder struct {
	sRC, sRCLBRR           rangeCoderState
	sNSQ, sNSQLBRR         *nsqState
	InHPState              [2]int32
	sLP                    _LP
	sVAD                   _VAD
	_LBRRprevLastGainIndex int32
	prevSigtype            int32
	typeOffsetPrev         int32
	prevLag                int32
	prevLagIndex           int32
	//APIfsHz                     int32
	prevAPIfsHz                 int32
	maxInternalfskHz            int32
	fskHz                       int32
	fskHzchanged                int32
	frameLength                 int32
	subfrLength                 int32
	laPitch                     int32
	laShape                     int32
	shapeWinLength              int32
	_TargetRateBps              int32
	_PacketSizeMS               int32
	_PacketLossPerc             int32
	frameCounter                int32
	_Complexity                 int32
	nStatesDelayedDecision      int32
	useInterpolatedNLSFs        int32
	shapingLPCOrder             int32
	predictLPCOrder             int32
	pitchEstimationComplexity   int32
	pitchEstimationLPCOrder     int32
	pitchEstimationThresholdQ16 int32
	_LTPQuantLowComplexity      int32
	_NLSFMSVQSurvivors          int32
	firstFrameAfterReset        int32
	controlledSinceLastPayload  int32
	warpingQ16                  int32
	inputBuf                    [MaxFrameLength]int16
	inputBufIx                  int32
	nFramesInPayloadBuf         int32
	nBytesInPayloadBuf          int32
	framesSinceOnset            int32
	psNLSFCB                    [2]NLSFCB
	_LBRRBuffer                 [MaxLBRRDelay]_LBRR
	oldestLBRRIdx               int32
	//useInBandFEC                int32
	_LBRREnabled              int32
	_LBRRGainIncreases        int32
	bitrateDiff               int32
	bitrateThresholdUp        int32
	bitrateThresholdDown      int32
	resamplerState            *resampler
	noSpeechCounter           int32
	useDTX                    int32
	inDTX                     int32
	vadFlag                   int32
	sSWBdetect                detectSWB
	q                         [MaxFrameLength]byte
	qLBRR                     [MaxFrameLength]byte
	variableHPsmth1Q15        int32
	variableHPsmth2Q15        int32
	sShape                    *shape
	sPrefilt                  *prefilter
	sPred                     *predict
	xBuf                      [2*MaxFrameLength + LAShapeMax]int16
	_LTPCorrQ15               int32
	muLTPQ8                   int32
	_SNRdBQ7                  int32
	avgGainQ16                int32
	avgGainQ16OneBitPerSample int32
	_BufferedInChannelMS      int32
	speechActivityQ8          int32
	prevLTPredCodGainQ7       int32
	_HPLTPredCodGainQ7        int32
	inBandFECSNRCompQ8        int32
	opts                      *EncodeOption
}

func NewEncoder(opt *EncodeOption) (enc *Encoder, err error) {
	if opt == nil {
		opt = &EncodeOption{}
	}
	opt.init()

	switch opt.SampleRate {
	case KHz8000, KHz12000, KHz16000, KHz44100, KHz24000, KHz32000, KHz48000:
	default:
		return nil, ErrEncodeSampleRateNotSupported
	}

	switch opt.MaxSampleRate {
	case KHz8000, KHz12000, KHz16000, KHz24000:
	default:
		return nil, ErrEncodeSampleRateNotSupported
	}

	encoder := &Encoder{opts: opt}
	encoder.maxInternalfskHz = (opt.MaxSampleRate >> 10) + 1

	return encoder, nil
}

func biquadAlt(in []int16, BQ28, AQ28, S []int32, out []int16, length int32) {
	var k, inVal, A0UQ28, A0LQ28, A1UQ28, A1LQ28, out32Q14 int32

	A0LQ28 = (-AQ28[0]) & 0x00003FFF
	A0UQ28 = rshift(-AQ28[0], 14)
	A1LQ28 = (-AQ28[1]) & 0x00003FFF
	A1UQ28 = rshift(-AQ28[1], 14)

	for k = 0; k < length; k++ {
		inVal = int32(in[k])
		out32Q14 = lshift(smlawb(S[0], BQ28[0], inVal), 2)

		S[0] = S[1] + rrshift(smulwb(out32Q14, A0LQ28), 14)
		S[0] = smlawb(S[0], out32Q14, A0UQ28)
		S[0] = smlawb(S[0], BQ28[1], inVal)

		S[1] = rrshift(smulwb(out32Q14, A1LQ28), 14)
		S[1] = smlawb(S[1], out32Q14, A1UQ28)
		S[1] = smlawb(S[1], BQ28[2], inVal)

		out[k] = sat16(rshift(out32Q14+(1<<14)-1, 14))
	}
}

func (e *Encoder) setupResampler(fskHz int32) (err error) {
	if e.fskHz != fskHz || e.prevAPIfsHz != e.opts.SampleRate {
		if e.fskHz == 0 {
			if e.resamplerState, err = newResampler(e.opts.SampleRate, fskHz*1000); err != nil {
				return
			}
		} else {
			size := (2*MaxFrameLength + LAShapeMax) * (MaxApiFSKHZ / 8)
			xBuf := make([]int16, size, size)

			nSamplesTemp := lshift(e.frameLength, 1) + LAShapeMS*e.fskHz

			if smulbb(fskHz, 1000) < e.opts.SampleRate && e.fskHz != 0 {
				var resamplerState *resampler
				if resamplerState, err = newResampler(smulbb(e.fskHz, 1000), e.opts.SampleRate); err != nil {
					return
				}

				if err = resamplerState.resample(xBuf, e.xBuf[:nSamplesTemp]); err != nil {
					return
				}

				nSamplesTemp = (nSamplesTemp * e.opts.SampleRate) / smulbb(e.fskHz, 1000)

				if e.resamplerState, err = newResampler(e.opts.SampleRate, smulbb(fskHz, 1000)); err != nil {
					return
				}
			} else {
				memcpy(xBuf, e.xBuf[:], int(nSamplesTemp))
			}

			if 100*fskHz != e.opts.SampleRate {
				if err = e.resamplerState.resample(e.xBuf[:], xBuf[:nSamplesTemp]); err != nil {
					return
				}
			}
		}
	}

	e.prevAPIfsHz = e.opts.SampleRate

	return
}

func (e *Encoder) setupComplexity(complexity int32) error {
	if complexity == 0 {
		e._Complexity = 0
		e.pitchEstimationComplexity = PitchESTComplexityLCMode
		e.pitchEstimationThresholdQ16 = fixConst(FindPitchCorrelationThresholdLCMode, 16)
		e.pitchEstimationLPCOrder = 6
		e.shapingLPCOrder = 8
		e.laShape = 3 * e.fskHz
		e.nStatesDelayedDecision = 1
		e.useInterpolatedNLSFs = 0
		e._LTPQuantLowComplexity = 1
		e._NLSFMSVQSurvivors = MaxNLSFMSVQSurvivorsLCMode
		e.warpingQ16 = 0
	} else if complexity == 1 {
		e._Complexity = 1
		e.pitchEstimationComplexity = PitchESTComplexityMCMode
		e.pitchEstimationThresholdQ16 = fixConst(FindPitchCorrelationThresholdMCMode, 16)
		e.pitchEstimationLPCOrder = 12
		e.shapeWinLength = 12
		e.laShape = 5 * e.fskHz
		e.nStatesDelayedDecision = 2
		e.useInterpolatedNLSFs = 0
		e._LTPQuantLowComplexity = 0
		e._NLSFMSVQSurvivors = MaxNLSFMSVQSurvivorsMCMode
		e.warpingQ16 = e.fskHz * fixConst(WarpingMultiplier, 16)
	} else if complexity == 2 {
		e._Complexity = 2
		e.pitchEstimationComplexity = PitchESTComplexityHCMode
		e.pitchEstimationThresholdQ16 = fixConst(FindPitchCorrelationThresholdHCMode, 16)
		e.pitchEstimationLPCOrder = 16
		e.shapeWinLength = 16
		e.laShape = 5 * e.fskHz
		e.nStatesDelayedDecision = MaxDELDECStates
		e.useInterpolatedNLSFs = 1
		e._LTPQuantLowComplexity = 0
		e._NLSFMSVQSurvivors = MaxNLSFMSVQSurvivors
		e.warpingQ16 = e.fskHz * fixConst(WarpingMultiplier, 16)
	} else {
		return ErrEncodeInvalidComplexitySetting
	}

	e.pitchEstimationLPCOrder = min(e.pitchEstimationLPCOrder, e.predictLPCOrder)
	e.shapeWinLength = 5*e.fskHz + 2*e.laShape

	assert(e.pitchEstimationLPCOrder <= MaxFindPitchLPCOrder)
	assert(e.shapingLPCOrder <= MaxShapeLPCOrder)
	assert(e.nStatesDelayedDecision <= MaxDELDECStates)
	assert(e.warpingQ16 <= 32767)
	assert(e.laShape <= LAShapeMax)
	assert(e.shapeWinLength <= ShapeLPCWINMax)

	return nil
}

func (e *Encoder) controlAudioBandwidth(targetRateBps int32) int32 {
	fskHz := e.fskHz
	if fskHz == 0 {
		if targetRateBps >= SWB2WBBitrateBPS {
			fskHz = 24
		} else if targetRateBps >= WB2MBBitrateBPS {
			fskHz = 16
		} else if targetRateBps >= MB2NBBitrateBPS {
			fskHz = 12
		} else {
			fskHz = 8
		}

		fskHz = min(fskHz, e.opts.SampleRate/1000)
		fskHz = min(fskHz, e.maxInternalfskHz)
	} else if smulbb(fskHz, 1000) > e.opts.SampleRate || fskHz > e.maxInternalfskHz {
		fskHz = e.opts.SampleRate / 1000
		fskHz = min(fskHz, e.maxInternalfskHz)
	} else {
		if e.opts.SampleRate > KHz8000 {
			e.bitrateDiff += mul(e._PacketSizeMS, targetRateBps-e.bitrateThresholdDown)
			e.bitrateDiff = min(e.bitrateDiff, 0)

			if e.vadFlag == NoVoiceActivity {
				if e.sLP.transitionFrameNo == 0 &&
					e.bitrateDiff <= -AccumBitsDiffThreshold ||
					e.sSWBdetect.WBDetected*e.fskHz == 24 {
					e.sLP.transitionFrameNo = 1
					e.sLP.mode = 0
				} else if e.sLP.transitionFrameNo >= TransitionFramesDown &&
					e.sLP.mode == 0 {
					e.sLP.transitionFrameNo = 0
					e.bitrateDiff = 0
					if e.fskHz == 24 {
						fskHz = 16
					} else if e.fskHz == 16 {
						fskHz = 12
					} else {
						assert(e.fskHz == 12)
					}
				}

				if (e.fskHz*1000 < e.opts.SampleRate &&
					targetRateBps >= e.bitrateThresholdUp &&
					e.sSWBdetect.WBDetected*e.fskHz < 16) &&
					((e.fskHz == 16 && e.maxInternalfskHz >= 24) ||
						(e.fskHz == 12 && e.maxInternalfskHz >= 16) ||
						(e.fskHz == 8 && e.maxInternalfskHz >= 12)) &&
					e.sLP.transitionFrameNo == 0 {

					e.sLP.mode = 1
					e.bitrateDiff = 0

					switch e.fskHz {
					case 8:
						fskHz = 12
					case 12:
						fskHz = 16
					default:
						assert(e.fskHz == 16)
						fskHz = 24
					}
				}
			}
		}

		if e.sLP.mode == 1 &&
			e.sLP.transitionFrameNo >= TransitionFramesUp &&
			e.vadFlag == NoVoiceActivity {
			e.sLP.transitionFrameNo = 0
			memset(e.sLP.InLPState[:], 0, 2)
		}
	}

	return fskHz
}

func (e *Encoder) _LBRRReset() {
	for i := 0; i < MaxLBRRDelay; i++ {
		e._LBRRBuffer[i].usage = NoLBRR
	}
}

func (e *Encoder) setupPacketSize(packetSizeMS int32) error {
	switch packetSizeMS {
	case 20, 40, 60, 80, 100:
		if packetSizeMS != e._PacketSizeMS {
			e._PacketSizeMS = packetSizeMS
			e._LBRRReset()
		}
	default:
		return ErrEncodePacketSizeNotSupported
	}
	return nil
}

func (e *Encoder) setupFs(fskHz int32) {
	if e.fskHz != fskHz {
		e.sShape = &shape{}
		e.sPrefilt = &prefilter{}
		e.sPred = &predict{}
		e.sNSQ = &nsqState{}
		e.sNSQLBRR = &nsqState{}
		if e.sLP.mode == 1 {
			e.sLP.transitionFrameNo = 1
		} else {
			e.sLP.transitionFrameNo = 0
		}
		e.prevLag = 100
		e.prevSigtype = SIGTypeUnvoiced
		e.firstFrameAfterReset = 1
		e.sPrefilt.lagPrev = 100
		e.sShape.LastGainIndex = 1
		e.sNSQ.lagPrev = 100
		e.sNSQ.prevInvGainQ16 = 65536
		e.sNSQLBRR.prevInvGainQ16 = 65536

		e.fskHz = fskHz
		if e.fskHz == 8 {
			e.predictLPCOrder = MinLPCOrder
			e.psNLSFCB[0] = NLSFCB010
			e.psNLSFCB[1] = NLSFCB110
		} else {
			e.predictLPCOrder = MaxLPCOrder
			e.psNLSFCB[0] = NLSFCB016
			e.psNLSFCB[1] = NLSFCB116
		}

		e.frameLength = smulbb(FrameLengthMS, fskHz)
		e.subfrLength = e.frameLength / NBSubFR
		e.laPitch = smulbb(LAPitchMS, fskHz)
		e.sPred.minPitchLag = smulbb(3, fskHz)
		e.sPred.maxPitchLag = smulbb(18, fskHz)
		e.sPred.pitchLPCWinLength = smulbb(FindPitchLPCWINMS, fskHz)

		if e.fskHz == 24 {
			e.muLTPQ8 = fixConst(MULTPQuantSWB, 8)
			e.bitrateThresholdUp = math.MaxInt32
			e.bitrateThresholdDown = SWB2WBBitrateBPS
		} else if e.fskHz == 16 {
			e.muLTPQ8 = fixConst(MULTPQuantWB, 8)
			e.bitrateThresholdUp = WB2SWBBitrateBPS
			e.bitrateThresholdDown = WB2MBBitrateBPS
		} else if e.fskHz == 12 {
			e.muLTPQ8 = fixConst(MULTPQuantMB, 8)
			e.bitrateThresholdUp = MB2WBBitrateBPS
			e.bitrateThresholdDown = MB2NBBitrateBPS
		} else {
			e.muLTPQ8 = fixConst(MULTPQuantNB, 8)
			e.bitrateThresholdUp = NB2MBBitrateBPS
			e.bitrateThresholdDown = 0
		}
		e.fskHzchanged = 1
		assert(e.subfrLength*NBSubFR == e.frameLength)
	}
}

func (e *Encoder) setupRate(targetRateBPS int32) {
	var rateTable []int32

	if targetRateBPS != e._TargetRateBps {
		e._TargetRateBps = targetRateBPS

		switch e.fskHz {
		case 8:
			rateTable = TargetRateTableNB
		case 12:
			rateTable = TargetRateTableMB
		case 16:
			rateTable = TargetRateTableWB
		default:

			rateTable = TargetRateTableSWB
		}

		for k := int32(1); k < TargetRateTabSZ; k++ {
			if targetRateBPS <= rateTable[k] {
				fracQ6 := lshift(targetRateBPS-rateTable[k-1], 6) / (rateTable[k] - rateTable[k-1])
				e._SNRdBQ7 = lshift(SNRTableQ1[k-1], 6) + mul(fracQ6, SNRTableQ1[k]-SNRTableQ1[k-1])
				break
			}
		}
	}
}

func (e *Encoder) setupLBRR() {
	var LBRRRateThresBps int32

	if e.opts.UseInBandFEC {
		e._LBRREnabled = 1
	}

	switch e.fskHz {
	case 8:
		LBRRRateThresBps = InBandFECMinRateBPS - 9000
	case 12:
		LBRRRateThresBps = InBandFECMinRateBPS - 6000
	case 16:
		LBRRRateThresBps = InBandFECMinRateBPS - 3000
	default:
		LBRRRateThresBps = InBandFECMinRateBPS
	}

	if e._TargetRateBps >= LBRRRateThresBps {
		e._LBRRGainIncreases = max(8-rshift(e._PacketLossPerc, 1), 0)

		if e._LBRREnabled > 0 && e._PacketLossPerc > LBRRLossThres {
			e.inBandFECSNRCompQ8 = fixConst(6.0, 8) - lshift(e._LBRRGainIncreases, 7)
		} else {
			e.inBandFECSNRCompQ8 = 0
			e._LBRREnabled = 0
		}
	} else {
		e.inBandFECSNRCompQ8 = 0
		e._LBRREnabled = 0
	}
}

func (e *Encoder) controlEncoder(packetSizeMS, targetRateBps,
	packetLossPerc int32, DTXEnabled bool, complexity int32) (err error) {
	if e.controlledSinceLastPayload != 0 {
		if e.opts.SampleRate != e.prevAPIfsHz && e.fskHz > 0 {
			if err = e.setupResampler(e.opts.SampleRate); err != nil {
				return
			}
		}
		return
	}

	var fskHz int32

	fskHz = e.controlAudioBandwidth(targetRateBps)

	if err = e.setupResampler(fskHz); err != nil {
		return
	}

	if err = e.setupPacketSize(packetSizeMS); err != nil {
		return
	}

	e.setupFs(fskHz)

	if err = e.setupComplexity(complexity); err != nil {
		return
	}

	e.setupRate(targetRateBps)

	if packetLossPerc < 0 || packetLossPerc > 100 {
		return ErrEncodeInvalidLossRate
	}
	e._PacketLossPerc = packetLossPerc

	e.setupLBRR()

	if DTXEnabled {
		e.useDTX = 1
	}
	e.controlledSinceLastPayload = 1

	return
}

func (e *Encoder) _HPVariableCutoff(psEncCtrl *encoderControl, out, in []int16) {
	var (
		qualityQ15, Fcq19, rQ28, rQ22               int32
		pitchFreqHzQ16, pitchFreqLogQ7, deltaFreqQ7 int32
		BQ28                                        [3]int32
		AQ28                                        [2]int32
	)

	if e.prevSigtype == SIGTypeVoiced {
		pitchFreqHzQ16 = div(lshift(mul(e.fskHz, 1000), 16), e.prevLag)
		pitchFreqLogQ7 = lin2log(pitchFreqHzQ16) - (16 << 7)

		qualityQ15 = psEncCtrl.inputQualityBandsQ15[0]
		pitchFreqLogQ7 = pitchFreqLogQ7 - smulwb(smulwb(lshift(qualityQ15, 2), qualityQ15),
			pitchFreqLogQ7-LOG2VariableHPMinFREQQ7)
		pitchFreqLogQ7 = pitchFreqLogQ7 + rshift(fixConst(0.6, 15)-qualityQ15, 9)

		deltaFreqQ7 = pitchFreqLogQ7 - rshift(e.variableHPsmth1Q15, 8)
		if deltaFreqQ7 < 0 {
			deltaFreqQ7 = mul(deltaFreqQ7, 3)
		}

		deltaFreqQ7 = limit(deltaFreqQ7, -fixConst(VariableHPMaxDeltaFREQ, 7),
			fixConst(VariableHPMaxDeltaFREQ, 7))

		e.variableHPsmth1Q15 = smlawb(e.variableHPsmth1Q15,
			mul(lshift(e.speechActivityQ8, 1), deltaFreqQ7), fixConst(VariableHPSMTHCoef1, 16))
	}
	e.variableHPsmth2Q15 = smlawb(e.variableHPsmth2Q15,
		e.variableHPsmth1Q15-e.variableHPsmth2Q15, fixConst(VariableHPSMTHCoef2, 16))

	psEncCtrl.pitchFreqLowHz = log2lin(rshift(e.variableHPsmth2Q15, 8))

	psEncCtrl.pitchFreqLowHz = limit(psEncCtrl.pitchFreqLowHz,
		fixConst(VariableHPMinFREQ, 0), fixConst(VariableHPMaxFREQ, 0))

	assert(psEncCtrl.pitchFreqLowHz <= math.MaxInt32/RadiansConstantQ19)
	Fcq19 = div(smulbb(RadiansConstantQ19, psEncCtrl.pitchFreqLowHz), e.fskHz)
	assert(Fcq19 >= 3704)
	assert(Fcq19 <= 27787)

	rQ28 = fixConst(1.0, 28) - mul(fixConst(0.92, 9), Fcq19)
	assert(rQ28 >= 255347779)
	assert(rQ28 <= 266690872)

	BQ28[0] = rQ28
	BQ28[1] = lshift(-rQ28, 1)
	BQ28[2] = rQ28

	rQ22 = rshift(rQ28, 6)
	AQ28[0] = smulww(rQ22, smulww(Fcq19, Fcq19)-fixConst(2.0, 22))
	AQ28[1] = smulww(rQ22, rQ22)

	biquadAlt(in, BQ28[:], AQ28[:], e.InHPState[:], out, e.frameLength)
}

func applySineWindow(pxWin, px []int16, winType, length int32) {
	var (
		k, fQ16, cQ16 int32
		S0Q16, S1Q16  int32
		px32          int32
	)

	assert(winType == 1 || winType == 2)

	assert(length >= 16 && length <= 120)
	assert(length&3 == 0)

	k = (length >> 2) - 4
	assert(k >= 0 && k <= 26)

	fQ16 = int32(freqTableQ16[k])

	cQ16 = smulwb(fQ16, -fQ16)
	assert(cQ16 >= -32768)

	if winType == 1 {
		S0Q16 = 0
		S1Q16 = fQ16 + rshift(length, 3)
	} else {
		S0Q16 = 1 << 16
		S1Q16 = (1 << 16) + rshift(cQ16, 1) + rshift(length, 4)
	}

	for k = 0; k < length; k += 4 {
		px32 = ua2i32(px[k:])
		pxWin[k] = int16(smulwb(rshift(S0Q16+S1Q16, 1), px32))
		pxWin[k+1] = int16(smulwt(S1Q16, px32))

		S0Q16 = smulwb(S1Q16, cQ16) + lshift(S1Q16, 1) - S0Q16 + 1
		S0Q16 = min(S0Q16, 1<<16)

		px32 = ua2i32(px[k+2:])
		pxWin[k+2] = int16(smulwb(rshift(S0Q16+S1Q16, 1), px32))
		pxWin[k+3] = int16(smulwt(S0Q16, px32))

		S1Q16 = smulwb(S0Q16, cQ16) + lshift(S0Q16, 1) - S1Q16
		S1Q16 = min(S1Q16, 1<<16)
	}
}

func innerProd16Aligned64(inVec1, inVec2 []int16, length int32) int64 {
	var sum int64
	for i := int32(0); i < length; i++ {
		sum = smlalbb(sum, inVec1[i], inVec2[i])
	}
	return sum
}

func innerProdAligned(inVec1, inVec2 []int16, length int32) int32 {
	var sum int32
	for i := int32(0); i < length; i++ {
		sum = smlabb(sum, int32(inVec1[i]), int32(inVec2[i]))
	}
	return sum
}

const (
	QS = 14
	QC = 10
)

func autoCorrelation(corr []int32, scale *int32, input []int16, warpingQ16 int16, length, order int32) {
	var n, i, lsh int32
	var tmp1QS, tmp2QS int32
	var (
		stateQS [MaxShapeLPCOrder + 1]int32
		corrQC  [MaxShapeLPCOrder + 1]int64
	)

	assert(order&1 == 0)

	for n = 0; n < length; n++ {
		tmp1QS = lshift(int32(input[n]), QS)
		for i = 0; i < order; i += 2 {
			tmp2QS = smlawb(stateQS[i], stateQS[i+1]-tmp1QS, int32(warpingQ16))
			stateQS[i] = tmp1QS
			corrQC[i] += smull(tmp1QS, stateQS[0]) >> (2*QS - QC)

			tmp1QS = smlawb(stateQS[i+1], stateQS[i+2]-tmp2QS, int32(warpingQ16))
			stateQS[i+1] = tmp2QS
			corrQC[i+1] += smull(tmp2QS, stateQS[0]) >> (2*QS - QC)
		}
		stateQS[order] = tmp1QS
		corrQC[order] += smull(tmp1QS, stateQS[0]) >> (2*QS - QC)
	}

	lsh = clz64(corrQC[0]) - 35
	lsh = limit(lsh, -12-QC, 30-QC)

	*scale = -(QC + lsh)
	assert(*scale >= -30 && *scale <= 12)

	if lsh >= 0 {
		for i = 0; i < order+1; i++ {
			corr[i] = corrQC[i] << lsh
		}
	} else {
		for i = 0; i < order+1; i++ {
			corr[i] = int32(corrQC[i] >> -lsh)
		}
	}
	assert(corrQC[0] >= 0)
}

func autocorr(results []int32, scale *int32, inputData []int16, inputDataSize, correlationCount int32) {
	var (
		i, lz, nRightShifts, corrCount int32
		corr64                         int64
	)

	corrCount = min(inputDataSize, correlationCount)

	corr64 = innerProd16Aligned64(inputData, inputData, inputDataSize)

	corr64 += 1

	lz = clz64(corr64)

	nRightShifts = 35 - lz
	*scale = nRightShifts

	if nRightShifts <= 0 {
		results[0] = lshift(int32(corr64), -nRightShifts)

		for i = 1; i < corrCount; i++ {
			results[i] = lshift(innerProdAligned(inputData, inputData[i:], inputDataSize-i), -nRightShifts)
		}
	} else {
		results[0] = int32(corr64 >> nRightShifts)

		for i = 1; i < corrCount; i++ {
			results[i] = int32(innerProd16Aligned64(inputData, inputData[i:], inputDataSize-i) >> nRightShifts)
		}
	}
}

func schur64(rcQ16 []int32, c []int32, order int32) int32 {
	var k, n int32
	var C [MaxLPCOrder + 1][2]int32
	var Ctmp1Q30, Ctmp2Q30, rcTmpQ31 int32

	if c[0] <= 0 {
		memset(rcQ16, 0, int(order))
		return 0
	}

	for k = 0; k < order+1; k++ {
		C[k][0] = c[k]
		C[k][1] = C[k][0]
	}

	for k = 0; k < order; k++ {
		rcTmpQ31 = div32varQ(-C[k+1][0], C[0][1], 31)
		rcQ16[k] = rrshift(rcTmpQ31, 15)

		for n = 0; n < order-k; n++ {
			Ctmp1Q30 = C[n+k+1][0]
			Ctmp2Q30 = C[n][1]

			C[n+k+1][0] = Ctmp1Q30 + smmul(lshift(Ctmp2Q30, 1), rcTmpQ31)
			C[n][1] = Ctmp2Q30 + smmul(lshift(Ctmp1Q30, 1), rcTmpQ31)
		}
	}

	return C[0][1]
}

func schur(rcQ15 []int16, c []int32, order int32) int32 {
	var k, n, lz int32
	var (
		C                      [MaxLPCOrder + 1][2]int32
		CTmp1, CTmp2, rcTmpQ15 int32
	)
	lz = clz32(c[0])
	if lz < 2 {
		for k = 0; k < order+1; k++ {
			C[k][0] = rshift(c[k], 1)
			C[k][1] = C[k][0]
		}
	} else if lz > 2 {
		lz -= 2
		for k = 0; k < order+1; k++ {
			C[k][0] = lshift(c[k], lz)
			C[k][1] = C[k][0]
		}
	} else {
		for k = 0; k < order+1; k++ {
			C[k][0] = c[k]
			C[k][1] = C[k][0]
		}
	}

	for k = 0; k < order; k++ {
		rcTmpQ15 = -(C[k+1][0] / max(rshift(C[0][1], 15), 1))
		rcTmpQ15 = int32(sat16(rcTmpQ15))
		rcQ15[k] = int16(rcTmpQ15)

		for n = 0; n < order-k; n++ {
			CTmp1 = C[n+k+1][0]
			CTmp2 = C[n][1]
			C[n+k+1][0] = smlawb(CTmp1, lshift(CTmp2, 1), rcTmpQ15)
			C[n][1] = smlawb(CTmp2, lshift(CTmp1, 1), rcTmpQ15)
		}
	}

	return C[0][1]
}

func k2aQ16(AQ24 []int32, rcQ16 []int32, order int32) {
	var k, n int32
	var atmp [MaxLPCOrder]int32

	for k = 0; k < order; k++ {
		for n = 0; n < k; n++ {
			atmp[n] = AQ24[n]
		}
		for n = 0; n < k; n++ {
			AQ24[n] = smlaww(AQ24[n], atmp[k-n-1], rcQ16[k])
		}
		AQ24[k] = -lshift(rcQ16[k], 8)
	}
}

func k2a(AQ24 []int32, rcQ15 []int16, order int32) {
	var k, n int32
	var atmp [MaxLPCOrder]int32

	for k = 0; k < order; k++ {
		for n = 0; n < k; n++ {
			atmp[n] = AQ24[n]
		}
		for n = 0; n < k; n++ {
			AQ24[n] = smlawb(AQ24[n], lshift(atmp[k-n-1], 1), int32(rcQ15[k]))
		}
		AQ24[k] = -lshift(int32(rcQ15[k]), 9)
	}
}

func pitchAnalysisCore(signal []int16, pitchOut []int32,
	lagIndex, contourIndex, LTPCorrQ15 *int32,
	prevLag, searchThres1Q16, searchThres2Q15, fskHz, complexity,
	forLJC int32) int32 {
	var (
		signal8kHz [PitchESTMaxFrameLengthST2]int16
		signal4kHz [PitchESTMaxFrameLengthST1]int16
		scratchMem [(3 * PitchESTMaxFrameLength) * 2]int16

		inputSignalPtr []int16
		filtState      [PitchESTMaxDecimateStateLength]int32

		i, k, d, j int32

		C                   [PitchESTNBSubFR][(PitchESTMaxLag >> 1) + 5]int16
		targetPtr, basisPtr []int16

		crossCorr, normalizer, energy, shift, energyBasis, energyTarget int32
		dSrch                                                           [PitchESTDSRCHLength]int32
		dComp                                                           [(PitchESTMaxLag >> 1) + 5]int16
		Cmax, lengthDSrch, lengthDComp                                  int32

		sum, threshold, temp32 int32

		CBimax, CBimaxNew, CBimaxOld, lag, startLag, endLag, lagNew int32

		CC                                                                        [PitchESTNBCBKSStage2Ext]int32
		CCmax, CCmaxb, CCmaxnewb, CCmaxnew                                        int32
		energiesSt3                                                               [][][]int32
		crosscorrSt3                                                              [][][]int32
		lagCounter                                                                int32
		frameLength, frameLength8kHz, frameLength4kHz, maxSumSqLength             int32
		sfLength, sfLength8kHz                                                    int32
		minLag, minLag8kHz, minLag4kHz                                            int32
		maxLag, maxLag8kHz, maxLag4kHz                                            int32
		contourBias, diff                                                         int32
		lz, ls                                                                    int32
		cbkOffset, cbkSize, nbCbksStage2                                          int32
		deltaLagLog2sqrQ7, lagLog2Q7, prevLagLog2Q7, prevLagBiasQ15, corrThresQ15 int32
	)

	energiesSt3 = make([][][]int32, PitchESTNBSubFR, PitchESTNBSubFR)
	crosscorrSt3 = make([][][]int32, PitchESTNBSubFR, PitchESTNBSubFR)
	for k = 0; k < PitchESTNBSubFR; k++ {
		energiesSt3[k] = make([][]int32, PitchESTNBCBKSStage3Max, PitchESTNBCBKSStage3Max)
		crosscorrSt3[k] = make([][]int32, PitchESTNBCBKSStage3Max, PitchESTNBCBKSStage3Max)
		for d = 0; d < PitchESTNBCBKSStage3Max; d++ {
			energiesSt3[k][d] = make([]int32, PitchESTNBStage3Lags, PitchESTNBStage3Lags)
			crosscorrSt3[k][d] = make([]int32, PitchESTNBStage3Lags, PitchESTNBStage3Lags)
		}
	}
	k, d = 0, 0

	assert(fskHz == 8 || fskHz == 12 || fskHz == 16 || fskHz == 24)
	assert(complexity >= PitchESTMinComplex)
	assert(complexity <= PitchESTMaxComplex)

	assert(searchThres1Q16 >= 0 && searchThres1Q16 <= (1<<16))
	assert(searchThres2Q15 >= 0 && searchThres2Q15 <= (1<<15))

	frameLength = PitchESTFrameLengthMS * fskHz
	frameLength4kHz = PitchESTFrameLengthMS * 4
	frameLength8kHz = PitchESTFrameLengthMS * 8

	sfLength = rshift(frameLength, 3)
	sfLength8kHz = rshift(frameLength8kHz, 3)

	minLag = PitchESTMinLagMS * fskHz
	minLag4kHz = PitchESTMinLagMS * 4
	minLag8kHz = PitchESTMinLagMS * 8

	maxLag = PitchESTMaxLagMS * fskHz
	maxLag4kHz = PitchESTMaxLagMS * 4
	maxLag8kHz = PitchESTMaxLagMS * 8

	switch fskHz {
	case 16:
		resampleDown2(filtState[:], signal8kHz[:], signal[:frameLength])
	case 12:
		var R23 [6]int32
		resampleDown23(R23[:], signal8kHz[:], signal[:PitchESTFrameLengthMS*12])
	case 24:
		var filtStateFix [8]int32
		resampleDown3(filtStateFix[:], signal8kHz[:], signal[:24*PitchESTFrameLengthMS])
	default:
		assert(fskHz == 8)
		memcpy(signal8kHz[:], signal[:], int(frameLength8kHz))
	}

	memset(filtState[:], 0, 2)
	resampleDown2(filtState[:], signal4kHz[:], signal8kHz[:frameLength8kHz])

	for i = frameLength4kHz - 1; i > 0; i-- {
		signal4kHz[i] = sat16(int32(signal4kHz[i] + signal4kHz[i-1]))
	}

	maxSumSqLength = max(sfLength8kHz, rshift(frameLength4kHz, 1))
	shift = _PAnaFindScaling(signal4kHz[:], frameLength4kHz, maxSumSqLength)
	if shift > 0 {
		for i = 0; i < frameLength4kHz; i++ {
			signal4kHz[i] = int16(rshift(int32(signal4kHz[i]), shift))
		}
	}

	tp := rshift(frameLength4kHz, 1)
	targetPtr = signal4kHz[tp:]
	for k = 0; k < 2; k++ {
		basisPtr = signal4kHz[tp-minLag4kHz:]

		normalizer = 0
		crossCorr = 0

		crossCorr = innerProdAligned(targetPtr, basisPtr, sfLength8kHz)
		normalizer = innerProdAligned(basisPtr, basisPtr, sfLength8kHz)
		normalizer = addSAT32(normalizer, smulbb(sfLength8kHz, 4000))

		temp32 = div(crossCorr, sqrtApprox(normalizer)+1)
		C[k][minLag4kHz] = sat16(temp32)

		for d = minLag4kHz + 1; d <= maxLag4kHz; d++ {
			basisPtr = signal4kHz[tp-minLag4kHz-(k*sfLength8kHz):]

			crossCorr = innerProdAligned(targetPtr, basisPtr, sfLength8kHz)

			normalizer +=
				smulbb(int32(basisPtr[0]), int32(basisPtr[0])) -
					smulbb(int32(basisPtr[sfLength8kHz]), int32(basisPtr[sfLength8kHz]))

			temp32 = div(crossCorr, sqrtApprox(normalizer)+1)
			C[k][d] = sat16(temp32)
		}

		targetPtr = targetPtr[sfLength8kHz:]
	}

	for i = maxLag4kHz; i >= minLag4kHz; i-- {
		sum = int32(C[0][i]) + int32(C[1][i])
		assert(rshift(sum, 1) == int32(sat16(rshift(sum, 1))))

		sum = rshift(sum, 1)
		assert(lshift(-i, 4) == int32(sat16(lshift(-i, 4))))

		sum = smlawb(sum, sum, lshift(-i, 4))
		assert(sum == int32(sat16(sum)))

		C[0][i] = int16(sum)
	}

	lengthDSrch = 4 + 2*complexity
	assert(3*lengthDSrch <= PitchESTDSRCHLength)

	insertionSortDecreasingInt16(C[0][minLag4kHz:], dSrch[:], maxLag4kHz-minLag4kHz+1, lengthDSrch)

	targetPtr = signal4kHz[rshift(frameLength4kHz, 1):]
	energy = innerProdAligned(targetPtr, targetPtr, rshift(frameLength4kHz, 1))
	energy = addPosSAT32(energy, 1000)

	Cmax = int32(C[0][minLag4kHz])
	threshold = smulbb(Cmax, Cmax)

	if rshift(energy, 4+2) > threshold {
		memset(pitchOut, 0, PitchESTNBSubFR)
		*LTPCorrQ15 = 0
		*lagIndex = 0
		*contourIndex = 0
		return 1
	}

	threshold = smulwb(searchThres1Q16, Cmax)
	for i = 0; i < lengthDSrch; i++ {
		if int32(C[0][minLag4kHz+i]) > threshold {
			dSrch[i] = (dSrch[i] + minLag4kHz) << 1
		} else {
			lengthDSrch = i
			break
		}
	}
	assert(lengthDSrch > 0)

	for i = minLag8kHz - 5; i < maxLag8kHz+5; i++ {
		dComp[i] = 0
	}
	for i = 0; i < lengthDSrch; i++ {
		dComp[dSrch[i]] = 1
	}

	for i = maxLag8kHz + 3; i >= minLag8kHz; i-- {
		dComp[i] += dComp[i-1] + dComp[i-2]
	}

	lengthDSrch = 0
	for i = minLag8kHz; i < maxLag8kHz+1; i++ {
		if dComp[i+1] > 0 {
			dSrch[lengthDSrch] = i
			lengthDSrch++
		}
	}

	for i = maxLag8kHz + 3; i >= minLag8kHz; i-- {
		dComp[i] += dComp[i-1] + dComp[i-2] + dComp[i-3]
	}

	lengthDComp = 0
	for i = minLag8kHz; i < maxLag8kHz+4; i++ {
		if dComp[i] > 0 {
			dComp[lengthDComp] = int16(i - 2)
			lengthDComp++
		}
	}

	shift = _PAnaFindScaling(signal8kHz[:], frameLength8kHz, sfLength8kHz)
	if shift > 0 {
		for i = 0; i < frameLength8kHz; i++ {
			signal8kHz[i] = int16(rshift(int32(signal8kHz[i]), shift))
		}
	}

	for x := 0; x < len(C); x++ {
		for q := 0; q < len(C[x]); q++ {
			C[x][q] = 0
		}
	}

	targetPtr = signal8kHz[frameLength4kHz:]
	for k = 0; k < PitchESTNBSubFR; k++ {
		energyTarget = innerProdAligned(targetPtr, targetPtr, sfLength8kHz)

		for j = 0; j < lengthDComp; j++ {
			d = int32(dComp[j])
			basisPtr = signal8kHz[frameLength4kHz-d+(k*sfLength8kHz):]
			crossCorr = innerProdAligned(targetPtr, basisPtr, sfLength8kHz)
			energyBasis = innerProdAligned(basisPtr, basisPtr, sfLength8kHz)

			if crossCorr > 0 {
				energy = max(energyTarget, energyBasis)
				lz = clz32(crossCorr)
				ls = limit(lz-1, 0, 15)
				temp32 = div(lshift(crossCorr, ls), rshift(energy, 15-ls)+1)
				assert(temp32 == int32(sat16(temp32)))

				temp32 = smulwb(crossCorr, temp32)
				temp32 = addSAT32(temp32, temp32)
				lz = clz32(temp32)
				ls = limit(lz-1, 0, 15)
				energy = min(energyTarget, energyBasis)
				C[k][d] = int16(div(lshift(temp32, ls), rshift(energy, 15-ls)+1))
			} else {
				C[k][d] = 0
			}
		}
		targetPtr = targetPtr[sfLength8kHz:]
	}

	CCmax = math.MinInt32
	CCmaxb = math.MinInt32

	CBimax = 0
	lag = -1

	if prevLag > 0 {
		switch fskHz {
		case 12:
			prevLag = div(lshift(prevLag, 1), 3)
		case 16:
			prevLag = rshift(prevLag, 1)
		case 24:
			prevLag = div(prevLag, 3)
		}
		prevLagLog2Q7 = lin2log(prevLag)
	} else {
		prevLagLog2Q7 = 0
	}
	assert(searchThres2Q15 == int32(sat16(searchThres2Q15)))

	corrThresQ15 = rshift(smulbb(searchThres2Q15, searchThres2Q15), 13)

	if fskHz == 8 && complexity > PitchESTMinComplex {
		nbCbksStage2 = PitchESTNBCBKSStage2Ext
	} else {
		nbCbksStage2 = PitchESTNBCBKSStage2
	}

	for k = 0; k < lengthDSrch; k++ {
		d = dSrch[k]
		for j = 0; j < nbCbksStage2; j++ {
			CC[j] = 0
			for i = 0; i < PitchESTNBSubFR; i++ {
				CC[j] = CC[j] + int32(C[i][d+int32(CBLagsStage2[i][j])])
			}
		}

		CCmaxnew = math.MinInt32
		CBimaxNew = 0
		for i = 0; i < nbCbksStage2; i++ {
			if CC[i] > CCmaxnew {
				CCmaxnew = CC[i]
				CBimaxNew = i
			}
		}

		lagLog2Q7 = lin2log(d)
		assert(lagLog2Q7 == int32(sat16(lagLog2Q7)))

		if forLJC > 0 {
			CCmaxnewb = CCmaxnew
		} else {
			CCmaxnewb = CCmaxnew - rshift(smulbb(PitchESTNBSubFR*PitchESTShortLAGBIASQ15, lagLog2Q7), 7)
		}

		if prevLag > 0 {
			deltaLagLog2sqrQ7 = lagLog2Q7 - prevLagLog2Q7
			assert(deltaLagLog2sqrQ7 == int32(sat16(deltaLagLog2sqrQ7)))
			deltaLagLog2sqrQ7 = rshift(smulbb(deltaLagLog2sqrQ7, deltaLagLog2sqrQ7), 7)
			prevLagBiasQ15 = rshift(smulbb(PitchESTNBSubFR*PitchESTPrevLAGBIASQ15, *LTPCorrQ15), 15)
			prevLagBiasQ15 = div(mul(prevLagBiasQ15, deltaLagLog2sqrQ7), deltaLagLog2sqrQ7+(1<<6))
			CCmaxnewb -= prevLagBiasQ15
		}

		if CCmaxnewb > CCmaxb &&
			CCmaxnew > corrThresQ15 &&
			int32(CBLagsStage2[0][CBimaxNew]) <= minLag8kHz {
			CCmaxb = CCmaxnewb
			CCmax = CCmaxnew
			lag = d
			CBimax = CBimaxNew
		}
	}

	if lag == -1 {
		memset(pitchOut, 0, PitchESTNBSubFR)
		*LTPCorrQ15 = 0
		*lagIndex = 0
		*contourIndex = 0
		return 1
	}

	if fskHz > 8 {
		shift = _PAnaFindScaling(signal, frameLength, sfLength)
		if shift > 0 {
			inputSignalPtr = scratchMem[:]
			for i = 0; i < frameLength; i++ {
				inputSignalPtr[i] = int16(rshift(int32(signal[i]), shift))
			}
		} else {
			inputSignalPtr = signal
		}

		CBimaxOld = CBimax
		assert(lag == int32(sat16(lag)))

		if fskHz == 12 {
			lag = rshift(smulbb(lag, 3), 1)
		} else if fskHz == 16 {
			lag = lshift(lag, 1)
		} else {
			lag = smulbb(lag, 3)
		}

		lag = limit(lag, minLag, maxLag)
		startLag = max(lag-2, minLag)
		endLag = min(lag+2, maxLag)
		lagNew = lag
		CBimax = 0
		assert(lshift(CCmax, 13) >= 0)
		*LTPCorrQ15 = sqrtApprox(lshift(CCmax, 13))

		CCmax = math.MinInt32
		for k = 0; k < PitchESTNBSubFR; k++ {
			pitchOut[k] = lag + 2*int32(CBLagsStage2[k][CBimaxOld])
		}

		_PAnaCalcCorrSt3(crosscorrSt3, inputSignalPtr, startLag, sfLength, complexity)
		_PAnaCalcEnergySt3(energiesSt3, inputSignalPtr, startLag, sfLength, complexity)

		lagCounter = 0
		assert(lag == int32(sat16(lag)))
		contourBias = div(PitchESTFlatContourBIASQ20, lag)

		cbkSize = int32(cbkSizesStage3[complexity])
		cbkOffset = int32(cbkOffsetsStage3[complexity])

		for d = startLag; d <= endLag; d++ {
			for j = cbkOffset; j < cbkOffset+cbkSize; j++ {
				crossCorr = 0
				energy = 0
				for k = 0; k < PitchESTNBSubFR; k++ {
					energy += rshift(energiesSt3[k][j][lagCounter], 2)
					assert(energy >= 0)
					crossCorr += rshift(crosscorrSt3[k][j][lagCounter], 2)
				}
				if crossCorr > 0 {
					lz = clz32(crossCorr)
					ls = limit(lz-1, 0, 13)
					CCmaxnew = div(lshift(crossCorr, ls), rshift(energy, 13-ls)+1)
					CCmaxnew = int32(sat16(CCmaxnew))
					CCmaxnew = smulwb(crossCorr, CCmaxnew)

					if CCmaxnew > rshift(math.MaxInt32, 3) {
						CCmaxnew = math.MaxInt32
					} else {
						CCmaxnew = lshift(CCmaxnew, 3)
					}

					diff = j - rshift(PitchESTNBCBKSStage3Max, 1)
					diff = mul(diff, diff)
					diff = math.MaxInt16 - rshift(mul(contourBias, diff), 5)
					assert(diff == int32(sat16(diff)))
					CCmaxnew = lshift(smulwb(CCmaxnew, diff), 1)
				} else {
					CCmaxnew = 0
				}

				if CCmaxnew > CCmax &&
					(d+int32(CBLagsStage3[0][j])) <= maxLag {
					CCmax = CCmaxnew
					lagNew = d
					CBimax = j
				}
			}
			lagCounter++
		}

		for k = 0; k < PitchESTNBSubFR; k++ {
			pitchOut[k] = lagNew + int32(CBLagsStage3[k][CBimax])
		}
		*lagIndex = lagNew - minLag
		*contourIndex = CBimax
	} else {
		CCmax = max(CCmax, 0)
		*LTPCorrQ15 = sqrtApprox(lshift(CCmax, 13))
		for k = 0; k < PitchESTNBSubFR; k++ {
			pitchOut[k] = lag + int32(CBLagsStage2[k][CBimax])
		}
		*lagIndex = lag - minLag8kHz
		*contourIndex = CBimax
	}
	assert(*lagIndex >= 0)
	return 0
}

func (e *Encoder) findPitchLags(psEncCtrl *encoderControl, res, x []int16) {
	psPredSt := e.sPred
	var (
		bufLen, i, scale  int32
		thrhldQ15, resNrg int32
		xBuf, xBufPtr     []int16
	)
	var (
		Wsig      [FindPitchLPCWINMax]int16
		WsigPtr   []int16
		autoCorr  [MaxFindPitchLPCOrder + 1]int32
		rcQ15     [MaxFindPitchLPCOrder]int16
		AQ24      [MaxFindPitchLPCOrder]int32
		FiltState [MaxFindPitchLPCOrder]int32
		AQ12      [MaxFindPitchLPCOrder]int16
	)

	bufLen = e.laPitch + lshift(e.frameLength, 1)

	assert(bufLen >= psPredSt.pitchLPCWinLength)

	xBuf = x[:]
	xBufPtr = xBuf[bufLen-psPredSt.pitchLPCWinLength:]
	WsigPtr = Wsig[:]
	applySineWindow(WsigPtr, xBufPtr, 1, e.laPitch)

	WsigPtr = WsigPtr[e.laPitch:]
	xBufPtr = xBufPtr[e.laPitch:]
	memcpy(WsigPtr, xBufPtr, int(psPredSt.pitchLPCWinLength-lshift(e.laPitch, 1)))

	WsigPtr = WsigPtr[psPredSt.pitchLPCWinLength-lshift(e.laPitch, 1):]
	xBufPtr = xBufPtr[psPredSt.pitchLPCWinLength-lshift(e.laPitch, 1):]
	applySineWindow(WsigPtr, xBufPtr, 2, e.laPitch)

	autocorr(autoCorr[:], &scale, Wsig[:], psPredSt.pitchLPCWinLength, e.pitchEstimationLPCOrder+1)

	autoCorr[0] = smlawb(autoCorr[0], autoCorr[0], fixConst(FindPitchWhiteNoiseFraction, 16))

	resNrg = schur(rcQ15[:], autoCorr[:], e.pitchEstimationLPCOrder)

	psEncCtrl.predGainQ16 = div32varQ(autoCorr[0], max(resNrg, 1), 16)

	k2a(AQ24[:], rcQ15[:], e.pitchEstimationLPCOrder)

	for i = 0; i < e.pitchEstimationLPCOrder; i++ {
		AQ12[i] = sat16(rshift(AQ24[i], 12))
	}

	bwexpander(AQ12[:], e.pitchEstimationLPCOrder, fixConst(FindPitchBandWithExpansion, 16))

	_MAPrediction(xBuf, AQ12[:], FiltState[:], nil, bufLen, e.pitchEstimationLPCOrder)

	thrhldQ15 = fixConst(0.45, 15)
	thrhldQ15 = smlabb(thrhldQ15, fixConst(-0.004, 15), e.pitchEstimationLPCOrder)
	thrhldQ15 = smlabb(thrhldQ15, fixConst(-0.1, 7), e.speechActivityQ8)
	thrhldQ15 = smlabb(thrhldQ15, fixConst(0.15, 15), e.prevSigtype)
	thrhldQ15 = smlawb(thrhldQ15, fixConst(-0.1, 16), psEncCtrl.inputTiltQ15)
	thrhldQ15 = int32(sat16(thrhldQ15))

	psEncCtrl.sigType = pitchAnalysisCore(res, psEncCtrl.pitchL[:], &psEncCtrl.lagIndex,
		&psEncCtrl.contourIndex, &e._LTPCorrQ15, e.prevLag, e.pitchEstimationThresholdQ16,
		thrhldQ15, e.fskHz, e.pitchEstimationComplexity, 0)
}

func (e *Encoder) noiseShapeAnalysis(psEncCtrl *encoderControl, pitchRes, x []int16) {
	psShapeSt := e.sShape

	var (
		k, i, nSamples, Qnrg, bQ14, warpingQ16, scale                             int32
		SNRadjdBQ7, HarmBoostQ16, HarmShapeGainQ16, TiltQ16, tmp32                int32
		nrg, preNrgQ30, logEnergyQ7, logEnergyPrevQ7, energyVariationQ7           int32
		deltaQ16, BWExp1Q16, BWExp2Q16, gainMultQ16, gainAddQ16, strengthQ16, bQ8 int32
		autoCorr                                                                  [MaxShapeLPCOrder + 1]int32
		reflCoefQ16                                                               [MaxShapeLPCOrder]int32
		AR1Q24                                                                    [MaxShapeLPCOrder]int32
		AR2Q24                                                                    [MaxShapeLPCOrder]int32
		xWindowed                                                                 [ShapeLPCWINMax]int16
		xPtr, pitchResPtr                                                         []int16
	)

	xPtr = x

	psEncCtrl.currentSNRdBQ7 = e._SNRdBQ7 - smulwb(lshift(e._BufferedInChannelMS, 7),
		fixConst(0.05, 16))
	if e.speechActivityQ8 > fixConst(LBRRSpeechActivityThres, 8) {
		psEncCtrl.currentSNRdBQ7 -= rshift(e.inBandFECSNRCompQ8, 1)
	}

	psEncCtrl.inputQualityQ14 = rshift(psEncCtrl.inputQualityBandsQ15[0]+psEncCtrl.inputQualityBandsQ15[1], 2)
	psEncCtrl.codingQualityQ14 = rshift(sigmQ15(rrshift(
		psEncCtrl.currentSNRdBQ7-fixConst(18.0, 7), 4)), 1)

	bQ8 = fixConst(1.0, 8) - e.speechActivityQ8
	bQ8 = smulwb(lshift(bQ8, 8), bQ8)
	SNRadjdBQ7 = smlawb(psEncCtrl.currentSNRdBQ7,
		smulbb(fixConst(-BGSNRDECRdB, 7)>>(4+1), bQ8),
		smulwb(fixConst(1.0, 14)+psEncCtrl.inputQualityQ14, psEncCtrl.codingQualityQ14))

	if psEncCtrl.sigType == SIGTypeVoiced {
		SNRadjdBQ7 = smlawb(SNRadjdBQ7, fixConst(HarmSNRINCRdB, 8), e._LTPCorrQ15)
	} else {
		SNRadjdBQ7 = smlawb(SNRadjdBQ7,
			smlawb(fixConst(6.0, 9), -fixConst(0.4, 18), psEncCtrl.currentSNRdBQ7),
			fixConst(1.0, 14)-psEncCtrl.inputQualityQ14)
	}

	if psEncCtrl.sigType == SIGTypeVoiced {
		psEncCtrl.QuantOffsetType = 0
		psEncCtrl.sparsenessQ8 = 0
	} else {
		nSamples = lshift(e.fskHz, 1)
		energyVariationQ7 = 0
		logEnergyPrevQ7 = 0
		pitchResPtr = pitchRes
		for k = 0; k < FrameLengthMS/2; k++ {
			sumSqrShift(&nrg, &scale, pitchResPtr, nSamples)
			nrg += rshift(nSamples, scale)

			logEnergyQ7 = lin2log(nrg)
			if k > 0 {
				energyVariationQ7 += abs(logEnergyQ7 - logEnergyPrevQ7)
			}
			logEnergyPrevQ7 = logEnergyQ7
			pitchResPtr = pitchResPtr[nSamples:]
		}

		psEncCtrl.sparsenessQ8 = rshift(sigmQ15(smulwb(energyVariationQ7-fixConst(5.0, 7),
			fixConst(0.1, 16))), 7)
		if psEncCtrl.sparsenessQ8 > fixConst(SparsenessThresholdQNTOffset, 8) {
			psEncCtrl.QuantOffsetType = 0
		} else {
			psEncCtrl.QuantOffsetType = 1
		}

		SNRadjdBQ7 = smlawb(SNRadjdBQ7, fixConst(SparseSNRINCRdB, 15),
			psEncCtrl.sparsenessQ8-fixConst(0.5, 8))
	}

	strengthQ16 = smulwb(psEncCtrl.predGainQ16, fixConst(FindPitchWhiteNoiseFraction, 16))
	BWExp1Q16 = div32varQ(fixConst(BandWidthExpansion, 16),
		smlaww(fixConst(1.0, 16), strengthQ16, strengthQ16), 16)
	BWExp2Q16 = BWExp1Q16
	deltaQ16 = smulwb(fixConst(1.0, 16)-smulbb(3, psEncCtrl.codingQualityQ14),
		fixConst(LowRateBandWidthExpansionDelta, 16))
	BWExp1Q16 = BWExp1Q16 - deltaQ16
	BWExp2Q16 = BWExp2Q16 - deltaQ16

	BWExp1Q16 = div(lshift(BWExp1Q16, 14), rshift(BWExp2Q16, 2))

	if e.warpingQ16 > 0 {
		warpingQ16 = smlawb(e.warpingQ16, psEncCtrl.codingQualityQ14, fixConst(0.01, 18))
	} else {
		warpingQ16 = 0
	}

	for k = 0; k < NBSubFR; k++ {
		var shift, slopePart, flatPart int32
		flatPart = e.fskHz * 5
		slopePart = rshift(e.shapeWinLength-flatPart, 1)

		applySineWindow(xWindowed[:], xPtr, 1, slopePart)
		shift = slopePart
		memcpy(xWindowed[shift:], xPtr[shift:], int(flatPart))

		shift += flatPart
		applySineWindow(xWindowed[shift:], xPtr[shift:], 2, slopePart)

		xPtr = xPtr[e.subfrLength:]

		if e.warpingQ16 > 0 {
			autoCorrelation(autoCorr[:], &scale, xWindowed[:], int16(warpingQ16), e.shapeWinLength, e.shapingLPCOrder)
		} else {
			autocorr(autoCorr[:], &scale, xWindowed[:], e.shapeWinLength, e.shapingLPCOrder+1)
		}

		autoCorr[0] = autoCorr[0] + max(smulwb(rshift(autoCorr[0], 4),
			fixConst(ShapeWhiteNoiseFraction, 20)), 1)
		nrg = schur64(reflCoefQ16[:], autoCorr[:], e.shapingLPCOrder)
		assert(nrg >= 0)

		k2aQ16(AR2Q24[:], reflCoefQ16[:], e.shapingLPCOrder)

		Qnrg = -scale
		assert(Qnrg >= -12)
		assert(Qnrg <= 30)

		if Qnrg&1 > 0 {
			Qnrg -= 1
			nrg >>= 1
		}

		tmp32 = sqrtApprox(nrg)
		Qnrg >>= 1

		psEncCtrl.GainsQ16[k] = lshiftSAT32(tmp32, 16-Qnrg)

		if e.warpingQ16 > 0 {
			gainMultQ16 = warpedGain(AR2Q24[:], warpingQ16, e.shapingLPCOrder)
			assert(psEncCtrl.GainsQ16[k] >= 0)

			psEncCtrl.GainsQ16[k] = smulww(psEncCtrl.GainsQ16[k], gainMultQ16)
			if psEncCtrl.GainsQ16[k] < 0 {
				psEncCtrl.GainsQ16[k] = math.MaxInt32
			}
		}

		bwexpander32(AR2Q24[:], e.shapingLPCOrder, BWExp2Q16)

		memcpy(AR1Q24[:], AR2Q24[:], int(e.shapingLPCOrder))
		assert(BWExp1Q16 <= fixConst(1.0, 16))

		bwexpander32(AR1Q24[:], e.shapingLPCOrder, BWExp1Q16)

		_LPCInversePredGainQ24(&preNrgQ30, AR2Q24[:], e.shapingLPCOrder)
		_LPCInversePredGainQ24(&nrg, AR1Q24[:], e.shapingLPCOrder)

		preNrgQ30 = lshift(smulwb(preNrgQ30, fixConst(0.7, 15)), 1)
		psEncCtrl.GainsPreQ14[k] = fixConst(0.3, 14) + div32varQ(preNrgQ30, nrg, 14)

		limitWarpedCoefs(AR2Q24[:], AR1Q24[:], warpingQ16, fixConst(3.999, 24), e.shapingLPCOrder)

		for i = 0; i < e.shapingLPCOrder; i++ {
			psEncCtrl.AR1Q13[k*MaxShapeLPCOrder+i] = sat16(rrshift(AR1Q24[i], 11))
			psEncCtrl.AR2Q13[k*MaxShapeLPCOrder+i] = sat16(rrshift(AR2Q24[i], 11))
		}
	}

	gainMultQ16 = log2lin(-smlawb(-fixConst(16.0, 7), SNRadjdBQ7, fixConst(0.16, 16)))
	gainAddQ16 = log2lin(smlawb(fixConst(16.0, 7), fixConst(NoiseFloordB, 7), fixConst(0.16, 16)))
	tmp32 = log2lin(smlawb(fixConst(16.0, 7), fixConst(RelativeMinGaindB, 7), fixConst(0.16, 16)))
	tmp32 = smulww(e.avgGainQ16, tmp32)
	gainAddQ16 = addSAT32(gainAddQ16, tmp32)
	assert(gainMultQ16 >= 0)

	for k = 0; k < NBSubFR; k++ {
		psEncCtrl.GainsQ16[k] = smulww(psEncCtrl.GainsQ16[k], gainMultQ16)
		if psEncCtrl.GainsQ16[k] < 0 {
			psEncCtrl.GainsQ16[k] = math.MaxInt32
		}
	}

	for k = 0; k < NBSubFR; k++ {
		psEncCtrl.GainsQ16[k] = addPosSAT32(psEncCtrl.GainsQ16[k], gainAddQ16)
		e.avgGainQ16 = addSAT32(
			e.avgGainQ16,
			smulwb(
				psEncCtrl.GainsQ16[k]-e.avgGainQ16,
				rrshift(smulbb(e.speechActivityQ8, fixConst(GainSmoothingCoef, 10)), 2),
			),
		)
	}

	gainMultQ16 = fixConst(1.0, 16) + rrshift(mla(fixConst(InputTilt, 26),
		psEncCtrl.codingQualityQ14, fixConst(HighRateInputTilt, 12)), 10)

	if psEncCtrl.inputTiltQ15 <= 0 && psEncCtrl.sigType == SIGTypeUnvoiced {
		if e.fskHz == 24 {
			essStrengthQ15 := smulww(-psEncCtrl.inputTiltQ15,
				smulbb(e.speechActivityQ8, fixConst(1.0, 8)-psEncCtrl.sparsenessQ8))
			tmp32 = log2lin(fixConst(16.0, 7) - smulwb(essStrengthQ15,
				smulwb(fixConst(DEESSERCoefSWBdB, 7), fixConst(0.16, 17))))
			gainMultQ16 = smulww(gainMultQ16, tmp32)
		} else if e.fskHz == 16 {
			essStrengthQ15 := smulww(-psEncCtrl.inputTiltQ15,
				smulbb(e.speechActivityQ8, fixConst(1.0, 8)-psEncCtrl.sparsenessQ8))
			tmp32 = log2lin(fixConst(16.0, 7) - smulwb(essStrengthQ15,
				smulwb(fixConst(DEESSERCoefWBdB, 7), fixConst(0.16, 17))))
			gainMultQ16 = smulww(gainMultQ16, tmp32)
		} else {
			assert(e.fskHz == 12 || e.fskHz == 8)
		}
	}

	for k = 0; k < NBSubFR; k++ {
		psEncCtrl.GainsPreQ14[k] = smulwb(gainMultQ16, psEncCtrl.GainsPreQ14[k])
	}

	strengthQ16 = mul(fixConst(LowFreqShaping, 0), fixConst(1.0, 16)+
		smulbb(fixConst(LowQualityLowFreqShapingDECR, 1),
			psEncCtrl.inputQualityBandsQ15[0]-fixConst(1.0, 15)))
	if psEncCtrl.sigType == SIGTypeVoiced {
		fskHzinv := div(fixConst(0.2, 14), e.fskHz)
		for k = 0; k < NBSubFR; k++ {
			bQ14 = fskHzinv + div(fixConst(3.0, 14), psEncCtrl.pitchL[k])
			psEncCtrl.LFShpQ14[k] = lshift(fixConst(1.0, 14)-bQ14-smulwb(strengthQ16, bQ14), 16)
			psEncCtrl.LFShpQ14[k] |= bQ14 - fixConst(1.0, 14)
		}
		TiltQ16 = -fixConst(HPNoiseCoef, 16) -
			smulwb(fixConst(1.0, 16)-fixConst(HPNoiseCoef, 16),
				smulwb(fixConst(HarmHPNoiseCoef, 24), e.speechActivityQ8))
	} else {
		bQ14 = div(21299, e.fskHz)

		psEncCtrl.LFShpQ14[0] = lshift(fixConst(1.0, 14)-bQ14-
			smulwb(strengthQ16, smulwb(fixConst(0.6, 16), bQ14)), 16)
		psEncCtrl.LFShpQ14[0] |= bQ14 - fixConst(1.0, 14)
		for k = 1; k < NBSubFR; k++ {
			psEncCtrl.LFShpQ14[k] = psEncCtrl.LFShpQ14[0]
		}
		TiltQ16 = -fixConst(HPNoiseCoef, 16)
	}

	HarmBoostQ16 = smulwb(smulwb(fixConst(1.0, 17)-lshift(psEncCtrl.codingQualityQ14, 3),
		e._LTPCorrQ15), fixConst(LowRateHarmonicBoost, 16))
	HarmBoostQ16 = smlawb(HarmBoostQ16,
		fixConst(1.0, 16)-lshift(psEncCtrl.inputQualityQ14, 2), fixConst(LowInputQualityHarmonicBoost, 16))

	if psEncCtrl.sigType == SIGTypeVoiced {
		HarmShapeGainQ16 = smlawb(fixConst(HarmonicShaping, 16),
			fixConst(1.0, 16)-smulwb(fixConst(1.0, 18)-lshift(psEncCtrl.codingQualityQ14, 4),
				psEncCtrl.inputQualityQ14), fixConst(HighRateOrLowRateHarmonicShaping, 16))
		HarmShapeGainQ16 = smulwb(lshift(HarmShapeGainQ16, 1),
			sqrtApprox(lshift(e._LTPCorrQ15, 15)))
	} else {
		HarmShapeGainQ16 = 0
	}

	for k = 0; k < NBSubFR; k++ {
		psShapeSt.HarmBoostSmthQ16 = smlawb(
			psShapeSt.HarmBoostSmthQ16, HarmBoostQ16-psShapeSt.HarmBoostSmthQ16,
			fixConst(SubFRSmthCoef, 16))
		psShapeSt.HarmShapeGainSmthQ16 = smlawb(
			psShapeSt.HarmShapeGainSmthQ16, HarmShapeGainQ16-psShapeSt.HarmShapeGainSmthQ16,
			fixConst(SubFRSmthCoef, 16))
		psShapeSt.TiltSmthQ16 = smlawb(
			psShapeSt.TiltSmthQ16,
			TiltQ16-psShapeSt.TiltSmthQ16,
			fixConst(SubFRSmthCoef, 16))

		psEncCtrl.HarmBoostQ14[k] = rrshift(psShapeSt.HarmBoostSmthQ16, 2)
		psEncCtrl.HarmShapeGainQ14[k] = rrshift(psShapeSt.HarmShapeGainSmthQ16, 2)
		psEncCtrl.TiltQ14[k] = rrshift(psShapeSt.TiltSmthQ16, 2)
	}
}

func (e *Encoder) prefilter(psEncCtrl *encoderControl, xw, x []int16) {
	P := e.sPrefilt

	var (
		j, k, lag, tmp32                int32
		AR1shpQ13, px, pxw              []int16
		HarmShapeGainQ12, TiltQ14       int32
		HarmShapeFIRPackedQ12, LFshpQ14 int32
		xFiltQ12                        [MaxFrameLength / NBSubFR]int32
		stRes                           [(MaxFrameLength / NBSubFR) + MaxShapeLPCOrder]int16
		BQ12                            int32
	)

	px = x
	pxw = xw
	lag = P.lagPrev
	for k = 0; k < NBSubFR; k++ {
		if psEncCtrl.sigType == SIGTypeVoiced {
			lag = psEncCtrl.pitchL[k]
		}

		HarmShapeGainQ12 = smulwb(psEncCtrl.HarmShapeGainQ14[k], 16384-psEncCtrl.HarmShapeGainQ14[k])
		assert(HarmShapeGainQ12 >= 0)
		HarmShapeFIRPackedQ12 = rshift(HarmShapeGainQ12, 2)
		HarmShapeFIRPackedQ12 |= lshift(rshift(HarmShapeGainQ12, 1), 16)
		TiltQ14 = psEncCtrl.TiltQ14[k]
		LFshpQ14 = psEncCtrl.LFShpQ14[k]
		AR1shpQ13 = psEncCtrl.AR1Q13[k*MaxShapeLPCOrder:]

		warpedLPCAnalysisFilter(P.sARShp[:], stRes[:], AR1shpQ13, px,
			int16(e.warpingQ16), e.subfrLength, e.shapingLPCOrder)

		BQ12 = rrshift(psEncCtrl.GainsPreQ14[k], 2)
		tmp32 = smlabb(fixConst(InputTilt, 26), psEncCtrl.HarmBoostQ14[k], HarmShapeGainQ12)
		tmp32 = smlabb(tmp32, psEncCtrl.codingQualityQ14, fixConst(HighRateInputTilt, 12))
		tmp32 = smulwb(tmp32, -psEncCtrl.GainsPreQ14[k])
		tmp32 = rrshift(tmp32, 12)
		BQ12 |= lshift(int32(sat16(tmp32)), 16)

		xFiltQ12[0] = smlabt(smulbb(int32(stRes[0]), BQ12), P.sHarmHP, BQ12)
		for j = 1; j < e.subfrLength; j++ {
			xFiltQ12[j] = smlabt(smulbb(int32(stRes[j]), BQ12), int32(stRes[j-1]), BQ12)
		}

		P.sHarmHP = int32(stRes[e.subfrLength-1])

		P.prefilt(xFiltQ12[:], pxw, HarmShapeFIRPackedQ12, TiltQ14,
			LFshpQ14, lag, e.subfrLength)

		px = px[e.subfrLength:]
		pxw = pxw[e.subfrLength:]
	}

	P.lagPrev = psEncCtrl.pitchL[NBSubFR-1]
}

func scaleCopyVector16(dataOut, dataIn []int16, gainQ16, dataSize int32) {
	var i, tmp32 int32

	for i = 0; i < dataSize; i++ {
		tmp32 = smulwb(gainQ16, int32(dataIn[i]))
		dataOut[i] = int16(tmp32)
	}
}

func (e *Encoder) _LTPScaleCtrl(psEncCtrl *encoderControl) {
	var (
		roundLoss, framesPerPacket              int32
		gOutQ5, gLimitQ15, thrld1Q15, thrld2Q15 int32
	)

	e._HPLTPredCodGainQ7 = max(psEncCtrl.LTPredCodGainQ7-e.prevLTPredCodGainQ7, 0) +
		rrshift(e._HPLTPredCodGainQ7, 1)

	e.prevLTPredCodGainQ7 = psEncCtrl.LTPredCodGainQ7

	gOutQ5 = rrshift(rshift(psEncCtrl.LTPredCodGainQ7, 1)+rshift(e._HPLTPredCodGainQ7, 1), 3)
	gLimitQ15 = sigmQ15(gOutQ5 - (3 << 5))

	psEncCtrl.LTPScaleIndex = 0

	roundLoss = e._PacketLossPerc

	if e.nFramesInPayloadBuf == 0 {
		framesPerPacket = div(e._PacketSizeMS, FrameLengthMS)

		roundLoss += framesPerPacket - 1
		thrld1Q15 = int32(LTPScaleThresholdsQ15[min(roundLoss, NBThresholds-1)])
		thrld2Q15 = int32(LTPScaleThresholdsQ15[min(roundLoss+1, NBThresholds-1)])

		if gLimitQ15 > thrld1Q15 {
			psEncCtrl.LTPScaleIndex = 2
		} else if gLimitQ15 > thrld2Q15 {
			psEncCtrl.LTPScaleIndex = 1
		}
	}
	psEncCtrl.LTPScaleQ14 = int32(LTPScalesTableQ14[psEncCtrl.LTPScaleIndex])
}

func (e *Encoder) findPredCoefs(psEncCtrl *encoderControl, resPitch []int16) {
	var i int32
	var (
		WLTP            [NBSubFR * LTPOrder * LTPOrder]int32
		invGainsQ16     [NBSubFR]int32
		localGains      [NBSubFR]int32
		WghtQ15         [NBSubFR]int32
		NLSFQ15         [MaxLPCOrder]int32
		xPtr            []int16
		xPrePtr         []int16
		LPCInPre        [NBSubFR*MaxLPCOrder + MaxFrameLength]int16
		tmp, minGainQ16 int32
		LTPCorrsRShift  [NBSubFR]int32
	)

	minGainQ16 = math.MaxInt32 >> 6
	for i = 0; i < NBSubFR; i++ {
		minGainQ16 = min(minGainQ16, psEncCtrl.GainsQ16[i])
	}

	for i = 0; i < NBSubFR; i++ {
		assert(psEncCtrl.GainsQ16[i] > 0)

		invGainsQ16[i] = div32varQ(minGainQ16, psEncCtrl.GainsQ16[i], 16-2)
		invGainsQ16[i] = max(invGainsQ16[i], 363)
		assert(invGainsQ16[i] == int32(sat16(invGainsQ16[i])))

		tmp = smulwb(invGainsQ16[i], invGainsQ16[i])
		WghtQ15[i] = rshift(tmp, 1)

		localGains[i] = div(1<<16, invGainsQ16[i])
	}

	if psEncCtrl.sigType == SIGTypeVoiced {
		assert(e.frameLength-e.predictLPCOrder >= psEncCtrl.pitchL[0]+LTPOrder/2)

		findLTP(psEncCtrl.LTPCoefQ14[:], WLTP[:], &psEncCtrl.LTPredCodGainQ7, resPitch,
			resPitch[rshift(e.frameLength, 1):], psEncCtrl.pitchL[:], WghtQ15[:],
			e.subfrLength, e.frameLength, LTPCorrsRShift[:])
		quantLTPGains(psEncCtrl.LTPCoefQ14[:], psEncCtrl.LTPIndex[:], &psEncCtrl.PERIndex,
			WLTP[:], e.muLTPQ8, e._LTPQuantLowComplexity)
		e._LTPScaleCtrl(psEncCtrl)
		_LTPAnalysisFilter(LPCInPre[:], e.xBuf[:], e.frameLength-e.predictLPCOrder,
			psEncCtrl.LTPCoefQ14[:], psEncCtrl.pitchL[:], invGainsQ16[:],
			e.subfrLength, e.predictLPCOrder)
	} else {
		xPtr = e.xBuf[e.frameLength-e.predictLPCOrder:]
		xPrePtr = LPCInPre[:]
		for i = 0; i < NBSubFR; i++ {
			scaleCopyVector16(xPrePtr, xPtr, invGainsQ16[i],
				e.subfrLength+e.predictLPCOrder)
			xPrePtr = xPrePtr[e.subfrLength+e.predictLPCOrder:]
			xPtr = xPtr[e.subfrLength:]
		}

		memset(psEncCtrl.LTPCoefQ14[:], 0, NBSubFR)
		psEncCtrl.LTPredCodGainQ7 = 0
	}

	findLPC(NLSFQ15[:], &psEncCtrl.NLSFInterpCoefQ2, e.sPred.prevNLSFqQ15[:],
		e.useInterpolatedNLSFs*(1-e.firstFrameAfterReset), e.predictLPCOrder,
		LPCInPre[:], e.subfrLength+e.predictLPCOrder)

	processNLSFs(psEncCtrl, NLSFQ15)

	residualEnergy(psEncCtrl.ResNrg, psEncCtrl.ResNrgQ, LPCInPre, psEncCtrl.PredCoefQ12, localGains,
		e.subfrLength, e.predictLPCOrder)

	memcpy(e.sPred.prevNLSFqQ15[:], NLSFQ15[:], int(e.predictLPCOrder))
}

func (e *Encoder) encodeFrame(pCode, pIn []int16) {
	var (
		LBRRIdx, frameTerminator, SNRdBQ7 int32
		pInHP                             [MaxFrameLength]int16
		resPitch                          [2*MaxFrameLength + LAPitchMax]int16
		xfw                               [MaxFrameLength]int16
	)

	sEncCtrl := &encoderControl{}

	sEncCtrl.Seed = e.frameCounter & 3
	e.frameCounter++

	xFrame := e.xBuf[e.frameLength:]
	resPitchFrame := resPitch[e.frameLength:]

	e.sVAD.GetSAQ8(&e.speechActivityQ8, &SNRdBQ7,
		sEncCtrl.inputQualityBandsQ15[:], &sEncCtrl.inputTiltQ15,
		pIn, e.frameLength)

	e._HPVariableCutoff(sEncCtrl, pInHP[:], pIn)

	e.sLP.VariableCutoff(xFrame[LAShapeMS*e.fskHz:], pInHP[:], e.frameLength)

	e.findPitchLags(sEncCtrl, resPitch[:], e.xBuf[:])
	e.noiseShapeAnalysis(sEncCtrl, resPitchFrame, e.xBuf[e.frameLength-e.laShape:])
	e.prefilter(sEncCtrl, xfw[:], xFrame[:])
	e.findPredCoefs(sEncCtrl, resPitch[:])
}

func (e *Encoder) Encode(samplesIn []int16) (out []byte, err error) {
	nSamplesIn := int32(len(samplesIn))

	maxInternalfskHz := (e.opts.MaxSampleRate >> 10) + 1

	e.maxInternalfskHz = maxInternalfskHz

	input10ms := div(100*nSamplesIn, e.opts.SampleRate)
	if input10ms*e.opts.SampleRate != 100*int32(len(samplesIn)) || nSamplesIn < 0 {
		err = ErrEncodeInputInvalidNoOfSamples
		return
	}

	packetSizeMS := (1000 * e.opts.PacketSize) / e.opts.SampleRate

	targetRateBps := limit(e.opts.BitRate, MinTargetRateBPS, MaxTargetRateBPS)
	if err = e.controlEncoder(packetSizeMS, targetRateBps,
		e.opts.PacketLossPercentage, e.opts.UseDTX, e.opts.Complexity); err != nil {
		return
	}

	if 1000*nSamplesIn > e._PacketSizeMS*e.opts.SampleRate {
		err = ErrEncodeInputInvalidNoOfSamples
		return
	}

	if min(e.opts.SampleRate, 1000*maxInternalfskHz) == 24000 &&
		e.sSWBdetect.SWBDetected == 0 &&
		e.sSWBdetect.WBDetected == 0 {
		e.sSWBdetect.SWBInput(samplesIn[:nSamplesIn])
	}

	out = make([]byte, MaxBytesPerFrame*MaxInputFrames, MaxBytesPerFrame*MaxInputFrames)

	var nSamplesToBuffer, nSamplesFromInput int32

	maxBytesOut := int32(0)
	for {
		nSamplesIn = int32(len(samplesIn))
		nSamplesToBuffer = e.frameLength - e.inputBufIx
		if e.opts.SampleRate == smulbb(1000, e.fskHz) {
			nSamplesToBuffer = min(nSamplesToBuffer, nSamplesIn)
			nSamplesFromInput = nSamplesToBuffer
			memcpy(e.inputBuf[e.inputBufIx:], samplesIn, int(nSamplesFromInput))
		} else {
			nSamplesToBuffer = min(nSamplesToBuffer, 10*input10ms*e.fskHz)
			nSamplesFromInput = (nSamplesToBuffer * e.opts.SampleRate) / (e.fskHz * 1000)

			if err = e.resamplerState.resample(e.inputBuf[e.inputBufIx:],
				samplesIn[:nSamplesFromInput]); err != nil {
				return
			}
		}

		samplesIn = samplesIn[nSamplesFromInput:]
		e.inputBufIx += nSamplesToBuffer

		if e.inputBufIx >= e.frameLength {
			assert(e.inputBufIx == e.frameLength)

			if maxBytesOut == 0 {
				//SKP_Silk_encode_frame_FIX

			} else {
				//SKP_Silk_encode_frame_FIX
			}
			e.inputBufIx = 0
			e.controlledSinceLastPayload = 0

			if nSamplesIn == 0 {
				break
			}
		} else {
			break
		}
	}

	if e.opts.UseDTX && e.inDTX > 0 {
		out = nil
		return
	}

	out = out[:maxBytesOut]
	return
}

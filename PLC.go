package silk

type PLC struct {
	pitchLQ8        int32
	LTPCoefQ14      []int16
	prevLPCQ12      []int16
	lastFrameLost   int
	randSeed        int32
	randScaleQ14    int16
	concEnergy      int32
	concEnergyShift int32
	prevLTPScaleQ14 int16
	prevGainQ16     []int32
	fskHz           int32
}

func (p *PLC) init() {
	p.prevLPCQ12 = make([]int16, MaxLPCOrder, MaxLPCOrder)
	p.LTPCoefQ14 = make([]int16, LTPOrder, LTPOrder)
	p.prevGainQ16 = make([]int32, NBSubFR, NBSubFR)
}

func (s *decoderState) plcConceal(psDecCtrl *decoderControl,
	signal []int16, length int32) {

	var (
		energy1, energy2 int32
		shift1, shift2   int32
		randPtr          []int32
	)

	psPLC := s.sPLC
	excBuf := make([]int16, MaxFrameLength, MaxFrameLength)

	memcpy(s.sLTPQ16, s.sLTPQ16[s.frameLength:], int(s.frameLength))

	bwexpander(psPLC.prevLPCQ12, s.LPCOrder, BWECOEFQ16)

	excBufPtr := excBuf
	for k := int32(NBSubFR >> 1); k < NBSubFR; k++ {
		for i := int32(0); i < s.subfrLength; i++ {
			excBufPtr[i] = int16(rshift(smulww(s.excQ10[i+k*s.subfrLength], psPLC.prevGainQ16[k]), 10))
		}
		excBufPtr = excBufPtr[s.subfrLength:]
	}

	sumSqrShift(&energy1, &shift1, excBuf, s.subfrLength)
	sumSqrShift(&energy2, &shift2, excBuf[s.subfrLength:], s.subfrLength)

	if rshift(energy1, shift2) < rshift(energy2, shift1) {
		randPtr = s.excQ10[i32max(0, 3*s.subfrLength-RandBufSize):]
	} else {
		randPtr = s.excQ10[i32max(0, s.frameLength-RandBufSize):]
	}

	BQ14 := psPLC.LTPCoefQ14
	randScaleQ14 := psPLC.randScaleQ14

	var randGainQ15 int32

	harmGainQ15 := HarmATTQ15[i32min(NBATT-1, s.lossCnt)]
	if s.prevSigType == SIGTypeVoiced {
		randGainQ15 = int32(PLCRandAttenuateVQ15[i32min(NBATT-1, s.lossCnt)])
	} else {
		randGainQ15 = int32(PLCRandAttenuateUVQ15[i32min(NBATT-1, s.lossCnt)])
	}

	if s.lossCnt == 0 {
		randScaleQ14 = 1 << 14

		if s.prevSigType == SIGTypeVoiced {
			for i := 0; i < LTPOrder; i++ {
				randScaleQ14 -= BQ14[i]
			}
			randScaleQ14 = int16(i32max(3277, int32(randScaleQ14)))
			randScaleQ14 = int16(rshift(smulbb(int32(randScaleQ14), int32(psPLC.prevLTPScaleQ14)), 14))
		}

		if s.prevSigType == SIGTypeUnvoiced {
			var invGainQ30, downScaleQ30 int32

			_LPCInversePredGain(&invGainQ30, psPLC.prevLPCQ12, s.LPCOrder)

			downScaleQ30 = i32min(rshift(1<<30, LOG2INVLPCGainHighThres), invGainQ30)
			downScaleQ30 = i32min(rshift(1<<30, LOG2INVLPCGainLowThres), downScaleQ30)
			downScaleQ30 = lshift(downScaleQ30, LOG2INVLPCGainHighThres)

			randGainQ15 = rshift(smulwb(downScaleQ30, randGainQ15), 14)
		}
	}

	randSeed := psPLC.randSeed
	lag := rrshift(psPLC.pitchLQ8, 8)
	sLTPBufIdx := s.frameLength

	sigQ10 := make([]int32, MaxFrameLength, MaxFrameLength)

	var (
		sigQ10Ptr, predLagPtr      []int32
		idx, LTPPredQ14, LPCExcQ10 int32
	)

	sigQ10Ptr = sigQ10
	for k := int32(0); k < NBSubFR; k++ {
		predLagPtr = s.sLTPQ16[(sLTPBufIdx-lag+LTPOrder/2)-4:]
		for i := int32(0); i < s.subfrLength; i++ {
			randSeed = rand(randSeed)
			idx = rshift(randSeed, 25) & RandBufMask

			LTPPredQ14 = smulwb(predLagPtr[4], int32(BQ14[0]))
			LTPPredQ14 = smlawb(LTPPredQ14, predLagPtr[3], int32(BQ14[1]))
			LTPPredQ14 = smlawb(LTPPredQ14, predLagPtr[2], int32(BQ14[2]))
			LTPPredQ14 = smlawb(LTPPredQ14, predLagPtr[1], int32(BQ14[3]))
			LTPPredQ14 = smlawb(LTPPredQ14, predLagPtr[0], int32(BQ14[4]))
			predLagPtr = predLagPtr[1:]

			LPCExcQ10 = lshift(smulwb(randPtr[idx], int32(randScaleQ14)), 2)
			LPCExcQ10 = LPCExcQ10 + rrshift(LTPPredQ14, 4)

			s.sLTPQ16[sLTPBufIdx] = lshift(LPCExcQ10, 6)
			sLTPBufIdx++

			sigQ10Ptr[i] = LPCExcQ10
		}
		sigQ10Ptr = sigQ10Ptr[s.subfrLength:]

		for j := 0; j < LTPOrder; j++ {
			BQ14[j] = int16(rshift(smulbb(int32(harmGainQ15), int32(BQ14[j])), 15))
		}

		randScaleQ14 = int16(rshift(smulbb(int32(randScaleQ14), randGainQ15), 15))

		psPLC.pitchLQ8 += smulwb(psPLC.pitchLQ8, PitchDriftFACQ16)
		psPLC.pitchLQ8 = i32min(psPLC.pitchLQ8, lshift(smulbb(MaxPitchLagMS, s.fskHz), 8))
		lag = rrshift(psPLC.pitchLQ8, 8)
	}

	sigQ10Ptr = sigQ10

	AQ12Tmp := make([]int16, MaxLPCOrder, MaxLPCOrder)

	for i := int32(0); i < s.LPCOrder; i++ {
		AQ12Tmp[i] = psPLC.prevLPCQ12[i]
	}
	assert(s.LPCOrder >= 10)

	for k := int32(0); k < NBSubFR; k++ {
		for i := int32(0); i < s.subfrLength; i++ {
			atmp := ua2i32(AQ12Tmp[0:])
			LPCPredQ10 := smulwb(s.sLPCQ14[MaxLPCOrder+i-1], atmp)
			LPCPredQ10 = smlawt(LPCPredQ10, s.sLPCQ14[MaxLPCOrder+i-2], atmp)

			atmp = ua2i32(AQ12Tmp[2:])
			LPCPredQ10 = smlawb(LPCPredQ10, s.sLPCQ14[MaxLPCOrder+i-3], atmp)
			LPCPredQ10 = smlawt(LPCPredQ10, s.sLPCQ14[MaxLPCOrder+i-4], atmp)

			atmp = ua2i32(AQ12Tmp[4:])
			LPCPredQ10 = smlawb(LPCPredQ10, s.sLPCQ14[MaxLPCOrder+i-5], atmp)
			LPCPredQ10 = smlawt(LPCPredQ10, s.sLPCQ14[MaxLPCOrder+i-6], atmp)

			atmp = ua2i32(AQ12Tmp[6:])
			LPCPredQ10 = smlawb(LPCPredQ10, s.sLPCQ14[MaxLPCOrder+i-7], atmp)
			LPCPredQ10 = smlawt(LPCPredQ10, s.sLPCQ14[MaxLPCOrder+i-8], atmp)

			atmp = ua2i32(AQ12Tmp[8:])
			LPCPredQ10 = smlawb(LPCPredQ10, s.sLPCQ14[MaxLPCOrder+i-9], atmp)
			LPCPredQ10 = smlawt(LPCPredQ10, s.sLPCQ14[MaxLPCOrder+i-10], atmp)

			for j := int32(10); j < s.LPCOrder; j += 2 {
				atmp = ua2i32(AQ12Tmp[j/2:])
				LPCPredQ10 = smlawb(LPCPredQ10, s.sLPCQ14[MaxLPCOrder+i-1-j], atmp)
				LPCPredQ10 = smlawt(LPCPredQ10, s.sLPCQ14[MaxLPCOrder+i-2-j], atmp)
			}

			sigQ10Ptr[i] += LPCPredQ10
			s.sLPCQ14[MaxLPCOrder+i] = lshift(sigQ10Ptr[i], 4)
		}
		sigQ10Ptr = sigQ10Ptr[s.subfrLength:]

		memcpy(s.sLPCQ14, s.sLPCQ14[s.subfrLength:], MaxLPCOrder)
	}

	for i := int32(0); i < s.frameLength; i++ {
		signal[i] = sat16(rrshift(smulww(sigQ10[i], psPLC.prevGainQ16[NBSubFR-1]), 10))
	}

	psPLC.randSeed = randSeed
	psPLC.randScaleQ14 = randScaleQ14
	for i := 0; i < NBSubFR; i++ {
		psDecCtrl.pitchL[i] = lag
	}

	return
}

func (s *decoderState) plcUpdate(psDecCtrl *decoderControl,
	signal []int16, length int32) {

	psPLC := s.sPLC

	s.prevSigType = psDecCtrl.sigType

	LTPGainQ14 := int32(0)

	if psDecCtrl.sigType == SIGTypeVoiced {
		for j := int32(0); j*s.subfrLength < psDecCtrl.pitchL[NBSubFR-1]; j++ {
			tempLTPGainQ14 := int32(0)
			for i := int32(0); i < LTPOrder; i++ {
				tempLTPGainQ14 += int32(psDecCtrl.LTPCoefQ14[(NBSubFR-1-j)*LTPOrder+i])
			}
			if tempLTPGainQ14 > LTPGainQ14 {
				LTPGainQ14 = tempLTPGainQ14
				memcpy(psPLC.LTPCoefQ14, psDecCtrl.LTPCoefQ14[smulbb(NBSubFR-1-j, LTPOrder):], LTPOrder)
				psPLC.pitchLQ8 = lshift(psDecCtrl.pitchL[NBSubFR-1-j], 8)
			}
		}

		memset(psPLC.LTPCoefQ14, 0, LTPOrder)
		psPLC.LTPCoefQ14[LTPOrder/2] = int16(LTPGainQ14)

		if LTPGainQ14 < VPitchGainStartMinQ14 {
			tmp := lshift(VPitchGainStartMinQ14, 10)
			scaleQ10 := tmp / i32max(LTPGainQ14, 1)
			for i := 0; i < LTPOrder; i++ {
				psPLC.LTPCoefQ14[i] = int16(rshift(smulbb(int32(psPLC.LTPCoefQ14[i]), scaleQ10), 10))
			}
		} else if LTPGainQ14 > VPitchGainStartMaxQ14 {
			tmp := lshift(VPitchGainStartMaxQ14, 14)
			scaleQ14 := tmp / i32max(LTPGainQ14, 1)
			for i := 0; i < LTPOrder; i++ {
				psPLC.LTPCoefQ14[i] = int16(rshift(smulbb(int32(psPLC.LTPCoefQ14[i]), scaleQ14), 14))
			}
		}
	} else {
		psPLC.pitchLQ8 = lshift(smulbb(s.fskHz, 18), 8)
		memset(psPLC.LTPCoefQ14, 0, LTPOrder)
	}

	memcpy(psPLC.prevLPCQ12, psDecCtrl.PredCoefQ12[1], int(s.LPCOrder))
	psPLC.prevLTPScaleQ14 = int16(psDecCtrl.LTPScaleQ14)

	memcpy(psPLC.prevGainQ16, psDecCtrl.GainsQ16, NBSubFR)
}

func (s *decoderState) PLCGlueFrames(psDecCtrl *decoderControl, signal []int16, length int32) {
	psPLC := s.sPLC

	var energy, energyShift int32

	if s.lossCnt > 0 {
		sumSqrShift(&psPLC.concEnergy, &psPLC.concEnergyShift, signal, length)
		psPLC.lastFrameLost = 1
	} else {
		if psPLC.lastFrameLost > 0 {
			sumSqrShift(&energy, &energyShift, signal, length)

			if energyShift > psPLC.concEnergyShift {
				psPLC.concEnergy = rshift(psPLC.concEnergy, energyShift-psPLC.concEnergyShift)
			} else if energyShift < psPLC.concEnergyShift {
				energy = rshift(energy, psPLC.concEnergyShift-energyShift)
			}

			if energy > psPLC.concEnergy {
				LZ := clz32(psPLC.concEnergy)
				LZ = LZ - 1
				psPLC.concEnergy = lshift(psPLC.concEnergy, LZ)
				energy = rshift(energy, i32max(24-LZ, 0))

				fracQ24 := psPLC.concEnergy / i32max(energy, 1)

				gainQ12 := sqrtApprox(fracQ24)
				slopeQ12 := ((1 << 12) - gainQ12) / length

				for i := int32(0); i < length; i++ {
					signal[i] = int16(rshift(mul(gainQ12, int32(signal[i])), 12))
					gainQ12 += slopeQ12
					gainQ12 = i32min(gainQ12, 1<<12)
				}
			}
		}
		psPLC.lastFrameLost = 0
	}
}

func (s *decoderState) PLC(psDecCtrl *decoderControl, signal []int16, length int32, lost bool) {
	if s.fskHz != s.sPLC.fskHz {
		s.PLCReset()
		s.sPLC.fskHz = s.fskHz
	}

	if lost {
		s.plcConceal(psDecCtrl, signal, length)
		s.lossCnt++
	} else {
		s.plcUpdate(psDecCtrl, signal, length)
	}
}

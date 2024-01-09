package silk

type CNG struct {
	excBufQ10   []int32
	smthNLSFQ15 []int32
	synthState  []int32
	smthGainQ16 int32
	randSeed    int32
	fskHz       int32
}

func (c *CNG) init() {
	c.excBufQ10 = make([]int32, MaxFrameLength, MaxFrameLength)
	c.smthNLSFQ15 = make([]int32, MaxLPCOrder, MaxLPCOrder)
	c.synthState = make([]int32, MaxLPCOrder, MaxLPCOrder)
}

func _CNGExc(residual []int16, excBufQ10 []int32, GainQ16 int32, length int32, randSeed *int32) {
	excMask := int32(CNGBufMaskMax)
	for excMask > length {
		excMask = excMask >> 1
	}

	seed := *randSeed
	for i := int32(0); i < length; i++ {
		seed = rand(seed)
		idx := rshift(seed, 24) & excMask
		assert(idx >= 0)
		assert(idx <= CNGBufMaskMax)
		residual[i] = sat16(rrshift(smulww(excBufQ10[idx], GainQ16), 10))
	}
	*randSeed = seed
}

func (s *decoderState) CNG(psDecCtrl *decoderControl, signal []int16, length int32) (err error) {
	LPCBuf := make([]int16, MaxLPCOrder, MaxLPCOrder)
	CNGSig := make([]int16, MaxFrameLength, MaxFrameLength)

	var maxGainQ16, subfr int32

	psCNG := s.sCNG

	if s.fskHz != psCNG.fskHz {
		s.CNGReset()
		psCNG.fskHz = s.fskHz
	}

	if s.lossCnt == 0 && s.vadFlag == NoVoiceActivity {
		for i := int32(0); i < s.LPCOrder; i++ {
			psCNG.smthNLSFQ15[i] += smulwb(s.prevNLSFQ15[i]-psCNG.smthNLSFQ15[i], CNGNLSFSMTHQ16)
		}

		maxGainQ16 = 0
		subfr = 0
		for i := int32(0); i < NBSubFR; i++ {
			if psDecCtrl.GainsQ16[i] > maxGainQ16 {
				maxGainQ16 = psDecCtrl.GainsQ16[i]
				subfr = i
			}
		}

		// memmove
		for i := int32(0); i < (NBSubFR-1)*s.subfrLength; i++ {
			psCNG.excBufQ10[s.subfrLength+i] = psCNG.excBufQ10[i]
		}

		memcpy(psCNG.excBufQ10, psCNG.excBufQ10[(subfr*s.subfrLength):], int(s.subfrLength))

		for i := int32(0); i < NBSubFR; i++ {
			psCNG.smthGainQ16 += smulwb(psDecCtrl.GainsQ16[i]-psCNG.smthGainQ16, CNGGainSMTHQ16)
		}
	}

	if s.lossCnt > 0 {
		_CNGExc(CNGSig, psCNG.excBufQ10,
			psCNG.smthGainQ16, length, &psCNG.randSeed)

		_NLSF2AStable(LPCBuf, psCNG.smthNLSFQ15, s.LPCOrder)

		GainQ26 := int32(1 << 26)

		if s.LPCOrder == 16 {
			_LPCSynthesisOrder16(CNGSig, LPCBuf,
				GainQ26, psCNG.synthState, CNGSig, length)
		} else {
			_LPCSynthesisFilter(CNGSig, LPCBuf,
				GainQ26, psCNG.synthState, CNGSig, length, s.LPCOrder)
		}

		for i := int32(0); i < length; i++ {
			tmp32 := int32(signal[i] + CNGSig[i])
			signal[i] = sat16(tmp32)
		}
	} else {
		memset(psCNG.synthState, 0, int(s.LPCOrder))
	}

	return
}

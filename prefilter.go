package silk

type prefilter struct {
	sLTPShp       [LTPBufLength]int16
	sARShp        [MaxShapeLPCOrder + 1]int32
	sLTPShpBufIdx int32
	sLFARShpQ12   int32
	sLFMAShpQ12   int32
	sHarmHP       int32
	randSeed      int32
	lagPrev       int32
}

func (P *prefilter) prefilt(stResQ12 []int32, xw []int16, HarmShapeFIRPackedQ12 int32,
	TiltQ14, LFshpQ14, lag, length int32) {
	var (
		i, idx, LTPShpBufIdx      int32
		nLTPQ12, nTiltQ10, nLFQ10 int32
		sLFMAShpQ12, sLFARShpQ12  int32
		LTPShpBuf                 []int16
	)

	LTPShpBuf = P.sLTPShp[:]
	LTPShpBufIdx = P.sLTPShpBufIdx
	sLFARShpQ12 = P.sLFARShpQ12
	sLFMAShpQ12 = P.sLFMAShpQ12

	for i = 0; i < length; i++ {
		if lag > 0 {
			idx = lag + LTPShpBufIdx
			nLTPQ12 = smulbb(int32(LTPShpBuf[(idx-HarmShapeFIRTaps/2-1)&LTPMask]), HarmShapeFIRPackedQ12)
			nLTPQ12 = smlabt(nLTPQ12, int32(LTPShpBuf[(idx-HarmShapeFIRTaps/2)&LTPMask]), HarmShapeFIRPackedQ12)
			nLTPQ12 = smlabb(nLTPQ12, int32(LTPShpBuf[(idx-HarmShapeFIRTaps/2+1)&LTPMask]), HarmShapeFIRPackedQ12)
		} else {
			nLTPQ12 = 0
		}

		nTiltQ10 = smulwb(sLFARShpQ12, TiltQ14)
		nLFQ10 = smlawb(smulwt(sLFARShpQ12, LFshpQ14), sLFMAShpQ12, LFshpQ14)

		sLFARShpQ12 = stResQ12[i] - lshift(nTiltQ10, 2)
		sLFMAShpQ12 = sLFARShpQ12 - lshift(nLFQ10, 2)

		LTPShpBufIdx = (LTPShpBufIdx - 1) & LTPMask
		LTPShpBuf[LTPShpBufIdx] = sat16(rrshift(sLFMAShpQ12, 12))

		xw[i] = sat16(rrshift(sLFMAShpQ12-nLTPQ12, 12))
	}

	P.sLFARShpQ12 = sLFARShpQ12
	P.sLFMAShpQ12 = sLFMAShpQ12
	P.sLTPShpBufIdx = LTPShpBufIdx
}

func warpedLPCAnalysisFilter(state []int32, res, coefQ13, input []int16,
	lambdaQ16 int16, length, order int32) {
	var n, i, accQ11, tmp1, tmp2 int32

	assert(order&1 == 0)

	for n = 0; n < length; n++ {
		tmp2 = smlawb(state[0], state[1], int32(lambdaQ16))
		state[0] = lshift(int32(input[n]), 14)

		tmp1 = smlawb(state[1], state[2]-tmp2, int32(lambdaQ16))
		state[1] = tmp2
		accQ11 = smulwb(tmp2, int32(coefQ13[0]))

		for i = 2; i < order; i += 2 {
			tmp2 = smlawb(state[i], state[i+1]-tmp1, int32(lambdaQ16))
			state[i] = tmp1
			accQ11 = smlawb(accQ11, tmp1, int32(coefQ13[i-1]))
			tmp1 = smlawb(state[i+1], state[i+2]-tmp2, int32(lambdaQ16))
			state[i+1] = tmp2
			accQ11 = smlawb(accQ11, tmp2, int32(coefQ13[i]))
		}
		state[order] = tmp1
		accQ11 = smlawb(accQ11, tmp1, int32(coefQ13[order-1]))
		res[n] = sat16(int32(input[n]) - rrshift(accQ11, 11))
	}
}

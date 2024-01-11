package silk

import "math"

type _VAD struct {
	AnaState       [2]int32
	AnaState1      [2]int32
	AnaState2      [2]int32
	XnrgSubfr      [VADNBands]int32
	NrgRatioSmthQ8 [VADNBands]int32
	HPstate        int16
	NL             [VADNBands]int32
	invNL          [VADNBands]int32
	NoiseLevelBias [VADNBands]int32
	counter        int32
}

var (
	Afb120      = []int32{5394 << 1}
	Afb121      = []int32{20623 << 1}
	tiltWeights = []int32{30000, 6000, -12000, -12000}

	sigmLUTSlopeQ10 = []int32{
		237, 153, 73, 30, 12, 7,
	}
	sigmLUTPosQ15 = []int32{
		16384, 23955, 28861, 31213, 32178, 32548,
	}
	sigmLUTNegQ15 = []int32{
		16384, 8812, 3906, 1554, 589, 219,
	}
)

func anaFiltBank1(in []int16, S []int32, outL, outH []int16, scratch []int32, N int32) {
	var k, in32, X, Y, out1, out2 int32
	N2 := rshift(N, 1)

	for k = 0; k < N2; k++ {
		in32 = lshift(int32(in[2*k]), 10)

		Y = in32 - S[0]
		X = smlawb(Y, Y, Afb121[0])
		out1 = S[0] + X
		S[0] = in32 + X

		in32 = lshift(int32(in[2*k+1]), 10)

		Y = in32 - S[1]
		X = smulwb(Y, Afb120[0])
		out2 = S[1] + X
		S[1] = in32 + X

		outL[k] = sat16(rrshift(out2+out1, 11))
		outH[k] = sat16(rrshift(out2-out1, 11))
	}
}

func sigmQ15(inQ5 int32) int32 {
	var ind int32

	if inQ5 < 0 {
		inQ5 = -inQ5
		if inQ5 >= 6*32 {
			return 0
		} else {
			ind = rshift(inQ5, 5)
			return sigmLUTNegQ15[ind] - smulbb(sigmLUTSlopeQ10[ind], inQ5&0x1f)
		}
	} else {
		if inQ5 >= 6*32 {
			return 32767
		} else {
			ind = rshift(inQ5, 5)
			return sigmLUTPosQ15[ind] + smulbb(sigmLUTSlopeQ10[ind], inQ5&0x1f)
		}
	}
}

func (v *_VAD) GetNoiseLevels(pX []int32) {
	var nl, nrg, invNrg, k, coef, minCoef int32

	if v.counter < 1000 {
		minCoef = div(math.MaxInt16, rshift(v.counter, 4)+1)
	} else {
		minCoef = 0
	}

	for k = 0; k < VADNBands; k++ {
		nl = v.NL[k]
		assert(nl >= 0)

		nrg = addPosSAT32(pX[k], v.NoiseLevelBias[k])
		assert(nrg > 0)

		invNrg = div(math.MaxInt32, nrg)
		assert(invNrg >= 0)

		if nrg > lshift(nl, 3) {
			coef = VADNoiseLevelSmoothCOEFQ16 >> 3
		} else if nrg < nl {
			coef = VADNoiseLevelSmoothCOEFQ16
		} else {
			coef = smulwb(smulww(invNrg, nl), VADNoiseLevelSmoothCOEFQ16<<1)
		}

		coef = max(coef, minCoef)

		v.invNL[k] = smlawb(v.invNL[k], invNrg-v.invNL[k], coef)
		assert(v.invNL[k] >= 0)

		nl = div(math.MaxInt32, v.invNL[k])
		assert(nl >= 0)

		nl = min(nl, 0x00FFFFFF)

		v.NL[k] = nl
	}

	v.counter++
}

func (v *_VAD) GetSAQ8(pSAQ8, pSNRdBQ7 *int32, pQualityQ15 []int32, pTiltQ15 *int32,
	pIn []int16, frameLength int32) {
	var SAQ15, inputTilt int32
	var (
		decimatedFrameLength, decSubFrameLength, decSubframeOffset int32
		SNRQ7, i, b, s                                             int32
		sumSquared, smoothCoefQ16                                  int32
		HPStateTmp                                                 int16
	)
	var (
		scratch           [3 * MaxFrameLength / 2]int32
		X                 [VADNBands][MaxFrameLength / 2]int16
		Xnrg              [VADNBands]int32
		NrgToNoiseRatioQ8 [VADNBands]int32
		speechNrg, xTmp   int32
	)

	assert(MaxFrameLength >= frameLength)
	assert(frameLength <= 512)

	anaFiltBank1(pIn, v.AnaState[:], X[0][:], X[3][:], scratch[:], frameLength)
	anaFiltBank1(X[0][:], v.AnaState1[:], X[0][:], X[2][:], scratch[:], rshift(frameLength, 1))
	anaFiltBank1(X[0][:], v.AnaState2[:], X[0][:], X[1][:], scratch[:], rshift(frameLength, 2))

	decimatedFrameLength = rshift(frameLength, 3)
	X[0][decimatedFrameLength-1] = int16(rshift(int32(X[0][decimatedFrameLength-1]), 1))
	HPStateTmp = X[0][decimatedFrameLength-1]

	for i = decimatedFrameLength - 1; i > 0; i-- {
		X[0][i-1] = int16(rshift(int32(X[0][i-1]), 1))
		X[0][i] -= X[0][i-1]
	}
	X[0][0] -= v.HPstate
	v.HPstate = HPStateTmp

	for b = 0; b < VADNBands; b++ {
		decimatedFrameLength = rshift(frameLength, min(VADNBands-b, VADNBands-1))

		decSubFrameLength = rshift(decimatedFrameLength, VADInternalSubFramesLog2)
		decSubframeOffset = 0

		Xnrg[b] = v.XnrgSubfr[b]
		for s = 0; s < VADInternalSubFrames; s++ {
			sumSquared = 0
			for i = 0; i < decSubFrameLength; i++ {
				xTmp = rshift(int32(X[b][i+decSubframeOffset]), 3)
				sumSquared = smlabb(sumSquared, xTmp, xTmp)

				assert(sumSquared >= 0)
			}

			if s < VADInternalSubFrames-1 {
				Xnrg[b] = addPosSAT32(Xnrg[b], sumSquared)
			} else {
				Xnrg[b] = addPosSAT32(Xnrg[b], rshift(sumSquared, 1))
			}

			decSubframeOffset += decSubFrameLength
		}

		v.XnrgSubfr[b] = sumSquared
	}

	v.GetNoiseLevels(Xnrg[:])

	sumSquared = 0
	inputTilt = 0
	for b = 0; b < VADNBands; b++ {
		speechNrg = Xnrg[b] - v.NL[b]
		if speechNrg > 0 {
			if int(Xnrg[b])&0xFF800000 == 0 {
				NrgToNoiseRatioQ8[b] = div(lshift(Xnrg[b], 8), v.NL[b]+1)
			} else {
				NrgToNoiseRatioQ8[b] = div(Xnrg[b], rshift(v.NL[b], 8)+1)
			}

			SNRQ7 = lin2log(NrgToNoiseRatioQ8[b]) - 8*128
			sumSquared = smlabb(sumSquared, SNRQ7, SNRQ7)

			if speechNrg < (1 << 20) {
				SNRQ7 = smulwb(lshift(sqrtApprox(speechNrg), 6), SNRQ7)
			}
			inputTilt = smlawb(inputTilt, tiltWeights[b], SNRQ7)
		} else {
			NrgToNoiseRatioQ8[b] = 256
		}
	}

	sumSquared = div(sumSquared, VADNBands)
	*pSNRdBQ7 = 3 * sqrtApprox(sumSquared)
	SAQ15 = sigmQ15(smulwb(VADSNRFactorQ16, *pSNRdBQ7) - VADNegativeOffsetQ5)

	*pTiltQ15 = lshift(sigmQ15(inputTilt)-16384, 1)

	speechNrg = 0
	for b = 0; b < VADNBands; b++ {
		speechNrg += (b + 1) * rshift(Xnrg[b]-v.NL[b], 4)
	}

	if speechNrg <= 0 {
		SAQ15 = rshift(SAQ15, 1)
	} else if speechNrg < 32768 {
		speechNrg = sqrtApprox(lshift(speechNrg, 15))
		SAQ15 = smulwb(32768+speechNrg, SAQ15)
	}

	*pSAQ8 = min(rshift(SAQ15, 7), math.MaxUint8)

	smoothCoefQ16 = smulwb(VADSNRSmoothCOEFQ18, smulwb(SAQ15, SAQ15))
	for b = 0; b < VADNBands; b++ {
		v.NrgRatioSmthQ8[b] = smlawb(v.NrgRatioSmthQ8[b],
			NrgToNoiseRatioQ8[b]-v.NrgRatioSmthQ8[b], smoothCoefQ16)

		SNRQ7 = 3 * (lin2log(v.NrgRatioSmthQ8[b]) - 8*128)

		pQualityQ15[b] = sigmQ15(rshift(SNRQ7-16*128, 4))
	}
}

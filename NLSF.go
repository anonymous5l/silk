package silk

import (
	"math"
)

type NLSFCBS struct {
	nVectors  int32
	CBNLSFQ15 []int16
	RatesQ5   []int16
}

type NLSFCB struct {
	nStages      int32
	CBStages     []NLSFCBS
	NDeltaMinQ15 []int32
	CDF          []uint16
	StartPtr     [][]uint16
	MiddleIx     []int32
}

func (n *NLSFCB) Stabilize(NLSFQ15 []int32, L int32) {
	assert(n.NDeltaMinQ15[L] >= 1)

	var (
		diffQ15, minDiffQ15                       int32
		centerFreqQ15, minCenterQ15, maxCenterQ15 int32
		I                                         int32
	)

	for loops := 0; loops < MaxLoops; loops++ {
		minDiffQ15 = NLSFQ15[0] - n.NDeltaMinQ15[0]
		I = 0
		for i := int32(1); i <= L-1; i++ {
			diffQ15 = NLSFQ15[i] - (NLSFQ15[i-1] + n.NDeltaMinQ15[i])
			if diffQ15 < minDiffQ15 {
				minDiffQ15 = diffQ15
				I = i
			}
		}

		diffQ15 = (1 << 15) - (NLSFQ15[L-1] + n.NDeltaMinQ15[L])
		if diffQ15 < minDiffQ15 {
			minDiffQ15 = diffQ15
			I = L
		}

		if minDiffQ15 >= 0 {
			return
		}

		switch I {
		case 0:
			NLSFQ15[0] = n.NDeltaMinQ15[0]
		case L:
			NLSFQ15[L-1] = (1 << 15) - n.NDeltaMinQ15[L]
		default:
			minCenterQ15 = 0
			for k := int32(0); k < I; k++ {
				minCenterQ15 += n.NDeltaMinQ15[k]
			}
			minCenterQ15 += rshift(n.NDeltaMinQ15[I], 1)

			maxCenterQ15 = 1 << 15
			for k := L; k > I; k-- {
				maxCenterQ15 -= n.NDeltaMinQ15[k]
			}
			maxCenterQ15 -= n.NDeltaMinQ15[I] - rshift(n.NDeltaMinQ15[I], 1)

			centerFreqQ15 = limit(rrshift(NLSFQ15[I-1]+NLSFQ15[I], 1),
				minCenterQ15, maxCenterQ15)
			NLSFQ15[I-1] = centerFreqQ15 - rshift(n.NDeltaMinQ15[I], 1)
			NLSFQ15[I] = NLSFQ15[I-1] + n.NDeltaMinQ15[I]
		}
	}

	insertionSortIncreasingAllValues(NLSFQ15, L)

	NLSFQ15[0] = i32max(NLSFQ15[0], n.NDeltaMinQ15[0])

	for i := int32(1); i < L; i++ {
		NLSFQ15[i] = i32max(NLSFQ15[i], NLSFQ15[i-1]+n.NDeltaMinQ15[i])
	}

	NLSFQ15[L-1] = i32min(NLSFQ15[L-1], (1<<15)-n.NDeltaMinQ15[L])

	for i := L - 2; i >= 0; i-- {
		NLSFQ15[i] = i32min(NLSFQ15[i], NLSFQ15[i+1]-n.NDeltaMinQ15[i+1])
	}
}

func (n *NLSFCB) MSVQDecode(pNLSFQ15 []int32, NLSFIndices []int32, LPCOrder int32) {
	var (
		pCBElement []int16
		i          int32
	)

	assert(0 <= NLSFIndices[0] && NLSFIndices[0] < n.CBStages[0].nVectors)

	pCBElement = n.CBStages[0].CBNLSFQ15[mul(NLSFIndices[0], LPCOrder):]

	for i = 0; i < LPCOrder; i++ {
		pNLSFQ15[i] = int32(pCBElement[i])
	}

	for s := int32(1); s < n.nStages; s++ {
		assert(0 <= NLSFIndices[s] && NLSFIndices[s] < n.CBStages[s].nVectors)

		switch LPCOrder {
		case 16:
			pCBElement = n.CBStages[s].CBNLSFQ15[NLSFIndices[s]<<4:]

			for x := 0; x < 16; x++ {
				pNLSFQ15[x] += int32(pCBElement[x])
			}
		default:
			pCBElement = n.CBStages[s].CBNLSFQ15[smulbb(NLSFIndices[s], LPCOrder):]

			for x := int32(0); x < LPCOrder; x++ {
				pNLSFQ15[x] += int32(pCBElement[x])
			}
		}
	}

	n.Stabilize(pNLSFQ15, LPCOrder)
}

func bwexpander(ar []int16, d int32, chirpQ16 int32) {
	chirpMinusOneQ16 := chirpQ16 - 65536
	for i := int32(0); i < d-1; i++ {
		ar[i] = int16(rrshift(mul(chirpQ16, int32(ar[i])), 16))
		chirpQ16 += rrshift(mul(chirpQ16, chirpMinusOneQ16), 16)
	}
	ar[d-1] = int16(rrshift(mul(chirpQ16, int32(ar[d-1])), 16))
}

func bwexpander32(ar []int32, d int32, chirpQ16 int32) {
	tmpChirpQ16 := chirpQ16
	for i := int32(0); i < d-1; i++ {
		ar[i] = smulww(ar[i], tmpChirpQ16)
		tmpChirpQ16 = smulww(chirpQ16, tmpChirpQ16)
	}
	ar[d-1] = smulww(ar[d-1], tmpChirpQ16)
}

func _NLSF2AFindPoly(out []int32, cLSF []int32, dd int32) {
	out[0] = lshift(1, 20)
	out[1] = -cLSF[0]
	for k := int32(1); k < dd; k++ {
		ftmp := cLSF[2*k]
		out[k+1] = lshift(out[k-1], 1) - int32(i64rrshift(smull(ftmp, out[k]), 20))
		for n := k; n > 1; n-- {
			out[n] += out[n-2] - int32(i64rrshift(smull(ftmp, out[n-1]), 20))
		}
		out[1] -= ftmp
	}
}

func _NLSF2A(a []int16, NLSF []int32, d int32) {
	var (
		fInt, fFrac, cosVal, delta int32
		P                          = make([]int32, MaxLPCOrder/2+1)
		Q                          = make([]int32, MaxLPCOrder/2+1)
		aInt32                     = make([]int32, MaxLPCOrder, MaxLPCOrder)
		cosLSFQ20                  = make([]int32, MaxLPCOrder, MaxLPCOrder)
	)

	for k := int32(0); k < d; k++ {
		assert(NLSF[k] >= 0)
		assert(NLSF[k] <= 32767)

		fInt = rshift(NLSF[k], 15-7)

		fFrac = NLSF[k] - lshift(fInt, 15-7)

		assert(fInt >= 0)
		assert(fInt < LSFCosTabSZFix)

		cosVal = LSFCosTabFixQ12[fInt]
		delta = LSFCosTabFixQ12[fInt+1] - cosVal

		cosLSFQ20[k] = lshift(cosVal, 8) + mul(delta, fFrac)
	}

	dd := rshift(d, 1)

	_NLSF2AFindPoly(P, cosLSFQ20[0:], dd)
	_NLSF2AFindPoly(Q, cosLSFQ20[1:], dd)

	for k := int32(0); k < dd; k++ {
		PTmp := P[k+1] + P[k]
		QTmp := Q[k+1] - Q[k]

		aInt32[k] = -rrshift(PTmp+QTmp, 9)
		aInt32[d-k-1] = rrshift(QTmp-PTmp, 9)
	}

	var (
		maxabs, absval int32
		idx            int32
		i              int32
	)

	for i = 0; i < 10; i++ {
		maxabs = 0
		for k := int32(0); k < d; k++ {
			absval = i32abs(aInt32[k])
			if absval > maxabs {
				maxabs = absval
				idx = k
			}
		}

		if maxabs > math.MaxInt16 {
			maxabs = i32min(maxabs, 98369)

			scQ16 := 65470 - div(mul(65470>>2, maxabs-math.MaxInt16),
				rshift(mul(maxabs, idx+1), 2))
			bwexpander32(aInt32, d, scQ16)
		} else {
			break
		}
	}

	if i == 10 {
		return
	}

	for i = 0; i < d; i++ {
		a[i] = int16(aInt32[i])
	}
}

func _NLSF2AStable(pARQ12 []int16, pNLSF []int32, LPCOrder int32) {
	_NLSF2A(pARQ12, pNLSF, LPCOrder)

	var i, invGainQ30 int32

	for i = 0; i < MaxLPCStabilizeIterations; i++ {
		if _LPCInversePredGain(&invGainQ30, pARQ12, LPCOrder) == 1 {
			bwexpander(pARQ12, LPCOrder, 65536-smulbb(10+i, i))
		} else {
			break
		}
	}
}

package silk

import (
	"math"
)

const (
	QA     = int32(16)
	ALimit = 65520
)

func _LPCInversePredGainQA(invGainQ30 *int32, AQA [][]int32, order int32) int32 {
	var (
		rcQ31, rcMult1Q30, rcMult2Q16 int32
		aoldQA                        []int32
	)

	anewQA := AQA[order&1]
	*invGainQ30 = 1 << 30
	for k := order - 1; k > 0; k-- {
		if anewQA[k] > ALimit || anewQA[k] < -ALimit {
			return 1
		}

		rcQ31 = -lshift(anewQA[k], 31-QA)

		rcMult1Q30 = (math.MaxInt32 >> 1) - smmul(rcQ31, rcQ31)
		assert(rcMult1Q30 > (1 << 15))
		assert(rcMult1Q30 < (1 << 30))

		rcMult2Q16 = inverse32varQ(rcMult1Q30, 46)

		*invGainQ30 = lshift(smmul(*invGainQ30, rcMult1Q30), 2)
		assert(*invGainQ30 >= 0)
		assert(*invGainQ30 <= (1 << 30))

		aoldQA = anewQA
		anewQA = AQA[k&1]

		headrm := clz32(rcMult2Q16) - 1
		rcMult2Q16 = lshift(rcMult2Q16, headrm)

		for n := int32(0); n < k; n++ {
			tmpQA := aoldQA[n] - lshift(smmul(aoldQA[k-n-1], rcQ31), 1)
			anewQA[n] = lshift(smmul(tmpQA, rcMult2Q16), 16-headrm)
		}
	}

	if anewQA[0] > ALimit || anewQA[0] < -ALimit {
		return 1
	}

	rcQ31 = -lshift(anewQA[0], 31-QA)
	rcMult1Q30 = (math.MaxInt32 >> 1) - smmul(rcQ31, rcQ31)

	*invGainQ30 = lshift(smmul(*invGainQ30, rcMult1Q30), 2)
	assert(*invGainQ30 >= 0)
	assert(*invGainQ30 <= 1<<30)

	return 0
}

func _LPCInversePredGain(invGainQ30 *int32, AQ12 []int16, order int32) int32 {
	atmpQA := make([][]int32, 2)
	atmpQA[0] = make([]int32, MaxLPCOrder, MaxLPCOrder)
	atmpQA[1] = make([]int32, MaxLPCOrder, MaxLPCOrder)

	anewQA := atmpQA[order&1]
	for k := int32(0); k < order; k++ {
		anewQA[k] = lshift(int32(AQ12[k]), QA-12)
	}

	return _LPCInversePredGainQA(invGainQ30, atmpQA, order)
}

func _LPCSynthesisOrder16(in, AQ12 []int16, GainQ26 int32, S []int32, out []int16, length int32) {
	AAlignQ12 := make([]int32, 8, 8)

	var SA, SB, out32Q10, out32 int32

	for k := int32(0); k < 8; k++ {
		AAlignQ12[k] = (int32(AQ12[2*k]) & 0x0000ffff) | lshift(int32(AQ12[2*k+1]), 16)
	}

	for k := int32(0); k < length; k++ {
		SA = S[15]
		atmp := AAlignQ12[0]

		SB, S[14] = S[14], SA
		out32Q10 = smulwb(SA, atmp)
		out32Q10 = smlawt(out32Q10, SB, atmp)
		SA, S[13] = S[13], SB

		atmp = AAlignQ12[1]
		SB, S[12] = S[12], SA
		out32Q10 = smlawb(out32Q10, SA, atmp)
		out32Q10 = smlawt(out32Q10, SB, atmp)
		SA, S[11] = S[11], SB

		atmp = AAlignQ12[2]
		SB, S[10] = S[10], SA
		out32Q10 = smlawb(out32Q10, SA, atmp)
		out32Q10 = smlawt(out32Q10, SB, atmp)
		SA, S[9] = S[9], SB

		atmp = AAlignQ12[3]
		SB, S[8] = S[8], SA
		out32Q10 = smlawb(out32Q10, SA, atmp)
		out32Q10 = smlawt(out32Q10, SB, atmp)
		SA, S[7] = S[7], SB

		atmp = AAlignQ12[4]
		SB, S[6] = S[6], SA
		out32Q10 = smlawb(out32Q10, SA, atmp)
		out32Q10 = smlawt(out32Q10, SB, atmp)
		SA, S[5] = S[5], SB

		atmp = AAlignQ12[5]
		SB, S[4] = S[4], SA
		out32Q10 = smlawb(out32Q10, SA, atmp)
		out32Q10 = smlawt(out32Q10, SB, atmp)
		SA, S[3] = S[3], SB

		atmp = AAlignQ12[6]
		SB, S[2] = S[2], SA
		out32Q10 = smlawb(out32Q10, SA, atmp)
		out32Q10 = smlawt(out32Q10, SB, atmp)
		SA, S[1] = S[1], SB

		atmp = AAlignQ12[7]
		SB, S[0] = S[0], SA
		out32Q10 = smlawb(out32Q10, SA, atmp)
		out32Q10 = smlawt(out32Q10, SB, atmp)

		out32Q10 = addSAT32(out32Q10, smulwb(GainQ26, int32(in[k])))
		out32 = rrshift(out32Q10, 10)

		out[k] = sat16(out32)

		S[15] = lshiftSAT32(out32Q10, 4)
	}
}

func _LPCSynthesisFilter(in, AQ12 []int16, GainQ26 int32, S []int32, out []int16, length, order int32) {
	var SA, SB, out32Q10, out32 int32
	var k, j, idx int32

	OrderHalf := order >> 1

	var atmp int32

	AAlignQ12 := make([]int32, MaxLPCOrder>>1, MaxLPCOrder>>1)

	for k = 0; k < OrderHalf; k++ {
		idx = smulbb(2, k)
		AAlignQ12[k] = (int32(AQ12[idx]) & 0x0000ffff) | lshift(int32(AQ12[idx+1]), 16)
	}

	for k = 0; k < length; k++ {
		SA = S[order-1]
		out32Q10 = 0
		for j = 0; j < OrderHalf-1; j++ {
			idx = smulbb(2, j) + 1
			atmp = AAlignQ12[j]
			SB, S[order-1-idx] = S[order-1-idx], SA
			out32Q10 = smlawb(out32Q10, SA, atmp)
			out32Q10 = smlawt(out32Q10, SB, atmp)
			SA, S[order-2-idx] = S[order-2-idx], SB
		}

		atmp = AAlignQ12[OrderHalf-1]
		SB, S[0] = S[0], SA

		out32Q10 = smlawb(out32Q10, SA, atmp)
		out32Q10 = smlawt(out32Q10, SB, atmp)

		out32Q10 = addSAT32(out32Q10, smulwb(GainQ26, int32(in[k])))

		out32 = rrshift(out32Q10, 10)

		out[k] = sat16(out32)

		S[order-1] = lshiftSAT32(out32Q10, 4)
	}
}

func _LPCInversePredGainQ24(invGainQ30 *int32, AQ24 []int32, order int32) int32 {
	var k int32
	var (
		atmpQA [2][]int32
		anewQA []int32
	)

	for i := 0; i < 2; i++ {
		atmpQA[i] = make([]int32, MaxLPCOrder, MaxLPCOrder)
	}

	anewQA = atmpQA[order&1]

	for k = 0; k < order; k++ {
		anewQA[k] = rrshift(AQ24[k], 24-QA)
	}

	return _LPCInversePredGainQA(invGainQ30, atmpQA[:], order)
}

const (
	FindLPCChirp   = 0.99995
	FindLPCCondFAC = 2.5e-5
)

func findLPC(
	NLSFQ15 []int32,
	interpIndex *int32,
	prevNLSFqQ15 []int32,
	useInterpolatedNLSFs int32,
	LPCorder int32,
	x []int16,
	subfrLength int32,
) {

}

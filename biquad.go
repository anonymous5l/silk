package silk

func biquad(in *slice[int16], B, A []int16, S *slice[int32], out *slice[int16], len int32) {
	var (
		k, in16                              int32
		A0_neg, A1_neg, S0, S1, out32, tmp32 int32
	)

	S0 = S.idx(0)
	S1 = S.idx(1)
	A0_neg = int32(-A[0])
	A1_neg = int32(-A[1])
	for k = 0; k < len; k++ {

		in16 = int32(in.idx(int(k)))
		out32 = SMLABB(S0, in16, int32(B[0]))

		S0 = SMLABB(S1, in16, int32(B[1]))
		S0 += LSHIFT(SMULWB(out32, A0_neg), 3)

		S1 = LSHIFT(SMULWB(out32, A1_neg), 3)
		S1 = SMLABB(S1, in16, int32(B[2]))
		tmp32 = RSHIFT_ROUND(out32, 13) + 1
		*out.ptr(int(k)) = int16(SAT16(tmp32))
	}
	*S.ptr(0) = S0
	*S.ptr(1) = S1
}

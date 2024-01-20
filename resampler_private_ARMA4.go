package silk

func resampler_private_ARMA4(S *slice[int32], out, in, Coef *slice[int16], inLen int32) {
	var k, in_Q8, out1_Q8, out2_Q8, X int32

	for k = 0; k < inLen; k++ {
		in_Q8 = LSHIFT32(int32(in.idx(int(k))), 8)

		out1_Q8 = ADD_LSHIFT32(in_Q8, S.idx(0), 2)
		out2_Q8 = ADD_LSHIFT32(out1_Q8, S.idx(2), 2)

		X = SMLAWB(S.idx(1), in_Q8, int32(Coef.idx(0)))
		*S.ptr(0) = SMLAWB(X, out1_Q8, int32(Coef.idx(2)))

		X = SMLAWB(S.idx(3), out1_Q8, int32(Coef.idx(1)))
		*S.ptr(2) = SMLAWB(X, out2_Q8, int32(Coef.idx(4)))

		*S.ptr(1) = SMLAWB(RSHIFT32(in_Q8, 2), out1_Q8, int32(Coef.idx(3)))
		*S.ptr(3) = SMLAWB(RSHIFT32(out1_Q8, 2), out2_Q8, int32(Coef.idx(5)))

		*out.ptr(int(k)) = int16(SAT16(RSHIFT32(SMLAWB(128, out2_Q8, int32(Coef.idx(6))), 8)))
	}
}

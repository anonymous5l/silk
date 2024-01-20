package silk

func resampler_private_AR2(S, out_Q8 *slice[int32], in, A_Q14 *slice[int16], inLen int32) {
	var k, out32 int32

	for k = 0; k < inLen; k++ {
		out32 = ADD_LSHIFT32(S.idx(0), int32(in.idx(int(k))), 8)
		*out_Q8.ptr(int(k)) = out32
		out32 = LSHIFT(out32, 2)
		*S.ptr(0) = SMLAWB(S.idx(1), out32, int32(A_Q14.idx(0)))
		*S.ptr(1) = SMULWB(out32, int32(A_Q14.idx(1)))
	}
}

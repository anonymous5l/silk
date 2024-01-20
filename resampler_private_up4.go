package silk

func resampler_private_up4(S *slice[int32], out, in *slice[int16], inLen int32) {
	var (
		k                 int32
		in32, out32, Y, X int32
		out16             int16
	)

	for k = 0; k < inLen; k++ {
		in32 = LSHIFT(int32(in.idx(int(k))), 10)

		Y = SUB32(in32, S.idx(0))
		X = SMULWB(Y, resampler_up2_lq_0)
		out32 = ADD32(S.idx(0), X)
		*S.ptr(0) = ADD32(in32, X)

		out16 = int16(SAT16(RSHIFT_ROUND(out32, 10)))
		*out.ptr(int(4 * k)) = out16
		*out.ptr(int(4*k + 1)) = out16

		Y = SUB32(in32, S.idx(0))
		X = SMLAWB(Y, Y, resampler_up2_lq_1)
		out32 = ADD32(S.idx(1), X)
		*S.ptr(1) = ADD32(in32, X)

		out16 = int16(SAT16(RSHIFT_ROUND(out32, 10)))
		*out.ptr(int(4*k + 2)) = out16
		*out.ptr(int(4*k + 3)) = out16
	}
}

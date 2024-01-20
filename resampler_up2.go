package silk

func resampler_up2(S *slice[int32], out, in *slice[int16], inLen int32) {
	var (
		k                 int32
		in32, out32, Y, X int32
	)

	for k = 0; k < inLen; k++ {
		in32 = LSHIFT(int32(in.idx(int(k))), 10)

		Y = SUB32(in32, S.idx(0))
		X = SMULWB(Y, resampler_up2_lq_0)
		out32 = ADD32(S.idx(0), X)
		*S.ptr(0) = ADD32(in32, X)

		*out.ptr(int(2 * k)) = int16(SAT16(RSHIFT_ROUND(out32, 10)))

		Y = SUB32(in32, S.idx(1))
		X = SMLAWB(Y, Y, resampler_up2_lq_1)
		out32 = ADD32(S.idx(1), X)
		*S.ptr(1) = ADD32(in32, X)

		*out.ptr(int(2*k + 1)) = int16(SAT16(RSHIFT_ROUND(out32, 10)))
	}
}

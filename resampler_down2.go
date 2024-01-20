package silk

func resampler_down2(S *slice[int32], out, in *slice[int16], inLen int32) {
	var (
		k                 int32
		len2              = RSHIFT32(inLen, 1)
		in32, out32, Y, X int32
	)

	for k = 0; k < len2; k++ {
		in32 = LSHIFT(int32(in.idx(int(2*k))), 10)

		Y = SUB32(in32, S.idx(0))
		X = SMLAWB(Y, Y, resampler_down2_1)
		out32 = ADD32(S.idx(0), X)
		*S.ptr(0) = ADD32(in32, X)

		in32 = LSHIFT(int32(in.idx(int(2*k+1))), 10)

		Y = SUB32(in32, S.idx(1))
		X = SMULWB(Y, resampler_down2_0)
		out32 = ADD32(out32, S.idx(1))
		out32 = ADD32(out32, X)
		*S.ptr(1) = ADD32(in32, X)

		*out.ptr(int(k)) = int16(SAT16(RSHIFT_ROUND(out32, 11)))
	}
}

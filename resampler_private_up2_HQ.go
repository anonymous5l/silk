package silk

func resampler_private_up2_HQ(S *slice[int32], out, in *slice[int16], inLen int32) {
	var (
		k                            int32
		in32, out32_1, out32_2, Y, X int32
	)

	for k = 0; k < inLen; k++ {
		in32 = LSHIFT(int32(in.idx(int(k))), 10)

		Y = SUB32(in32, S.idx(0))
		X = SMULWB(Y, int32(resampler_up2_hq_0[0]))
		out32_1 = ADD32(S.idx(0), X)
		*S.ptr(0) = ADD32(in32, X)

		Y = SUB32(out32_1, S.idx(1))
		X = SMLAWB(Y, Y, int32(resampler_up2_hq_0[1]))
		out32_2 = ADD32(S.idx(1), X)
		*S.ptr(1) = ADD32(out32_1, X)

		out32_2 = SMLAWB(out32_2, S.idx(5), int32(resampler_up2_hq_notch[2]))
		out32_2 = SMLAWB(out32_2, S.idx(4), int32(resampler_up2_hq_notch[1]))
		out32_1 = SMLAWB(out32_2, S.idx(4), int32(resampler_up2_hq_notch[0]))
		*S.ptr(5) = SUB32(out32_2, S.idx(5))

		*out.ptr(int(2 * k)) = int16(SAT16(RSHIFT32(
			SMLAWB(256, out32_1, int32(resampler_up2_hq_notch[3])), 9)))

		Y = SUB32(in32, S.idx(2))
		X = SMULWB(Y, int32(resampler_up2_hq_1[0]))
		out32_1 = ADD32(S.idx(2), X)
		*S.ptr(2) = ADD32(in32, X)

		Y = SUB32(out32_1, S.idx(3))
		X = SMLAWB(Y, Y, int32(resampler_up2_hq_1[1]))
		out32_2 = ADD32(S.idx(3), X)
		*S.ptr(3) = ADD32(out32_1, X)

		out32_2 = SMLAWB(out32_2, S.idx(4), int32(resampler_up2_hq_notch[2]))
		out32_2 = SMLAWB(out32_2, S.idx(5), int32(resampler_up2_hq_notch[1]))
		out32_1 = SMLAWB(out32_2, S.idx(5), int32(resampler_up2_hq_notch[0]))
		*S.ptr(4) = SUB32(out32_2, S.idx(4))

		*out.ptr(int(2*k + 1)) = int16(SAT16(RSHIFT32(
			SMLAWB(256, out32_1, int32(resampler_up2_hq_notch[3])), 9)))
	}
}

func resampler_private_up2_HQ_wrapper(S *resampler_state_struct, out, in *slice[int16], inLen int32) {
	resampler_private_up2_HQ(S.sIIR, out, in, inLen)
}

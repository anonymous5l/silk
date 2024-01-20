package silk

func biquad_alt(in *slice[int16], B_Q28, A_Q28, S []int32, out *slice[int16], len int32) {

	var (
		k                                                        int32
		inval, A0_U_Q28, A0_L_Q28, A1_U_Q28, A1_L_Q28, out32_Q14 int32
	)

	A0_L_Q28 = (-A_Q28[0]) & 0x00003FFF
	A0_U_Q28 = RSHIFT(-A_Q28[0], 14)
	A1_L_Q28 = (-A_Q28[1]) & 0x00003FFF
	A1_U_Q28 = RSHIFT(-A_Q28[1], 14)

	for k = 0; k < len; k++ {
		inval = int32(in.idx(int(k)))
		out32_Q14 = LSHIFT(SMLAWB(S[0], B_Q28[0], inval), 2)

		S[0] = S[1] + RSHIFT_ROUND(SMULWB(out32_Q14, A0_L_Q28), 14)
		S[0] = SMLAWB(S[0], out32_Q14, A0_U_Q28)
		S[0] = SMLAWB(S[0], B_Q28[1], inval)

		S[1] = RSHIFT_ROUND(SMULWB(out32_Q14, A1_L_Q28), 14)
		S[1] = SMLAWB(S[1], out32_Q14, A1_U_Q28)
		S[1] = SMLAWB(S[1], B_Q28[2], inval)

		*out.ptr(int(k)) = int16(SAT16(RSHIFT(out32_Q14+(1<<14)-1, 14)))
	}
}

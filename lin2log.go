package silk

func lin2log(inLin int32) int32 {
	var lz, frac_Q7 int32

	CLZ_FRAC(inLin, &lz, &frac_Q7)

	return LSHIFT(31-lz, 7) + SMLAWB(frac_Q7, MUL(frac_Q7, 128-frac_Q7), 179)
}

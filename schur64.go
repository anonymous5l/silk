package silk

func schur64(rc_Q16, c *slice[int32], order int32) int32 {
	var (
		k, n                             int32
		C                                [MAX_ORDER_LPC + 1][2]int32
		Ctmp1_Q30, Ctmp2_Q30, rc_tmp_Q31 int32
	)

	if c.idx(0) <= 0 {
		memset(rc_Q16, 0, int(order))
		return 0
	}

	for k = 0; k < order+1; k++ {
		C[k][0] = c.idx(int(k))
		C[k][1] = C[k][0]
	}

	for k = 0; k < order; k++ {
		rc_tmp_Q31 = DIV32_varQ(-C[k+1][0], C[0][1], 31)

		*rc_Q16.ptr(int(k)) = RSHIFT_ROUND(rc_tmp_Q31, 15)

		for n = 0; n < order-k; n++ {
			Ctmp1_Q30 = C[n+k+1][0]
			Ctmp2_Q30 = C[n][1]

			C[n+k+1][0] = Ctmp1_Q30 + SMMUL(LSHIFT(Ctmp2_Q30, 1), rc_tmp_Q31)
			C[n][1] = Ctmp2_Q30 + SMMUL(LSHIFT(Ctmp1_Q30, 1), rc_tmp_Q31)
		}
	}

	return C[0][1]
}

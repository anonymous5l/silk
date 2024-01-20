package silk

func schur(rc_Q15 *slice[int16], c *slice[int32], order int32) int32 {
	var (
		k, n, lz                 int32
		C                        [MAX_ORDER_LPC + 1][2]int32
		Ctmp1, Ctmp2, rc_tmp_Q15 int32
	)

	lz = CLZ32(c.idx(0))

	if lz < 2 {
		for k = 0; k < order+1; k++ {
			C[k][0] = RSHIFT(c.idx(int(k)), 1)
			C[k][1] = C[k][0]
		}
	} else if lz > 2 {
		lz -= 2
		for k = 0; k < order+1; k++ {
			C[k][0] = LSHIFT(c.idx(int(k)), lz)
			C[k][1] = C[k][0]
		}
	} else {
		for k = 0; k < order+1; k++ {
			C[k][0] = c.idx(int(k))
			C[k][1] = C[k][0]
		}
	}

	for k = 0; k < order; k++ {

		rc_tmp_Q15 = -DIV32_16(C[k+1][0], int16(max(RSHIFT(C[0][1], 15), 1)))

		rc_tmp_Q15 = SAT16(rc_tmp_Q15)

		*rc_Q15.ptr(int(k)) = int16(rc_tmp_Q15)

		for n = 0; n < order-k; n++ {
			Ctmp1 = C[n+k+1][0]
			Ctmp2 = C[n][1]
			C[n+k+1][0] = SMLAWB(Ctmp1, LSHIFT(Ctmp2, 1), rc_tmp_Q15)
			C[n][1] = SMLAWB(Ctmp2, LSHIFT(Ctmp1, 1), rc_tmp_Q15)
		}
	}

	return C[0][1]
}

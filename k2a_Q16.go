package silk

func k2a_Q16(A_Q24, rc_Q16 *slice[int32], order int32) {

	var (
		k, n int32
		Atmp [MAX_ORDER_LPC]int32
	)

	for k = 0; k < order; k++ {
		for n = 0; n < k; n++ {
			Atmp[n] = A_Q24.idx(int(n))
		}
		for n = 0; n < k; n++ {
			*A_Q24.ptr(int(n)) = SMLAWW(A_Q24.idx(int(n)), Atmp[k-n-1], rc_Q16.idx(int(k)))
		}
		*A_Q24.ptr(int(k)) = -LSHIFT(rc_Q16.idx(int(k)), 8)
	}
}

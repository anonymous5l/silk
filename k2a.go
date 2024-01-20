package silk

func k2a(A_Q24 *slice[int32], rc_Q15 *slice[int16], order int32) {
	var (
		k, n int32
		Atmp [MAX_ORDER_LPC]int32
	)

	for k = 0; k < order; k++ {
		for n = 0; n < k; n++ {
			Atmp[n] = A_Q24.idx(int(n))
		}
		for n = 0; n < k; n++ {
			*A_Q24.ptr(int(n)) = SMLAWB(A_Q24.idx(int(n)), LSHIFT(Atmp[k-n-1], 1), int32(rc_Q15.idx(int(k))))
		}
		*A_Q24.ptr(int(n)) = -LSHIFT(int32(rc_Q15.idx(int(k))), 9)
	}
}

package silk

func NLSF_VQ_sum_error_FIX(
	err_Q20, in_Q15, w_Q6 *slice[int32],
	pCB_Q15 *slice[int16],
	N, K, LPC_order int32) {
	var (
		i, n, m                      int32
		diff_Q15, sum_error, Wtmp_Q6 int32
		Wcpy_Q6                      [MAX_LPC_ORDER / 2]int32
		cb_vec_Q15                   *slice[int16]
	)

	for m = 0; m < RSHIFT(LPC_order, 1); m++ {
		Wcpy_Q6[m] = w_Q6.idx(int(2*m)) | LSHIFT(int32(w_Q6.idx(int(2*m+1))), 16)
	}

	for n = 0; n < N; n++ {
		cb_vec_Q15 = pCB_Q15
		for i = 0; i < K; i++ {
			sum_error = 0
			for m = 0; m < LPC_order; m += 2 {
				Wtmp_Q6 = Wcpy_Q6[RSHIFT(m, 1)]

				diff_Q15 = in_Q15.idx(int(m)) - int32(cb_vec_Q15.idx(0))
				cb_vec_Q15 = cb_vec_Q15.off(1)
				sum_error = SMLAWB(sum_error, SMULBB(diff_Q15, diff_Q15), Wtmp_Q6)

				diff_Q15 = in_Q15.idx(int(m+1)) - int32(cb_vec_Q15.idx(0))
				cb_vec_Q15 = cb_vec_Q15.off(1)
				sum_error = SMLAWT(sum_error, SMULBB(diff_Q15, diff_Q15), Wtmp_Q6)
			}

			*err_Q20.ptr(int(i)) = sum_error
		}
		err_Q20 = err_Q20.off(int(K))
		in_Q15 = in_Q15.off(int(LPC_order))
	}
}

package silk

func NLSF_VQ_rate_distortion_FIX(
	pRD_Q20 *slice[int32],
	psNLSF_CBS *NLSF_CBS,
	in_Q15, w_Q6, rate_acc_Q5 *slice[int32],
	mu_Q15, N, LPC_order int32) {

	var (
		i, n        int32
		pRD_vec_Q20 *slice[int32]
	)

	NLSF_VQ_sum_error_FIX(pRD_Q20, in_Q15, w_Q6, psNLSF_CBS.CB_NLSF_Q15,
		N, psNLSF_CBS.nVectors, LPC_order)

	pRD_vec_Q20 = pRD_Q20
	for n = 0; n < N; n++ {
		for i = 0; i < psNLSF_CBS.nVectors; i++ {
			*pRD_vec_Q20.ptr(int(i)) = SMLABB(pRD_vec_Q20.idx(int(i)), rate_acc_Q5.idx(int(n))+int32(psNLSF_CBS.Rates_Q5.idx(int(i))), mu_Q15)
		}
		pRD_vec_Q20 = pRD_vec_Q20.off(int(psNLSF_CBS.nVectors))
	}
}

package silk

func LTP_analysis_filter_FIX(
	LTP_res, x, LTPCoef_Q14 *slice[int16],
	pitchL, invGains_Q16 *slice[int32],
	subfr_length, pre_length int32) {

	var (
		x_ptr, x_lag_ptr *slice[int16]
		Btmp_Q14         = alloc[int16](LTP_ORDER)
		LTP_res_ptr      *slice[int16]
		k, i, j, LTP_est int32
	)

	x_ptr = x
	LTP_res_ptr = LTP_res
	for k = 0; k < NB_SUBFR; k++ {

		x_lag_ptr = x_ptr.off(-int(pitchL.idx(int(k))))
		for i = 0; i < LTP_ORDER; i++ {
			*Btmp_Q14.ptr(int(i)) = LTPCoef_Q14.idx(int(k*LTP_ORDER + i))
		}

		for i = 0; i < subfr_length+pre_length; i++ {
			*LTP_res_ptr.ptr(int(i)) = x_ptr.idx(int(i))

			LTP_est = SMULBB(int32(x_lag_ptr.idx(LTP_ORDER/2)), int32(Btmp_Q14.idx(0)))
			for j = 1; j < LTP_ORDER; j++ {
				LTP_est = SMLABB_ovflw(LTP_est, int32(x_lag_ptr.idx(int(LTP_ORDER/2-j))), int32(Btmp_Q14.idx(int(j))))
			}
			LTP_est = RSHIFT_ROUND(LTP_est, 14)

			*LTP_res_ptr.ptr(int(i)) = int16(SAT16(int32(x_ptr.idx(int(i))) - LTP_est))

			*LTP_res_ptr.ptr(int(i)) = int16(SMULWB(invGains_Q16.idx(int(k)), int32(LTP_res_ptr.idx(int(i)))))

			x_lag_ptr = x_lag_ptr.off(1)
		}

		LTP_res_ptr = LTP_res_ptr.off(int(subfr_length + pre_length))
		x_ptr = x_ptr.off(int(subfr_length))
	}
}

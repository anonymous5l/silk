package silk

const LTP_CORRS_HEAD_ROOM = 2

func find_LTP_FIX(b_Q14 *slice[int16], WLTP *slice[int32], LTPredCodGain_Q7 *int32,
	r_first *slice[int16], r_last *slice[int16], lag *slice[int32],
	Wght_Q15 *slice[int32], subfr_length, mem_offset int32, corr_rshifts *slice[int32]) {

	var (
		i, k, lshift   int32
		r_ptr, lag_ptr *slice[int16]
		b_Q14_ptr      *slice[int16]

		regu                                         int32
		WLTP_ptr                                     *slice[int32]
		b_Q16                                        = alloc[int32](LTP_ORDER)
		delta_b_Q14                                  = alloc[int32](LTP_ORDER)
		d_Q14                                        = alloc[int32](NB_SUBFR)
		nrg                                          = alloc[int32](NB_SUBFR)
		g_Q26                                        int32
		w                                            = alloc[int32](NB_SUBFR)
		WLTP_max, max_abs_d_Q14, max_w_bits          int32
		temp32, denom32                              int32
		extra_shifts                                 int32
		rr_shifts, maxRshifts, maxRshifts_wxtra, LZs int32
		LPC_res_nrg, LPC_LTP_res_nrg, div_Q16        int32
		Rr                                           = alloc[int32](LTP_ORDER)
		rr                                           = alloc[int32](NB_SUBFR)
		wd, m_Q12                                    int32
	)

	b_Q14_ptr = b_Q14
	WLTP_ptr = WLTP
	r_ptr = r_first.off(int(mem_offset))
	for k = 0; k < NB_SUBFR; k++ {
		if k == (NB_SUBFR >> 1) {
			r_ptr = r_last.off(int(mem_offset))
		}
		lag_ptr = r_ptr.off(-int(lag.idx(int(k)) + LTP_ORDER/2))

		sum_sqr_shift(rr.ptr(int(k)), &rr_shifts, r_ptr, subfr_length)

		LZs = CLZ32(rr.idx(int(k)))
		if LZs < LTP_CORRS_HEAD_ROOM {
			*rr.ptr(int(k)) = RSHIFT_ROUND(rr.idx(int(k)), LTP_CORRS_HEAD_ROOM-LZs)
			rr_shifts += LTP_CORRS_HEAD_ROOM - LZs
		}
		*corr_rshifts.ptr(int(k)) = rr_shifts
		corrMatrix_FIX(lag_ptr, subfr_length, LTP_ORDER, LTP_CORRS_HEAD_ROOM, WLTP_ptr, corr_rshifts.ptr(int(k)))

		corrVector_FIX(lag_ptr, r_ptr, subfr_length, LTP_ORDER, Rr, corr_rshifts.idx(int(k)))
		if corr_rshifts.idx(int(k)) > rr_shifts {
			*rr.ptr(int(k)) = RSHIFT(rr.idx(int(k)), corr_rshifts.idx(int(k))-rr_shifts)
		}

		regu = 1
		regu = SMLAWB(regu, rr.idx(int(k)), FIX_CONST(LTP_DAMPING/3, 16))
		regu = SMLAWB(regu, *matrix_ptr(WLTP_ptr, 0, 0, LTP_ORDER), FIX_CONST(LTP_DAMPING/3, 16))
		regu = SMLAWB(regu, *matrix_ptr(WLTP_ptr, LTP_ORDER-1, LTP_ORDER-1, LTP_ORDER), FIX_CONST(LTP_DAMPING/3, 16))
		regularize_correlations_FIX(WLTP_ptr, rr.off(int(k)), regu, LTP_ORDER)

		solve_LDL_FIX(WLTP_ptr, LTP_ORDER, Rr, b_Q16)

		fit_LTP(b_Q16, b_Q14_ptr)

		*nrg.ptr(int(k)) = residual_energy16_covar_FIX(b_Q14_ptr, WLTP_ptr, Rr, rr.idx(int(k)), LTP_ORDER, 14)

		extra_shifts = min(corr_rshifts.idx(int(k)), LTP_CORRS_HEAD_ROOM)
		denom32 = LSHIFT_SAT32(SMULWB(nrg.idx(int(k)), Wght_Q15.idx(int(k))), 1+extra_shifts) +
			RSHIFT(SMULWB(subfr_length, 655), corr_rshifts.idx(int(k))-extra_shifts)
		denom32 = max(denom32, 1)

		temp32 = DIV32(LSHIFT(Wght_Q15.idx(int(k)), 16), denom32)
		temp32 = RSHIFT(temp32, 31+corr_rshifts.idx(int(k))-extra_shifts-26)

		WLTP_max = 0
		for i = 0; i < LTP_ORDER*LTP_ORDER; i++ {
			WLTP_max = max(WLTP_ptr.idx(int(i)), WLTP_max)
		}
		lshift = CLZ32(WLTP_max) - 1 - 3

		if 26-18+lshift < 31 {
			temp32 = min(temp32, LSHIFT(1, 26-18+lshift))
		}

		scale_vector32_Q26_lshift_18(WLTP_ptr, temp32, LTP_ORDER*LTP_ORDER)

		*w.ptr(int(k)) = *matrix_ptr(WLTP_ptr, LTP_ORDER>>1, LTP_ORDER>>1, LTP_ORDER)

		r_ptr = r_ptr.off(int(subfr_length))
		b_Q14_ptr = b_Q14_ptr.off(LTP_ORDER)
		WLTP_ptr = WLTP_ptr.off(LTP_ORDER * LTP_ORDER)
	}

	maxRshifts = 0
	for k = 0; k < NB_SUBFR; k++ {
		maxRshifts = max(corr_rshifts.idx(int(k)), maxRshifts)
	}

	if LTPredCodGain_Q7 != nil {
		LPC_LTP_res_nrg = 0
		LPC_res_nrg = 0

		for k = 0; k < NB_SUBFR; k++ {
			LPC_res_nrg = ADD32(LPC_res_nrg, RSHIFT(ADD32(SMULWB(rr.idx(int(k)), Wght_Q15.idx(int(k))), 1), 1+(maxRshifts-corr_rshifts.idx(int(k)))))
			LPC_LTP_res_nrg = ADD32(LPC_LTP_res_nrg, RSHIFT(ADD32(SMULWB(nrg.idx(int(k)), Wght_Q15.idx(int(k))), 1), 1+(maxRshifts-corr_rshifts.idx(int(k)))))
		}
		LPC_LTP_res_nrg = max(LPC_LTP_res_nrg, 1)

		div_Q16 = DIV32_varQ(LPC_res_nrg, LPC_LTP_res_nrg, 16)
		*LTPredCodGain_Q7 = SMULBB(3, lin2log(div_Q16)-(16<<7))

	}

	b_Q14_ptr = b_Q14
	for k = 0; k < NB_SUBFR; k++ {
		*d_Q14.ptr(int(k)) = 0
		for i = 0; i < LTP_ORDER; i++ {
			*d_Q14.ptr(int(k)) += int32(b_Q14_ptr.idx(int(i)))
		}
		b_Q14_ptr = b_Q14_ptr.off(LTP_ORDER)
	}

	max_abs_d_Q14 = 0
	max_w_bits = 0
	for k = 0; k < NB_SUBFR; k++ {
		max_abs_d_Q14 = max(max_abs_d_Q14, abs(d_Q14.idx(int(k))))
		max_w_bits = max(max_w_bits, 32-CLZ32(w.idx(int(k)))+corr_rshifts.idx(int(k))-maxRshifts)
	}

	extra_shifts = max_w_bits + 32 - CLZ32(max_abs_d_Q14) - 14

	extra_shifts -= 32 - 1 - 2 + maxRshifts
	extra_shifts = max(extra_shifts, 0)

	maxRshifts_wxtra = maxRshifts + extra_shifts

	temp32 = RSHIFT(262, maxRshifts+extra_shifts) + 1
	wd = 0
	for k = 0; k < NB_SUBFR; k++ {
		temp32 = ADD32(temp32, RSHIFT(w.idx(int(k)), maxRshifts_wxtra-corr_rshifts.idx(int(k))))
		wd = ADD32(wd, LSHIFT(SMULWW(RSHIFT(w.idx(int(k)), maxRshifts_wxtra-corr_rshifts.idx(int(k))), d_Q14.idx(int(k))), 2))
	}
	m_Q12 = DIV32_varQ(wd, temp32, 12)

	b_Q14_ptr = b_Q14
	for k = 0; k < NB_SUBFR; k++ {
		if 2-corr_rshifts.idx(int(k)) > 0 {
			temp32 = RSHIFT(w.idx(int(k)), 2-corr_rshifts.idx(int(k)))
		} else {
			temp32 = LSHIFT_SAT32(w.idx(int(k)), corr_rshifts.idx(int(k))-2)
		}

		g_Q26 = MUL(
			DIV32(
				FIX_CONST(LTP_SMOOTHING, 26),
				RSHIFT(FIX_CONST(LTP_SMOOTHING, 26), 10)+temp32),
			LSHIFT_SAT32(SUB_SAT32(m_Q12, RSHIFT(d_Q14.idx(int(k)), 2)), 4))

		temp32 = 0
		for i = 0; i < LTP_ORDER; i++ {
			*delta_b_Q14.ptr(int(i)) = int32(max(b_Q14_ptr.idx(int(i)), 1638))
			temp32 += delta_b_Q14.idx(int(i))
		}
		temp32 = DIV32(g_Q26, temp32)
		for i = 0; i < LTP_ORDER; i++ {
			*b_Q14_ptr.ptr(int(i)) = int16(LIMIT(int32(b_Q14_ptr.idx(int(i)))+SMULWB(LSHIFT_SAT32(temp32, 4),
				delta_b_Q14.idx(int(i))), -16000, 28000))
		}
		b_Q14_ptr = b_Q14_ptr.off(LTP_ORDER)
	}
}

func fit_LTP(LTP_coefs_Q16 *slice[int32], LTP_coefs_Q14 *slice[int16]) {
	var i int32
	for i = 0; i < LTP_ORDER; i++ {
		*LTP_coefs_Q14.ptr(int(i)) = int16(SAT16(RSHIFT_ROUND(LTP_coefs_Q16.idx(int(i)), 2)))
	}
}

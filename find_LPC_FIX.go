package silk

func find_LPC_FIX(
	NLSF_Q15 *slice[int32],
	interpIndex *int32,
	prev_NLSFq_Q15 *slice[int32],
	useInterpolatedNLSFs, LPC_order int32,
	x *slice[int16],
	subfr_length int32) {
	var (
		k                                          int32
		a_Q16                                      = alloc[int32](MAX_LPC_ORDER)
		isInterpLower, shift                       int32
		S                                          = alloc[int16](MAX_LPC_ORDER)
		res_nrg0, res_nrg1                         int32
		rshift0, rshift1                           int32
		a_tmp_Q16                                  = alloc[int32](MAX_LPC_ORDER)
		res_nrg_interp, res_nrg, res_tmp_nrg       int32
		res_nrg_interp_Q, res_nrg_Q, res_tmp_nrg_Q int32
		a_tmp_Q12                                  = alloc[int16](MAX_LPC_ORDER)
		NLSF0_Q15                                  = alloc[int32](MAX_LPC_ORDER)
		LPC_res                                    = alloc[int16]((MAX_FRAME_LENGTH + NB_SUBFR*MAX_LPC_ORDER) / 2)
	)

	*interpIndex = 4

	burg_modified(&res_nrg, &res_nrg_Q, a_Q16, x, subfr_length, NB_SUBFR, FIX_CONST(FIND_LPC_COND_FAC, 32), LPC_order)

	bwexpander_32(a_Q16, LPC_order, FIX_CONST(FIND_LPC_CHIRP, 16))

	if useInterpolatedNLSFs == 1 {

		burg_modified(&res_tmp_nrg, &res_tmp_nrg_Q, a_tmp_Q16, x.off(int((NB_SUBFR>>1)*subfr_length)),
			subfr_length, (NB_SUBFR >> 1), FIX_CONST(FIND_LPC_COND_FAC, 32), LPC_order)

		bwexpander_32(a_tmp_Q16, LPC_order, FIX_CONST(FIND_LPC_CHIRP, 16))

		shift = res_tmp_nrg_Q - res_nrg_Q
		if shift >= 0 {
			if shift < 32 {
				res_nrg = res_nrg - RSHIFT(res_tmp_nrg, shift)
			}
		} else {
			res_nrg = RSHIFT(res_nrg, -shift) - res_tmp_nrg
			res_nrg_Q = res_tmp_nrg_Q
		}

		A2NLSF(NLSF_Q15, a_tmp_Q16, LPC_order)

		for k = 3; k >= 0; k-- {
			interpolate(NLSF0_Q15, prev_NLSFq_Q15, NLSF_Q15, k, LPC_order)

			NLSF2A_stable(a_tmp_Q12, NLSF0_Q15, LPC_order)

			memset(S, 0, int(LPC_order))
			LPC_analysis_filter(x, a_tmp_Q12, S, LPC_res, 2*subfr_length, LPC_order)

			sum_sqr_shift(&res_nrg0, &rshift0, LPC_res.off(int(LPC_order)), subfr_length-LPC_order)
			sum_sqr_shift(&res_nrg1, &rshift1, LPC_res.off(int(LPC_order+subfr_length)), subfr_length-LPC_order)

			shift = rshift0 - rshift1
			if shift >= 0 {
				res_nrg1 = RSHIFT(res_nrg1, shift)
				res_nrg_interp_Q = -rshift0
			} else {
				res_nrg0 = RSHIFT(res_nrg0, -shift)
				res_nrg_interp_Q = -rshift1
			}
			res_nrg_interp = ADD32(res_nrg0, res_nrg1)

			shift = res_nrg_interp_Q - res_nrg_Q
			if shift >= 0 {
				if RSHIFT(res_nrg_interp, shift) < res_nrg {
					isInterpLower = 1
				} else {
					isInterpLower = 0
				}
			} else {
				if -shift < 32 {
					if res_nrg_interp < RSHIFT(res_nrg, -shift) {
						isInterpLower = 1
					} else {
						isInterpLower = 0
					}
				} else {
					isInterpLower = 0
				}
			}

			if isInterpLower == 1 {
				res_nrg = res_nrg_interp
				res_nrg_Q = res_nrg_interp_Q
				*interpIndex = k
			}
		}
	}

	if *interpIndex == 4 {
		A2NLSF(NLSF_Q15, a_Q16, LPC_order)
	}
}

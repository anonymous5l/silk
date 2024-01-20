package silk

import (
	"math"
)

func decode_core(psDec *decoder_state, psDecCtrl *decoder_control, xq *slice[int16], q *slice[int32]) {
	var (
		i, k, lag, start_idx, sLTP_buf_idx, NLSF_interpolation_flag, sigtype                            int32
		A_Q12, B_Q14, pxq                                                                               *slice[int16]
		A_Q12_tmp                                                                                       = alloc[int16](MAX_LPC_ORDER)
		sLTP                                                                                            = alloc[int16](MAX_FRAME_LENGTH)
		LTP_pred_Q14, Gain_Q16, inv_gain_Q16, inv_gain_Q32, gain_adj_Q16, rand_seed, offset_Q10, dither int32
		pred_lag_ptr, pexc_Q10, pres_Q10                                                                *slice[int32]
		vec_Q10                                                                                         = alloc[int32](MAX_FRAME_LENGTH / NB_SUBFR)
		FiltState                                                                                       = alloc[int32](MAX_LPC_ORDER)
	)

	offset_Q10 = int32(Quantization_Offsets_Q10[psDecCtrl.sigtype][psDecCtrl.QuantOffsetType])

	if psDecCtrl.NLSFInterpCoef_Q2 < (1 << 2) {
		NLSF_interpolation_flag = 1
	} else {
		NLSF_interpolation_flag = 0
	}

	rand_seed = psDecCtrl.Seed
	for i = 0; i < psDec.frame_length; i++ {
		rand_seed = RAND(rand_seed)
		dither = RSHIFT(rand_seed, 31)

		*psDec.exc_Q10.ptr(int(i)) = LSHIFT(q.idx(int(i)), 10) + offset_Q10
		*psDec.exc_Q10.ptr(int(i)) = (psDec.exc_Q10.idx(int(i)) ^ dither) - dither

		rand_seed += q.idx(int(i))
	}

	pexc_Q10 = psDec.exc_Q10.off(0)
	pres_Q10 = psDec.res_Q10.off(0)
	pxq = psDec.outBuf.off(int(psDec.frame_length))
	sLTP_buf_idx = psDec.frame_length
	for k = 0; k < NB_SUBFR; k++ {
		A_Q12 = psDecCtrl.PredCoef_Q12[k>>1].off(0)

		A_Q12.copy(A_Q12_tmp, int(psDec.LPC_order))

		B_Q14 = psDecCtrl.LTPCoef_Q14.off(int(k * LTP_ORDER))
		Gain_Q16 = psDecCtrl.Gains_Q16.idx(int(k))
		sigtype = psDecCtrl.sigtype

		inv_gain_Q16 = INVERSE32_varQ(max(Gain_Q16, 1), 32)
		inv_gain_Q16 = min(inv_gain_Q16, math.MaxInt16)

		gain_adj_Q16 = 1 << 16
		if inv_gain_Q16 != psDec.prev_inv_gain_Q16 {
			gain_adj_Q16 = DIV32_varQ(inv_gain_Q16, psDec.prev_inv_gain_Q16, 16)
		}

		if psDec.lossCnt != 0 && psDec.prev_sigtype == SIG_TYPE_VOICED &&
			psDecCtrl.sigtype == SIG_TYPE_UNVOICED && k < (NB_SUBFR>>1) {

			memset(B_Q14, 0, LTP_ORDER)
			*B_Q14.ptr(LTP_ORDER / 2) = 1 << 12

			sigtype = SIG_TYPE_VOICED
			*psDecCtrl.pitchL.ptr(int(k)) = psDec.lagPrev
		}

		if sigtype == SIG_TYPE_VOICED {

			lag = psDecCtrl.pitchL.idx(int(k))
			if (k & (3 - LSHIFT(NLSF_interpolation_flag, 1))) == 0 {
				start_idx = psDec.frame_length - lag - psDec.LPC_order - LTP_ORDER/2

				memset(FiltState, 0, int(psDec.LPC_order))

				MA_Prediction(psDec.outBuf.off(int(start_idx+k*(psDec.frame_length>>2))),
					A_Q12, FiltState, sLTP.off(int(start_idx)),
					psDec.frame_length-start_idx, psDec.LPC_order)

				inv_gain_Q32 = LSHIFT(inv_gain_Q16, 16)
				if k == 0 {
					inv_gain_Q32 = LSHIFT(SMULWB(inv_gain_Q32, int32(psDecCtrl.LTP_scale_Q14)), 2)
				}

				for i = 0; i < (lag + LTP_ORDER/2); i++ {
					*psDec.sLTP_Q16.ptr(int(sLTP_buf_idx - i - 1)) = SMULWB(inv_gain_Q32, int32(sLTP.idx(int(psDec.frame_length-i-1))))
				}

			} else {
				if gain_adj_Q16 != 1<<16 {
					for i = 0; i < (lag + LTP_ORDER/2); i++ {
						*psDec.sLTP_Q16.ptr(int(sLTP_buf_idx - i - 1)) = SMULWW(gain_adj_Q16, psDec.sLTP_Q16.idx(int(sLTP_buf_idx-i-1)))
					}
				}
			}
		}

		for i = 0; i < MAX_LPC_ORDER; i++ {
			*psDec.sLPC_Q14.ptr(int(i)) = SMULWW(gain_adj_Q16, psDec.sLPC_Q14.idx(int(i)))
		}

		psDec.prev_inv_gain_Q16 = inv_gain_Q16

		if sigtype == SIG_TYPE_VOICED {
			pred_lag_ptr = psDec.sLTP_Q16.off(int(sLTP_buf_idx - lag + LTP_ORDER/2))

			for i = 0; i < psDec.subfr_length; i++ {
				LTP_pred_Q14 = SMULWB(pred_lag_ptr.idx(0), int32(B_Q14.idx(0)))
				LTP_pred_Q14 = SMLAWB(LTP_pred_Q14, pred_lag_ptr.idx(-1), int32(B_Q14.idx(1)))
				LTP_pred_Q14 = SMLAWB(LTP_pred_Q14, pred_lag_ptr.idx(-2), int32(B_Q14.idx(2)))
				LTP_pred_Q14 = SMLAWB(LTP_pred_Q14, pred_lag_ptr.idx(-3), int32(B_Q14.idx(3)))
				LTP_pred_Q14 = SMLAWB(LTP_pred_Q14, pred_lag_ptr.idx(-4), int32(B_Q14.idx(4)))

				pred_lag_ptr = pred_lag_ptr.off(1)
				*pres_Q10.ptr(int(i)) = ADD32(pexc_Q10.idx(int(i)), RSHIFT_ROUND(LTP_pred_Q14, 4))

				*psDec.sLTP_Q16.ptr(int(sLTP_buf_idx)) = LSHIFT(pres_Q10.idx(int(i)), 6)
				sLTP_buf_idx++
			}
		} else {
			pexc_Q10.copy(pres_Q10, int(psDec.subfr_length))
		}

		decode_short_term_prediction(vec_Q10, pres_Q10, psDec.sLPC_Q14, A_Q12_tmp, psDec.LPC_order, psDec.subfr_length)

		for i = 0; i < psDec.subfr_length; i++ {
			*pxq.ptr(int(i)) = int16(SAT16(RSHIFT_ROUND(SMULWW(vec_Q10.idx(int(i)), Gain_Q16), 10)))
		}

		a := psDec.sLPC_Q14.off(int(psDec.subfr_length))
		a.copy(psDec.sLPC_Q14, MAX_LPC_ORDER)

		pexc_Q10 = pexc_Q10.off(int(psDec.subfr_length))
		pres_Q10 = pres_Q10.off(int(psDec.subfr_length))
		pxq = pxq.off(int(psDec.subfr_length))
	}

	psDec.outBuf.off(int(psDec.frame_length)).copy(xq, int(psDec.frame_length))
}

func decode_short_term_prediction(vec_Q10, pres_Q10, sLPC_Q14 *slice[int32],
	A_Q12_tmp *slice[int16], LPC_order, subfr_length int32) {

	var i, LPC_pred_Q10, Atmp int32

	A_Q12_tmp32 := slice2[int32](A_Q12_tmp)

	if LPC_order == 16 {
		for i = 0; i < subfr_length; i++ {
			Atmp = A_Q12_tmp32.idx(0)
			LPC_pred_Q10 = SMULWB(sLPC_Q14.idx(int(MAX_LPC_ORDER+i-1)), Atmp)
			LPC_pred_Q10 = SMLAWT(LPC_pred_Q10, sLPC_Q14.idx(int(MAX_LPC_ORDER+i-2)), Atmp)
			Atmp = A_Q12_tmp32.idx(1)
			LPC_pred_Q10 = SMLAWB(LPC_pred_Q10, sLPC_Q14.idx(int(MAX_LPC_ORDER+i-3)), Atmp)
			LPC_pred_Q10 = SMLAWT(LPC_pred_Q10, sLPC_Q14.idx(int(MAX_LPC_ORDER+i-4)), Atmp)
			Atmp = A_Q12_tmp32.idx(2)
			LPC_pred_Q10 = SMLAWB(LPC_pred_Q10, sLPC_Q14.idx(int(MAX_LPC_ORDER+i-5)), Atmp)
			LPC_pred_Q10 = SMLAWT(LPC_pred_Q10, sLPC_Q14.idx(int(MAX_LPC_ORDER+i-6)), Atmp)
			Atmp = A_Q12_tmp32.idx(3)
			LPC_pred_Q10 = SMLAWB(LPC_pred_Q10, sLPC_Q14.idx(int(MAX_LPC_ORDER+i-7)), Atmp)
			LPC_pred_Q10 = SMLAWT(LPC_pred_Q10, sLPC_Q14.idx(int(MAX_LPC_ORDER+i-8)), Atmp)
			Atmp = A_Q12_tmp32.idx(4)
			LPC_pred_Q10 = SMLAWB(LPC_pred_Q10, sLPC_Q14.idx(int(MAX_LPC_ORDER+i-9)), Atmp)
			LPC_pred_Q10 = SMLAWT(LPC_pred_Q10, sLPC_Q14.idx(int(MAX_LPC_ORDER+i-10)), Atmp)
			Atmp = A_Q12_tmp32.idx(5)
			LPC_pred_Q10 = SMLAWB(LPC_pred_Q10, sLPC_Q14.idx(int(MAX_LPC_ORDER+i-11)), Atmp)
			LPC_pred_Q10 = SMLAWT(LPC_pred_Q10, sLPC_Q14.idx(int(MAX_LPC_ORDER+i-12)), Atmp)
			Atmp = A_Q12_tmp32.idx(6)
			LPC_pred_Q10 = SMLAWB(LPC_pred_Q10, sLPC_Q14.idx(int(MAX_LPC_ORDER+i-13)), Atmp)
			LPC_pred_Q10 = SMLAWT(LPC_pred_Q10, sLPC_Q14.idx(int(MAX_LPC_ORDER+i-14)), Atmp)
			Atmp = A_Q12_tmp32.idx(7)
			LPC_pred_Q10 = SMLAWB(LPC_pred_Q10, sLPC_Q14.idx(int(MAX_LPC_ORDER+i-15)), Atmp)
			LPC_pred_Q10 = SMLAWT(LPC_pred_Q10, sLPC_Q14.idx(int(MAX_LPC_ORDER+i-16)), Atmp)

			*vec_Q10.ptr(int(i)) = ADD32(pres_Q10.idx(int(i)), LPC_pred_Q10)

			*sLPC_Q14.ptr(int(MAX_LPC_ORDER + i)) = int32(LSHIFT_ovflw(uint32(vec_Q10.idx(int(i))), 4))
		}
	} else {
		for i = 0; i < subfr_length; i++ {
			Atmp = A_Q12_tmp32.idx(0)
			LPC_pred_Q10 = SMULWB(sLPC_Q14.idx(int(MAX_LPC_ORDER+i-1)), Atmp)
			LPC_pred_Q10 = SMLAWT(LPC_pred_Q10, sLPC_Q14.idx(int(MAX_LPC_ORDER+i-2)), Atmp)
			Atmp = A_Q12_tmp32.idx(1)
			LPC_pred_Q10 = SMLAWB(LPC_pred_Q10, sLPC_Q14.idx(int(MAX_LPC_ORDER+i-3)), Atmp)
			LPC_pred_Q10 = SMLAWT(LPC_pred_Q10, sLPC_Q14.idx(int(MAX_LPC_ORDER+i-4)), Atmp)
			Atmp = A_Q12_tmp32.idx(2)
			LPC_pred_Q10 = SMLAWB(LPC_pred_Q10, sLPC_Q14.idx(int(MAX_LPC_ORDER+i-5)), Atmp)
			LPC_pred_Q10 = SMLAWT(LPC_pred_Q10, sLPC_Q14.idx(int(MAX_LPC_ORDER+i-6)), Atmp)
			Atmp = A_Q12_tmp32.idx(3)
			LPC_pred_Q10 = SMLAWB(LPC_pred_Q10, sLPC_Q14.idx(int(MAX_LPC_ORDER+i-7)), Atmp)
			LPC_pred_Q10 = SMLAWT(LPC_pred_Q10, sLPC_Q14.idx(int(MAX_LPC_ORDER+i-8)), Atmp)
			Atmp = A_Q12_tmp32.idx(4)
			LPC_pred_Q10 = SMLAWB(LPC_pred_Q10, sLPC_Q14.idx(int(MAX_LPC_ORDER+i-9)), Atmp)
			LPC_pred_Q10 = SMLAWT(LPC_pred_Q10, sLPC_Q14.idx(int(MAX_LPC_ORDER+i-10)), Atmp)

			*vec_Q10.ptr(int(i)) = ADD32(pres_Q10.idx(int(i)), LPC_pred_Q10)

			*sLPC_Q14.ptr(int(MAX_LPC_ORDER + i)) = int32(LSHIFT_ovflw(uint32(vec_Q10.idx(int(i))), 4))
		}
	}
}

package silk

import (
	"math"
)

func NSQ(
	psEncC *encoder_state,
	psEncCtrlC *encoder_control,
	NSQ *nsq_state,
	x *slice[int16],
	q *slice[int8],
	LSFInterpFactor_Q2 int32,
	PredCoef_Q12 [2]*slice[int16],
	LTPCoef_Q14, AR2_Q13 *slice[int16],
	HarmShapeGain_Q14 *slice[int32],
	Tilt_Q14, LF_shp_Q14, Gains_Q16 *slice[int32],
	Lambda_Q10, LTP_scale_Q14 int32) {
	var (
		k, lag, start_idx, LSF_interpolation_flag int32
		A_Q12, B_Q14, AR_shp_Q13                  *slice[int16]
		pxq                                       *slice[int16]
		sLTP_Q16                                  = alloc[int32](2 * MAX_FRAME_LENGTH)
		sLTP                                      = alloc[int16](2 * MAX_FRAME_LENGTH)
		HarmShapeFIRPacked_Q14                    int32
		offset_Q10                                int32
		FiltState                                 = alloc[int32](MAX_LPC_ORDER)
		x_sc_Q10                                  = alloc[int32](MAX_FRAME_LENGTH / NB_SUBFR)
	)

	NSQ.rand_seed = psEncCtrlC.Seed
	lag = NSQ.lagPrev

	offset_Q10 = int32(Quantization_Offsets_Q10[psEncCtrlC.sigtype][psEncCtrlC.QuantOffsetType])

	if LSFInterpFactor_Q2 == (1 << 2) {
		LSF_interpolation_flag = 0
	} else {
		LSF_interpolation_flag = 1
	}

	NSQ.sLTP_shp_buf_idx = psEncC.frame_length
	NSQ.sLTP_buf_idx = psEncC.frame_length
	pxq = NSQ.xq.off(int(psEncC.frame_length))
	for k = 0; k < NB_SUBFR; k++ {
		A_Q12 = PredCoef_Q12[(k>>1)|(1-LSF_interpolation_flag)]
		B_Q14 = LTPCoef_Q14.off(int(k * LTP_ORDER))
		AR_shp_Q13 = AR2_Q13.off(int(k * MAX_SHAPE_LPC_ORDER))

		HarmShapeFIRPacked_Q14 = RSHIFT(HarmShapeGain_Q14.idx(int(k)), 2)
		HarmShapeFIRPacked_Q14 |= LSHIFT(RSHIFT(HarmShapeGain_Q14.idx(int(k)), 1), 16)

		NSQ.rewhite_flag = 0
		if psEncCtrlC.sigtype == SIG_TYPE_VOICED {
			lag = psEncCtrlC.pitchL.idx(int(k))

			if (k & (3 - LSHIFT(LSF_interpolation_flag, 1))) == 0 {

				start_idx = psEncC.frame_length - lag - psEncC.predictLPCOrder - LTP_ORDER/2

				memset(FiltState, 0, int(psEncC.predictLPCOrder))
				MA_Prediction(NSQ.xq.off(int(start_idx+k*(psEncC.frame_length>>2))),
					A_Q12, FiltState, sLTP.off(int(start_idx)), psEncC.frame_length-start_idx, psEncC.predictLPCOrder)

				NSQ.rewhite_flag = 1
				NSQ.sLTP_buf_idx = psEncC.frame_length
			}
		}

		nsq_scale_states(NSQ, x, x_sc_Q10, psEncC.subfr_length, sLTP,
			sLTP_Q16, k, LTP_scale_Q14, Gains_Q16, psEncCtrlC.pitchL)

		noise_shape_quantizer(NSQ, psEncCtrlC.sigtype, x_sc_Q10, q, pxq, sLTP_Q16, A_Q12, B_Q14,
			AR_shp_Q13, lag, HarmShapeFIRPacked_Q14, Tilt_Q14.idx(int(k)), LF_shp_Q14.idx(int(k)), Gains_Q16.idx(int(k)), Lambda_Q10,
			offset_Q10, psEncC.subfr_length, psEncC.shapingLPCOrder, psEncC.predictLPCOrder)

		x = x.off(int(psEncC.subfr_length))
		q = q.off(int(psEncC.subfr_length))
		pxq = pxq.off(int(psEncC.subfr_length))
	}

	NSQ.lagPrev = psEncCtrlC.pitchL.idx(NB_SUBFR - 1)

	NSQ.xq.copy(NSQ.xq, int(psEncC.frame_length))
	NSQ.sLTP_shp_Q10.copy(NSQ.sLTP_shp_Q10, int(psEncC.frame_length))
}

func noise_shape_quantizer(
	NSQ *nsq_state,
	sigtype int32,
	x_sc_Q10 *slice[int32],
	q *slice[int8],
	xq *slice[int16],
	sLTP_Q16 *slice[int32],
	a_Q12 *slice[int16],
	b_Q14 *slice[int16],
	AR_shp_Q13 *slice[int16],
	lag, HarmShapeFIRPacked_Q14, Tilt_Q14, LF_shp_Q14, Gain_Q16, Lambda_Q10 int32,
	offset_Q10, length, shapingLPCOrder, predictLPCOrder int32) {
	var (
		i, j                                            int32
		LTP_pred_Q14, LPC_pred_Q10, n_AR_Q10, n_LTP_Q14 int32
		n_LF_Q10, r_Q10, q_Q0, q_Q10                    int32
		thr1_Q10, thr2_Q10, thr3_Q10                    int32
		dither, exc_Q10, LPC_exc_Q10, xq_Q10            int32
		tmp1, tmp2, sLF_AR_shp_Q10                      int32
		psLPC_Q14, shp_lag_ptr, pred_lag_ptr            *slice[int32]
		a_Q12_tmp                                       = alloc[int32](MAX_LPC_ORDER / 2)
		Atmp                                            int32
	)

	a_Q12.copy(slice2[int16](a_Q12_tmp), int(predictLPCOrder))

	shp_lag_ptr = NSQ.sLTP_shp_Q10.off(int(NSQ.sLTP_shp_buf_idx - lag + HARM_SHAPE_FIR_TAPS/2))
	pred_lag_ptr = sLTP_Q16.off(int(NSQ.sLTP_buf_idx - lag + LTP_ORDER/2))

	psLPC_Q14 = NSQ.sLPC_Q14.off(NSQ_LPC_BUF_LENGTH - 1)

	thr1_Q10 = SUB_RSHIFT32(-1536, Lambda_Q10, 1)
	thr2_Q10 = SUB_RSHIFT32(-512, Lambda_Q10, 1)
	thr2_Q10 = ADD_RSHIFT32(thr2_Q10, SMULBB(offset_Q10, Lambda_Q10), 10)
	thr3_Q10 = ADD_RSHIFT32(512, Lambda_Q10, 1)

	for i = 0; i < length; i++ {
		NSQ.rand_seed = RAND(NSQ.rand_seed)

		dither = RSHIFT(NSQ.rand_seed, 31)

		Atmp = a_Q12_tmp.idx(0)
		LPC_pred_Q10 = SMULWB(psLPC_Q14.idx(0), Atmp)
		LPC_pred_Q10 = SMLAWT(LPC_pred_Q10, psLPC_Q14.idx(-1), Atmp)
		Atmp = a_Q12_tmp.idx(1)
		LPC_pred_Q10 = SMLAWB(LPC_pred_Q10, psLPC_Q14.idx(-2), Atmp)
		LPC_pred_Q10 = SMLAWT(LPC_pred_Q10, psLPC_Q14.idx(-3), Atmp)
		Atmp = a_Q12_tmp.idx(2)
		LPC_pred_Q10 = SMLAWB(LPC_pred_Q10, psLPC_Q14.idx(-4), Atmp)
		LPC_pred_Q10 = SMLAWT(LPC_pred_Q10, psLPC_Q14.idx(-5), Atmp)
		Atmp = a_Q12_tmp.idx(3)
		LPC_pred_Q10 = SMLAWB(LPC_pred_Q10, psLPC_Q14.idx(-6), Atmp)
		LPC_pred_Q10 = SMLAWT(LPC_pred_Q10, psLPC_Q14.idx(-7), Atmp)
		Atmp = a_Q12_tmp.idx(4)
		LPC_pred_Q10 = SMLAWB(LPC_pred_Q10, psLPC_Q14.idx(-8), Atmp)
		LPC_pred_Q10 = SMLAWT(LPC_pred_Q10, psLPC_Q14.idx(-9), Atmp)
		for j = 10; j < predictLPCOrder; j += 2 {
			Atmp = a_Q12_tmp.idx(int(j >> 1))
			LPC_pred_Q10 = SMLAWB(LPC_pred_Q10, psLPC_Q14.idx(int(-j)), Atmp)
			LPC_pred_Q10 = SMLAWT(LPC_pred_Q10, psLPC_Q14.idx(int(-j-1)), Atmp)
		}

		if sigtype == SIG_TYPE_VOICED {
			LTP_pred_Q14 = SMULWB(pred_lag_ptr.idx(0), int32(b_Q14.idx(0)))
			LTP_pred_Q14 = SMLAWB(LTP_pred_Q14, pred_lag_ptr.idx(-1), int32(b_Q14.idx(1)))
			LTP_pred_Q14 = SMLAWB(LTP_pred_Q14, pred_lag_ptr.idx(-2), int32(b_Q14.idx(2)))
			LTP_pred_Q14 = SMLAWB(LTP_pred_Q14, pred_lag_ptr.idx(-3), int32(b_Q14.idx(3)))
			LTP_pred_Q14 = SMLAWB(LTP_pred_Q14, pred_lag_ptr.idx(-4), int32(b_Q14.idx(4)))
			pred_lag_ptr = pred_lag_ptr.off(1)
		} else {
			LTP_pred_Q14 = 0
		}

		tmp2 = psLPC_Q14.idx(0)
		tmp1 = NSQ.sAR2_Q14.idx(0)
		*NSQ.sAR2_Q14.ptr(0) = tmp2
		n_AR_Q10 = SMULWB(tmp2, int32(AR_shp_Q13.idx(0)))
		for j = 2; j < shapingLPCOrder; j += 2 {
			tmp2 = NSQ.sAR2_Q14.idx(int(j - 1))
			*NSQ.sAR2_Q14.ptr(int(j - 1)) = tmp1
			n_AR_Q10 = SMLAWB(n_AR_Q10, tmp1, int32(AR_shp_Q13.idx(int(j-1))))
			tmp1 = NSQ.sAR2_Q14.idx(int(j + 0))
			*NSQ.sAR2_Q14.ptr(int(j + 0)) = tmp2
			n_AR_Q10 = SMLAWB(n_AR_Q10, tmp2, int32(AR_shp_Q13.idx(int(j))))
		}
		*NSQ.sAR2_Q14.ptr(int(shapingLPCOrder - 1)) = tmp1
		n_AR_Q10 = SMLAWB(n_AR_Q10, tmp1, int32(AR_shp_Q13.idx(int(shapingLPCOrder-1))))

		n_AR_Q10 = RSHIFT(n_AR_Q10, 1)
		n_AR_Q10 = SMLAWB(n_AR_Q10, NSQ.sLF_AR_shp_Q12, Tilt_Q14)

		n_LF_Q10 = LSHIFT(SMULWB(NSQ.sLTP_shp_Q10.idx(int(NSQ.sLTP_shp_buf_idx-1)), LF_shp_Q14), 2)
		n_LF_Q10 = SMLAWT(n_LF_Q10, NSQ.sLF_AR_shp_Q12, LF_shp_Q14)

		if lag > 0 {
			n_LTP_Q14 = SMULWB(ADD32(shp_lag_ptr.idx(0), shp_lag_ptr.idx(-2)), HarmShapeFIRPacked_Q14)
			n_LTP_Q14 = SMLAWT(n_LTP_Q14, shp_lag_ptr.idx(-1), HarmShapeFIRPacked_Q14)
			n_LTP_Q14 = LSHIFT(n_LTP_Q14, 6)
			shp_lag_ptr = shp_lag_ptr.off(1)
		} else {
			n_LTP_Q14 = 0
		}

		tmp1 = SUB32(LTP_pred_Q14, n_LTP_Q14)
		tmp1 = RSHIFT(tmp1, 4)
		tmp1 = ADD32(tmp1, LPC_pred_Q10)
		tmp1 = SUB32(tmp1, n_AR_Q10)
		tmp1 = SUB32(tmp1, n_LF_Q10)
		r_Q10 = SUB32(x_sc_Q10.idx(int(i)), tmp1)

		r_Q10 = (r_Q10 ^ dither) - dither
		r_Q10 = SUB32(r_Q10, offset_Q10)
		r_Q10 = LIMIT_32(r_Q10, -64<<10, 64<<10)

		q_Q0 = 0
		q_Q10 = 0
		if r_Q10 < thr2_Q10 {
			if r_Q10 < thr1_Q10 {
				q_Q0 = RSHIFT_ROUND(ADD_RSHIFT32(r_Q10, Lambda_Q10, 1), 10)
				q_Q10 = LSHIFT(q_Q0, 10)
			} else {
				q_Q0 = -1
				q_Q10 = -1024
			}
		} else {
			if r_Q10 > thr3_Q10 {
				q_Q0 = RSHIFT_ROUND(SUB_RSHIFT32(r_Q10, Lambda_Q10, 1), 10)
				q_Q10 = LSHIFT(q_Q0, 10)
			}
		}
		*q.ptr(int(i)) = int8(q_Q0)

		exc_Q10 = ADD32(q_Q10, offset_Q10)
		exc_Q10 = (exc_Q10 ^ dither) - dither

		LPC_exc_Q10 = ADD32(exc_Q10, RSHIFT_ROUND(LTP_pred_Q14, 4))
		xq_Q10 = ADD32(LPC_exc_Q10, LPC_pred_Q10)

		*xq.ptr(int(i)) = int16(SAT16(RSHIFT_ROUND(SMULWW(xq_Q10, Gain_Q16), 10)))

		psLPC_Q14 = psLPC_Q14.off(1)
		*psLPC_Q14.ptr(0) = LSHIFT(xq_Q10, 4)
		sLF_AR_shp_Q10 = SUB32(xq_Q10, n_AR_Q10)
		NSQ.sLF_AR_shp_Q12 = LSHIFT(sLF_AR_shp_Q10, 2)

		*NSQ.sLTP_shp_Q10.ptr(int(NSQ.sLTP_shp_buf_idx)) = SUB32(sLF_AR_shp_Q10, n_LF_Q10)
		*sLTP_Q16.ptr(int(NSQ.sLTP_buf_idx)) = LSHIFT(LPC_exc_Q10, 6)
		NSQ.sLTP_shp_buf_idx++
		NSQ.sLTP_buf_idx++

		NSQ.rand_seed += int32(q.idx(int(i)))
	}

	NSQ.sLPC_Q14.off(int(length)).copy(NSQ.sLPC_Q14, NSQ_LPC_BUF_LENGTH)
}

func nsq_scale_states(
	NSQ *nsq_state,
	x *slice[int16],
	x_sc_Q10 *slice[int32],
	subfr_length int32,
	sLTP *slice[int16],
	sLTP_Q16 *slice[int32],
	subfr int32,
	LTP_scale_Q14 int32,
	Gains_Q16 *slice[int32],
	pitchL *slice[int32]) {
	var (
		i, lag                                   int32
		inv_gain_Q16, gain_adj_Q16, inv_gain_Q32 int32
	)

	inv_gain_Q16 = INVERSE32_varQ(max(Gains_Q16.idx(int(subfr)), 1), 32)
	inv_gain_Q16 = min(inv_gain_Q16, math.MaxInt16)
	lag = pitchL.idx(int(subfr))

	if NSQ.rewhite_flag != 0 {
		inv_gain_Q32 = LSHIFT(inv_gain_Q16, 16)
		if subfr == 0 {
			inv_gain_Q32 = LSHIFT(SMULWB(inv_gain_Q32, LTP_scale_Q14), 2)
		}
		for i = NSQ.sLTP_buf_idx - lag - LTP_ORDER/2; i < NSQ.sLTP_buf_idx; i++ {
			*sLTP_Q16.ptr(int(i)) = SMULWB(inv_gain_Q32, int32(sLTP.idx(int(i))))
		}
	}

	if inv_gain_Q16 != NSQ.prev_inv_gain_Q16 {
		gain_adj_Q16 = DIV32_varQ(inv_gain_Q16, NSQ.prev_inv_gain_Q16, 16)

		for i = NSQ.sLTP_shp_buf_idx - subfr_length*NB_SUBFR; i < NSQ.sLTP_shp_buf_idx; i++ {
			*NSQ.sLTP_shp_Q10.ptr(int(i)) = SMULWW(gain_adj_Q16, NSQ.sLTP_shp_Q10.idx(int(i)))
		}

		if NSQ.rewhite_flag == 0 {
			for i = NSQ.sLTP_buf_idx - lag - LTP_ORDER/2; i < NSQ.sLTP_buf_idx; i++ {
				*sLTP_Q16.ptr(int(i)) = SMULWW(gain_adj_Q16, sLTP_Q16.idx(int(i)))
			}
		}

		NSQ.sLF_AR_shp_Q12 = SMULWW(gain_adj_Q16, NSQ.sLF_AR_shp_Q12)

		for i = 0; i < NSQ_LPC_BUF_LENGTH; i++ {
			*NSQ.sLPC_Q14.ptr(int(i)) = SMULWW(gain_adj_Q16, NSQ.sLPC_Q14.idx(int(i)))
		}
		for i = 0; i < MAX_SHAPE_LPC_ORDER; i++ {
			*NSQ.sAR2_Q14.ptr(int(i)) = SMULWW(gain_adj_Q16, NSQ.sAR2_Q14.idx(int(i)))
		}
	}

	for i = 0; i < subfr_length; i++ {
		*x_sc_Q10.ptr(int(i)) = RSHIFT(SMULBB(int32(x.idx(int(i))), inv_gain_Q16), 6)
	}

	NSQ.prev_inv_gain_Q16 = inv_gain_Q16
}

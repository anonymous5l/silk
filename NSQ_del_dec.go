package silk

import (
	"math"
)

type NSQ_del_dec_struct struct {
	RandState *slice[int32]
	Q_Q10     *slice[int32]
	Xq_Q10    *slice[int32]
	Pred_Q16  *slice[int32]
	Shape_Q10 *slice[int32]
	Gain_Q16  *slice[int32]
	sAR2_Q14  *slice[int32]
	sLPC_Q14  *slice[int32]
	LF_AR_Q12 int32
	Seed      int32
	SeedInit  int32
	RD_Q10    int32
}

func (n *NSQ_del_dec_struct) init() {
	n.RandState = alloc[int32](DECISION_DELAY)
	n.Q_Q10 = alloc[int32](DECISION_DELAY)
	n.Xq_Q10 = alloc[int32](DECISION_DELAY)
	n.Pred_Q16 = alloc[int32](DECISION_DELAY)
	n.Shape_Q10 = alloc[int32](DECISION_DELAY)
	n.Gain_Q16 = alloc[int32](DECISION_DELAY)
	n.sAR2_Q14 = alloc[int32](MAX_SHAPE_LPC_ORDER)
	n.sLPC_Q14 = alloc[int32](MAX_FRAME_LENGTH/NB_SUBFR + NSQ_LPC_BUF_LENGTH)
}

type NSQ_sample_struct struct {
	Q_Q10        int32
	RD_Q10       int32
	xq_Q14       int32
	LF_AR_Q12    int32
	sLTP_shp_Q10 int32
	LPC_exc_Q16  int32
}

func (n *NSQ_sample_struct) copy(dest *NSQ_sample_struct) {
	dest.Q_Q10 = n.Q_Q10
	dest.RD_Q10 = n.RD_Q10
	dest.xq_Q14 = n.xq_Q14
	dest.LF_AR_Q12 = n.LF_AR_Q12
	dest.sLTP_shp_Q10 = n.sLTP_shp_Q10
	dest.LPC_exc_Q16 = n.LPC_exc_Q16
}

func NSQ_del_dec(
	psEncC *encoder_state,
	psEncCtrlC *encoder_control,
	NSQ *nsq_state,
	x *slice[int16],
	q *slice[int8],
	LSFInterpFactor_Q2 int32,
	PredCoef_Q12 [2]*slice[int16], LTPCoef_Q14, AR2_Q13 *slice[int16],
	HarmShapeGain_Q14 *slice[int32],
	Tilt_Q14 *slice[int32],
	LF_shp_Q14 *slice[int32],
	Gains_Q16 *slice[int32],
	Lambda_Q10, LTP_scale_Q14 int32) {
	var (
		i, k, lag, start_idx, LSF_interpolation_flag, Winner_ind, subfr int32
		last_smple_idx, smpl_buf_idx, decisionDelay, subfr_length       int32
		A_Q12, B_Q14, AR_shp_Q13                                        *slice[int16]
		pxq                                                             *slice[int16]
		sLTP_Q16                                                        = alloc[int32](2 * MAX_FRAME_LENGTH)
		sLTP                                                            = alloc[int16](2 * MAX_FRAME_LENGTH)
		HarmShapeFIRPacked_Q14, offset_Q10                              int32
		FiltState                                                       = alloc[int32](MAX_LPC_ORDER)
		RDmin_Q10                                                       int32
		x_sc_Q10                                                        = alloc[int32](MAX_FRAME_LENGTH / NB_SUBFR)
		psDelDec                                                        [MAX_DEL_DEC_STATES]NSQ_del_dec_struct
		psDD                                                            *NSQ_del_dec_struct
	)

	for p := 0; p < MAX_DEL_DEC_STATES; p++ {
		psDelDec[p].init()
	}

	subfr_length = psEncC.frame_length / NB_SUBFR

	lag = NSQ.lagPrev

	for k = 0; k < psEncC.nStatesDelayedDecision; k++ {
		psDD = &psDelDec[k]
		psDD.Seed = (k + psEncCtrlC.Seed) & 3
		psDD.SeedInit = psDD.Seed
		psDD.RD_Q10 = 0
		psDD.LF_AR_Q12 = NSQ.sLF_AR_shp_Q12
		*psDD.Shape_Q10.ptr(0) = NSQ.sLTP_shp_Q10.idx(int(psEncC.frame_length - 1))
		NSQ.sLPC_Q14.copy(psDD.sLPC_Q14, NSQ_LPC_BUF_LENGTH)
		NSQ.sAR2_Q14.copy(psDD.sAR2_Q14, MAX_SHAPE_LPC_ORDER)
	}

	offset_Q10 = int32(Quantization_Offsets_Q10[psEncCtrlC.sigtype][psEncCtrlC.QuantOffsetType])
	smpl_buf_idx = 0

	decisionDelay = min(DECISION_DELAY, subfr_length)

	if psEncCtrlC.sigtype == SIG_TYPE_VOICED {
		for k = 0; k < NB_SUBFR; k++ {
			decisionDelay = min(decisionDelay, psEncCtrlC.pitchL.idx(int(k))-LTP_ORDER/2-1)
		}
	} else {
		if lag > 0 {
			decisionDelay = min(decisionDelay, lag-LTP_ORDER/2-1)
		}
	}

	if LSFInterpFactor_Q2 == (1 << 2) {
		LSF_interpolation_flag = 0
	} else {
		LSF_interpolation_flag = 1
	}

	pxq = NSQ.xq.off(int(psEncC.frame_length))
	NSQ.sLTP_shp_buf_idx = psEncC.frame_length
	NSQ.sLTP_buf_idx = psEncC.frame_length
	subfr = 0
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
				if k == 2 {
					RDmin_Q10 = psDelDec[0].RD_Q10
					Winner_ind = 0
					for i = 1; i < psEncC.nStatesDelayedDecision; i++ {
						if psDelDec[i].RD_Q10 < RDmin_Q10 {
							RDmin_Q10 = psDelDec[i].RD_Q10
							Winner_ind = i
						}
					}
					for i = 0; i < psEncC.nStatesDelayedDecision; i++ {
						if i != Winner_ind {
							psDelDec[i].RD_Q10 += math.MaxInt32 >> 4
						}
					}

					psDD = &psDelDec[Winner_ind]
					last_smple_idx = smpl_buf_idx + decisionDelay
					for i = 0; i < decisionDelay; i++ {
						last_smple_idx = (last_smple_idx - 1) & DECISION_DELAY_MASK
						*q.ptr(int(i - decisionDelay)) = int8(RSHIFT(psDD.Q_Q10.idx(int(last_smple_idx)), 10))
						*pxq.ptr(int(i - decisionDelay)) = int16(SAT16(RSHIFT_ROUND(
							SMULWW(psDD.Xq_Q10.idx(int(last_smple_idx)),
								psDD.Gain_Q16.idx(int(last_smple_idx))), 10)))
						*NSQ.sLTP_shp_Q10.ptr(int(NSQ.sLTP_shp_buf_idx - decisionDelay + i)) = psDD.Shape_Q10.idx(int(last_smple_idx))
					}

					subfr = 0
				}

				start_idx = psEncC.frame_length - lag - psEncC.predictLPCOrder - LTP_ORDER/2

				memset(FiltState, 0, int(psEncC.predictLPCOrder))
				MA_Prediction(NSQ.xq.off(int(start_idx+k*psEncC.subfr_length)),
					A_Q12, FiltState, sLTP.off(int(start_idx)), psEncC.frame_length-start_idx, psEncC.predictLPCOrder)

				NSQ.sLTP_buf_idx = psEncC.frame_length
				NSQ.rewhite_flag = 1
			}
		}

		nsq_del_dec_scale_states(NSQ, psDelDec[:], x, x_sc_Q10,
			subfr_length, sLTP, sLTP_Q16, k, psEncC.nStatesDelayedDecision, smpl_buf_idx,
			LTP_scale_Q14, Gains_Q16, psEncCtrlC.pitchL)

		noise_shape_quantizer_del_dec(NSQ, psDelDec[:], psEncCtrlC.sigtype, x_sc_Q10, q, pxq, sLTP_Q16,
			A_Q12, B_Q14, AR_shp_Q13, lag, HarmShapeFIRPacked_Q14, Tilt_Q14.idx(int(k)), LF_shp_Q14.idx(int(k)), Gains_Q16.idx(int(k)),
			Lambda_Q10, offset_Q10, psEncC.subfr_length, subfr, psEncC.shapingLPCOrder, psEncC.predictLPCOrder,
			psEncC.warping_Q16, psEncC.nStatesDelayedDecision, &smpl_buf_idx, decisionDelay)
		subfr++

		x = x.off(int(psEncC.subfr_length))
		q = q.off(int(psEncC.subfr_length))
		pxq = pxq.off(int(psEncC.subfr_length))
	}

	RDmin_Q10 = psDelDec[0].RD_Q10
	Winner_ind = 0
	for k = 1; k < psEncC.nStatesDelayedDecision; k++ {
		if psDelDec[k].RD_Q10 < RDmin_Q10 {
			RDmin_Q10 = psDelDec[k].RD_Q10
			Winner_ind = k
		}
	}

	psDD = &psDelDec[Winner_ind]
	psEncCtrlC.Seed = psDD.SeedInit
	last_smple_idx = smpl_buf_idx + decisionDelay
	for i = 0; i < decisionDelay; i++ {
		last_smple_idx = (last_smple_idx - 1) & DECISION_DELAY_MASK
		*q.ptr(int(i - decisionDelay)) = int8(RSHIFT(psDD.Q_Q10.idx(int(last_smple_idx)), 10))
		*pxq.ptr(int(i - decisionDelay)) = int16(SAT16(RSHIFT_ROUND(
			SMULWW(psDD.Xq_Q10.idx(int(last_smple_idx)), psDD.Gain_Q16.idx(int(last_smple_idx))), 10)))
		*NSQ.sLTP_shp_Q10.ptr(int(NSQ.sLTP_shp_buf_idx - decisionDelay + i)) = psDD.Shape_Q10.idx(int(last_smple_idx))
		*sLTP_Q16.ptr(int(NSQ.sLTP_buf_idx - decisionDelay + i)) = psDD.Pred_Q16.idx(int(last_smple_idx))
	}

	psDD.sLPC_Q14.off(int(psEncC.subfr_length)).copy(NSQ.sLPC_Q14, NSQ_LPC_BUF_LENGTH)
	psDD.sAR2_Q14.copy(NSQ.sAR2_Q14, MAX_SHAPE_LPC_ORDER)

	NSQ.sLF_AR_shp_Q12 = psDD.LF_AR_Q12
	NSQ.lagPrev = psEncCtrlC.pitchL.idx(NB_SUBFR - 1)

	NSQ.xq.off(int(psEncC.frame_length)).copy(NSQ.xq, int(psEncC.frame_length))
	NSQ.sLTP_shp_Q10.off(int(psEncC.frame_length)).copy(NSQ.sLTP_shp_Q10, int(psEncC.frame_length))
}

func noise_shape_quantizer_del_dec(
	NSQ *nsq_state,
	psDelDec []NSQ_del_dec_struct,
	sigtype int32,
	x_Q10 *slice[int32],
	q *slice[int8],
	xq *slice[int16],
	sLTP_Q16 *slice[int32],
	a_Q12 *slice[int16],
	b_Q14 *slice[int16],
	AR_shp_Q13 *slice[int16],
	lag int32,
	HarmShapeFIRPacked_Q14 int32,
	Tilt_Q14 int32,
	LF_shp_Q14 int32,
	Gain_Q16 int32,
	Lambda_Q10 int32,
	offset_Q10 int32,
	length int32,
	subfr int32,
	shapingLPCOrder int32,
	predictLPCOrder int32,
	warping_Q16 int32,
	nStatesDelayedDecision int32,
	smpl_buf_idx *int32,
	decisionDelay int32) {
	var (
		i, j, k, Winner_ind, RDmin_ind, RDmax_ind, last_smple_idx       int32
		Winner_rand_state                                               int32
		LTP_pred_Q14, LPC_pred_Q10, n_AR_Q10, n_LTP_Q14                 int32
		n_LF_Q10, r_Q10, rr_Q20, rd1_Q10, rd2_Q10, RDmin_Q10, RDmax_Q10 int32
		q1_Q10, q2_Q10, dither, exc_Q10, LPC_exc_Q10, xq_Q10            int32
		tmp1, tmp2, sLF_AR_shp_Q10                                      int32
		pred_lag_ptr, shp_lag_ptr, psLPC_Q14                            *slice[int32]
		psSampleState                                                   [MAX_DEL_DEC_STATES][2]NSQ_sample_struct
		psDD                                                            *NSQ_del_dec_struct
		psSS                                                            []NSQ_sample_struct
		a_Q12_tmp                                                       = alloc[int32](MAX_LPC_ORDER / 2)
		Atmp                                                            int32
	)

	a_Q12.copy(slice2[int16](a_Q12_tmp), int(predictLPCOrder))

	shp_lag_ptr = NSQ.sLTP_shp_Q10.off(int(NSQ.sLTP_shp_buf_idx - lag + HARM_SHAPE_FIR_TAPS/2))
	pred_lag_ptr = sLTP_Q16.off(int(NSQ.sLTP_buf_idx - lag + LTP_ORDER/2))

	for i = 0; i < length; i++ {

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

		if lag > 0 {
			n_LTP_Q14 = SMULWB(ADD32(shp_lag_ptr.idx(0), shp_lag_ptr.idx(-2)), HarmShapeFIRPacked_Q14)
			n_LTP_Q14 = SMLAWT(n_LTP_Q14, shp_lag_ptr.idx(-1), HarmShapeFIRPacked_Q14)
			n_LTP_Q14 = LSHIFT(n_LTP_Q14, 6)
			shp_lag_ptr = shp_lag_ptr.off(1)
		} else {
			n_LTP_Q14 = 0
		}

		for k = 0; k < nStatesDelayedDecision; k++ {
			psDD = &psDelDec[k]

			psSS = psSampleState[k][:]

			psDD.Seed = RAND(psDD.Seed)

			dither = RSHIFT(psDD.Seed, 31)

			psLPC_Q14 = psDD.sLPC_Q14.off(int(NSQ_LPC_BUF_LENGTH - 1 + i))

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

			tmp2 = SMLAWB(psLPC_Q14.idx(0), psDD.sAR2_Q14.idx(0), warping_Q16)
			tmp1 = SMLAWB(psDD.sAR2_Q14.idx(0), psDD.sAR2_Q14.idx(1)-tmp2, warping_Q16)
			*psDD.sAR2_Q14.ptr(0) = tmp2
			n_AR_Q10 = SMULWB(tmp2, int32(AR_shp_Q13.idx(0)))
			for j = 2; j < shapingLPCOrder; j += 2 {
				tmp2 = SMLAWB(psDD.sAR2_Q14.idx(int(j-1)), psDD.sAR2_Q14.idx(int(j+0))-tmp1, warping_Q16)
				*psDD.sAR2_Q14.ptr(int(j - 1)) = tmp1
				n_AR_Q10 = SMLAWB(n_AR_Q10, tmp1, int32(AR_shp_Q13.idx(int(j-1))))
				tmp1 = SMLAWB(psDD.sAR2_Q14.idx(int(j+0)), psDD.sAR2_Q14.idx(int(j+1))-tmp2, warping_Q16)
				*psDD.sAR2_Q14.ptr(int(j + 0)) = tmp2
				n_AR_Q10 = SMLAWB(n_AR_Q10, tmp2, int32(AR_shp_Q13.idx(int(j))))
			}
			*psDD.sAR2_Q14.ptr(int(shapingLPCOrder - 1)) = tmp1
			n_AR_Q10 = SMLAWB(n_AR_Q10, tmp1, int32(AR_shp_Q13.idx(int(shapingLPCOrder-1))))

			n_AR_Q10 = RSHIFT(n_AR_Q10, 1)
			n_AR_Q10 = SMLAWB(n_AR_Q10, psDD.LF_AR_Q12, Tilt_Q14)

			n_LF_Q10 = LSHIFT(SMULWB(psDD.Shape_Q10.idx(int(*smpl_buf_idx)), LF_shp_Q14), 2)
			n_LF_Q10 = SMLAWT(n_LF_Q10, psDD.LF_AR_Q12, LF_shp_Q14)

			tmp1 = SUB32(LTP_pred_Q14, n_LTP_Q14)
			tmp1 = RSHIFT(tmp1, 4)
			tmp1 = ADD32(tmp1, LPC_pred_Q10)
			tmp1 = SUB32(tmp1, n_AR_Q10)
			tmp1 = SUB32(tmp1, n_LF_Q10)
			r_Q10 = SUB32(x_Q10.idx(int(i)), tmp1)

			r_Q10 = (r_Q10 ^ dither) - dither
			r_Q10 = SUB32(r_Q10, offset_Q10)
			r_Q10 = LIMIT_32(r_Q10, -64<<10, 64<<10)

			if r_Q10 < -1536 {
				q1_Q10 = LSHIFT(RSHIFT_ROUND(r_Q10, 10), 10)
				r_Q10 = SUB32(r_Q10, q1_Q10)
				rd1_Q10 = RSHIFT(SMLABB(MUL(-ADD32(q1_Q10, offset_Q10), Lambda_Q10), r_Q10, r_Q10), 10)
				rd2_Q10 = ADD32(rd1_Q10, 1024)
				rd2_Q10 = SUB32(rd2_Q10, ADD_LSHIFT32(Lambda_Q10, r_Q10, 1))
				q2_Q10 = ADD32(q1_Q10, 1024)
			} else if r_Q10 > 512 {
				q1_Q10 = LSHIFT(RSHIFT_ROUND(r_Q10, 10), 10)
				r_Q10 = SUB32(r_Q10, q1_Q10)
				rd1_Q10 = RSHIFT(SMLABB(MUL(ADD32(q1_Q10, offset_Q10), Lambda_Q10), r_Q10, r_Q10), 10)
				rd2_Q10 = ADD32(rd1_Q10, 1024)
				rd2_Q10 = SUB32(rd2_Q10, SUB_LSHIFT32(Lambda_Q10, r_Q10, 1))
				q2_Q10 = SUB32(q1_Q10, 1024)
			} else {
				rr_Q20 = SMULBB(offset_Q10, Lambda_Q10)
				rd2_Q10 = RSHIFT(SMLABB(rr_Q20, r_Q10, r_Q10), 10)
				rd1_Q10 = ADD32(rd2_Q10, 1024)
				rd1_Q10 = ADD32(rd1_Q10, SUB_RSHIFT32(ADD_LSHIFT32(Lambda_Q10, r_Q10, 1), rr_Q20, 9))
				q1_Q10 = -1024
				q2_Q10 = 0
			}

			if rd1_Q10 < rd2_Q10 {
				psSS[0].RD_Q10 = ADD32(psDD.RD_Q10, rd1_Q10)
				psSS[1].RD_Q10 = ADD32(psDD.RD_Q10, rd2_Q10)
				psSS[0].Q_Q10 = q1_Q10
				psSS[1].Q_Q10 = q2_Q10
			} else {
				psSS[0].RD_Q10 = ADD32(psDD.RD_Q10, rd2_Q10)
				psSS[1].RD_Q10 = ADD32(psDD.RD_Q10, rd1_Q10)
				psSS[0].Q_Q10 = q2_Q10
				psSS[1].Q_Q10 = q1_Q10
			}

			exc_Q10 = ADD32(offset_Q10, psSS[0].Q_Q10)
			exc_Q10 = (exc_Q10 ^ dither) - dither

			LPC_exc_Q10 = exc_Q10 + RSHIFT_ROUND(LTP_pred_Q14, 4)
			xq_Q10 = ADD32(LPC_exc_Q10, LPC_pred_Q10)

			sLF_AR_shp_Q10 = SUB32(xq_Q10, n_AR_Q10)
			psSS[0].sLTP_shp_Q10 = SUB32(sLF_AR_shp_Q10, n_LF_Q10)
			psSS[0].LF_AR_Q12 = LSHIFT(sLF_AR_shp_Q10, 2)
			psSS[0].xq_Q14 = LSHIFT(xq_Q10, 4)
			psSS[0].LPC_exc_Q16 = LSHIFT(LPC_exc_Q10, 6)

			exc_Q10 = ADD32(offset_Q10, psSS[1].Q_Q10)
			exc_Q10 = (exc_Q10 ^ dither) - dither

			LPC_exc_Q10 = exc_Q10 + RSHIFT_ROUND(LTP_pred_Q14, 4)
			xq_Q10 = ADD32(LPC_exc_Q10, LPC_pred_Q10)

			sLF_AR_shp_Q10 = SUB32(xq_Q10, n_AR_Q10)
			psSS[1].sLTP_shp_Q10 = SUB32(sLF_AR_shp_Q10, n_LF_Q10)
			psSS[1].LF_AR_Q12 = LSHIFT(sLF_AR_shp_Q10, 2)
			psSS[1].xq_Q14 = LSHIFT(xq_Q10, 4)
			psSS[1].LPC_exc_Q16 = LSHIFT(LPC_exc_Q10, 6)
		}

		*smpl_buf_idx = (*smpl_buf_idx - 1) & DECISION_DELAY_MASK
		last_smple_idx = (*smpl_buf_idx + decisionDelay) & DECISION_DELAY_MASK

		RDmin_Q10 = psSampleState[0][0].RD_Q10
		Winner_ind = 0
		for k = 1; k < nStatesDelayedDecision; k++ {
			if psSampleState[k][0].RD_Q10 < RDmin_Q10 {
				RDmin_Q10 = psSampleState[k][0].RD_Q10
				Winner_ind = k
			}
		}

		Winner_rand_state = psDelDec[Winner_ind].RandState.idx(int(last_smple_idx))
		for k = 0; k < nStatesDelayedDecision; k++ {
			if psDelDec[k].RandState.idx(int(last_smple_idx)) != Winner_rand_state {
				psSampleState[k][0].RD_Q10 = ADD32(psSampleState[k][0].RD_Q10, math.MaxInt32>>4)
				psSampleState[k][1].RD_Q10 = ADD32(psSampleState[k][1].RD_Q10, math.MaxInt32>>4)
			}
		}

		RDmax_Q10 = psSampleState[0][0].RD_Q10
		RDmin_Q10 = psSampleState[0][1].RD_Q10
		RDmax_ind = 0
		RDmin_ind = 0
		for k = 1; k < nStatesDelayedDecision; k++ {
			if psSampleState[k][0].RD_Q10 > RDmax_Q10 {
				RDmax_Q10 = psSampleState[k][0].RD_Q10
				RDmax_ind = k
			}
			if psSampleState[k][1].RD_Q10 < RDmin_Q10 {
				RDmin_Q10 = psSampleState[k][1].RD_Q10
				RDmin_ind = k
			}
		}

		if RDmin_Q10 < RDmax_Q10 {
			copy_del_dec_state(&psDelDec[RDmax_ind], &psDelDec[RDmin_ind], i)
			psSampleState[RDmin_ind][1].copy(&psSampleState[RDmax_ind][0])
		}

		psDD = &psDelDec[Winner_ind]
		if subfr > 0 || i >= decisionDelay {
			*q.ptr(int(i - decisionDelay)) = int8(RSHIFT(psDD.Q_Q10.idx(int(last_smple_idx)), 10))
			*xq.ptr(int(i - decisionDelay)) = int16(SAT16(RSHIFT_ROUND(
				SMULWW(psDD.Xq_Q10.idx(int(last_smple_idx)), psDD.Gain_Q16.idx(int(last_smple_idx))), 10)))
			*NSQ.sLTP_shp_Q10.ptr(int(NSQ.sLTP_shp_buf_idx - decisionDelay)) = psDD.Shape_Q10.idx(int(last_smple_idx))
			*sLTP_Q16.ptr(int(NSQ.sLTP_buf_idx - decisionDelay)) = psDD.Pred_Q16.idx(int(last_smple_idx))
		}
		NSQ.sLTP_shp_buf_idx++
		NSQ.sLTP_buf_idx++

		for k = 0; k < nStatesDelayedDecision; k++ {
			psDD = &psDelDec[k]
			psSS = psSampleState[k][:]
			psDD.LF_AR_Q12 = psSS[0].LF_AR_Q12
			*psDD.sLPC_Q14.ptr(int(NSQ_LPC_BUF_LENGTH + i)) = psSS[0].xq_Q14
			*psDD.Xq_Q10.ptr(int(*smpl_buf_idx)) = RSHIFT(psSS[0].xq_Q14, 4)
			*psDD.Q_Q10.ptr(int(*smpl_buf_idx)) = psSS[0].Q_Q10
			*psDD.Pred_Q16.ptr(int(*smpl_buf_idx)) = psSS[0].LPC_exc_Q16
			*psDD.Shape_Q10.ptr(int(*smpl_buf_idx)) = psSS[0].sLTP_shp_Q10
			psDD.Seed = ADD_RSHIFT32(psDD.Seed, psSS[0].Q_Q10, 10)
			*psDD.RandState.ptr(int(*smpl_buf_idx)) = psDD.Seed
			psDD.RD_Q10 = psSS[0].RD_Q10
			*psDD.Gain_Q16.ptr(int(*smpl_buf_idx)) = Gain_Q16
		}
	}
	for k = 0; k < nStatesDelayedDecision; k++ {
		psDD = &psDelDec[k]
		psDD.sLPC_Q14.off(int(length)).copy(psDD.sLPC_Q14, NSQ_LPC_BUF_LENGTH)
	}
}

func nsq_del_dec_scale_states(
	NSQ *nsq_state,
	psDelDec []NSQ_del_dec_struct,
	x *slice[int16],
	x_sc_Q10 *slice[int32],
	subfr_length int32,
	sLTP *slice[int16],
	sLTP_Q16 *slice[int32],
	subfr int32,
	nStatesDelayedDecision int32,
	smpl_buf_idx int32,
	LTP_scale_Q14 int32,
	Gains_Q16 *slice[int32],
	pitchL *slice[int32]) {
	var (
		i, k, lag                                int32
		inv_gain_Q16, gain_adj_Q16, inv_gain_Q32 int32
		psDD                                     *NSQ_del_dec_struct
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

		for k = 0; k < nStatesDelayedDecision; k++ {
			psDD = &psDelDec[k]

			psDD.LF_AR_Q12 = SMULWW(gain_adj_Q16, psDD.LF_AR_Q12)

			for i = 0; i < NSQ_LPC_BUF_LENGTH; i++ {
				*psDD.sLPC_Q14.ptr(int(i)) = SMULWW(gain_adj_Q16, psDD.sLPC_Q14.idx(int(i)))
			}
			for i = 0; i < MAX_SHAPE_LPC_ORDER; i++ {
				*psDD.sAR2_Q14.ptr(int(i)) = SMULWW(gain_adj_Q16, psDD.sAR2_Q14.idx(int(i)))
			}
			for i = 0; i < DECISION_DELAY; i++ {
				*psDD.Pred_Q16.ptr(int(i)) = SMULWW(gain_adj_Q16, psDD.Pred_Q16.idx(int(i)))
				*psDD.Shape_Q10.ptr(int(i)) = SMULWW(gain_adj_Q16, psDD.Shape_Q10.idx(int(i)))
			}
		}
	}

	for i = 0; i < subfr_length; i++ {
		*x_sc_Q10.ptr(int(i)) = RSHIFT(SMULBB(int32(x.idx(int(i))), inv_gain_Q16), 6)
	}

	NSQ.prev_inv_gain_Q16 = inv_gain_Q16
}

func copy_del_dec_state(DD_dst *NSQ_del_dec_struct, DD_src *NSQ_del_dec_struct, LPC_state_idx int32) {
	DD_src.RandState.copy(DD_dst.RandState, DECISION_DELAY)
	DD_src.Q_Q10.copy(DD_dst.Q_Q10, DECISION_DELAY)
	DD_src.Pred_Q16.copy(DD_dst.Pred_Q16, DECISION_DELAY)
	DD_src.Shape_Q10.copy(DD_dst.Shape_Q10, DECISION_DELAY)
	DD_src.Xq_Q10.copy(DD_dst.Xq_Q10, DECISION_DELAY)
	DD_src.sAR2_Q14.copy(DD_dst.sAR2_Q14, MAX_SHAPE_LPC_ORDER)
	DD_src.sLPC_Q14.off(int(LPC_state_idx)).copy(DD_dst.sLPC_Q14.off(int(LPC_state_idx)), NSQ_LPC_BUF_LENGTH)

	DD_dst.LF_AR_Q12 = DD_src.LF_AR_Q12
	DD_dst.Seed = DD_src.Seed
	DD_dst.SeedInit = DD_src.SeedInit
	DD_dst.RD_Q10 = DD_src.RD_Q10
}

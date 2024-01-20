package silk

func warped_LPC_analysis_filter_FIX(state *slice[int32], res *slice[int16], coef_Q13 *slice[int16],
	input *slice[int16], lambda_Q16 int16, length, order int32) {
	var (
		n, i                int32
		acc_Q11, tmp1, tmp2 int32
	)

	for n = 0; n < length; n++ {
		tmp2 = SMLAWB(state.idx(0), state.idx(1), int32(lambda_Q16))
		*state.ptr(0) = LSHIFT(int32(input.idx(int(n))), 14)
		tmp1 = SMLAWB(state.idx(1), state.idx(2)-tmp2, int32(lambda_Q16))
		*state.ptr(1) = tmp2
		acc_Q11 = SMULWB(tmp2, int32(coef_Q13.idx(0)))
		for i = 2; i < order; i += 2 {
			tmp2 = SMLAWB(state.idx(int(i)), state.idx(int(i+1))-tmp1, int32(lambda_Q16))
			*state.ptr(int(i)) = tmp1
			acc_Q11 = SMLAWB(acc_Q11, tmp1, int32(coef_Q13.idx(int(i-1))))
			tmp1 = SMLAWB(state.idx(int(i+1)), state.idx(int(i+2))-tmp2, int32(lambda_Q16))
			*state.ptr(int(i + 1)) = tmp2
			acc_Q11 = SMLAWB(acc_Q11, tmp2, int32(coef_Q13.idx(int(i))))
		}
		*state.ptr(int(order)) = tmp1
		acc_Q11 = SMLAWB(acc_Q11, tmp1, int32(coef_Q13.idx(int(order-1))))
		*res.ptr(int(n)) = int16(SAT16(int32(input.idx(int(n))) - RSHIFT_ROUND(acc_Q11, 11)))
	}
}

func prefilter_FIX(psEnc *encoder_state_FIX, psEncCtrl *encoder_control_FIX, xw, x *slice[int16]) {
	P := psEnc.sPrefilt
	var (
		j, k, lag                          int32
		tmp_32                             int32
		AR1_shp_Q13, px, pxw               *slice[int16]
		HarmShapeGain_Q12, Tilt_Q14        int32
		HarmShapeFIRPacked_Q12, LF_shp_Q14 int32
		x_filt_Q12                         = alloc[int32](MAX_FRAME_LENGTH / NB_SUBFR)
		st_res                             = alloc[int16]((MAX_FRAME_LENGTH / NB_SUBFR) + MAX_SHAPE_LPC_ORDER)
		B_Q12                              int32
	)

	px = x
	pxw = xw
	lag = P.lagPrev
	for k = 0; k < NB_SUBFR; k++ {
		if psEncCtrl.sCmn.sigtype == SIG_TYPE_VOICED {
			lag = psEncCtrl.sCmn.pitchL.idx(int(k))
		}

		HarmShapeGain_Q12 = SMULWB(psEncCtrl.HarmShapeGain_Q14.idx(int(k)), 16384-psEncCtrl.HarmBoost_Q14[k])
		HarmShapeFIRPacked_Q12 = RSHIFT(HarmShapeGain_Q12, 2)
		HarmShapeFIRPacked_Q12 |= LSHIFT(RSHIFT(HarmShapeGain_Q12, 1), 16)
		Tilt_Q14 = psEncCtrl.Tilt_Q14.idx(int(k))
		LF_shp_Q14 = psEncCtrl.LF_shp_Q14.idx(int(k))
		AR1_shp_Q13 = psEncCtrl.AR1_Q13.off(int(k * MAX_SHAPE_LPC_ORDER))

		warped_LPC_analysis_filter_FIX(P.sAR_shp, st_res, AR1_shp_Q13, px,
			int16(psEnc.sCmn.warping_Q16), psEnc.sCmn.subfr_length, psEnc.sCmn.shapingLPCOrder)

		B_Q12 = RSHIFT_ROUND(psEncCtrl.GainsPre_Q14[k], 2)
		tmp_32 = SMLABB(FIX_CONST(INPUT_TILT, 26), psEncCtrl.HarmBoost_Q14[k], HarmShapeGain_Q12)
		tmp_32 = SMLABB(tmp_32, psEncCtrl.coding_quality_Q14, FIX_CONST(HIGH_RATE_INPUT_TILT, 12))
		tmp_32 = SMULWB(tmp_32, -psEncCtrl.GainsPre_Q14[k])
		tmp_32 = RSHIFT_ROUND(tmp_32, 12)
		B_Q12 |= LSHIFT(SAT16(tmp_32), 16)

		*x_filt_Q12.ptr(0) = SMLABT(SMULBB(int32(st_res.idx(0)), B_Q12), P.sHarmHP, B_Q12)
		for j = 1; j < psEnc.sCmn.subfr_length; j++ {
			*x_filt_Q12.ptr(int(j)) = SMLABT(SMULBB(int32(st_res.idx(int(j))), B_Q12), int32(st_res.idx(int(j-1))), B_Q12)
		}

		P.sHarmHP = int32(st_res.idx(int(psEnc.sCmn.subfr_length - 1)))

		prefilt_FIX(P, x_filt_Q12, pxw, HarmShapeFIRPacked_Q12, Tilt_Q14,
			LF_shp_Q14, lag, psEnc.sCmn.subfr_length)

		px = px.off(int(psEnc.sCmn.subfr_length))
		pxw = pxw.off(int(psEnc.sCmn.subfr_length))
	}

	P.lagPrev = psEncCtrl.sCmn.pitchL.idx(NB_SUBFR - 1)
}

func prefilt_FIX(P *prefilter_state_FIX, st_res_Q12 *slice[int32], xw *slice[int16],
	HarmShapeFIRPacked_Q12 int32, Tilt_Q14, LF_shp_Q14, lag, length int32) {
	var (
		i, idx, LTP_shp_buf_idx         int32
		n_LTP_Q12, n_Tilt_Q10, n_LF_Q10 int32
		sLF_MA_shp_Q12, sLF_AR_shp_Q12  int32
		LTP_shp_buf                     *slice[int16]
	)

	LTP_shp_buf = P.sLTP_shp
	LTP_shp_buf_idx = P.sLTP_shp_buf_idx
	sLF_AR_shp_Q12 = P.sLF_AR_shp_Q12
	sLF_MA_shp_Q12 = P.sLF_MA_shp_Q12

	for i = 0; i < length; i++ {
		if lag > 0 {
			idx = lag + LTP_shp_buf_idx
			n_LTP_Q12 = SMULBB(int32(LTP_shp_buf.idx(int((idx-HARM_SHAPE_FIR_TAPS/2-1)&LTP_MASK))), HarmShapeFIRPacked_Q12)
			n_LTP_Q12 = SMLABT(n_LTP_Q12, int32(LTP_shp_buf.idx(int((idx-HARM_SHAPE_FIR_TAPS/2)&LTP_MASK))), HarmShapeFIRPacked_Q12)
			n_LTP_Q12 = SMLABB(n_LTP_Q12, int32(LTP_shp_buf.idx(int((idx-HARM_SHAPE_FIR_TAPS/2+1)&LTP_MASK))), HarmShapeFIRPacked_Q12)
		} else {
			n_LTP_Q12 = 0
		}

		n_Tilt_Q10 = SMULWB(sLF_AR_shp_Q12, Tilt_Q14)
		n_LF_Q10 = SMLAWB(SMULWT(sLF_AR_shp_Q12, LF_shp_Q14), sLF_MA_shp_Q12, LF_shp_Q14)

		sLF_AR_shp_Q12 = SUB32(st_res_Q12.idx(int(i)), LSHIFT(n_Tilt_Q10, 2))
		sLF_MA_shp_Q12 = SUB32(sLF_AR_shp_Q12, LSHIFT(n_LF_Q10, 2))

		LTP_shp_buf_idx = (LTP_shp_buf_idx - 1) & LTP_MASK

		*LTP_shp_buf.ptr(int(LTP_shp_buf_idx)) = int16(SAT16(RSHIFT_ROUND(sLF_MA_shp_Q12, 12)))

		*xw.ptr(int(i)) = int16(SAT16(RSHIFT_ROUND(SUB32(sLF_MA_shp_Q12, n_LTP_Q12), 12)))
	}

	P.sLF_AR_shp_Q12 = sLF_AR_shp_Q12
	P.sLF_MA_shp_Q12 = sLF_MA_shp_Q12
	P.sLTP_shp_buf_idx = LTP_shp_buf_idx
}

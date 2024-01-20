package silk

func find_pitch_lags_FIX(psEnc *encoder_state_FIX, psEncCtrl *encoder_control_FIX, res, x *slice[int16]) {
	var (
		psPredSt            = psEnc.sPred
		buf_len, i, scale   int32
		thrhld_Q15, res_nrg int32
		x_buf, x_buf_ptr    *slice[int16]

		Wsig      = alloc[int16](FIND_PITCH_LPC_WIN_MAX)
		Wsig_ptr  *slice[int16]
		auto_corr = alloc[int32](MAX_FIND_PITCH_LPC_ORDER + 1)
		rc_Q15    = alloc[int16](MAX_FIND_PITCH_LPC_ORDER)
		A_Q24     = alloc[int32](MAX_FIND_PITCH_LPC_ORDER)
		FiltState = alloc[int32](MAX_FIND_PITCH_LPC_ORDER)
		A_Q12     = alloc[int16](MAX_FIND_PITCH_LPC_ORDER)
	)

	buf_len = ADD_LSHIFT(psEnc.sCmn.la_pitch, psEnc.sCmn.frame_length, 1)

	x_buf = x.off(-int(psEnc.sCmn.frame_length))

	x_buf_ptr = x_buf.off(int(buf_len - psPredSt.pitch_LPC_win_length))
	Wsig_ptr = Wsig.off(0)
	apply_sine_window(Wsig_ptr, x_buf_ptr, 1, psEnc.sCmn.la_pitch)

	Wsig_ptr = Wsig_ptr.off(int(psEnc.sCmn.la_pitch))
	x_buf_ptr = x_buf_ptr.off(int(psEnc.sCmn.la_pitch))
	x_buf_ptr.copy(Wsig_ptr, int(psPredSt.pitch_LPC_win_length-LSHIFT(psEnc.sCmn.la_pitch, 1)))

	Wsig_ptr = Wsig_ptr.off(int(psPredSt.pitch_LPC_win_length - LSHIFT(psEnc.sCmn.la_pitch, 1)))
	x_buf_ptr = x_buf_ptr.off(int(psPredSt.pitch_LPC_win_length - LSHIFT(psEnc.sCmn.la_pitch, 1)))
	apply_sine_window(Wsig_ptr, x_buf_ptr, 2, psEnc.sCmn.la_pitch)

	autocorr(auto_corr, &scale, Wsig, psPredSt.pitch_LPC_win_length, psEnc.sCmn.pitchEstimationLPCOrder+1)

	*auto_corr.ptr(0) = SMLAWB(auto_corr.idx(0), auto_corr.idx(0), FIX_CONST(FIND_PITCH_WHITE_NOISE_FRACTION, 16))

	res_nrg = schur(rc_Q15, auto_corr, psEnc.sCmn.pitchEstimationLPCOrder)

	psEncCtrl.predGain_Q16 = DIV32_varQ(auto_corr.idx(0), max(res_nrg, 1), 16)

	k2a(A_Q24, rc_Q15, psEnc.sCmn.pitchEstimationLPCOrder)

	for i = 0; i < psEnc.sCmn.pitchEstimationLPCOrder; i++ {
		*A_Q12.ptr(int(i)) = int16(SAT16(RSHIFT(A_Q24.idx(int(i)), 12)))
	}

	bwexpander(A_Q12, psEnc.sCmn.pitchEstimationLPCOrder, FIX_CONST(FIND_PITCH_BANDWITH_EXPANSION, 16))

	MA_Prediction(x_buf, A_Q12, FiltState, res, buf_len, psEnc.sCmn.pitchEstimationLPCOrder)
	memset(res, 0, int(psEnc.sCmn.pitchEstimationLPCOrder))

	thrhld_Q15 = FIX_CONST(0.45, 15)
	thrhld_Q15 = SMLABB(thrhld_Q15, FIX_CONST(-0.004, 15), psEnc.sCmn.pitchEstimationLPCOrder)
	thrhld_Q15 = SMLABB(thrhld_Q15, FIX_CONST(-0.1, 7), psEnc.speech_activity_Q8)
	thrhld_Q15 = SMLABB(thrhld_Q15, FIX_CONST(0.15, 15), psEnc.sCmn.prev_sigtype)
	thrhld_Q15 = SMLAWB(thrhld_Q15, FIX_CONST(-0.1, 16), psEncCtrl.input_tilt_Q15)
	thrhld_Q15 = SAT16(thrhld_Q15)

	psEncCtrl.sCmn.sigtype = pitch_analysis_core(res, psEncCtrl.sCmn.pitchL, &psEncCtrl.sCmn.lagIndex,
		&psEncCtrl.sCmn.contourIndex, &psEnc.LTPCorr_Q15, psEnc.sCmn.prevLag, psEnc.sCmn.pitchEstimationThreshold_Q16,
		thrhld_Q15, psEnc.sCmn.fs_kHz, psEnc.sCmn.pitchEstimationComplexity, 0)
}

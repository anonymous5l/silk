package silk

func process_NLSFs_FIX(psEnc *encoder_state_FIX, psEncCtrl *encoder_control_FIX, pNLSF_Q15 *slice[int32]) {
	var (
		doInterpolate                     bool
		pNLSFW_Q6                         = alloc[int32](MAX_LPC_ORDER)
		NLSF_mu_Q15, NLSF_mu_fluc_red_Q16 int32
		i_sqr_Q15                         int32
		psNLSF_CB                         *NLSF_CB_struct
		pNLSF0_temp_Q15                   = alloc[int32](MAX_LPC_ORDER)
		pNLSFW0_temp_Q6                   = alloc[int32](MAX_LPC_ORDER)
		i                                 int32
	)

	if psEncCtrl.sCmn.sigtype == SIG_TYPE_VOICED {
		NLSF_mu_Q15 = SMLAWB(66, -8388, psEnc.speech_activity_Q8)
		NLSF_mu_fluc_red_Q16 = SMLAWB(6554, -838848, psEnc.speech_activity_Q8)
	} else {
		NLSF_mu_Q15 = SMLAWB(164, -33554, psEnc.speech_activity_Q8)
		NLSF_mu_fluc_red_Q16 = SMLAWB(13107, -1677696, psEnc.speech_activity_Q8+psEncCtrl.sparseness_Q8)
	}

	NLSF_mu_Q15 = max(NLSF_mu_Q15, 1)

	NLSF_VQ_weights_laroia(pNLSFW_Q6, pNLSF_Q15, psEnc.sCmn.predictLPCOrder)

	doInterpolate = (psEnc.sCmn.useInterpolatedNLSFs == 1) && (psEncCtrl.sCmn.NLSFInterpCoef_Q2 < (1 << 2))
	if doInterpolate {

		interpolate(pNLSF0_temp_Q15, psEnc.sPred.prev_NLSFq_Q15, pNLSF_Q15,
			psEncCtrl.sCmn.NLSFInterpCoef_Q2, psEnc.sCmn.predictLPCOrder)

		NLSF_VQ_weights_laroia(pNLSFW0_temp_Q6, pNLSF0_temp_Q15, psEnc.sCmn.predictLPCOrder)

		i_sqr_Q15 = LSHIFT(SMULBB(psEncCtrl.sCmn.NLSFInterpCoef_Q2, psEncCtrl.sCmn.NLSFInterpCoef_Q2), 11)
		for i = 0; i < psEnc.sCmn.predictLPCOrder; i++ {
			*pNLSFW_Q6.ptr(int(i)) = SMLAWB(RSHIFT(pNLSFW_Q6.idx(int(i)), 1), pNLSFW0_temp_Q6.idx(int(i)), i_sqr_Q15)
		}
	}

	psNLSF_CB = &psEnc.sCmn.psNLSF_CB[psEncCtrl.sCmn.sigtype]

	NLSF_MSVQ_encode_FIX(psEncCtrl.sCmn.NLSFIndices, pNLSF_Q15, psNLSF_CB,
		psEnc.sPred.prev_NLSFq_Q15, pNLSFW_Q6, NLSF_mu_Q15, NLSF_mu_fluc_red_Q16,
		psEnc.sCmn.NLSF_MSVQ_Survivors, psEnc.sCmn.predictLPCOrder, psEnc.sCmn.first_frame_after_reset)

	NLSF2A_stable(psEncCtrl.PredCoef_Q12[1], pNLSF_Q15, psEnc.sCmn.predictLPCOrder)

	if doInterpolate {
		interpolate(pNLSF0_temp_Q15, psEnc.sPred.prev_NLSFq_Q15, pNLSF_Q15,
			psEncCtrl.sCmn.NLSFInterpCoef_Q2, psEnc.sCmn.predictLPCOrder)

		NLSF2A_stable(psEncCtrl.PredCoef_Q12[0], pNLSF0_temp_Q15, psEnc.sCmn.predictLPCOrder)

	} else {
		psEncCtrl.PredCoef_Q12[1].copy(psEncCtrl.PredCoef_Q12[0], int(psEnc.sCmn.predictLPCOrder))
	}
}

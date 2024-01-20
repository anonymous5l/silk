package silk

import (
	"math"
)

func find_pred_coefs_FIX(psEnc *encoder_state_FIX, psEncCtrl *encoder_control_FIX, res_pitch *slice[int16]) {
	var (
		i                 int32
		WLTP              = alloc[int32](NB_SUBFR * LTP_ORDER * LTP_ORDER)
		invGains_Q16      = alloc[int32](NB_SUBFR)
		local_gains       = alloc[int32](NB_SUBFR)
		Wght_Q15          = alloc[int32](NB_SUBFR)
		NLSF_Q15          = alloc[int32](MAX_LPC_ORDER)
		x_ptr, x_pre_ptr  *slice[int16]
		LPC_in_pre        = alloc[int16](NB_SUBFR*MAX_LPC_ORDER + MAX_FRAME_LENGTH)
		tmp, min_gain_Q16 int32
		LTP_corrs_rshift  = alloc[int32](NB_SUBFR)
	)

	min_gain_Q16 = math.MaxInt32 >> 6
	for i = 0; i < NB_SUBFR; i++ {
		min_gain_Q16 = min(min_gain_Q16, psEncCtrl.Gains_Q16.idx(int(i)))
	}
	for i = 0; i < NB_SUBFR; i++ {
		*invGains_Q16.ptr(int(i)) = DIV32_varQ(min_gain_Q16, psEncCtrl.Gains_Q16.idx(int(i)), 16-2)

		*invGains_Q16.ptr(int(i)) = max(invGains_Q16.idx(int(i)), 363)

		tmp = SMULWB(invGains_Q16.idx(int(i)), invGains_Q16.idx(int(i)))
		*Wght_Q15.ptr(int(i)) = RSHIFT(tmp, 1)

		*local_gains.ptr(int(i)) = DIV32(1<<16, invGains_Q16.idx(int(i)))
	}

	if psEncCtrl.sCmn.sigtype == SIG_TYPE_VOICED {
		find_LTP_FIX(psEncCtrl.LTPCoef_Q14, WLTP, &psEncCtrl.LTPredCodGain_Q7, res_pitch,
			res_pitch.off(int(RSHIFT(psEnc.sCmn.frame_length, 1))), psEncCtrl.sCmn.pitchL, Wght_Q15,
			psEnc.sCmn.subfr_length, psEnc.sCmn.frame_length, LTP_corrs_rshift)

		quant_LTP_gains_FIX(psEncCtrl.LTPCoef_Q14, psEncCtrl.sCmn.LTPIndex, &psEncCtrl.sCmn.PERIndex,
			WLTP, psEnc.mu_LTP_Q8, psEnc.sCmn.LTPQuantLowComplexity)

		LTP_scale_ctrl_FIX(psEnc, psEncCtrl)

		LTP_analysis_filter_FIX(LPC_in_pre, psEnc.x_buf.off(int(psEnc.sCmn.frame_length-psEnc.sCmn.predictLPCOrder)),
			psEncCtrl.LTPCoef_Q14, psEncCtrl.sCmn.pitchL, invGains_Q16, psEnc.sCmn.subfr_length, psEnc.sCmn.predictLPCOrder)
	} else {
		x_ptr = psEnc.x_buf.off(int(psEnc.sCmn.frame_length - psEnc.sCmn.predictLPCOrder))
		x_pre_ptr = LPC_in_pre.off(0)
		for i = 0; i < NB_SUBFR; i++ {
			scale_copy_vector16(x_pre_ptr, x_ptr, invGains_Q16.idx(int(i)),
				psEnc.sCmn.subfr_length+psEnc.sCmn.predictLPCOrder)
			x_pre_ptr = x_pre_ptr.off(int(psEnc.sCmn.subfr_length + psEnc.sCmn.predictLPCOrder))
			x_ptr = x_ptr.off(int(psEnc.sCmn.subfr_length))
		}

		memset(psEncCtrl.LTPCoef_Q14, 0, NB_SUBFR*LTP_ORDER)
		psEncCtrl.LTPredCodGain_Q7 = 0
	}

	find_LPC_FIX(NLSF_Q15, &psEncCtrl.sCmn.NLSFInterpCoef_Q2, psEnc.sPred.prev_NLSFq_Q15,
		psEnc.sCmn.useInterpolatedNLSFs*(1-psEnc.sCmn.first_frame_after_reset), psEnc.sCmn.predictLPCOrder,
		LPC_in_pre, psEnc.sCmn.subfr_length+psEnc.sCmn.predictLPCOrder)

	process_NLSFs_FIX(psEnc, psEncCtrl, NLSF_Q15)

	residual_energy_FIX(psEncCtrl.ResNrg, psEncCtrl.ResNrgQ, LPC_in_pre, psEncCtrl.PredCoef_Q12, local_gains,
		psEnc.sCmn.subfr_length, psEnc.sCmn.predictLPCOrder)

	NLSF_Q15.copy(psEnc.sPred.prev_NLSFq_Q15, int(psEnc.sCmn.predictLPCOrder))
}

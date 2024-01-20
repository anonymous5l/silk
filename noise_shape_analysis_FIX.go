package silk

import (
	"math"
)

func warped_gain(coefs_Q24 *slice[int32], lambda_Q16, order int32) int32 {
	var i, gain_Q24 int32

	lambda_Q16 = -lambda_Q16
	gain_Q24 = coefs_Q24.idx(int(order - 1))
	for i = order - 2; i >= 0; i-- {
		gain_Q24 = SMLAWB(coefs_Q24.idx(int(i)), gain_Q24, lambda_Q16)
	}
	gain_Q24 = SMLAWB(FIX_CONST(1.0, 24), gain_Q24, -lambda_Q16)
	return INVERSE32_varQ(gain_Q24, 40)
}

func limit_warped_coefs(coefs_syn_Q24, coefs_ana_Q24 *slice[int32], lambda_Q16, limit_Q24, order int32) {
	var (
		i, iter, ind                                           int32
		tmp, maxabs_Q24, chirp_Q16, gain_syn_Q16, gain_ana_Q16 int32
		nom_Q16, den_Q24                                       int32
	)

	lambda_Q16 = -lambda_Q16
	for i = order - 1; i > 0; i-- {
		*coefs_syn_Q24.ptr(int(i - 1)) = SMLAWB(coefs_syn_Q24.idx(int(i-1)), coefs_syn_Q24.idx(int(i)), lambda_Q16)
		*coefs_ana_Q24.ptr(int(i - 1)) = SMLAWB(coefs_ana_Q24.idx(int(i-1)), coefs_ana_Q24.idx(int(i)), lambda_Q16)
	}
	lambda_Q16 = -lambda_Q16
	nom_Q16 = SMLAWB(FIX_CONST(1.0, 16), -lambda_Q16, lambda_Q16)
	den_Q24 = SMLAWB(FIX_CONST(1.0, 24), coefs_syn_Q24.idx(0), lambda_Q16)
	gain_syn_Q16 = DIV32_varQ(nom_Q16, den_Q24, 24)
	den_Q24 = SMLAWB(FIX_CONST(1.0, 24), coefs_ana_Q24.idx(0), lambda_Q16)
	gain_ana_Q16 = DIV32_varQ(nom_Q16, den_Q24, 24)
	for i = 0; i < order; i++ {
		*coefs_syn_Q24.ptr(int(i)) = SMULWW(gain_syn_Q16, coefs_syn_Q24.idx(int(i)))
		*coefs_ana_Q24.ptr(int(i)) = SMULWW(gain_ana_Q16, coefs_ana_Q24.idx(int(i)))
	}

	for iter = 0; iter < 10; iter++ {
		maxabs_Q24 = -1
		for i = 0; i < order; i++ {
			tmp = max(abs_int32(coefs_syn_Q24.idx(int(i))), abs_int32(coefs_ana_Q24.idx(int(i))))
			if tmp > maxabs_Q24 {
				maxabs_Q24 = tmp
				ind = i
			}
		}
		if maxabs_Q24 <= limit_Q24 {
			return
		}

		for i = 1; i < order; i++ {
			*coefs_syn_Q24.ptr(int(i - 1)) = SMLAWB(coefs_syn_Q24.idx(int(i-1)), coefs_syn_Q24.idx(int(i)), lambda_Q16)
			*coefs_ana_Q24.ptr(int(i - 1)) = SMLAWB(coefs_ana_Q24.idx(int(i-1)), coefs_ana_Q24.idx(int(i)), lambda_Q16)
		}
		gain_syn_Q16 = INVERSE32_varQ(gain_syn_Q16, 32)
		gain_ana_Q16 = INVERSE32_varQ(gain_ana_Q16, 32)
		for i = 0; i < order; i++ {
			*coefs_syn_Q24.ptr(int(i)) = SMULWW(gain_syn_Q16, coefs_syn_Q24.idx(int(i)))
			*coefs_ana_Q24.ptr(int(i)) = SMULWW(gain_ana_Q16, coefs_ana_Q24.idx(int(i)))
		}

		chirp_Q16 = FIX_CONST(0.99, 16) - DIV32_varQ(
			SMULWB(maxabs_Q24-limit_Q24, SMLABB(FIX_CONST(0.8, 10), FIX_CONST(0.1, 10), iter)),
			MUL(maxabs_Q24, ind+1), 22)
		bwexpander_32(coefs_syn_Q24, order, chirp_Q16)
		bwexpander_32(coefs_ana_Q24, order, chirp_Q16)

		lambda_Q16 = -lambda_Q16
		for i = order - 1; i > 0; i-- {
			*coefs_syn_Q24.ptr(int(i - 1)) = SMLAWB(coefs_syn_Q24.idx(int(i-1)), coefs_syn_Q24.idx(int(i)), lambda_Q16)
			*coefs_syn_Q24.ptr(int(i - 1)) = SMLAWB(coefs_ana_Q24.idx(int(i-1)), coefs_ana_Q24.idx(int(i)), lambda_Q16)
		}
		lambda_Q16 = -lambda_Q16
		nom_Q16 = SMLAWB(FIX_CONST(1.0, 16), -lambda_Q16, lambda_Q16)
		den_Q24 = SMLAWB(FIX_CONST(1.0, 24), coefs_syn_Q24.idx(0), lambda_Q16)
		gain_syn_Q16 = DIV32_varQ(nom_Q16, den_Q24, 24)
		den_Q24 = SMLAWB(FIX_CONST(1.0, 24), coefs_ana_Q24.idx(0), lambda_Q16)
		gain_ana_Q16 = DIV32_varQ(nom_Q16, den_Q24, 24)
		for i = 0; i < order; i++ {
			*coefs_syn_Q24.ptr(int(i)) = SMULWW(gain_syn_Q16, coefs_syn_Q24.idx(int(i)))
			*coefs_ana_Q24.ptr(int(i)) = SMULWW(gain_ana_Q16, coefs_ana_Q24.idx(int(i)))
		}
	}
}

func noise_shape_analysis_FIX(psEnc *encoder_state_FIX, psEncCtrl *encoder_control_FIX, pitch_res *slice[int16], x *slice[int16]) {

	psShapeSt := psEnc.sShape
	var (
		k, i, nSamples, Qnrg, b_Q14, warping_Q16, scale                                    int32
		SNR_adj_dB_Q7, HarmBoost_Q16, HarmShapeGain_Q16, Tilt_Q16, tmp32                   int32
		nrg, pre_nrg_Q30, log_energy_Q7, log_energy_prev_Q7, energy_variation_Q7           int32
		delta_Q16, BWExp1_Q16, BWExp2_Q16, gain_mult_Q16, gain_add_Q16, strength_Q16, b_Q8 int32
		auto_corr                                                                          = alloc[int32](MAX_SHAPE_LPC_ORDER + 1)
		refl_coef_Q16                                                                      = alloc[int32](MAX_SHAPE_LPC_ORDER)
		AR1_Q24                                                                            = alloc[int32](MAX_SHAPE_LPC_ORDER)
		AR2_Q24                                                                            = alloc[int32](MAX_SHAPE_LPC_ORDER)
		x_windowed                                                                         = alloc[int16](SHAPE_LPC_WIN_MAX)
		x_ptr, pitch_res_ptr                                                               *slice[int16]
	)

	x_ptr = x.off(-int(psEnc.sCmn.la_shape))

	psEncCtrl.current_SNR_dB_Q7 = psEnc.SNR_dB_Q7 - SMULWB(LSHIFT(psEnc.BufferedInChannel_ms, 7),
		FIX_CONST(0.05, 16))

	if psEnc.speech_activity_Q8 > FIX_CONST(LBRR_SPEECH_ACTIVITY_THRES, 8) {
		psEncCtrl.current_SNR_dB_Q7 -= RSHIFT(psEnc.inBandFEC_SNR_comp_Q8, 1)
	}

	psEncCtrl.input_quality_Q14 = RSHIFT(psEncCtrl.input_quality_bands_Q15.idx(0)+psEncCtrl.input_quality_bands_Q15.idx(1), 2)

	psEncCtrl.coding_quality_Q14 = RSHIFT(sigm_Q15(RSHIFT_ROUND(psEncCtrl.current_SNR_dB_Q7-
		FIX_CONST(18.0, 7), 4)), 1)

	b_Q8 = FIX_CONST(1.0, 8) - psEnc.speech_activity_Q8
	b_Q8 = SMULWB(LSHIFT(b_Q8, 8), b_Q8)
	SNR_adj_dB_Q7 = SMLAWB(psEncCtrl.current_SNR_dB_Q7,
		SMULBB(FIX_CONST(-BG_SNR_DECR_dB, 7)>>(4+1), b_Q8),
		SMULWB(FIX_CONST(1.0, 14)+psEncCtrl.input_quality_Q14, psEncCtrl.coding_quality_Q14))

	if psEncCtrl.sCmn.sigtype == SIG_TYPE_VOICED {
		SNR_adj_dB_Q7 = SMLAWB(SNR_adj_dB_Q7, FIX_CONST(HARM_SNR_INCR_dB, 8), psEnc.LTPCorr_Q15)
	} else {
		SNR_adj_dB_Q7 = SMLAWB(SNR_adj_dB_Q7,
			SMLAWB(FIX_CONST(6.0, 9), -FIX_CONST(0.4, 18), psEncCtrl.current_SNR_dB_Q7),
			FIX_CONST(1.0, 14)-psEncCtrl.input_quality_Q14)
	}

	if psEncCtrl.sCmn.sigtype == SIG_TYPE_VOICED {
		psEncCtrl.sCmn.QuantOffsetType = 0
		psEncCtrl.sparseness_Q8 = 0
	} else {
		nSamples = LSHIFT(psEnc.sCmn.fs_kHz, 1)
		energy_variation_Q7 = 0
		log_energy_prev_Q7 = 0
		pitch_res_ptr = pitch_res.off(0)
		for k = 0; k < FRAME_LENGTH_MS/2; k++ {
			sum_sqr_shift(&nrg, &scale, pitch_res_ptr, nSamples)
			nrg += RSHIFT(nSamples, scale)

			log_energy_Q7 = lin2log(nrg)
			if k > 0 {
				energy_variation_Q7 += abs(log_energy_Q7 - log_energy_prev_Q7)
			}
			log_energy_prev_Q7 = log_energy_Q7
			pitch_res_ptr = pitch_res_ptr.off(int(nSamples))
		}

		psEncCtrl.sparseness_Q8 = RSHIFT(sigm_Q15(SMULWB(energy_variation_Q7-
			FIX_CONST(5.0, 7), FIX_CONST(0.1, 16))), 7)

		if psEncCtrl.sparseness_Q8 > FIX_CONST(SPARSENESS_THRESHOLD_QNT_OFFSET, 8) {
			psEncCtrl.sCmn.QuantOffsetType = 0
		} else {
			psEncCtrl.sCmn.QuantOffsetType = 1
		}

		SNR_adj_dB_Q7 = SMLAWB(SNR_adj_dB_Q7, FIX_CONST(SPARSE_SNR_INCR_dB, 15), psEncCtrl.sparseness_Q8-FIX_CONST(0.5, 8))
	}

	strength_Q16 = SMULWB(psEncCtrl.predGain_Q16, FIX_CONST(FIND_PITCH_WHITE_NOISE_FRACTION, 16))
	BWExp1_Q16 = DIV32_varQ(FIX_CONST(BANDWIDTH_EXPANSION, 16),
		SMLAWW(FIX_CONST(1.0, 16), strength_Q16, strength_Q16), 16)
	BWExp2_Q16 = BWExp1_Q16
	delta_Q16 = SMULWB(FIX_CONST(1.0, 16)-SMULBB(3, psEncCtrl.coding_quality_Q14),
		FIX_CONST(LOW_RATE_BANDWIDTH_EXPANSION_DELTA, 16))
	BWExp1_Q16 = SUB32(BWExp1_Q16, delta_Q16)
	BWExp2_Q16 = ADD32(BWExp2_Q16, delta_Q16)
	BWExp1_Q16 = DIV32_16(LSHIFT(BWExp1_Q16, 14), int16(RSHIFT(BWExp2_Q16, 2)))

	if psEnc.sCmn.warping_Q16 > 0 {
		warping_Q16 = SMLAWB(psEnc.sCmn.warping_Q16, psEncCtrl.coding_quality_Q14, FIX_CONST(0.01, 18))
	} else {
		warping_Q16 = 0
	}

	for k = 0; k < NB_SUBFR; k++ {
		var shift, slope_part, flat_part int32
		flat_part = psEnc.sCmn.fs_kHz * 5
		slope_part = RSHIFT(psEnc.sCmn.shapeWinLength-flat_part, 1)

		apply_sine_window(x_windowed, x_ptr, 1, slope_part)
		shift = slope_part
		x_ptr.off(int(shift)).copy(x_windowed.off(int(shift)), int(flat_part))
		shift += flat_part
		apply_sine_window(x_windowed.off(int(shift)), x_ptr.off(int(shift)), 2, slope_part)

		x_ptr = x_ptr.off(int(psEnc.sCmn.subfr_length))

		if psEnc.sCmn.warping_Q16 > 0 {
			warped_autocorrelation_FIX(auto_corr, &scale, x_windowed, int16(warping_Q16), psEnc.sCmn.shapeWinLength, psEnc.sCmn.shapingLPCOrder)
		} else {
			autocorr(auto_corr, &scale, x_windowed, psEnc.sCmn.shapeWinLength, psEnc.sCmn.shapingLPCOrder+1)
		}

		*auto_corr.ptr(0) = ADD32(auto_corr.idx(0), max(SMULWB(RSHIFT(auto_corr.idx(0), 4),
			FIX_CONST(SHAPE_WHITE_NOISE_FRACTION, 20)), 1))

		nrg = schur64(refl_coef_Q16, auto_corr, psEnc.sCmn.shapingLPCOrder)

		k2a_Q16(AR2_Q24, refl_coef_Q16, psEnc.sCmn.shapingLPCOrder)

		Qnrg = -scale

		if Qnrg&1 != 0 {
			Qnrg -= 1
			nrg >>= 1
		}

		tmp32 = SQRT_APPROX(nrg)
		Qnrg >>= 1

		*psEncCtrl.Gains_Q16.ptr(int(k)) = LSHIFT_SAT32(tmp32, 16-Qnrg)

		if psEnc.sCmn.warping_Q16 > 0 {
			gain_mult_Q16 = warped_gain(AR2_Q24, warping_Q16, psEnc.sCmn.shapingLPCOrder)
			*psEncCtrl.Gains_Q16.ptr(int(k)) = SMULWW(psEncCtrl.Gains_Q16.idx(int(k)), gain_mult_Q16)
			if psEncCtrl.Gains_Q16.idx(int(k)) < 0 {
				*psEncCtrl.Gains_Q16.ptr(int(k)) = math.MaxInt32
			}
		}

		bwexpander_32(AR2_Q24, psEnc.sCmn.shapingLPCOrder, BWExp2_Q16)

		AR2_Q24.copy(AR1_Q24, int(psEnc.sCmn.shapingLPCOrder))

		bwexpander_32(AR1_Q24, psEnc.sCmn.shapingLPCOrder, BWExp1_Q16)

		LPC_inverse_pred_gain_Q24(&pre_nrg_Q30, AR2_Q24, psEnc.sCmn.shapingLPCOrder)
		LPC_inverse_pred_gain_Q24(&nrg, AR1_Q24, psEnc.sCmn.shapingLPCOrder)

		pre_nrg_Q30 = LSHIFT32(SMULWB(pre_nrg_Q30, FIX_CONST(0.7, 15)), 1)
		psEncCtrl.GainsPre_Q14[k] = FIX_CONST(0.3, 14) + DIV32_varQ(pre_nrg_Q30, nrg, 14)

		limit_warped_coefs(AR2_Q24, AR1_Q24, warping_Q16, FIX_CONST(3.999, 24), psEnc.sCmn.shapingLPCOrder)

		for i = 0; i < psEnc.sCmn.shapingLPCOrder; i++ {
			*psEncCtrl.AR1_Q13.ptr(int(k*MAX_SHAPE_LPC_ORDER + i)) = int16(SAT16(RSHIFT_ROUND(AR1_Q24.idx(int(i)), 11)))
			*psEncCtrl.AR2_Q13.ptr(int(k*MAX_SHAPE_LPC_ORDER + i)) = int16(SAT16(RSHIFT_ROUND(AR2_Q24.idx(int(i)), 11)))
		}
	}

	gain_mult_Q16 = log2lin(-SMLAWB(-FIX_CONST(16.0, 7), SNR_adj_dB_Q7, FIX_CONST(0.16, 16)))
	gain_add_Q16 = log2lin(SMLAWB(FIX_CONST(16.0, 7), FIX_CONST(NOISE_FLOOR_dB, 7), FIX_CONST(0.16, 16)))
	tmp32 = log2lin(SMLAWB(FIX_CONST(16.0, 7), FIX_CONST(RELATIVE_MIN_GAIN_dB, 7), FIX_CONST(0.16, 16)))
	tmp32 = SMULWW(psEnc.avgGain_Q16, tmp32)
	gain_add_Q16 = ADD_SAT32(gain_add_Q16, tmp32)

	for k = 0; k < NB_SUBFR; k++ {
		*psEncCtrl.Gains_Q16.ptr(int(k)) = SMULWW(psEncCtrl.Gains_Q16.idx(int(k)), gain_mult_Q16)
		if psEncCtrl.Gains_Q16.idx(int(k)) < 0 {
			*psEncCtrl.Gains_Q16.ptr(int(k)) = math.MaxInt32
		}
	}

	for k = 0; k < NB_SUBFR; k++ {
		*psEncCtrl.Gains_Q16.ptr(int(k)) = ADD_POS_SAT32(psEncCtrl.Gains_Q16.idx(int(k)), gain_add_Q16)
		psEnc.avgGain_Q16 = ADD_SAT32(
			psEnc.avgGain_Q16,
			SMULWB(
				psEncCtrl.Gains_Q16.idx(int(k))-psEnc.avgGain_Q16,
				RSHIFT_ROUND(SMULBB(psEnc.speech_activity_Q8, FIX_CONST(GAIN_SMOOTHING_COEF, 10)), 2)))
	}

	gain_mult_Q16 = FIX_CONST(1.0, 16) + RSHIFT_ROUND(MLA(FIX_CONST(INPUT_TILT, 26),
		psEncCtrl.coding_quality_Q14, FIX_CONST(HIGH_RATE_INPUT_TILT, 12)), 10)

	if psEncCtrl.input_tilt_Q15 <= 0 && psEncCtrl.sCmn.sigtype == SIG_TYPE_UNVOICED {
		if psEnc.sCmn.fs_kHz == 24 {
			essStrength_Q15 := SMULWW(-psEncCtrl.input_tilt_Q15,
				SMULBB(psEnc.speech_activity_Q8, FIX_CONST(1.0, 8)-psEncCtrl.sparseness_Q8))
			tmp32 = log2lin(FIX_CONST(16.0, 7) - SMULWB(essStrength_Q15,
				SMULWB(FIX_CONST(DE_ESSER_COEF_SWB_dB, 7), FIX_CONST(0.16, 17))))
			gain_mult_Q16 = SMULWW(gain_mult_Q16, tmp32)
		} else if psEnc.sCmn.fs_kHz == 16 {
			essStrength_Q15 := SMULWW(-psEncCtrl.input_tilt_Q15,
				SMULBB(psEnc.speech_activity_Q8, FIX_CONST(1.0, 8)-psEncCtrl.sparseness_Q8))
			tmp32 = log2lin(FIX_CONST(16.0, 7) - SMULWB(essStrength_Q15,
				SMULWB(FIX_CONST(DE_ESSER_COEF_WB_dB, 7), FIX_CONST(0.16, 17))))
			gain_mult_Q16 = SMULWW(gain_mult_Q16, tmp32)
		}
	}

	for k = 0; k < NB_SUBFR; k++ {
		psEncCtrl.GainsPre_Q14[k] = SMULWB(gain_mult_Q16, psEncCtrl.GainsPre_Q14[k])
	}

	strength_Q16 = MUL(FIX_CONST(LOW_FREQ_SHAPING, 0), FIX_CONST(1.0, 16)+
		SMULBB(FIX_CONST(LOW_QUALITY_LOW_FREQ_SHAPING_DECR, 1),
			psEncCtrl.input_quality_bands_Q15.idx(0)-FIX_CONST(1.0, 15)))
	if psEncCtrl.sCmn.sigtype == SIG_TYPE_VOICED {
		fs_kHz_inv := DIV32_16(FIX_CONST(0.2, 14), int16(psEnc.sCmn.fs_kHz))
		for k = 0; k < NB_SUBFR; k++ {
			b_Q14 = fs_kHz_inv + DIV32_16(FIX_CONST(3.0, 14), int16(psEncCtrl.sCmn.pitchL.idx(int(k))))
			*psEncCtrl.LF_shp_Q14.ptr(int(k)) = LSHIFT(FIX_CONST(1.0, 14)-b_Q14-SMULWB(strength_Q16, b_Q14), 16)
			*psEncCtrl.LF_shp_Q14.ptr(int(k)) |= int32(uint16(b_Q14 - FIX_CONST(1.0, 14)))
		}

		Tilt_Q16 = -FIX_CONST(HP_NOISE_COEF, 16) -
			SMULWB(FIX_CONST(1.0, 16)-FIX_CONST(HP_NOISE_COEF, 16),
				SMULWB(FIX_CONST(HARM_HP_NOISE_COEF, 24), psEnc.speech_activity_Q8))
	} else {
		b_Q14 = DIV32_16(21299, int16(psEnc.sCmn.fs_kHz))
		*psEncCtrl.LF_shp_Q14.ptr(0) = LSHIFT(FIX_CONST(1.0, 14)-b_Q14-
			SMULWB(strength_Q16, SMULWB(FIX_CONST(0.6, 16), b_Q14)), 16)
		*psEncCtrl.LF_shp_Q14.ptr(0) |= int32(uint16(b_Q14 - FIX_CONST(1.0, 14)))
		for k = 1; k < NB_SUBFR; k++ {
			*psEncCtrl.LF_shp_Q14.ptr(int(k)) = psEncCtrl.LF_shp_Q14.idx(0)
		}
		Tilt_Q16 = -FIX_CONST(HP_NOISE_COEF, 16)
	}

	HarmBoost_Q16 = SMULWB(SMULWB(FIX_CONST(1.0, 17)-LSHIFT(psEncCtrl.coding_quality_Q14, 3),
		psEnc.LTPCorr_Q15), FIX_CONST(LOW_RATE_HARMONIC_BOOST, 16))

	HarmBoost_Q16 = SMLAWB(HarmBoost_Q16,
		FIX_CONST(1.0, 16)-LSHIFT(psEncCtrl.input_quality_Q14, 2), FIX_CONST(LOW_INPUT_QUALITY_HARMONIC_BOOST, 16))

	if psEncCtrl.sCmn.sigtype == SIG_TYPE_VOICED {
		HarmShapeGain_Q16 = SMLAWB(FIX_CONST(HARMONIC_SHAPING, 16),
			FIX_CONST(1.0, 16)-SMULWB(FIX_CONST(1.0, 18)-LSHIFT(psEncCtrl.coding_quality_Q14, 4),
				psEncCtrl.input_quality_Q14), FIX_CONST(HIGH_RATE_OR_LOW_QUALITY_HARMONIC_SHAPING, 16))

		HarmShapeGain_Q16 = SMULWB(LSHIFT(HarmShapeGain_Q16, 1),
			SQRT_APPROX(LSHIFT(psEnc.LTPCorr_Q15, 15)))
	} else {
		HarmShapeGain_Q16 = 0
	}

	for k = 0; k < NB_SUBFR; k++ {
		psShapeSt.HarmBoost_smth_Q16 =
			SMLAWB(psShapeSt.HarmBoost_smth_Q16, HarmBoost_Q16-psShapeSt.HarmBoost_smth_Q16, FIX_CONST(SUBFR_SMTH_COEF, 16))
		psShapeSt.HarmShapeGain_smth_Q16 =
			SMLAWB(psShapeSt.HarmShapeGain_smth_Q16, HarmShapeGain_Q16-psShapeSt.HarmShapeGain_smth_Q16, FIX_CONST(SUBFR_SMTH_COEF, 16))
		psShapeSt.Tilt_smth_Q16 =
			SMLAWB(psShapeSt.Tilt_smth_Q16, Tilt_Q16-psShapeSt.Tilt_smth_Q16, FIX_CONST(SUBFR_SMTH_COEF, 16))

		psEncCtrl.HarmBoost_Q14[k] = RSHIFT_ROUND(psShapeSt.HarmBoost_smth_Q16, 2)
		*psEncCtrl.HarmShapeGain_Q14.ptr(int(k)) = RSHIFT_ROUND(psShapeSt.HarmShapeGain_smth_Q16, 2)
		*psEncCtrl.Tilt_Q14.ptr(int(k)) = RSHIFT_ROUND(psShapeSt.Tilt_smth_Q16, 2)
	}
}

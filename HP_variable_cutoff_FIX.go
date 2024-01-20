package silk

const (
	RADIANS_CONSTANT_Q19         = 1482
	LOG2_VARIABLE_HP_MIN_FREQ_Q7 = 809
)

func HP_variable_cutoff_FIX(psEnc *encoder_state_FIX, psEncCtrl *encoder_control_FIX, out, in *slice[int16]) {
	var (
		quality_Q15                                         int32
		B_Q28                                               [3]int32
		A_Q28                                               [2]int32
		Fc_Q19, r_Q28, r_Q22                                int32
		pitch_freq_Hz_Q16, pitch_freq_log_Q7, delta_freq_Q7 int32
	)

	if psEnc.sCmn.prev_sigtype == SIG_TYPE_VOICED {
		pitch_freq_Hz_Q16 = DIV32_16(LSHIFT(MUL(psEnc.sCmn.fs_kHz, 1000), 16), int16(psEnc.sCmn.prevLag))
		pitch_freq_log_Q7 = lin2log(pitch_freq_Hz_Q16) - (16 << 7)

		quality_Q15 = psEncCtrl.input_quality_bands_Q15.idx(0)
		pitch_freq_log_Q7 = SUB32(pitch_freq_log_Q7, SMULWB(SMULWB(LSHIFT(quality_Q15, 2), quality_Q15),
			pitch_freq_log_Q7-LOG2_VARIABLE_HP_MIN_FREQ_Q7))
		pitch_freq_log_Q7 = ADD32(pitch_freq_log_Q7, RSHIFT(FIX_CONST(0.6, 15)-quality_Q15, 9))

		delta_freq_Q7 = pitch_freq_log_Q7 - RSHIFT(psEnc.variable_HP_smth1_Q15, 8)
		if delta_freq_Q7 < 0 {
			delta_freq_Q7 = MUL(delta_freq_Q7, 3)
		}

		delta_freq_Q7 = LIMIT_32(delta_freq_Q7, -FIX_CONST(VARIABLE_HP_MAX_DELTA_FREQ, 7), FIX_CONST(VARIABLE_HP_MAX_DELTA_FREQ, 7))

		psEnc.variable_HP_smth1_Q15 = SMLAWB(psEnc.variable_HP_smth1_Q15,
			MUL(LSHIFT(psEnc.speech_activity_Q8, 1), delta_freq_Q7), FIX_CONST(VARIABLE_HP_SMTH_COEF1, 16))
	}
	psEnc.variable_HP_smth2_Q15 = SMLAWB(psEnc.variable_HP_smth2_Q15,
		psEnc.variable_HP_smth1_Q15-psEnc.variable_HP_smth2_Q15, FIX_CONST(VARIABLE_HP_SMTH_COEF2, 16))

	psEncCtrl.pitch_freq_low_Hz = log2lin(RSHIFT(psEnc.variable_HP_smth2_Q15, 8))

	psEncCtrl.pitch_freq_low_Hz = LIMIT_32(psEncCtrl.pitch_freq_low_Hz,
		FIX_CONST(VARIABLE_HP_MIN_FREQ, 0), FIX_CONST(VARIABLE_HP_MAX_FREQ, 0))

	Fc_Q19 = DIV32_16(SMULBB(RADIANS_CONSTANT_Q19, psEncCtrl.pitch_freq_low_Hz), int16(psEnc.sCmn.fs_kHz))

	r_Q28 = FIX_CONST(1.0, 28) - MUL(FIX_CONST(0.92, 9), Fc_Q19)

	B_Q28[0] = r_Q28
	B_Q28[1] = LSHIFT(-r_Q28, 1)
	B_Q28[2] = r_Q28

	r_Q22 = RSHIFT(r_Q28, 6)
	A_Q28[0] = SMULWW(r_Q22, SMULWW(Fc_Q19, Fc_Q19)-FIX_CONST(2.0, 22))
	A_Q28[1] = SMULWW(r_Q22, r_Q22)

	biquad_alt(in, B_Q28[:], A_Q28[:], psEnc.sCmn.In_HP_State[:], out, psEnc.sCmn.frame_length)
}

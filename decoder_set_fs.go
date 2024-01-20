package silk

func decoder_set_fs(psDec *decoder_state, fs_kHz int32) {
	if psDec.fs_kHz != fs_kHz {
		psDec.fs_kHz = fs_kHz
		psDec.frame_length = SMULBB(FRAME_LENGTH_MS, fs_kHz)
		psDec.subfr_length = SMULBB(FRAME_LENGTH_MS/NB_SUBFR, fs_kHz)
		if psDec.fs_kHz == 8 {
			psDec.LPC_order = MIN_LPC_ORDER
			psDec.psNLSF_CB[0] = NLSF_CB0_10
			psDec.psNLSF_CB[1] = NLSF_CB1_10
		} else {
			psDec.LPC_order = MAX_LPC_ORDER
			psDec.psNLSF_CB[0] = NLSF_CB0_16
			psDec.psNLSF_CB[1] = NLSF_CB1_16
		}

		psDec.lagPrev = 100
		psDec.LastGainIndex = 1
		psDec.prev_sigtype = 0
		psDec.first_frame_after_reset = 1

		if fs_kHz == 24 {
			psDec.HP_A = Dec_A_HP_24
			psDec.HP_B = Dec_B_HP_24
		} else if fs_kHz == 16 {
			psDec.HP_A = Dec_A_HP_16
			psDec.HP_B = Dec_B_HP_16
		} else if fs_kHz == 12 {
			psDec.HP_A = Dec_A_HP_12
			psDec.HP_B = Dec_B_HP_12
		} else if fs_kHz == 8 {
			psDec.HP_A = Dec_A_HP_8
			psDec.HP_B = Dec_B_HP_8
		}
	}
}

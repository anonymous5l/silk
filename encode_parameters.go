package silk

func encode_parameters(psEncC *encoder_state, psEncCtrlC *encoder_control, psRC *range_coder_state, q *slice[int8]) {
	var (
		i, k, typeOffset int32
		psNLSF_CB        *NLSF_CB_struct
	)

	if psEncC.nFramesInPayloadBuf == 0 {
		for i = 0; i < 3; i++ {
			if SamplingRates_table[i] == psEncC.fs_kHz {
				break
			}
		}
		range_encoder(psRC, i, SamplingRates_CDF)
	}

	typeOffset = 2*psEncCtrlC.sigtype + psEncCtrlC.QuantOffsetType
	if psEncC.nFramesInPayloadBuf == 0 {
		range_encoder(psRC, typeOffset, type_offset_CDF)
	} else {
		range_encoder(psRC, typeOffset, type_offset_joint_CDF[psEncC.typeOffsetPrev])
	}
	psEncC.typeOffsetPrev = typeOffset

	if psEncC.nFramesInPayloadBuf == 0 {
		range_encoder(psRC, psEncCtrlC.GainsIndices.idx(0), gain_CDF[psEncCtrlC.sigtype])
	} else {
		range_encoder(psRC, psEncCtrlC.GainsIndices.idx(0), delta_gain_CDF)
	}

	for i = 1; i < NB_SUBFR; i++ {
		range_encoder(psRC, psEncCtrlC.GainsIndices.idx(int(i)), delta_gain_CDF)
	}

	psNLSF_CB = &psEncC.psNLSF_CB[psEncCtrlC.sigtype]
	range_encoder_multi(psRC, psEncCtrlC.NLSFIndices, psNLSF_CB.StartPtr, psNLSF_CB.nStages)

	range_encoder(psRC, psEncCtrlC.NLSFInterpCoef_Q2, NLSF_interpolation_factor_CDF)

	if psEncCtrlC.sigtype == SIG_TYPE_VOICED {

		if psEncC.fs_kHz == 8 {
			range_encoder(psRC, psEncCtrlC.lagIndex, pitch_lag_NB_CDF)
		} else if psEncC.fs_kHz == 12 {
			range_encoder(psRC, psEncCtrlC.lagIndex, pitch_lag_MB_CDF)
		} else if psEncC.fs_kHz == 16 {
			range_encoder(psRC, psEncCtrlC.lagIndex, pitch_lag_WB_CDF)
		} else {
			range_encoder(psRC, psEncCtrlC.lagIndex, pitch_lag_SWB_CDF)
		}

		if psEncC.fs_kHz == 8 {
			range_encoder(psRC, psEncCtrlC.contourIndex, pitch_contour_NB_CDF)
		} else {
			range_encoder(psRC, psEncCtrlC.contourIndex, pitch_contour_CDF)
		}

		range_encoder(psRC, psEncCtrlC.PERIndex, LTP_per_index_CDF)

		for k = 0; k < NB_SUBFR; k++ {
			range_encoder(psRC, psEncCtrlC.LTPIndex.idx(int(k)), LTP_gain_CDF_ptrs[psEncCtrlC.PERIndex])
		}

		range_encoder(psRC, psEncCtrlC.LTP_scaleIndex, LTPscale_CDF)
	}

	range_encoder(psRC, psEncCtrlC.Seed, Seed_CDF)

	encode_pulses(psRC, psEncCtrlC.sigtype, psEncCtrlC.QuantOffsetType, q, psEncC.frame_length)

	range_encoder(psRC, psEncC.vadFlag, vadflag_CDF)
}

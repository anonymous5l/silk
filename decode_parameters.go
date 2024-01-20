package silk

func decode_parameters(psDec *decoder_state, psDecCtrl *decoder_control, q *slice[int32], fullDecoding int32) {
	var (
		i, k, Ix, fs_kHz_dec, nBytesUsed int32
		Ixs                              [NB_SUBFR]int32
		GainsIndices                     = alloc[int32](NB_SUBFR)
		NLSFIndices                      = alloc[int32](NLSF_MSVQ_MAX_CB_STAGES)
		pNLSF_Q15                        = alloc[int32](MAX_LPC_ORDER)
		pNLSF0_Q15                       = alloc[int32](MAX_LPC_ORDER)

		cbk_ptr_Q14 []int16

		psNLSF_CB *NLSF_CB_struct
	)

	psRC := psDec.sRC

	if psDec.nFramesDecoded == 0 {
		range_decoder(&Ix, psRC, SamplingRates_CDF, SamplingRates_offset)

		if Ix < 0 || Ix > 3 {
			psRC.error = RANGE_CODER_ILLEGAL_SAMPLING_RATE
			return
		}
		fs_kHz_dec = SamplingRates_table[Ix]
		decoder_set_fs(psDec, fs_kHz_dec)
	}

	if psDec.nFramesDecoded == 0 {
		range_decoder(&Ix, psRC, type_offset_CDF, type_offset_CDF_offset)
	} else {
		range_decoder(&Ix, psRC, type_offset_joint_CDF[psDec.typeOffsetPrev],
			type_offset_CDF_offset)
	}

	psDecCtrl.sigtype = RSHIFT(Ix, 1)
	psDecCtrl.QuantOffsetType = Ix & 1
	psDec.typeOffsetPrev = Ix

	if psDec.nFramesDecoded == 0 {
		range_decoder(GainsIndices.ptr(0), psRC, gain_CDF[psDecCtrl.sigtype], gain_CDF_offset)
	} else {
		range_decoder(GainsIndices.ptr(0), psRC, delta_gain_CDF, delta_gain_CDF_offset)
	}

	for i = 1; i < NB_SUBFR; i++ {
		range_decoder(GainsIndices.ptr(int(i)), psRC, delta_gain_CDF, delta_gain_CDF_offset)
	}

	gains_dequant(psDecCtrl.Gains_Q16, GainsIndices.off(0), &psDec.LastGainIndex, psDec.nFramesDecoded)
	psNLSF_CB = &psDec.psNLSF_CB[psDecCtrl.sigtype]

	range_decoder_multi(NLSFIndices, psRC, psNLSF_CB.StartPtr, psNLSF_CB.MiddleIx, psNLSF_CB.nStages)

	NLSF_MSVQ_decode(pNLSF_Q15, psNLSF_CB, NLSFIndices, psDec.LPC_order)

	range_decoder(&psDecCtrl.NLSFInterpCoef_Q2, psRC, NLSF_interpolation_factor_CDF,
		NLSF_interpolation_factor_offset)

	if psDec.first_frame_after_reset == 1 {
		psDecCtrl.NLSFInterpCoef_Q2 = 4
	}

	if fullDecoding != 0 {
		NLSF2A_stable(psDecCtrl.PredCoef_Q12[1], pNLSF_Q15, psDec.LPC_order)
		if psDecCtrl.NLSFInterpCoef_Q2 < 4 {
			for i = 0; i < psDec.LPC_order; i++ {
				*pNLSF0_Q15.ptr(int(i)) = psDec.prevNLSF_Q15.idx(int(i)) + RSHIFT(MUL(psDecCtrl.NLSFInterpCoef_Q2,
					pNLSF_Q15.idx(int(i))-psDec.prevNLSF_Q15.idx(int(i))), 2)
			}

			NLSF2A_stable(psDecCtrl.PredCoef_Q12[0], pNLSF0_Q15, psDec.LPC_order)
		} else {
			psDecCtrl.PredCoef_Q12[1].copy(psDecCtrl.PredCoef_Q12[0], int(psDec.LPC_order))
		}
	}

	pNLSF_Q15.copy(psDec.prevNLSF_Q15, int(psDec.LPC_order))

	if psDec.lossCnt != 0 {
		bwexpander(psDecCtrl.PredCoef_Q12[0], psDec.LPC_order, BWE_AFTER_LOSS_Q16)
		bwexpander(psDecCtrl.PredCoef_Q12[1], psDec.LPC_order, BWE_AFTER_LOSS_Q16)
	}

	if psDecCtrl.sigtype == SIG_TYPE_VOICED {
		if psDec.fs_kHz == 8 {
			range_decoder(&Ixs[0], psRC, pitch_lag_NB_CDF, pitch_lag_NB_CDF_offset)
		} else if psDec.fs_kHz == 12 {
			range_decoder(&Ixs[0], psRC, pitch_lag_MB_CDF, pitch_lag_MB_CDF_offset)
		} else if psDec.fs_kHz == 16 {
			range_decoder(&Ixs[0], psRC, pitch_lag_WB_CDF, pitch_lag_WB_CDF_offset)
		} else {
			range_decoder(&Ixs[0], psRC, pitch_lag_SWB_CDF, pitch_lag_SWB_CDF_offset)
		}

		if psDec.fs_kHz == 8 {
			range_decoder(&Ixs[1], psRC, pitch_contour_NB_CDF, pitch_contour_NB_CDF_offset)
		} else {
			range_decoder(&Ixs[1], psRC, pitch_contour_CDF, pitch_contour_CDF_offset)
		}

		decode_pitch(Ixs[0], Ixs[1], psDecCtrl.pitchL, psDec.fs_kHz)

		range_decoder(&psDecCtrl.PERIndex, psRC, LTP_per_index_CDF,
			LTP_per_index_CDF_offset)

		cbk_ptr_Q14 = LTP_vq_ptrs_Q14[psDecCtrl.PERIndex]

		for k = 0; k < NB_SUBFR; k++ {
			range_decoder(&Ix, psRC, LTP_gain_CDF_ptrs[psDecCtrl.PERIndex],
				LTP_gain_CDF_offsets[psDecCtrl.PERIndex])

			for i = 0; i < LTP_ORDER; i++ {
				*psDecCtrl.LTPCoef_Q14.ptr(int(k*LTP_ORDER + i)) = cbk_ptr_Q14[Ix*LTP_ORDER+i]
			}
		}

		range_decoder(&Ix, psRC, LTPscale_CDF, LTPscale_offset)
		psDecCtrl.LTP_scale_Q14 = LTPScales_table_Q14[Ix]
	} else {
		memset(psDecCtrl.pitchL, 0, NB_SUBFR)
		memset(psDecCtrl.LTPCoef_Q14, 0, LTP_ORDER*NB_SUBFR)
		psDecCtrl.PERIndex = 0
		psDecCtrl.LTP_scale_Q14 = 0
	}

	range_decoder(&Ix, psRC, Seed_CDF, Seed_offset)
	psDecCtrl.Seed = Ix
	decode_pulses(psRC, psDecCtrl, q, psDec.frame_length)

	range_decoder(&psDec.vadFlag, psRC, vadflag_CDF, vadflag_offset)

	range_decoder(&psDec.FrameTermination, psRC, FrameTermination_CDF, FrameTermination_offset)

	range_coder_get_length(psRC, &nBytesUsed)
	psDec.nBytesLeft = psRC.bufferLength - nBytesUsed
	if psDec.nBytesLeft < 0 {
		psRC.error = RANGE_CODER_READ_BEYOND_BUFFER
	}

	if psDec.nBytesLeft == 0 {
		range_coder_check_after_decoding(psRC)
	}
}

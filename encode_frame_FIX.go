package silk

func encode_frame_FIX(psEnc *encoder_state_FIX, pCode *slice[byte], pnBytesOut *int16, pIn *slice[int16]) int {
	sEncCtrl := &encoder_control_FIX{}
	sEncCtrl.init()

	var (
		nBytes                                int32
		ret                                   int
		x_frame, res_pitch_frame              *slice[int16]
		xfw                                   = alloc[int16](MAX_FRAME_LENGTH)
		pIn_HP                                = alloc[int16](MAX_FRAME_LENGTH)
		res_pitch                             = alloc[int16](2*MAX_FRAME_LENGTH + LA_PITCH_MAX)
		LBRR_idx, frame_terminator, SNR_dB_Q7 int32
		LBRRpayload                           = alloc[byte](MAX_ARITHM_BYTES)
		nBytesLBRR                            int16
	)

	sEncCtrl.sCmn.Seed = psEnc.sCmn.frameCounter & 3
	psEnc.sCmn.frameCounter++
	x_frame = psEnc.x_buf.off(int(psEnc.sCmn.frame_length))
	res_pitch_frame = res_pitch.off(int(psEnc.sCmn.frame_length))

	ret = VAD_GetSA_Q8(psEnc.sCmn.sVAD, &psEnc.speech_activity_Q8, &SNR_dB_Q7,
		sEncCtrl.input_quality_bands_Q15, &sEncCtrl.input_tilt_Q15,
		pIn, psEnc.sCmn.frame_length)

	HP_variable_cutoff_FIX(psEnc, sEncCtrl, pIn_HP, pIn)

	LP_variable_cutoff(psEnc.sCmn.sLP, x_frame.off(int(LA_SHAPE_MS*psEnc.sCmn.fs_kHz)), pIn_HP, psEnc.sCmn.frame_length)

	find_pitch_lags_FIX(psEnc, sEncCtrl, res_pitch, x_frame)

	noise_shape_analysis_FIX(psEnc, sEncCtrl, res_pitch_frame, x_frame)

	prefilter_FIX(psEnc, sEncCtrl, xfw, x_frame)

	find_pred_coefs_FIX(psEnc, sEncCtrl, res_pitch)

	process_gains_FIX(psEnc, sEncCtrl)

	nBytesLBRR = MAX_ARITHM_BYTES
	LBRR_encode_FIX(psEnc, sEncCtrl, LBRRpayload, &nBytesLBRR, xfw)

	if psEnc.sCmn.nStatesDelayedDecision > 1 || psEnc.sCmn.warping_Q16 > 0 {
		NSQ_del_dec(psEnc.sCmn, &sEncCtrl.sCmn, psEnc.sCmn.sNSQ, xfw,
			psEnc.sCmn.q, sEncCtrl.sCmn.NLSFInterpCoef_Q2,
			sEncCtrl.PredCoef_Q12, sEncCtrl.LTPCoef_Q14, sEncCtrl.AR2_Q13, sEncCtrl.HarmShapeGain_Q14,
			sEncCtrl.Tilt_Q14, sEncCtrl.LF_shp_Q14, sEncCtrl.Gains_Q16, sEncCtrl.Lambda_Q10,
			sEncCtrl.LTP_scale_Q14)
	} else {
		NSQ(psEnc.sCmn, &sEncCtrl.sCmn, psEnc.sCmn.sNSQ, xfw,
			psEnc.sCmn.q, sEncCtrl.sCmn.NLSFInterpCoef_Q2,
			sEncCtrl.PredCoef_Q12, sEncCtrl.LTPCoef_Q14, sEncCtrl.AR2_Q13, sEncCtrl.HarmShapeGain_Q14,
			sEncCtrl.Tilt_Q14, sEncCtrl.LF_shp_Q14, sEncCtrl.Gains_Q16, sEncCtrl.Lambda_Q10,
			sEncCtrl.LTP_scale_Q14)
	}

	if psEnc.speech_activity_Q8 < FIX_CONST(SPEECH_ACTIVITY_DTX_THRES, 8) {
		psEnc.sCmn.vadFlag = NO_VOICE_ACTIVITY
		psEnc.sCmn.noSpeechCounter++
		if psEnc.sCmn.noSpeechCounter > NO_SPEECH_FRAMES_BEFORE_DTX {
			psEnc.sCmn.inDTX = 1
		}
		if psEnc.sCmn.noSpeechCounter > MAX_CONSECUTIVE_DTX+NO_SPEECH_FRAMES_BEFORE_DTX {
			psEnc.sCmn.noSpeechCounter = NO_SPEECH_FRAMES_BEFORE_DTX
			psEnc.sCmn.inDTX = 0
		}
	} else {
		psEnc.sCmn.noSpeechCounter = 0
		psEnc.sCmn.inDTX = 0
		psEnc.sCmn.vadFlag = VOICE_ACTIVITY
	}

	if psEnc.sCmn.nFramesInPayloadBuf == 0 {
		range_enc_init(psEnc.sCmn.sRC)
		psEnc.sCmn.nBytesInPayloadBuf = 0
	}

	encode_parameters(psEnc.sCmn, &sEncCtrl.sCmn, psEnc.sCmn.sRC, psEnc.sCmn.q)

	psEnc.x_buf.off(int(psEnc.sCmn.frame_length)).copy(psEnc.x_buf, int(psEnc.sCmn.frame_length+LA_SHAPE_MS*psEnc.sCmn.fs_kHz))

	psEnc.sCmn.prev_sigtype = sEncCtrl.sCmn.sigtype
	psEnc.sCmn.prevLag = sEncCtrl.sCmn.pitchL.idx(NB_SUBFR - 1)
	psEnc.sCmn.first_frame_after_reset = 0

	if psEnc.sCmn.sRC.error != 0 {
		psEnc.sCmn.nFramesInPayloadBuf = 0
	} else {
		psEnc.sCmn.nFramesInPayloadBuf++
	}

	if psEnc.sCmn.nFramesInPayloadBuf*FRAME_LENGTH_MS >= psEnc.sCmn.PacketSize_ms {

		LBRR_idx = (psEnc.sCmn.oldest_LBRR_idx + 1) & LBRR_IDX_MASK

		frame_terminator = LAST_FRAME
		if psEnc.sCmn.LBRR_buffer[LBRR_idx].usage == ADD_LBRR_TO_PLUS1 {
			frame_terminator = LBRR_VER1
		}
		if psEnc.sCmn.LBRR_buffer[psEnc.sCmn.oldest_LBRR_idx].usage == ADD_LBRR_TO_PLUS2 {
			frame_terminator = LBRR_VER2
			LBRR_idx = psEnc.sCmn.oldest_LBRR_idx
		}

		range_encoder(psEnc.sCmn.sRC, frame_terminator, FrameTermination_CDF)

		range_coder_get_length(psEnc.sCmn.sRC, &nBytes)

		if int32(*pnBytesOut) >= nBytes {
			range_enc_wrap_up(psEnc.sCmn.sRC)
			psEnc.sCmn.sRC.buffer.copy(pCode, int(nBytes))

			if frame_terminator > MORE_FRAMES &&
				int32(*pnBytesOut) >= nBytes+psEnc.sCmn.LBRR_buffer[LBRR_idx].nBytes {
				psEnc.sCmn.LBRR_buffer[LBRR_idx].payload.copy(
					pCode.off(int(nBytes)),
					int(psEnc.sCmn.LBRR_buffer[LBRR_idx].nBytes))
				nBytes += psEnc.sCmn.LBRR_buffer[LBRR_idx].nBytes
			}

			*pnBytesOut = int16(nBytes)

			LBRRpayload.copy(psEnc.sCmn.LBRR_buffer[psEnc.sCmn.oldest_LBRR_idx].payload, int(nBytesLBRR))

			psEnc.sCmn.LBRR_buffer[psEnc.sCmn.oldest_LBRR_idx].nBytes = int32(nBytesLBRR)
			psEnc.sCmn.LBRR_buffer[psEnc.sCmn.oldest_LBRR_idx].usage = sEncCtrl.sCmn.LBRR_usage
			psEnc.sCmn.oldest_LBRR_idx = (psEnc.sCmn.oldest_LBRR_idx + 1) & LBRR_IDX_MASK

		} else {
			*pnBytesOut = 0
			nBytes = 0
			ret = ENC_PAYLOAD_BUF_TOO_SHORT
		}

		psEnc.sCmn.nFramesInPayloadBuf = 0
	} else {
		*pnBytesOut = 0

		frame_terminator = MORE_FRAMES
		range_encoder(psEnc.sCmn.sRC, frame_terminator, FrameTermination_CDF)

		range_coder_get_length(psEnc.sCmn.sRC, &nBytes)
	}

	if psEnc.sCmn.sRC.error != 0 {
		ret = ENC_INTERNAL_ERROR
	}
	psEnc.BufferedInChannel_ms += DIV32(8*1000*(nBytes-psEnc.sCmn.nBytesInPayloadBuf), psEnc.sCmn.TargetRate_bps)
	psEnc.BufferedInChannel_ms -= FRAME_LENGTH_MS
	psEnc.BufferedInChannel_ms = LIMIT(psEnc.BufferedInChannel_ms, 0, 100)
	psEnc.sCmn.nBytesInPayloadBuf = nBytes

	if psEnc.speech_activity_Q8 > FIX_CONST(WB_DETECT_ACTIVE_SPEECH_LEVEL_THRES, 8) {
		psEnc.sCmn.sSWBdetect.ActiveSpeech_ms = ADD_POS_SAT32(psEnc.sCmn.sSWBdetect.ActiveSpeech_ms, FRAME_LENGTH_MS)
	}

	return ret
}

func LBRR_encode_FIX(psEnc *encoder_state_FIX, psEncCtrl *encoder_control_FIX, pCode *slice[byte],
	pnBytesOut *int16, xfw *slice[int16]) {
	var (
		TempGainsIndices                                 = alloc[int32](NB_SUBFR)
		frame_terminator                                 int32
		nBytes, nFramesInPayloadBuf                      int32
		TempGains_Q16                                    = alloc[int32](NB_SUBFR)
		typeOffset, LTP_scaleIndex, Rate_only_parameters int32
	)

	LBRR_ctrl_FIX(psEnc, &psEncCtrl.sCmn)

	if psEnc.sCmn.LBRR_enabled != 0 {
		psEncCtrl.sCmn.GainsIndices.copy(TempGainsIndices, NB_SUBFR)
		psEncCtrl.Gains_Q16.copy(TempGains_Q16, NB_SUBFR)

		typeOffset = psEnc.sCmn.typeOffsetPrev
		LTP_scaleIndex = psEncCtrl.sCmn.LTP_scaleIndex

		if psEnc.sCmn.fs_kHz == 8 {
			Rate_only_parameters = 13500
		} else if psEnc.sCmn.fs_kHz == 12 {
			Rate_only_parameters = 15500
		} else if psEnc.sCmn.fs_kHz == 16 {
			Rate_only_parameters = 17500
		} else if psEnc.sCmn.fs_kHz == 24 {
			Rate_only_parameters = 19500
		}

		if psEnc.sCmn.Complexity > 0 && psEnc.sCmn.TargetRate_bps > Rate_only_parameters {
			if psEnc.sCmn.nFramesInPayloadBuf == 0 {
				psEnc.sCmn.sNSQ.copy(psEnc.sCmn.sNSQ_LBRR)

				psEnc.sCmn.LBRRprevLastGainIndex = psEnc.sShape.LastGainIndex
				*psEncCtrl.sCmn.GainsIndices.ptr(0) = psEncCtrl.sCmn.GainsIndices.idx(0) + psEnc.sCmn.LBRR_GainIncreases
				*psEncCtrl.sCmn.GainsIndices.ptr(0) = LIMIT(psEncCtrl.sCmn.GainsIndices.idx(0), 0, N_LEVELS_QGAIN-1)
			}
			gains_dequant(psEncCtrl.Gains_Q16, psEncCtrl.sCmn.GainsIndices,
				&psEnc.sCmn.LBRRprevLastGainIndex, psEnc.sCmn.nFramesInPayloadBuf)

			if psEnc.sCmn.nStatesDelayedDecision > 1 || psEnc.sCmn.warping_Q16 > 0 {
				NSQ_del_dec(psEnc.sCmn, &psEncCtrl.sCmn, psEnc.sCmn.sNSQ_LBRR, xfw, psEnc.sCmn.q_LBRR,
					psEncCtrl.sCmn.NLSFInterpCoef_Q2, psEncCtrl.PredCoef_Q12, psEncCtrl.LTPCoef_Q14,
					psEncCtrl.AR2_Q13, psEncCtrl.HarmShapeGain_Q14, psEncCtrl.Tilt_Q14, psEncCtrl.LF_shp_Q14,
					psEncCtrl.Gains_Q16, psEncCtrl.Lambda_Q10, psEncCtrl.LTP_scale_Q14)
			} else {
				NSQ(psEnc.sCmn, &psEncCtrl.sCmn, psEnc.sCmn.sNSQ_LBRR, xfw, psEnc.sCmn.q_LBRR,
					psEncCtrl.sCmn.NLSFInterpCoef_Q2, psEncCtrl.PredCoef_Q12, psEncCtrl.LTPCoef_Q14,
					psEncCtrl.AR2_Q13, psEncCtrl.HarmShapeGain_Q14, psEncCtrl.Tilt_Q14, psEncCtrl.LF_shp_Q14,
					psEncCtrl.Gains_Q16, psEncCtrl.Lambda_Q10, psEncCtrl.LTP_scale_Q14)
			}
		} else {
			memset(psEnc.sCmn.q_LBRR, 0, int(psEnc.sCmn.frame_length))
			psEncCtrl.sCmn.LTP_scaleIndex = 0
		}
		if psEnc.sCmn.nFramesInPayloadBuf == 0 {
			range_enc_init(psEnc.sCmn.sRC_LBRR)
			psEnc.sCmn.nBytesInPayloadBuf = 0
		}

		encode_parameters(psEnc.sCmn, &psEncCtrl.sCmn,
			psEnc.sCmn.sRC_LBRR, psEnc.sCmn.q_LBRR)

		if psEnc.sCmn.sRC_LBRR.error != 0 {
			nFramesInPayloadBuf = 0
		} else {
			nFramesInPayloadBuf = psEnc.sCmn.nFramesInPayloadBuf + 1
		}

		if SMULBB(nFramesInPayloadBuf, FRAME_LENGTH_MS) >= psEnc.sCmn.PacketSize_ms {

			frame_terminator = LAST_FRAME

			range_encoder(psEnc.sCmn.sRC_LBRR, frame_terminator, FrameTermination_CDF)

			range_coder_get_length(psEnc.sCmn.sRC_LBRR, &nBytes)

			if int32(*pnBytesOut) >= nBytes {
				range_enc_wrap_up(psEnc.sCmn.sRC_LBRR)
				psEnc.sCmn.sRC_LBRR.buffer.copy(pCode, int(nBytes))

				*pnBytesOut = int16(nBytes)
			} else {
				*pnBytesOut = 0
			}
		} else {
			*pnBytesOut = 0

			frame_terminator = MORE_FRAMES
			range_encoder(psEnc.sCmn.sRC_LBRR, frame_terminator, FrameTermination_CDF)
		}

		TempGainsIndices.copy(psEncCtrl.sCmn.GainsIndices, NB_SUBFR)
		TempGains_Q16.copy(psEncCtrl.Gains_Q16, NB_SUBFR)

		psEncCtrl.sCmn.LTP_scaleIndex = LTP_scaleIndex
		psEnc.sCmn.typeOffsetPrev = typeOffset
	}
}

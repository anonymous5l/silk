package silk

import (
	"math"
)

func control_encoder_FIX(psEnc *encoder_state_FIX, PacketSize_ms int32,
	TargetRate_bps int32, PacketLoss_perc int32, DTX_enabled int32, Complexity int32) int {
	var fs_kHz int32
	var ret int

	if psEnc.sCmn.controlled_since_last_payload != 0 {
		if psEnc.sCmn.API_fs_Hz != psEnc.sCmn.prev_API_fs_Hz && psEnc.sCmn.fs_kHz > 0 {
			ret += setup_resamplers_FIX(psEnc, psEnc.sCmn.fs_kHz)
		}
		return ret
	}

	fs_kHz = control_audio_bandwidth(psEnc.sCmn, TargetRate_bps)

	ret += setup_resamplers_FIX(psEnc, fs_kHz)

	ret += setup_packetsize_FIX(psEnc, PacketSize_ms)

	ret += setup_fs_FIX(psEnc, fs_kHz)

	ret += setup_complexity(psEnc.sCmn, Complexity)

	ret += setup_rate_FIX(psEnc, TargetRate_bps)

	if (PacketLoss_perc < 0) || (PacketLoss_perc > 100) {
		ret = ENC_INVALID_LOSS_RATE
	}
	psEnc.sCmn.PacketLoss_perc = PacketLoss_perc

	ret += setup_LBRR_FIX(psEnc)

	if DTX_enabled < 0 || DTX_enabled > 1 {
		ret = ENC_INVALID_DTX_SETTING
	}
	psEnc.sCmn.useDTX = DTX_enabled
	psEnc.sCmn.controlled_since_last_payload = 1

	return ret
}

func LBRR_ctrl_FIX(psEnc *encoder_state_FIX, psEncCtrlC *encoder_control) {
	var LBRR_usage int32

	if psEnc.sCmn.LBRR_enabled != 0 {

		LBRR_usage = NO_LBRR
		if psEnc.speech_activity_Q8 > FIX_CONST(LBRR_SPEECH_ACTIVITY_THRES, 8) && psEnc.sCmn.PacketLoss_perc > LBRR_LOSS_THRES { // nb! maybe multiply loss prob and speech activity
			LBRR_usage = ADD_LBRR_TO_PLUS1
		}
		psEncCtrlC.LBRR_usage = LBRR_usage
	} else {
		psEncCtrlC.LBRR_usage = NO_LBRR
	}
}

func setup_resamplers_FIX(psEnc *encoder_state_FIX, fs_kHz int32) int {
	ret := NO_ERROR

	if psEnc.sCmn.fs_kHz != fs_kHz || psEnc.sCmn.prev_API_fs_Hz != psEnc.sCmn.API_fs_Hz {

		if psEnc.sCmn.fs_kHz == 0 {
			ret += resampler_init(&psEnc.sCmn.resampler_state, psEnc.sCmn.API_fs_Hz, fs_kHz*1000)
		} else {
			x_buf_API_fs_Hz := alloc[int16]((2*MAX_FRAME_LENGTH + LA_SHAPE_MAX) * (MAX_API_FS_KHZ / 8))

			nSamples_temp := LSHIFT(psEnc.sCmn.frame_length, 1) + LA_SHAPE_MS*psEnc.sCmn.fs_kHz

			if SMULBB(fs_kHz, 1000) < psEnc.sCmn.API_fs_Hz && psEnc.sCmn.fs_kHz != 0 {

				var temp_resampler_state resampler_state_struct

				ret += resampler_init(&temp_resampler_state, SMULBB(psEnc.sCmn.fs_kHz, 1000), psEnc.sCmn.API_fs_Hz)

				ret += resampler(&temp_resampler_state, x_buf_API_fs_Hz, psEnc.x_buf, nSamples_temp)

				nSamples_temp = DIV32_16(nSamples_temp*psEnc.sCmn.API_fs_Hz, int16(SMULBB(psEnc.sCmn.fs_kHz, 1000)))

				ret += resampler_init(&psEnc.sCmn.resampler_state, psEnc.sCmn.API_fs_Hz, SMULBB(fs_kHz, 1000))

			} else {
				psEnc.x_buf.copy(x_buf_API_fs_Hz, int(nSamples_temp))
			}

			if 1000*fs_kHz != psEnc.sCmn.API_fs_Hz {
				ret += resampler(&psEnc.sCmn.resampler_state, psEnc.x_buf, x_buf_API_fs_Hz, nSamples_temp)
			}
		}
	}

	psEnc.sCmn.prev_API_fs_Hz = psEnc.sCmn.API_fs_Hz

	return ret
}

func setup_packetsize_FIX(psEnc *encoder_state_FIX, PacketSize_ms int32) int {
	ret := NO_ERROR

	if (PacketSize_ms != 20) &&
		(PacketSize_ms != 40) &&
		(PacketSize_ms != 60) &&
		(PacketSize_ms != 80) &&
		(PacketSize_ms != 100) {
		ret = ENC_PACKET_SIZE_NOT_SUPPORTED
	} else {
		if PacketSize_ms != psEnc.sCmn.PacketSize_ms {
			psEnc.sCmn.PacketSize_ms = PacketSize_ms

			LBRR_reset(psEnc.sCmn)
		}
	}
	return ret
}

func setup_fs_FIX(psEnc *encoder_state_FIX, fs_kHz int32) int {
	ret := NO_ERROR

	if psEnc.sCmn.fs_kHz != fs_kHz {
		psEnc.sShape = &shape_state_FIX{}
		psEnc.sPrefilt = &prefilter_state_FIX{}
		psEnc.sPrefilt.init()
		psEnc.sPred = &predict_state_FIX{}
		psEnc.sPred.init()
		psEnc.sCmn.sNSQ = &nsq_state{}
		psEnc.sCmn.sNSQ.init()
		psEnc.sCmn.sNSQ_LBRR.init()

		psEnc.sCmn.sLP = &LP_state{}
		psEnc.sCmn.sLP.init()
		if psEnc.sCmn.sLP.mode == 1 {
			psEnc.sCmn.sLP.transition_frame_no = 1
		} else {
			psEnc.sCmn.sLP.transition_frame_no = 0
		}

		psEnc.sCmn.inputBufIx = 0
		psEnc.sCmn.nFramesInPayloadBuf = 0
		psEnc.sCmn.nBytesInPayloadBuf = 0
		psEnc.sCmn.oldest_LBRR_idx = 0
		psEnc.sCmn.TargetRate_bps = 0
		psEnc.sCmn.prevLag = 100
		psEnc.sCmn.prev_sigtype = SIG_TYPE_UNVOICED
		psEnc.sCmn.first_frame_after_reset = 1
		psEnc.sPrefilt.lagPrev = 100
		psEnc.sShape.LastGainIndex = 1
		psEnc.sCmn.sNSQ.lagPrev = 100
		psEnc.sCmn.sNSQ.prev_inv_gain_Q16 = 65536
		psEnc.sCmn.sNSQ_LBRR.prev_inv_gain_Q16 = 65536

		psEnc.sCmn.fs_kHz = fs_kHz
		if psEnc.sCmn.fs_kHz == 8 {
			psEnc.sCmn.predictLPCOrder = MIN_LPC_ORDER
			psEnc.sCmn.psNLSF_CB[0] = NLSF_CB0_10
			psEnc.sCmn.psNLSF_CB[1] = NLSF_CB1_10
		} else {
			psEnc.sCmn.predictLPCOrder = MAX_LPC_ORDER
			psEnc.sCmn.psNLSF_CB[0] = NLSF_CB0_16
			psEnc.sCmn.psNLSF_CB[1] = NLSF_CB1_16
		}
		psEnc.sCmn.frame_length = SMULBB(FRAME_LENGTH_MS, fs_kHz)
		psEnc.sCmn.subfr_length = DIV32_16(psEnc.sCmn.frame_length, NB_SUBFR)
		psEnc.sCmn.la_pitch = SMULBB(LA_PITCH_MS, fs_kHz)
		psEnc.sPred.min_pitch_lag = SMULBB(3, fs_kHz)
		psEnc.sPred.max_pitch_lag = SMULBB(18, fs_kHz)
		psEnc.sPred.pitch_LPC_win_length = SMULBB(FIND_PITCH_LPC_WIN_MS, fs_kHz)
		if psEnc.sCmn.fs_kHz == 24 {
			psEnc.mu_LTP_Q8 = FIX_CONST(MU_LTP_QUANT_SWB, 8)
			psEnc.sCmn.bitrate_threshold_up = math.MaxInt32
			psEnc.sCmn.bitrate_threshold_down = SWB2WB_BITRATE_BPS
		} else if psEnc.sCmn.fs_kHz == 16 {
			psEnc.mu_LTP_Q8 = FIX_CONST(MU_LTP_QUANT_WB, 8)
			psEnc.sCmn.bitrate_threshold_up = WB2SWB_BITRATE_BPS
			psEnc.sCmn.bitrate_threshold_down = WB2MB_BITRATE_BPS
		} else if psEnc.sCmn.fs_kHz == 12 {
			psEnc.mu_LTP_Q8 = FIX_CONST(MU_LTP_QUANT_MB, 8)
			psEnc.sCmn.bitrate_threshold_up = MB2WB_BITRATE_BPS
			psEnc.sCmn.bitrate_threshold_down = MB2NB_BITRATE_BPS
		} else {
			psEnc.mu_LTP_Q8 = FIX_CONST(MU_LTP_QUANT_NB, 8)
			psEnc.sCmn.bitrate_threshold_up = NB2MB_BITRATE_BPS
			psEnc.sCmn.bitrate_threshold_down = 0
		}
		psEnc.sCmn.fs_kHz_changed = 1

	}
	return (ret)
}

func setup_rate_FIX(psEnc *encoder_state_FIX, TargetRate_bps int32) int {
	var (
		ret       int
		k         int32
		frac_Q6   int32
		rateTable []int32
	)

	if TargetRate_bps != psEnc.sCmn.TargetRate_bps {
		psEnc.sCmn.TargetRate_bps = TargetRate_bps

		if psEnc.sCmn.fs_kHz == 8 {
			rateTable = TargetRate_table_NB
		} else if psEnc.sCmn.fs_kHz == 12 {
			rateTable = TargetRate_table_MB
		} else if psEnc.sCmn.fs_kHz == 16 {
			rateTable = TargetRate_table_WB
		} else {
			rateTable = TargetRate_table_SWB
		}
		for k = 1; k < TARGET_RATE_TAB_SZ; k++ {
			if TargetRate_bps <= rateTable[k] {
				frac_Q6 = DIV32(LSHIFT(TargetRate_bps-rateTable[k-1], 6),
					rateTable[k]-rateTable[k-1])
				psEnc.SNR_dB_Q7 = LSHIFT(SNR_table_Q1[k-1], 6) + MUL(frac_Q6, SNR_table_Q1[k]-SNR_table_Q1[k-1])
				break
			}
		}
	}
	return ret
}

func setup_LBRR_FIX(psEnc *encoder_state_FIX) int {
	ret := NO_ERROR

	var LBRRRate_thres_bps int32

	if psEnc.sCmn.useInBandFEC < 0 || psEnc.sCmn.useInBandFEC > 1 {
		ret = ENC_INVALID_INBAND_FEC_SETTING
	}

	psEnc.sCmn.LBRR_enabled = psEnc.sCmn.useInBandFEC
	if psEnc.sCmn.fs_kHz == 8 {
		LBRRRate_thres_bps = INBAND_FEC_MIN_RATE_BPS - 9000
	} else if psEnc.sCmn.fs_kHz == 12 {
		LBRRRate_thres_bps = INBAND_FEC_MIN_RATE_BPS - 6000
	} else if psEnc.sCmn.fs_kHz == 16 {
		LBRRRate_thres_bps = INBAND_FEC_MIN_RATE_BPS - 3000
	} else {
		LBRRRate_thres_bps = INBAND_FEC_MIN_RATE_BPS
	}

	if psEnc.sCmn.TargetRate_bps >= LBRRRate_thres_bps {
		psEnc.sCmn.LBRR_GainIncreases = max(8-RSHIFT(psEnc.sCmn.PacketLoss_perc, 1), 0)

		if psEnc.sCmn.LBRR_enabled != 0 && psEnc.sCmn.PacketLoss_perc > LBRR_LOSS_THRES {
			psEnc.inBandFEC_SNR_comp_Q8 = FIX_CONST(6.0, 8) - LSHIFT(psEnc.sCmn.LBRR_GainIncreases, 7)
		} else {
			psEnc.inBandFEC_SNR_comp_Q8 = 0
			psEnc.sCmn.LBRR_enabled = 0
		}
	} else {
		psEnc.inBandFEC_SNR_comp_Q8 = 0
		psEnc.sCmn.LBRR_enabled = 0
	}

	return ret
}

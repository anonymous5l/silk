package silk

func control_audio_bandwidth(psEncC *encoder_state, TargetRate_bps int32) int32 {
	var fs_kHz int32

	fs_kHz = psEncC.fs_kHz
	if fs_kHz == 0 {
		if TargetRate_bps >= SWB2WB_BITRATE_BPS {
			fs_kHz = 24
		} else if TargetRate_bps >= WB2MB_BITRATE_BPS {
			fs_kHz = 16
		} else if TargetRate_bps >= MB2NB_BITRATE_BPS {
			fs_kHz = 12
		} else {
			fs_kHz = 8
		}

		fs_kHz = min(fs_kHz, DIV32_16(psEncC.API_fs_Hz, 1000))
		fs_kHz = min(fs_kHz, psEncC.maxInternal_fs_kHz)
	} else if SMULBB(fs_kHz, 1000) > psEncC.API_fs_Hz || fs_kHz > psEncC.maxInternal_fs_kHz {

		fs_kHz = DIV32_16(psEncC.API_fs_Hz, 1000)
		fs_kHz = min(fs_kHz, psEncC.maxInternal_fs_kHz)
	} else {

		if psEncC.API_fs_Hz > 8000 {

			psEncC.bitrateDiff += MUL(psEncC.PacketSize_ms, TargetRate_bps-psEncC.bitrate_threshold_down)
			psEncC.bitrateDiff = min(psEncC.bitrateDiff, 0)

			if psEncC.vadFlag == NO_VOICE_ACTIVITY {

				if (psEncC.sLP.transition_frame_no == 0) &&
					(psEncC.bitrateDiff <= -ACCUM_BITS_DIFF_THRESHOLD ||
						(psEncC.sSWBdetect.WB_detected*psEncC.fs_kHz == 24)) {
					psEncC.sLP.transition_frame_no = 1
					psEncC.sLP.mode = 0
				} else if (psEncC.sLP.transition_frame_no >= TRANSITION_FRAMES_DOWN) &&
					(psEncC.sLP.mode == 0) {
					psEncC.sLP.transition_frame_no = 0
					psEncC.bitrateDiff = 0

					if psEncC.fs_kHz == 24 {
						fs_kHz = 16
					} else if psEncC.fs_kHz == 16 {
						fs_kHz = 12
					} else {

						fs_kHz = 8
					}
				}

				if ((psEncC.fs_kHz*1000 < psEncC.API_fs_Hz) &&
					(TargetRate_bps >= psEncC.bitrate_threshold_up) &&
					(psEncC.sSWBdetect.WB_detected*psEncC.fs_kHz < 16)) &&
					(((psEncC.fs_kHz == 16) && (psEncC.maxInternal_fs_kHz >= 24)) ||
						((psEncC.fs_kHz == 12) && (psEncC.maxInternal_fs_kHz >= 16)) ||
						((psEncC.fs_kHz == 8) && (psEncC.maxInternal_fs_kHz >= 12))) &&
					(psEncC.sLP.transition_frame_no == 0) {
					psEncC.sLP.mode = 1

					psEncC.bitrateDiff = 0

					if psEncC.fs_kHz == 8 {
						fs_kHz = 12
					} else if psEncC.fs_kHz == 12 {
						fs_kHz = 16
					} else {

						fs_kHz = 24
					}
				}
			}
		}

		if (psEncC.sLP.mode == 1) &&
			(psEncC.sLP.transition_frame_no >= TRANSITION_FRAMES_UP) &&
			(psEncC.vadFlag == NO_VOICE_ACTIVITY) {

			psEncC.sLP.transition_frame_no = 0

			psEncC.sLP.In_LP_State[0] = 0
			psEncC.sLP.In_LP_State[1] = 0
		}

	}

	return fs_kHz
}

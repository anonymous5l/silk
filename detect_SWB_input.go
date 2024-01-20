package silk

func detect_SWB_input(psSWBdetect *detect_SWB_state, samplesIn *slice[int16], nSamplesIn int32) {
	var (
		HP_8_kHz_len, i, shift int32
		in_HP_8_kHz            = alloc[int16](MAX_FRAME_LENGTH)
		energy_32              int32
	)

	HP_8_kHz_len = min(nSamplesIn, MAX_FRAME_LENGTH)
	HP_8_kHz_len = max(HP_8_kHz_len, 0)

	biquad(samplesIn, SWB_detect_B_HP_Q13[0], SWB_detect_A_HP_Q13[0],
		psSWBdetect.S_HP_8_kHz[0], in_HP_8_kHz, HP_8_kHz_len)
	for i = 1; i < NB_SOS; i++ {
		biquad(in_HP_8_kHz, SWB_detect_B_HP_Q13[i], SWB_detect_A_HP_Q13[i],
			psSWBdetect.S_HP_8_kHz[i], in_HP_8_kHz, HP_8_kHz_len)
	}

	sum_sqr_shift(&energy_32, &shift, in_HP_8_kHz, HP_8_kHz_len)

	if energy_32 > RSHIFT(SMULBB(HP_8_KHZ_THRES, HP_8_kHz_len), shift) {
		psSWBdetect.ConsecSmplsAboveThres += nSamplesIn
		if psSWBdetect.ConsecSmplsAboveThres > CONCEC_SWB_SMPLS_THRES {
			psSWBdetect.SWB_detected = 1
		}
	} else {
		psSWBdetect.ConsecSmplsAboveThres -= nSamplesIn
		psSWBdetect.ConsecSmplsAboveThres = max(psSWBdetect.ConsecSmplsAboveThres, 0)
	}

	if (psSWBdetect.ActiveSpeech_ms > WB_DETECT_ACTIVE_SPEECH_MS_THRES) && (psSWBdetect.SWB_detected == 0) {
		psSWBdetect.WB_detected = 1
	}
}

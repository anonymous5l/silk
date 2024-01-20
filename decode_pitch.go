package silk

func decode_pitch(lagIndex int32, contourIndex int32, pitch_lags *slice[int32], Fs_kHz int32) {
	var lag, i, min_lag int32

	min_lag = SMULBB(PITCH_EST_MIN_LAG_MS, Fs_kHz)

	lag = min_lag + lagIndex
	if Fs_kHz == 8 {
		for i = 0; i < PITCH_EST_NB_SUBFR; i++ {
			*pitch_lags.ptr(int(i)) = lag + int32(CB_lags_stage2[i][contourIndex])
		}
	} else {
		for i = 0; i < PITCH_EST_NB_SUBFR; i++ {
			*pitch_lags.ptr(int(i)) = lag + int32(CB_lags_stage3[i][contourIndex])
		}
	}
}

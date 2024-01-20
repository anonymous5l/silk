package silk

func setup_complexity(psEncC *encoder_state, Complexity int32) int {
	ret := NO_ERROR

	if Complexity == 0 {
		psEncC.Complexity = 0
		psEncC.pitchEstimationComplexity = PITCH_EST_COMPLEXITY_LC_MODE
		psEncC.pitchEstimationThreshold_Q16 = FIX_CONST(FIND_PITCH_CORRELATION_THRESHOLD_LC_MODE, 16)
		psEncC.pitchEstimationLPCOrder = 6
		psEncC.shapingLPCOrder = 8
		psEncC.la_shape = 3 * psEncC.fs_kHz
		psEncC.nStatesDelayedDecision = 1
		psEncC.useInterpolatedNLSFs = 0
		psEncC.LTPQuantLowComplexity = 1
		psEncC.NLSF_MSVQ_Survivors = MAX_NLSF_MSVQ_SURVIVORS_LC_MODE
		psEncC.warping_Q16 = 0
	} else if Complexity == 1 {
		psEncC.Complexity = 1
		psEncC.pitchEstimationComplexity = PITCH_EST_COMPLEXITY_MC_MODE
		psEncC.pitchEstimationThreshold_Q16 = FIX_CONST(FIND_PITCH_CORRELATION_THRESHOLD_MC_MODE, 16)
		psEncC.pitchEstimationLPCOrder = 12
		psEncC.shapingLPCOrder = 12
		psEncC.la_shape = 5 * psEncC.fs_kHz
		psEncC.nStatesDelayedDecision = 2
		psEncC.useInterpolatedNLSFs = 0
		psEncC.LTPQuantLowComplexity = 0
		psEncC.NLSF_MSVQ_Survivors = MAX_NLSF_MSVQ_SURVIVORS_MC_MODE
		psEncC.warping_Q16 = psEncC.fs_kHz * FIX_CONST(WARPING_MULTIPLIER, 16)
	} else if Complexity == 2 {
		psEncC.Complexity = 2
		psEncC.pitchEstimationComplexity = PITCH_EST_COMPLEXITY_HC_MODE
		psEncC.pitchEstimationThreshold_Q16 = FIX_CONST(FIND_PITCH_CORRELATION_THRESHOLD_HC_MODE, 16)
		psEncC.pitchEstimationLPCOrder = 16
		psEncC.shapingLPCOrder = 16
		psEncC.la_shape = 5 * psEncC.fs_kHz
		psEncC.nStatesDelayedDecision = MAX_DEL_DEC_STATES
		psEncC.useInterpolatedNLSFs = 1
		psEncC.LTPQuantLowComplexity = 0
		psEncC.NLSF_MSVQ_Survivors = MAX_NLSF_MSVQ_SURVIVORS
		psEncC.warping_Q16 = psEncC.fs_kHz * FIX_CONST(WARPING_MULTIPLIER, 16)
	} else {
		ret = ENC_INVALID_COMPLEXITY_SETTING
	}

	psEncC.pitchEstimationLPCOrder = min(psEncC.pitchEstimationLPCOrder, psEncC.predictLPCOrder)
	psEncC.shapeWinLength = 5*psEncC.fs_kHz + 2*psEncC.la_shape

	return (ret)
}

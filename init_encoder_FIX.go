package silk

func init_encoder_FIX(psEnc *encoder_state_FIX) int {
	ret := 0

	psEnc.init()

	psEnc.variable_HP_smth1_Q15 = 200844
	psEnc.variable_HP_smth2_Q15 = 200844

	psEnc.sCmn.first_frame_after_reset = 1

	ret += VAD_Init(psEnc.sCmn.sVAD)

	psEnc.sCmn.sNSQ.prev_inv_gain_Q16 = 65536
	psEnc.sCmn.sNSQ_LBRR.prev_inv_gain_Q16 = 65536

	return ret
}

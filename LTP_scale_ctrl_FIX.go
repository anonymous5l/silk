package silk

const NB_THRESHOLDS = 11

var LTPScaleThresholds_Q15 = []int16{
	31129, 26214, 16384, 13107, 9830, 6554,
	4915, 3276, 2621, 2458, 0,
}

func LTP_scale_ctrl_FIX(psEnc *encoder_state_FIX, psEncCtrl *encoder_control_FIX) {
	var (
		round_loss, frames_per_packet                 int32
		g_out_Q5, g_limit_Q15, thrld1_Q15, thrld2_Q15 int32
	)

	psEnc.HPLTPredCodGain_Q7 = max(psEncCtrl.LTPredCodGain_Q7-psEnc.prevLTPredCodGain_Q7, 0) +
		RSHIFT_ROUND(psEnc.HPLTPredCodGain_Q7, 1)

	psEnc.prevLTPredCodGain_Q7 = psEncCtrl.LTPredCodGain_Q7

	g_out_Q5 = RSHIFT_ROUND(RSHIFT(psEncCtrl.LTPredCodGain_Q7, 1)+RSHIFT(psEnc.HPLTPredCodGain_Q7, 1), 3)
	g_limit_Q15 = sigm_Q15(g_out_Q5 - (3 << 5))

	psEncCtrl.sCmn.LTP_scaleIndex = 0

	round_loss = psEnc.sCmn.PacketLoss_perc

	if psEnc.sCmn.nFramesInPayloadBuf == 0 {

		frames_per_packet = DIV32_16(psEnc.sCmn.PacketSize_ms, FRAME_LENGTH_MS)

		round_loss += frames_per_packet - 1
		thrld1_Q15 = int32(LTPScaleThresholds_Q15[min(round_loss, NB_THRESHOLDS-1)])
		thrld2_Q15 = int32(LTPScaleThresholds_Q15[min(round_loss+1, NB_THRESHOLDS-1)])

		if g_limit_Q15 > thrld1_Q15 {
			psEncCtrl.sCmn.LTP_scaleIndex = 2
		} else if g_limit_Q15 > thrld2_Q15 {
			psEncCtrl.sCmn.LTP_scaleIndex = 1
		}
	}
	psEncCtrl.LTP_scale_Q14 = int32(LTPScales_table_Q14[psEncCtrl.sCmn.LTP_scaleIndex])
}

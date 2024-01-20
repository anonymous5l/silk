package silk

import (
	"math"
)

func process_gains_FIX(psEnc *encoder_state_FIX, psEncCtrl *encoder_control_FIX) {
	psShapeSt := psEnc.sShape

	var (
		k                                                                                 int32
		s_Q16, InvMaxSqrVal_Q16, gain, gain_squared, ResNrg, ResNrgPart, quant_offset_Q10 int32
	)

	if psEncCtrl.sCmn.sigtype == SIG_TYPE_VOICED {
		s_Q16 = -sigm_Q15(RSHIFT_ROUND(psEncCtrl.LTPredCodGain_Q7-FIX_CONST(12.0, 7), 4))
		for k = 0; k < NB_SUBFR; k++ {
			*psEncCtrl.Gains_Q16.ptr(int(k)) = SMLAWB(psEncCtrl.Gains_Q16.idx(int(k)), psEncCtrl.Gains_Q16.idx(int(k)), s_Q16)
		}
	}

	InvMaxSqrVal_Q16 = DIV32_16(log2lin(
		SMULWB(FIX_CONST(70.0, 7)-psEncCtrl.current_SNR_dB_Q7, FIX_CONST(0.33, 16))), int16(psEnc.sCmn.subfr_length))

	for k = 0; k < NB_SUBFR; k++ {
		ResNrg = psEncCtrl.ResNrg.idx(int(k))
		ResNrgPart = SMULWW(ResNrg, InvMaxSqrVal_Q16)

		if psEncCtrl.ResNrgQ.idx(int(k)) > 0 {
			if psEncCtrl.ResNrgQ.idx(int(k)) < 32 {
				ResNrgPart = RSHIFT_ROUND(ResNrgPart, psEncCtrl.ResNrgQ.idx(int(k)))
			} else {
				ResNrgPart = 0
			}
		} else if psEncCtrl.ResNrgQ.idx(int(k)) != 0 {
			if ResNrgPart > RSHIFT(math.MaxInt32, -psEncCtrl.ResNrgQ.idx(int(k))) {
				ResNrgPart = math.MaxInt32
			} else {
				ResNrgPart = LSHIFT(ResNrgPart, -psEncCtrl.ResNrgQ.idx(int(k)))
			}
		}
		gain = psEncCtrl.Gains_Q16.idx(int(k))
		gain_squared = ADD_SAT32(ResNrgPart, SMMUL(gain, gain))
		if gain_squared < math.MaxInt16 {
			gain_squared = SMLAWW(LSHIFT(ResNrgPart, 16), gain, gain)
			gain = SQRT_APPROX(gain_squared)
			*psEncCtrl.Gains_Q16.ptr(int(k)) = LSHIFT_SAT32(gain, 8)
		} else {
			gain = SQRT_APPROX(gain_squared)
			*psEncCtrl.Gains_Q16.ptr(int(k)) = LSHIFT_SAT32(gain, 16)
		}
	}

	gains_quant(psEncCtrl.sCmn.GainsIndices, psEncCtrl.Gains_Q16,
		&psShapeSt.LastGainIndex, psEnc.sCmn.nFramesInPayloadBuf)
	if psEncCtrl.sCmn.sigtype == SIG_TYPE_VOICED {
		if psEncCtrl.LTPredCodGain_Q7+RSHIFT(psEncCtrl.input_tilt_Q15, 8) > FIX_CONST(1.0, 7) {
			psEncCtrl.sCmn.QuantOffsetType = 0
		} else {
			psEncCtrl.sCmn.QuantOffsetType = 1
		}
	}

	quant_offset_Q10 = int32(Quantization_Offsets_Q10[psEncCtrl.sCmn.sigtype][psEncCtrl.sCmn.QuantOffsetType])
	psEncCtrl.Lambda_Q10 = FIX_CONST(LAMBDA_OFFSET, 10) +
		SMULBB(FIX_CONST(LAMBDA_DELAYED_DECISIONS, 10), psEnc.sCmn.nStatesDelayedDecision) +
		SMULWB(FIX_CONST(LAMBDA_SPEECH_ACT, 18), psEnc.speech_activity_Q8) +
		SMULWB(FIX_CONST(LAMBDA_INPUT_QUALITY, 12), psEncCtrl.input_quality_Q14) +
		SMULWB(FIX_CONST(LAMBDA_CODING_QUALITY, 12), psEncCtrl.coding_quality_Q14) +
		SMULWB(FIX_CONST(LAMBDA_QUANT_OFFSET, 16), quant_offset_Q10)
}

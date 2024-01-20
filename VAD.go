package silk

import (
	"math"
)

func VAD_Init(psSilk_VAD *VAD_state) int {
	var (
		b   int32
		ret int
	)

	psSilk_VAD.init()

	for b = 0; b < VAD_N_BANDS; b++ {
		*psSilk_VAD.NoiseLevelBias.ptr(int(b)) = max(DIV32_16(VAD_NOISE_LEVELS_BIAS, int16(b+1)), 1)
	}

	for b = 0; b < VAD_N_BANDS; b++ {
		*psSilk_VAD.NL.ptr(int(b)) = MUL(100, psSilk_VAD.NoiseLevelBias.idx(int(b)))
		*psSilk_VAD.inv_NL.ptr(int(b)) = DIV32(math.MaxInt32, psSilk_VAD.NL.idx(int(b)))
	}
	psSilk_VAD.counter = 15

	for b = 0; b < VAD_N_BANDS; b++ {
		*psSilk_VAD.NrgRatioSmth_Q8.ptr(int(b)) = 100 * 256
	}

	return ret
}

var tiltWeights = []int32{30000, 6000, -12000, -12000}

func VAD_GetSA_Q8(psSilk_VAD *VAD_state, pSA_Q8, pSNR_dB_Q7 *int32,
	pQuality_Q15 *slice[int32], pTilt_Q15 *int32,
	pIn *slice[int16], framelength int32) int {

	var (
		SA_Q15, input_tilt                                                               int32
		scratch                                                                          = alloc[int32](3 * MAX_FRAME_LENGTH / 2)
		decimated_framelength, dec_subframe_length, dec_subframe_offset, SNR_Q7, i, b, s int32
		sumSquared, smooth_coef_Q16                                                      int32
		HPstateTmp                                                                       int16
		X                                                                                = [VAD_N_BANDS]*slice[int16]{
			alloc[int16](MAX_FRAME_LENGTH / 2),
			alloc[int16](MAX_FRAME_LENGTH / 2),
			alloc[int16](MAX_FRAME_LENGTH / 2),
			alloc[int16](MAX_FRAME_LENGTH / 2),
		}
		Xnrg               = alloc[int32](VAD_N_BANDS)
		NrgToNoiseRatio_Q8 [VAD_N_BANDS]int32
		speech_nrg, x_tmp  int32
		ret                int
	)

	ana_filt_bank_1(pIn, psSilk_VAD.AnaState.off(0), X[0].off(0), X[3].off(0),
		scratch.ptr(0), framelength)

	ana_filt_bank_1(X[0].off(0), psSilk_VAD.AnaState1.off(0), X[0].off(0), X[2].off(0),
		scratch.ptr(0), RSHIFT(framelength, 1))

	ana_filt_bank_1(X[0].off(0), psSilk_VAD.AnaState2.off(0), X[0].off(0), X[1].off(0),
		scratch.ptr(0), RSHIFT(framelength, 2))

	decimated_framelength = RSHIFT(framelength, 3)
	*X[0].ptr(int(decimated_framelength - 1)) = int16(RSHIFT(int32(X[0].idx(int(decimated_framelength-1))), 1))
	HPstateTmp = X[0].idx(int(decimated_framelength - 1))
	for i = decimated_framelength - 1; i > 0; i-- {
		*X[0].ptr(int(i - 1)) = int16(RSHIFT(int32(X[0].idx(int(i-1))), 1))
		*X[0].ptr(int(i)) -= X[0].idx(int(i - 1))
	}
	*X[0].ptr(0) -= psSilk_VAD.HPstate
	psSilk_VAD.HPstate = HPstateTmp

	for b = 0; b < VAD_N_BANDS; b++ {
		decimated_framelength = RSHIFT(framelength, min(VAD_N_BANDS-b, VAD_N_BANDS-1))

		dec_subframe_length = RSHIFT(decimated_framelength, VAD_INTERNAL_SUBFRAMES_LOG2)
		dec_subframe_offset = 0

		*Xnrg.ptr(int(b)) = psSilk_VAD.XnrgSubfr.idx(int(b))
		for s = 0; s < VAD_INTERNAL_SUBFRAMES; s++ {
			sumSquared = 0
			for i = 0; i < dec_subframe_length; i++ {
				x_tmp = RSHIFT(int32(X[b].idx(int(i+dec_subframe_offset))), 3)
				sumSquared = SMLABB(sumSquared, x_tmp, x_tmp)
			}

			if s < VAD_INTERNAL_SUBFRAMES-1 {
				*Xnrg.ptr(int(b)) = ADD_POS_SAT32(Xnrg.idx(int(b)), sumSquared)
			} else {
				*Xnrg.ptr(int(b)) = ADD_POS_SAT32(Xnrg.idx(int(b)), RSHIFT(sumSquared, 1))
			}

			dec_subframe_offset += dec_subframe_length
		}
		*psSilk_VAD.XnrgSubfr.ptr(int(b)) = sumSquared
	}

	VAD_GetNoiseLevels(Xnrg.off(0), psSilk_VAD)

	sumSquared = 0
	input_tilt = 0
	for b = 0; b < VAD_N_BANDS; b++ {
		speech_nrg = Xnrg.idx(int(b)) - psSilk_VAD.NL.idx(int(b))
		if speech_nrg > 0 {
			if (int(Xnrg.idx(int(b))) & 0xFF800000) == 0 {
				NrgToNoiseRatio_Q8[b] = DIV32(LSHIFT(Xnrg.idx(int(b)), 8), psSilk_VAD.NL.idx(int(b))+1)
			} else {
				NrgToNoiseRatio_Q8[b] = DIV32(Xnrg.idx(int(b)), RSHIFT(psSilk_VAD.NL.idx(int(b)), 8)+1)
			}

			SNR_Q7 = lin2log(NrgToNoiseRatio_Q8[b]) - 8*128

			sumSquared = SMLABB(sumSquared, SNR_Q7, SNR_Q7)

			if speech_nrg < (1 << 20) {
				SNR_Q7 = SMULWB(LSHIFT(SQRT_APPROX(speech_nrg), 6), SNR_Q7)
			}
			input_tilt = SMLAWB(input_tilt, tiltWeights[b], SNR_Q7)
		} else {
			NrgToNoiseRatio_Q8[b] = 256
		}
	}

	sumSquared = DIV32_16(sumSquared, VAD_N_BANDS)

	*pSNR_dB_Q7 = 3 * SQRT_APPROX(sumSquared)

	SA_Q15 = sigm_Q15(SMULWB(VAD_SNR_FACTOR_Q16, *pSNR_dB_Q7) - VAD_NEGATIVE_OFFSET_Q5)

	*pTilt_Q15 = LSHIFT(sigm_Q15(input_tilt)-16384, 1)

	speech_nrg = 0
	for b = 0; b < VAD_N_BANDS; b++ {
		speech_nrg += (b + 1) * RSHIFT(Xnrg.idx(int(b))-psSilk_VAD.NL.idx(int(b)), 4)
	}

	if speech_nrg <= 0 {
		SA_Q15 = RSHIFT(SA_Q15, 1)
	} else if speech_nrg < 32768 {
		speech_nrg = SQRT_APPROX(LSHIFT(speech_nrg, 15))
		SA_Q15 = SMULWB(32768+speech_nrg, SA_Q15)
	}

	*pSA_Q8 = min(RSHIFT(SA_Q15, 7), math.MaxUint8)

	smooth_coef_Q16 = SMULWB(VAD_SNR_SMOOTH_COEF_Q18, SMULWB(SA_Q15, SA_Q15))
	for b = 0; b < VAD_N_BANDS; b++ {
		*psSilk_VAD.NrgRatioSmth_Q8.ptr(int(b)) = SMLAWB(psSilk_VAD.NrgRatioSmth_Q8.idx(int(b)),
			NrgToNoiseRatio_Q8[b]-psSilk_VAD.NrgRatioSmth_Q8.idx(int(b)), smooth_coef_Q16)

		SNR_Q7 = 3 * (lin2log(psSilk_VAD.NrgRatioSmth_Q8.idx(int(b))) - 8*128)
		*pQuality_Q15.ptr(int(b)) = sigm_Q15(RSHIFT(SNR_Q7-16*128, 4))
	}

	return (ret)
}

func VAD_GetNoiseLevels(pX *slice[int32], psSilk_VAD *VAD_state) {
	var (
		k, nl, nrg, inv_nrg int32
		coef, min_coef      int32
	)
	if psSilk_VAD.counter < 1000 {
		min_coef = DIV32_16(math.MaxInt16, int16(RSHIFT(psSilk_VAD.counter, 4)+1))
	} else {
		min_coef = 0
	}

	for k = 0; k < VAD_N_BANDS; k++ {
		nl = psSilk_VAD.NL.idx(int(k))

		nrg = ADD_POS_SAT32(pX.idx(int(k)), psSilk_VAD.NoiseLevelBias.idx(int(k)))

		inv_nrg = DIV32(math.MaxInt32, nrg)

		if nrg > LSHIFT(nl, 3) {
			coef = VAD_NOISE_LEVEL_SMOOTH_COEF_Q16 >> 3
		} else if nrg < nl {
			coef = VAD_NOISE_LEVEL_SMOOTH_COEF_Q16
		} else {
			coef = SMULWB(SMULWW(inv_nrg, nl), VAD_NOISE_LEVEL_SMOOTH_COEF_Q16<<1)
		}

		coef = max(coef, min_coef)

		*psSilk_VAD.inv_NL.ptr(int(k)) = SMLAWB(psSilk_VAD.inv_NL.idx(int(k)), inv_nrg-psSilk_VAD.inv_NL.idx(int(k)), coef)

		nl = DIV32(math.MaxInt32, psSilk_VAD.inv_NL.idx(int(k)))

		nl = min(nl, 0x00FFFFFF)

		*psSilk_VAD.NL.ptr(int(k)) = nl
	}

	psSilk_VAD.counter++
}

package silk

const NB_ATT = 2

var (
	HARM_ATT_Q15              = []int16{32440, 31130}
	PLC_RAND_ATTENUATE_V_Q15  = []int16{31130, 26214}
	PLC_RAND_ATTENUATE_UV_Q15 = []int16{32440, 29491}
)

const (
	BWE_COEF_Q16                 = 64880
	V_PITCH_GAIN_START_MIN_Q14   = 11469
	V_PITCH_GAIN_START_MAX_Q14   = 15565
	MAX_PITCH_LAG_MS             = 18
	SA_THRES_Q8                  = 50
	USE_SINGLE_TAP               = 1
	RAND_BUF_SIZE                = 128
	RAND_BUF_MASK                = (RAND_BUF_SIZE - 1)
	LOG2_INV_LPC_GAIN_HIGH_THRES = 3
	LOG2_INV_LPC_GAIN_LOW_THRES  = 8
	PITCH_DRIFT_FAC_Q16          = 655
)

func PLC_Reset(psDec *decoder_state) {
	psDec.sPLC = &PLC_struct{}
	psDec.sPLC.init()

	psDec.sPLC.pitchL_Q8 = RSHIFT(psDec.frame_length, 1)
}

func PLC(
	psDec *decoder_state,
	psDecCtrl *decoder_control,
	signal *slice[int16],
	length int32, lost bool) {

	if psDec.fs_kHz != psDec.sPLC.fs_kHz {
		PLC_Reset(psDec)
		psDec.sPLC.fs_kHz = psDec.fs_kHz
	}

	if lost {
		PLC_conceal(psDec, psDecCtrl, signal, length)
		psDec.lossCnt++
	} else {
		PLC_update(psDec, psDecCtrl, signal, length)
	}
}

func PLC_update(psDec *decoder_state, psDecCtrl *decoder_control, signal *slice[int16], length int32) {
	var (
		LTP_Gain_Q14, temp_LTP_Gain_Q14 int32
		i, j                            int32
		psPLC                           *PLC_struct
	)

	psPLC = psDec.sPLC

	psDec.prev_sigtype = psDecCtrl.sigtype
	LTP_Gain_Q14 = 0
	if psDecCtrl.sigtype == SIG_TYPE_VOICED {
		for j = 0; j*psDec.subfr_length < psDecCtrl.pitchL.idx(NB_SUBFR-1); j++ {
			temp_LTP_Gain_Q14 = 0
			for i = 0; i < LTP_ORDER; i++ {
				temp_LTP_Gain_Q14 += int32(psDecCtrl.LTPCoef_Q14.idx(int((NB_SUBFR-1-j)*LTP_ORDER + i)))
			}
			if temp_LTP_Gain_Q14 > LTP_Gain_Q14 {
				LTP_Gain_Q14 = temp_LTP_Gain_Q14

				psDecCtrl.LTPCoef_Q14.off(int(SMULBB(NB_SUBFR-1-j, LTP_ORDER))).copy(psPLC.LTPCoef_Q14, LTP_ORDER)

				psPLC.pitchL_Q8 = LSHIFT(psDecCtrl.pitchL.idx(int(NB_SUBFR-1-j)), 8)
			}
		}
		memset(psPLC.LTPCoef_Q14, 0, LTP_ORDER)
		*psPLC.LTPCoef_Q14.ptr(LTP_ORDER / 2) = int16(LTP_Gain_Q14)

		if LTP_Gain_Q14 < V_PITCH_GAIN_START_MIN_Q14 {
			var scale_Q10, tmp int32

			tmp = LSHIFT(V_PITCH_GAIN_START_MIN_Q14, 10)
			scale_Q10 = DIV32(tmp, max(LTP_Gain_Q14, 1))
			for i = 0; i < LTP_ORDER; i++ {
				*psPLC.LTPCoef_Q14.ptr(int(i)) = int16(RSHIFT(SMULBB(int32(psPLC.LTPCoef_Q14.idx(int(i))), scale_Q10), 10))
			}
		} else if LTP_Gain_Q14 > V_PITCH_GAIN_START_MAX_Q14 {
			var scale_Q14, tmp int32

			tmp = LSHIFT(V_PITCH_GAIN_START_MAX_Q14, 14)
			scale_Q14 = DIV32(tmp, max(LTP_Gain_Q14, 1))
			for i = 0; i < LTP_ORDER; i++ {
				*psPLC.LTPCoef_Q14.ptr(int(i)) = int16(RSHIFT(SMULBB(int32(psPLC.LTPCoef_Q14.idx(int(i))), scale_Q14), 14))
			}
		}
	} else {
		psPLC.pitchL_Q8 = LSHIFT(SMULBB(psDec.fs_kHz, 18), 8)
		memset(psPLC.LTPCoef_Q14, 0, LTP_ORDER)
	}

	psDecCtrl.PredCoef_Q12[1].copy(psPLC.prevLPC_Q12, int(psDec.LPC_order))
	psPLC.prevLTP_scale_Q14 = psDecCtrl.LTP_scale_Q14

	psDecCtrl.Gains_Q16.copy(psPLC.prevGain_Q16, NB_SUBFR)
}

func PLC_conceal(psDec *decoder_state, psDecCtrl *decoder_control, signal *slice[int16], length int32) {
	var (
		i, j, k                                 int32
		B_Q14                                   *slice[int16]
		rand_scale_Q14                          int16
		exc_buf                                 = alloc[int16](MAX_FRAME_LENGTH)
		exc_buf_ptr                             *slice[int16]
		A_Q12_tmp16                             *slice[int16]
		A_Q12_tmp32                             *slice[int32]
		rand_seed, harm_Gain_Q15, rand_Gain_Q15 int32
		lag, idx, sLTP_buf_idx, shift1, shift2  int32
		energy1, energy2                        int32

		rand_ptr, pred_lag_ptr                  *slice[int32]
		sig_Q10                                 = alloc[int32](MAX_FRAME_LENGTH)
		sig_Q10_ptr                             *slice[int32]
		LPC_exc_Q10, LPC_pred_Q10, LTP_pred_Q14 int32
		psPLC                                   *PLC_struct
		Atmp                                    int32
	)

	psPLC = psDec.sPLC

	psDec.sLTP_Q16.off(int(psDec.frame_length)).
		copy(psDec.sLTP_Q16, int(psDec.frame_length))

	bwexpander(psPLC.prevLPC_Q12, psDec.LPC_order, BWE_COEF_Q16)

	exc_buf_ptr = exc_buf.off(0)
	for k = NB_SUBFR >> 1; k < NB_SUBFR; k++ {
		for i = 0; i < psDec.subfr_length; i++ {
			*exc_buf_ptr.ptr(int(i)) = int16(RSHIFT(
				SMULWW(psDec.exc_Q10.idx(int(i+k*psDec.subfr_length)), psPLC.prevGain_Q16.idx(int(k))), 10))
		}
		exc_buf_ptr = exc_buf_ptr.off(int(psDec.subfr_length))
	}

	sum_sqr_shift(&energy1, &shift1, exc_buf, psDec.subfr_length)
	sum_sqr_shift(&energy2, &shift2, exc_buf.off(int(psDec.subfr_length)), psDec.subfr_length)

	if RSHIFT(energy1, shift2) < RSHIFT(energy2, shift1) {
		rand_ptr = psDec.exc_Q10.off(int(max(0, 3*psDec.subfr_length-RAND_BUF_SIZE)))
	} else {
		rand_ptr = psDec.exc_Q10.off(int(max(0, psDec.frame_length-RAND_BUF_SIZE)))
	}

	B_Q14 = psPLC.LTPCoef_Q14
	rand_scale_Q14 = psPLC.randScale_Q14

	harm_Gain_Q15 = int32(HARM_ATT_Q15[min(NB_ATT-1, psDec.lossCnt)])
	if psDec.prev_sigtype == SIG_TYPE_VOICED {
		rand_Gain_Q15 = int32(PLC_RAND_ATTENUATE_V_Q15[min(NB_ATT-1, psDec.lossCnt)])
	} else {
		rand_Gain_Q15 = int32(PLC_RAND_ATTENUATE_UV_Q15[min(NB_ATT-1, psDec.lossCnt)])
	}

	if psDec.lossCnt == 0 {
		rand_scale_Q14 = 1 << 14

		if psDec.prev_sigtype == SIG_TYPE_VOICED {
			for i = 0; i < LTP_ORDER; i++ {
				rand_scale_Q14 -= B_Q14.idx(int(i))
			}
			rand_scale_Q14 = max(3277, rand_scale_Q14)
			rand_scale_Q14 = int16(RSHIFT(SMULBB(int32(rand_scale_Q14), int32(psPLC.prevLTP_scale_Q14)), 14))
		}

		if psDec.prev_sigtype == SIG_TYPE_UNVOICED {
			var invGain_Q30, down_scale_Q30 int32

			LPC_inverse_pred_gain(&invGain_Q30, psPLC.prevLPC_Q12, psDec.LPC_order)

			down_scale_Q30 = min(RSHIFT(1<<30, LOG2_INV_LPC_GAIN_HIGH_THRES), invGain_Q30)
			down_scale_Q30 = max(RSHIFT(1<<30, LOG2_INV_LPC_GAIN_LOW_THRES), down_scale_Q30)
			down_scale_Q30 = LSHIFT(down_scale_Q30, LOG2_INV_LPC_GAIN_HIGH_THRES)

			rand_Gain_Q15 = RSHIFT(SMULWB(down_scale_Q30, rand_Gain_Q15), 14)
		}
	}

	rand_seed = psPLC.rand_seed
	lag = RSHIFT_ROUND(psPLC.pitchL_Q8, 8)
	sLTP_buf_idx = psDec.frame_length

	sig_Q10_ptr = sig_Q10.off(0)
	for k = 0; k < NB_SUBFR; k++ {
		pred_lag_ptr = psDec.sLTP_Q16.off(int(sLTP_buf_idx - lag + LTP_ORDER/2))
		for i = 0; i < psDec.subfr_length; i++ {
			rand_seed = RAND(rand_seed)
			idx = RSHIFT(rand_seed, 25) & RAND_BUF_MASK

			LTP_pred_Q14 = SMULWB(pred_lag_ptr.idx(0), int32(B_Q14.idx(0)))
			LTP_pred_Q14 = SMLAWB(LTP_pred_Q14, pred_lag_ptr.idx(-1), int32(B_Q14.idx(1)))
			LTP_pred_Q14 = SMLAWB(LTP_pred_Q14, pred_lag_ptr.idx(-2), int32(B_Q14.idx(2)))
			LTP_pred_Q14 = SMLAWB(LTP_pred_Q14, pred_lag_ptr.idx(-3), int32(B_Q14.idx(3)))
			LTP_pred_Q14 = SMLAWB(LTP_pred_Q14, pred_lag_ptr.idx(-4), int32(B_Q14.idx(4)))
			pred_lag_ptr = pred_lag_ptr.off(1)

			LPC_exc_Q10 = LSHIFT(SMULWB(rand_ptr.idx(int(idx)), int32(rand_scale_Q14)), 2)
			LPC_exc_Q10 = ADD32(LPC_exc_Q10, RSHIFT_ROUND(LTP_pred_Q14, 4))

			*psDec.sLTP_Q16.ptr(int(sLTP_buf_idx)) = LSHIFT(LPC_exc_Q10, 6)
			sLTP_buf_idx++

			*sig_Q10_ptr.ptr(int(i)) = LPC_exc_Q10
		}
		sig_Q10_ptr = sig_Q10_ptr.off(int(psDec.subfr_length))
		for j = 0; j < LTP_ORDER; j++ {
			*B_Q14.ptr(int(j)) = int16(RSHIFT(SMULBB(harm_Gain_Q15, int32(B_Q14.idx(int(j)))), 15))
		}
		rand_scale_Q14 = int16(RSHIFT(SMULBB(int32(rand_scale_Q14), rand_Gain_Q15), 15))

		psPLC.pitchL_Q8 += SMULWB(psPLC.pitchL_Q8, PITCH_DRIFT_FAC_Q16)
		psPLC.pitchL_Q8 = min(psPLC.pitchL_Q8, LSHIFT(SMULBB(MAX_PITCH_LAG_MS, psDec.fs_kHz), 8))
		lag = RSHIFT_ROUND(psPLC.pitchL_Q8, 8)
	}

	sig_Q10_ptr = sig_Q10.off(0)

	A_Q12_tmp16 = alloc[int16](MAX_LPC_ORDER)
	A_Q12_tmp32 = slice2[int32](A_Q12_tmp16)

	psPLC.prevLPC_Q12.copy(A_Q12_tmp16, int(psDec.LPC_order))

	for k = 0; k < NB_SUBFR; k++ {
		for i = 0; i < psDec.subfr_length; i++ {

			Atmp = A_Q12_tmp32.idx(int(0))
			LPC_pred_Q10 = SMULWB(psDec.sLPC_Q14.idx(int(MAX_LPC_ORDER+i-1)), Atmp)
			LPC_pred_Q10 = SMLAWT(LPC_pred_Q10, psDec.sLPC_Q14.idx(int(MAX_LPC_ORDER+i-2)), Atmp)
			Atmp = A_Q12_tmp32.idx(int(1))
			LPC_pred_Q10 = SMLAWB(LPC_pred_Q10, psDec.sLPC_Q14.idx(int(MAX_LPC_ORDER+i-3)), Atmp)
			LPC_pred_Q10 = SMLAWT(LPC_pred_Q10, psDec.sLPC_Q14.idx(int(MAX_LPC_ORDER+i-4)), Atmp)
			Atmp = A_Q12_tmp32.idx(int(2))
			LPC_pred_Q10 = SMLAWB(LPC_pred_Q10, psDec.sLPC_Q14.idx(int(MAX_LPC_ORDER+i-5)), Atmp)
			LPC_pred_Q10 = SMLAWT(LPC_pred_Q10, psDec.sLPC_Q14.idx(int(MAX_LPC_ORDER+i-6)), Atmp)
			Atmp = A_Q12_tmp32.idx(int(3))
			LPC_pred_Q10 = SMLAWB(LPC_pred_Q10, psDec.sLPC_Q14.idx(int(MAX_LPC_ORDER+i-7)), Atmp)
			LPC_pred_Q10 = SMLAWT(LPC_pred_Q10, psDec.sLPC_Q14.idx(int(MAX_LPC_ORDER+i-8)), Atmp)
			Atmp = A_Q12_tmp32.idx(int(4))
			LPC_pred_Q10 = SMLAWB(LPC_pred_Q10, psDec.sLPC_Q14.idx(int(MAX_LPC_ORDER+i-9)), Atmp)
			LPC_pred_Q10 = SMLAWT(LPC_pred_Q10, psDec.sLPC_Q14.idx(int(MAX_LPC_ORDER+i-10)), Atmp)
			for j = 10; j < psDec.LPC_order; j += 2 {
				Atmp = A_Q12_tmp32.idx(int(j / 2))
				LPC_pred_Q10 = SMLAWB(LPC_pred_Q10, psDec.sLPC_Q14.idx(int(MAX_LPC_ORDER+i-1-j)), Atmp)
				LPC_pred_Q10 = SMLAWT(LPC_pred_Q10, psDec.sLPC_Q14.idx(int(MAX_LPC_ORDER+i-2-j)), Atmp)
			}

			*sig_Q10_ptr.ptr(int(i)) = ADD32(sig_Q10_ptr.idx(int(i)), LPC_pred_Q10)

			*psDec.sLPC_Q14.ptr(int(MAX_LPC_ORDER + i)) = LSHIFT(sig_Q10_ptr.idx(int(i)), 4)
		}
		sig_Q10_ptr = sig_Q10_ptr.off(int(psDec.subfr_length))
		psDec.sLPC_Q14.off(int(psDec.subfr_length)).copy(psDec.sLPC_Q14, MAX_LPC_ORDER)
	}

	for i = 0; i < psDec.frame_length; i++ {
		*signal.ptr(int(i)) = int16(SAT16(RSHIFT_ROUND(SMULWW(sig_Q10.idx(int(i)), psPLC.prevGain_Q16.idx(NB_SUBFR-1)), 10)))
	}

	psPLC.rand_seed = rand_seed
	psPLC.randScale_Q14 = rand_scale_Q14
	for i = 0; i < NB_SUBFR; i++ {
		*psDecCtrl.pitchL.ptr(int(i)) = lag
	}
}

func PLC_glue_frames(psDec *decoder_state, psDecCtrl *decoder_control,
	signal *slice[int16], length int32) {
	var (
		i, energy_shift, energy int32
	)
	var psPLC *PLC_struct
	psPLC = psDec.sPLC

	if psDec.lossCnt != 0 {
		sum_sqr_shift(&psPLC.conc_energy, &psPLC.conc_energy_shift, signal, length)

		psPLC.last_frame_lost = 1
	} else {
		if psDec.sPLC.last_frame_lost != 0 {
			sum_sqr_shift(&energy, &energy_shift, signal, length)

			if energy_shift > psPLC.conc_energy_shift {
				psPLC.conc_energy = RSHIFT(psPLC.conc_energy, energy_shift-psPLC.conc_energy_shift)
			} else if energy_shift < psPLC.conc_energy_shift {
				energy = RSHIFT(energy, psPLC.conc_energy_shift-energy_shift)
			}

			if energy > psPLC.conc_energy {
				var (
					frac_Q24, LZ        int32
					gain_Q12, slope_Q12 int32
				)

				LZ = CLZ32(psPLC.conc_energy)
				LZ = LZ - 1
				psPLC.conc_energy = LSHIFT(psPLC.conc_energy, LZ)
				energy = RSHIFT(energy, max(24-LZ, 0))

				frac_Q24 = DIV32(psPLC.conc_energy, max(energy, 1))

				gain_Q12 = SQRT_APPROX(frac_Q24)
				slope_Q12 = DIV32_16((1<<12)-gain_Q12, int16(length))

				for i = 0; i < length; i++ {
					*signal.ptr(int(i)) = int16(RSHIFT(MUL(gain_Q12, int32(signal.idx(int(i)))), 12))
					gain_Q12 += slope_Q12
					gain_Q12 = min(gain_Q12, 1<<12)
				}
			}
		}
		psPLC.last_frame_lost = 0
	}
}

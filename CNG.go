package silk

import "math"

func CNG_exc(residual *slice[int16], exc_buf_Q10 *slice[int32], Gain_Q16 int32, length int32, rand_seed *int32) {
	var (
		seed        int32
		i, exc_mask int32
	)

	exc_mask = CNG_BUF_MASK_MAX
	for exc_mask > length {
		exc_mask = RSHIFT(exc_mask, 1)
	}

	seed = *rand_seed
	for i = 0; i < length; i++ {
		seed = RAND(seed)
		idx := int(RSHIFT(seed, 24) & exc_mask)
		*residual.ptr(int(i)) = int16(SAT16(RSHIFT_ROUND(SMULWW(exc_buf_Q10.idx(idx), Gain_Q16), 10)))
	}
	*rand_seed = seed
}

func CNG_Reset(psDec *decoder_state) {
	var i, NLSF_step_Q15, NLSF_acc_Q15 int32

	psDec.sCNG = &CNG_struct{}
	psDec.sCNG.init()

	NLSF_step_Q15 = DIV32_16(math.MaxInt16, int16(psDec.LPC_order+1))
	NLSF_acc_Q15 = 0
	for i = 0; i < psDec.LPC_order; i++ {
		NLSF_acc_Q15 += NLSF_step_Q15
		*psDec.sCNG.CNG_smth_NLSF_Q15.ptr(int(i)) = NLSF_acc_Q15
	}
	psDec.sCNG.CNG_smth_Gain_Q16 = 0
	psDec.sCNG.rand_seed = 3176576
}

func CNG(psDec *decoder_state, psDecCtrl *decoder_control, signal *slice[int16], length int32) {
	var (
		i, subfr                       int32
		tmp_32, Gain_Q26, max_Gain_Q16 int32
		LPC_buf                        = alloc[int16](MAX_LPC_ORDER)
		CNG_sig                        = alloc[int16](MAX_FRAME_LENGTH)
		psCNG                          *CNG_struct
	)

	psCNG = psDec.sCNG

	if psDec.fs_kHz != psCNG.fs_kHz {
		CNG_Reset(psDec)
		psCNG.fs_kHz = psDec.fs_kHz
	}

	if psDec.lossCnt == 0 && psDec.vadFlag == NO_VOICE_ACTIVITY {
		for i = 0; i < psDec.LPC_order; i++ {
			*psCNG.CNG_smth_NLSF_Q15.ptr(int(i)) += SMULWB(
				psDec.prevNLSF_Q15.idx(int(i))-
					psCNG.CNG_smth_NLSF_Q15.idx(int(i)), CNG_NLSF_SMTH_Q16)
		}

		max_Gain_Q16 = 0
		subfr = 0
		for i = 0; i < NB_SUBFR; i++ {
			if psDecCtrl.Gains_Q16.idx(int(i)) > max_Gain_Q16 {
				max_Gain_Q16 = psDecCtrl.Gains_Q16.idx(int(i))
				subfr = i
			}
		}

		psCNG.CNG_exc_buf_Q10.copy(psCNG.CNG_exc_buf_Q10.off(int(psDec.subfr_length)),
			int((NB_SUBFR-1)*psDec.subfr_length))
		psDec.exc_Q10.off(int(subfr*psDec.subfr_length)).copy(psCNG.CNG_exc_buf_Q10,
			int(psDec.subfr_length))

		for i = 0; i < NB_SUBFR; i++ {
			psCNG.CNG_smth_Gain_Q16 += SMULWB(psDecCtrl.Gains_Q16.idx(int(i))-psCNG.CNG_smth_Gain_Q16, CNG_GAIN_SMTH_Q16)
		}
	}

	if psDec.lossCnt != 0 {

		CNG_exc(CNG_sig, psCNG.CNG_exc_buf_Q10,
			psCNG.CNG_smth_Gain_Q16, length, &psCNG.rand_seed)

		NLSF2A_stable(LPC_buf, psCNG.CNG_smth_NLSF_Q15, psDec.LPC_order)

		Gain_Q26 = 1 << 26

		if psDec.LPC_order == 16 {
			LPC_synthesis_order16(CNG_sig, LPC_buf,
				Gain_Q26, psCNG.CNG_synth_state, CNG_sig, length)
		} else {
			LPC_synthesis_filter(CNG_sig, LPC_buf,
				Gain_Q26, psCNG.CNG_synth_state, CNG_sig, length, psDec.LPC_order)
		}

		for i = 0; i < length; i++ {
			tmp_32 = int32(signal.idx(int(i)) + CNG_sig.idx(int(i)))
			*signal.ptr(int(i)) = int16(SAT16(tmp_32))
		}
	} else {
		memset(psCNG.CNG_synth_state, 0, int(psDec.LPC_order))
	}
}

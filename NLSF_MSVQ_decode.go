package silk

func NLSF_MSVQ_decode(pNLSF_Q15 *slice[int32], psNLSF_CB *NLSF_CB_struct, NLSFIndices *slice[int32], LPC_order int32) {
	var (
		pCB_element *slice[int16]
		s, i        int32
	)

	pCB_element = psNLSF_CB.CBStages[0].CB_NLSF_Q15.off(int(MUL(NLSFIndices.idx(0), LPC_order)))

	for i = 0; i < LPC_order; i++ {
		*pNLSF_Q15.ptr(int(i)) = int32(pCB_element.idx(int(i)))
	}

	for s = 1; s < psNLSF_CB.nStages; s++ {

		if LPC_order == 16 {
			pCB_element = psNLSF_CB.CBStages[s].CB_NLSF_Q15.off(int(LSHIFT(NLSFIndices.idx(int(s)), 4)))

			*pNLSF_Q15.ptr(0) += int32(pCB_element.idx(0))
			*pNLSF_Q15.ptr(1) += int32(pCB_element.idx(1))
			*pNLSF_Q15.ptr(2) += int32(pCB_element.idx(2))
			*pNLSF_Q15.ptr(3) += int32(pCB_element.idx(3))
			*pNLSF_Q15.ptr(4) += int32(pCB_element.idx(4))
			*pNLSF_Q15.ptr(5) += int32(pCB_element.idx(5))
			*pNLSF_Q15.ptr(6) += int32(pCB_element.idx(6))
			*pNLSF_Q15.ptr(7) += int32(pCB_element.idx(7))
			*pNLSF_Q15.ptr(8) += int32(pCB_element.idx(8))
			*pNLSF_Q15.ptr(9) += int32(pCB_element.idx(9))
			*pNLSF_Q15.ptr(10) += int32(pCB_element.idx(10))
			*pNLSF_Q15.ptr(11) += int32(pCB_element.idx(11))
			*pNLSF_Q15.ptr(12) += int32(pCB_element.idx(12))
			*pNLSF_Q15.ptr(13) += int32(pCB_element.idx(13))
			*pNLSF_Q15.ptr(14) += int32(pCB_element.idx(14))
			*pNLSF_Q15.ptr(15) += int32(pCB_element.idx(15))
		} else {
			pCB_element = psNLSF_CB.CBStages[s].CB_NLSF_Q15.off(int(SMULBB(NLSFIndices.idx(int(s)), LPC_order)))

			for i = 0; i < LPC_order; i++ {
				*pNLSF_Q15.ptr(int(i)) += int32(pCB_element.idx(int(i)))
			}
		}
	}

	NLSF_stabilize(pNLSF_Q15, psNLSF_CB.NDeltaMin_Q15, LPC_order)
}

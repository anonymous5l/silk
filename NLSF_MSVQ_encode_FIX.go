package silk

import "math"

func NLSF_MSVQ_encode_FIX(NLSFIndices, pNLSF_Q15 *slice[int32], psNLSF_CB *NLSF_CB_struct,
	pNLSF_q_Q15_prev, pW_Q6 *slice[int32], NLSF_mu_Q15, NLSF_mu_fluc_red_Q16, NLSF_MSVQ_Survivors, LPC_order, deactivate_fluc_red int32) {
	var (
		i, s, k, cur_survivors, prev_survivors, min_survivors, input_index, cb_index, bestIndex int32
		rateDistThreshold_Q18                                                                   int32
		se_Q15, wsse_Q20, bestRateDist_Q20                                                      int32
		pRateDist_Q18                                                                           = alloc[int32](NLSF_MSVQ_TREE_SEARCH_MAX_VECTORS_EVALUATED)
		pRate_Q5                                                                                = alloc[int32](MAX_NLSF_MSVQ_SURVIVORS)
		pRate_new_Q5                                                                            = alloc[int32](MAX_NLSF_MSVQ_SURVIVORS)
		pTempIndices                                                                            = alloc[int32](MAX_NLSF_MSVQ_SURVIVORS)
		pPath                                                                                   = alloc[int32](MAX_NLSF_MSVQ_SURVIVORS * NLSF_MSVQ_MAX_CB_STAGES)
		pPath_new                                                                               = alloc[int32](MAX_NLSF_MSVQ_SURVIVORS * NLSF_MSVQ_MAX_CB_STAGES)
		pRes_Q15                                                                                = alloc[int32](MAX_NLSF_MSVQ_SURVIVORS * MAX_LPC_ORDER)
		pRes_new_Q15                                                                            = alloc[int32](MAX_NLSF_MSVQ_SURVIVORS * MAX_LPC_ORDER)
		pConstInt, pInt                                                                         *slice[int32]
		pCB_element                                                                             *slice[int16]
		pCurrentCBStage                                                                         *NLSF_CBS
	)

	memset(pRate_Q5, 0, int(NLSF_MSVQ_Survivors))

	for i = 0; i < LPC_order; i++ {
		*pRes_Q15.ptr(int(i)) = pNLSF_Q15.idx(int(i))
	}

	prev_survivors = 1

	min_survivors = NLSF_MSVQ_Survivors / 2

	for s = 0; s < psNLSF_CB.nStages; s++ {

		pCurrentCBStage = &psNLSF_CB.CBStages[s]

		cur_survivors = min(NLSF_MSVQ_Survivors, SMULBB(prev_survivors, pCurrentCBStage.nVectors))

		NLSF_VQ_rate_distortion_FIX(pRateDist_Q18, pCurrentCBStage, pRes_Q15, pW_Q6,
			pRate_Q5, NLSF_mu_Q15, prev_survivors, LPC_order)

		insertion_sort_increasing(pRateDist_Q18, pTempIndices,
			prev_survivors*pCurrentCBStage.nVectors, cur_survivors)

		if pRateDist_Q18.idx(0) < math.MaxInt32/MAX_NLSF_MSVQ_SURVIVORS {
			rateDistThreshold_Q18 = SMLAWB(pRateDist_Q18.idx(0),
				MUL(NLSF_MSVQ_Survivors, pRateDist_Q18.idx(0)), FIX_CONST(NLSF_MSVQ_SURV_MAX_REL_RD, 16))
			for pRateDist_Q18.idx(int(cur_survivors-1)) > rateDistThreshold_Q18 && cur_survivors > min_survivors {
				cur_survivors--
			}
		}
		for k = 0; k < cur_survivors; k++ {
			if s > 0 {
				if pCurrentCBStage.nVectors == 8 {
					input_index = RSHIFT(pTempIndices.idx(int(k)), 3)
					cb_index = pTempIndices.idx(int(k)) & 7
				} else {
					input_index = DIV32_16(pTempIndices.idx(int(k)), int16(pCurrentCBStage.nVectors))
					cb_index = pTempIndices.idx(int(k)) - SMULBB(input_index, pCurrentCBStage.nVectors)
				}
			} else {
				input_index = 0
				cb_index = pTempIndices.idx(int(k))
			}

			pConstInt = pRes_Q15.off(int(SMULBB(input_index, LPC_order)))
			pCB_element = pCurrentCBStage.CB_NLSF_Q15.off(int(SMULBB(cb_index, LPC_order)))
			pInt = pRes_new_Q15.off(int(SMULBB(k, LPC_order)))
			for i = 0; i < LPC_order; i++ {
				*pInt.ptr(int(i)) = pConstInt.idx(int(i)) - int32(pCB_element.idx(int(i)))
			}

			*pRate_new_Q5.ptr(int(k)) = pRate_Q5.idx(int(input_index)) + int32(pCurrentCBStage.Rates_Q5.idx(int(cb_index)))

			pConstInt = pPath.off(int(SMULBB(input_index, psNLSF_CB.nStages)))
			pInt = pPath_new.off(int(SMULBB(k, psNLSF_CB.nStages)))
			for i = 0; i < s; i++ {
				*pInt.ptr(int(i)) = pConstInt.idx(int(i))
			}
			*pInt.ptr(int(s)) = cb_index
		}

		if s < psNLSF_CB.nStages-1 {
			pRes_new_Q15.copy(pRes_Q15, int(SMULBB(cur_survivors, LPC_order)))
			pRate_new_Q5.copy(pRate_Q5, int(cur_survivors))
			pPath_new.copy(pPath, int(SMULBB(cur_survivors, psNLSF_CB.nStages)))
		}

		prev_survivors = cur_survivors
	}

	bestIndex = 0

	if deactivate_fluc_red != 1 {

		bestRateDist_Q20 = math.MaxInt32
		for s = 0; s < cur_survivors; s++ {
			NLSF_MSVQ_decode(pNLSF_Q15, psNLSF_CB, pPath_new.off(int(SMULBB(s, psNLSF_CB.nStages))), LPC_order)

			wsse_Q20 = 0
			for i = 0; i < LPC_order; i += 2 {
				se_Q15 = pNLSF_Q15.idx(int(i)) - pNLSF_q_Q15_prev.idx(int(i))
				wsse_Q20 = SMLAWB(wsse_Q20, SMULBB(se_Q15, se_Q15), pW_Q6.idx(int(i)))

				se_Q15 = pNLSF_Q15.idx(int(i+1)) - pNLSF_q_Q15_prev.idx(int(i+1))
				wsse_Q20 = SMLAWB(wsse_Q20, SMULBB(se_Q15, se_Q15), pW_Q6.idx(int(i+1)))
			}

			wsse_Q20 = ADD_POS_SAT32(pRateDist_Q18.idx(int(s)), SMULWB(wsse_Q20, NLSF_mu_fluc_red_Q16))

			if wsse_Q20 < bestRateDist_Q20 {
				bestRateDist_Q20 = wsse_Q20
				bestIndex = s
			}
		}
	}

	pPath_new.off(int(SMULBB(bestIndex, psNLSF_CB.nStages))).copy(NLSFIndices, int(psNLSF_CB.nStages))

	NLSF_MSVQ_decode(pNLSF_Q15, psNLSF_CB, NLSFIndices, LPC_order)
}

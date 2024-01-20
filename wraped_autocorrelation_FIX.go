package silk

const QC = 10
const QS = 14

func warped_autocorrelation_FIX(corr *slice[int32], scale *int32, input *slice[int16], warping_Q16 int16, length, order int32) {
	var (
		n, i, lsh        int32
		tmp1_QS, tmp2_QS int32
		state_QS         = alloc[int32](MAX_SHAPE_LPC_ORDER + 1)
		corr_QC          = alloc[int64](MAX_SHAPE_LPC_ORDER + 1)
	)

	for n = 0; n < length; n++ {
		tmp1_QS = LSHIFT32(int32(input.idx(int(n))), QS)
		for i = 0; i < order; i += 2 {
			tmp2_QS = SMLAWB(state_QS.idx(int(i)), state_QS.idx(int(i+1))-tmp1_QS, int32(warping_Q16))
			*state_QS.ptr(int(i)) = tmp1_QS
			*corr_QC.ptr(int(i)) += RSHIFT64(SMULL(tmp1_QS, state_QS.idx(0)), 2*QS-QC)
			tmp1_QS = SMLAWB(state_QS.idx(int(i+1)), state_QS.idx(int(i+2))-tmp2_QS, int32(warping_Q16))
			*state_QS.ptr(int(i + 1)) = tmp2_QS
			*corr_QC.ptr(int(i + 1)) += RSHIFT64(SMULL(tmp2_QS, state_QS.idx(0)), 2*QS-QC)
		}
		*state_QS.ptr(int(order)) = tmp1_QS
		*corr_QC.ptr(int(order)) += RSHIFT64(SMULL(tmp1_QS, state_QS.idx(0)), 2*QS-QC)
	}

	lsh = CLZ64(corr_QC.idx(0)) - 35
	lsh = LIMIT(lsh, -12-QC, 30-QC)
	*scale = -(QC + lsh)

	if lsh >= 0 {
		for i = 0; i < order+1; i++ {
			*corr.ptr(int(i)) = int32(LSHIFT64(corr_QC.idx(int(i)), lsh))
		}
	} else {
		for i = 0; i < order+1; i++ {
			*corr.ptr(int(i)) = int32(RSHIFT64(corr_QC.idx(int(i)), -lsh))
		}
	}
}

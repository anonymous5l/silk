package silk

import "math"

func VQ_WMat_EC_FIX(ind, rate_dist_Q14 *int32,
	in_Q14 *slice[int16], W_Q18 *slice[int32], cb_Q14, cl_Q6 *slice[int16], mu_Q8, L int32) {
	var (
		k                                                        int32
		cb_row_Q14                                               *slice[int16]
		sum1_Q14, sum2_Q16, diff_Q14_01, diff_Q14_23, diff_Q14_4 int32
	)

	*rate_dist_Q14 = math.MaxInt32
	cb_row_Q14 = cb_Q14.off(0)
	for k = 0; k < L; k++ {
		diff_Q14_01 = int32(uint16(in_Q14.idx(0))-uint16(cb_row_Q14.idx(0))) |
			LSHIFT(int32(in_Q14.idx(1))-int32(cb_row_Q14.idx(1)), 16)
		diff_Q14_23 = int32(uint16(in_Q14.idx(2))-uint16(cb_row_Q14.idx(2))) |
			LSHIFT(int32(in_Q14.idx(3))-int32(cb_row_Q14.idx(3)), 16)
		diff_Q14_4 = int32(in_Q14.idx(4) - cb_row_Q14.idx(4))

		sum1_Q14 = SMULBB(mu_Q8, int32(cl_Q6.idx(int(k))))

		sum2_Q16 = SMULWT(W_Q18.idx(1), diff_Q14_01)
		sum2_Q16 = SMLAWB(sum2_Q16, W_Q18.idx(2), diff_Q14_23)
		sum2_Q16 = SMLAWT(sum2_Q16, W_Q18.idx(3), diff_Q14_23)
		sum2_Q16 = SMLAWB(sum2_Q16, W_Q18.idx(4), diff_Q14_4)
		sum2_Q16 = LSHIFT(sum2_Q16, 1)
		sum2_Q16 = SMLAWB(sum2_Q16, W_Q18.idx(0), diff_Q14_01)
		sum1_Q14 = SMLAWB(sum1_Q14, sum2_Q16, diff_Q14_01)

		sum2_Q16 = SMULWB(W_Q18.idx(7), diff_Q14_23)
		sum2_Q16 = SMLAWT(sum2_Q16, W_Q18.idx(8), diff_Q14_23)
		sum2_Q16 = SMLAWB(sum2_Q16, W_Q18.idx(9), diff_Q14_4)
		sum2_Q16 = LSHIFT(sum2_Q16, 1)
		sum2_Q16 = SMLAWT(sum2_Q16, W_Q18.idx(6), diff_Q14_01)
		sum1_Q14 = SMLAWT(sum1_Q14, sum2_Q16, diff_Q14_01)

		sum2_Q16 = SMULWT(W_Q18.idx(13), diff_Q14_23)
		sum2_Q16 = SMLAWB(sum2_Q16, W_Q18.idx(14), diff_Q14_4)
		sum2_Q16 = LSHIFT(sum2_Q16, 1)
		sum2_Q16 = SMLAWB(sum2_Q16, W_Q18.idx(12), diff_Q14_23)
		sum1_Q14 = SMLAWB(sum1_Q14, sum2_Q16, diff_Q14_23)

		sum2_Q16 = SMULWB(W_Q18.idx(19), diff_Q14_4)
		sum2_Q16 = LSHIFT(sum2_Q16, 1)
		sum2_Q16 = SMLAWT(sum2_Q16, W_Q18.idx(18), diff_Q14_23)
		sum1_Q14 = SMLAWT(sum1_Q14, sum2_Q16, diff_Q14_23)

		sum2_Q16 = SMULWB(W_Q18.idx(24), diff_Q14_4)
		sum1_Q14 = SMLAWB(sum1_Q14, sum2_Q16, diff_Q14_4)

		if sum1_Q14 < *rate_dist_Q14 {
			*rate_dist_Q14 = sum1_Q14
			*ind = k
		}

		cb_row_Q14 = cb_row_Q14.off(LTP_ORDER)
	}
}

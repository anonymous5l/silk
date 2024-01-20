package silk

import "math"

func quant_LTP_gains_FIX(
	B_Q14 *slice[int16], cbk_index *slice[int32],
	periodicity_index *int32, W_Q18 *slice[int32], mu_Q8, lowComplexity int32) {
	var (
		j, k                                      int32
		temp_idx                                  = alloc[int32](NB_SUBFR)
		cbk_size                                  int32
		cl_ptr                                    *slice[int16]
		cbk_ptr_Q14                               *slice[int16]
		b_Q14_ptr                                 *slice[int16]
		W_Q18_ptr                                 *slice[int32]
		rate_dist_subfr, rate_dist, min_rate_dist int32
	)
	min_rate_dist = math.MaxInt32
	for k = 0; k < 3; k++ {
		cl_ptr = mem2Slice[int16](LTP_gain_BITS_Q6_ptrs[k])
		cbk_ptr_Q14 = mem2Slice[int16](LTP_vq_ptrs_Q14[k])
		cbk_size = LTP_vq_sizes[k]

		W_Q18_ptr = W_Q18
		b_Q14_ptr = B_Q14

		rate_dist = 0
		for j = 0; j < NB_SUBFR; j++ {
			VQ_WMat_EC_FIX(
				temp_idx.ptr(int(j)),
				&rate_dist_subfr,
				b_Q14_ptr,
				W_Q18_ptr,
				cbk_ptr_Q14,
				cl_ptr,
				mu_Q8,
				cbk_size)

			rate_dist = ADD_POS_SAT32(rate_dist, rate_dist_subfr)

			b_Q14_ptr = b_Q14_ptr.off(LTP_ORDER)
			W_Q18_ptr = W_Q18_ptr.off(LTP_ORDER * LTP_ORDER)
		}

		rate_dist = min(math.MaxInt32-1, rate_dist)

		if rate_dist < min_rate_dist {
			min_rate_dist = rate_dist
			temp_idx.copy(cbk_index, NB_SUBFR)
			*periodicity_index = k
		}

		if lowComplexity != 0 && (rate_dist < LTP_gain_middle_avg_RD_Q14) {
			break
		}
	}

	cbk_ptr_Q14 = mem2Slice[int16](LTP_vq_ptrs_Q14[*periodicity_index])
	for j = 0; j < NB_SUBFR; j++ {
		for k = 0; k < LTP_ORDER; k++ {
			*B_Q14.ptr(int(j*LTP_ORDER + k)) = cbk_ptr_Q14.idx(int(MLA(k, cbk_index.idx(int(j)), LTP_ORDER)))
		}
	}
}

package silk

var (
	sigm_LUT_slope_Q10 = []int32{
		237, 153, 73, 30, 12, 7,
	}
	sigm_LUT_pos_Q15 = []int32{
		16384, 23955, 28861, 31213, 32178, 32548,
	}
	sigm_LUT_neg_Q15 = []int32{
		16384, 8812, 3906, 1554, 589, 219,
	}
)

func sigm_Q15(in_Q5 int32) int32 {
	var ind int32

	if in_Q5 < 0 {
		in_Q5 = -in_Q5
		if in_Q5 >= 6*32 {
			return 0
		} else {
			ind = RSHIFT(in_Q5, 5)
			return sigm_LUT_neg_Q15[ind] - SMULBB(sigm_LUT_slope_Q10[ind], in_Q5&0x1F)
		}
	} else {
		if in_Q5 >= 6*32 {
			return 32767
		} else {
			ind = RSHIFT(in_Q5, 5)
			return sigm_LUT_pos_Q15[ind] + SMULBB(sigm_LUT_slope_Q10[ind], in_Q5&0x1F)
		}
	}
}

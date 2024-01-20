package silk

func inner_prod_aligned(inVec1, inVec2 *slice[int16], len int32) int32 {
	var i, sum int32
	for i = 0; i < len; i++ {
		sum = SMLABB(sum, int32(inVec1.idx(int(i))), int32(inVec2.idx(int(i))))
	}
	return sum
}

func inner_prod16_aligned_64(inVec1, inVec2 *slice[int16], len int32) int64 {
	var (
		i   int32
		sum int64
	)

	for i = 0; i < len; i++ {
		sum = SMLALBB(sum, inVec1.idx(int(i)), inVec2.idx(int(i)))
	}
	return sum
}

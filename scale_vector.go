package silk

func scale_vector32_Q26_lshift_18(data1 *slice[int32], gain_Q26, dataSize int32) {
	var i int32
	for i = 0; i < dataSize; i++ {
		*data1.ptr(int(i)) = int32(RSHIFT64(SMULL(data1.idx(int(i)), gain_Q26), 8))
	}
}

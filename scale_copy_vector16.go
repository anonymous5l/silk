package silk

func scale_copy_vector16(
	data_out,
	data_in *slice[int16],
	gain_Q16, dataSize int32) {
	var i, tmp32 int32

	for i = 0; i < dataSize; i++ {
		tmp32 = SMULWB(gain_Q16, int32(data_in.idx(int(i))))
		*data_out.ptr(int(i)) = int16(tmp32)
	}
}

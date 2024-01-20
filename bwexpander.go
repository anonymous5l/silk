package silk

func bwexpander(ar *slice[int16], d, chirp_Q16 int32) {
	var i, chirp_minus_one_Q16 int32

	chirp_minus_one_Q16 = chirp_Q16 - 65536

	for i = 0; i < d-1; i++ {
		*ar.ptr(int(i)) = int16(RSHIFT_ROUND(MUL(chirp_Q16, int32(ar.idx(int(i)))), 16))
		chirp_Q16 += RSHIFT_ROUND(MUL(chirp_Q16, chirp_minus_one_Q16), 16)
	}
	*ar.ptr(int(d - 1)) = int16(RSHIFT_ROUND(MUL(chirp_Q16, int32(ar.idx(int(d-1)))), 16))
}

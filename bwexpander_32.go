package silk

func bwexpander_32(ar *slice[int32], d, chirp_Q16 int32) {
	var i, tmp_chirp_Q16 int32

	tmp_chirp_Q16 = chirp_Q16
	for i = 0; i < d-1; i++ {
		*ar.ptr(int(i)) = SMULWW(ar.idx(int(i)), tmp_chirp_Q16)
		tmp_chirp_Q16 = SMULWW(chirp_Q16, tmp_chirp_Q16)
	}
	*ar.ptr(int(d - 1)) = SMULWW(ar.idx(int(d-1)), tmp_chirp_Q16)
}

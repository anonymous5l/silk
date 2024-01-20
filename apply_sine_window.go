package silk

var (
	freq_table_Q16 = []int16{
		12111, 9804, 8235, 7100, 6239, 5565, 5022, 4575, 4202,
		3885, 3612, 3375, 3167, 2984, 2820, 2674, 2542, 2422,
		2313, 2214, 2123, 2038, 1961, 1889, 1822, 1760, 1702,
	}
)

func apply_sine_window(
	px_win, px *slice[int16],
	win_type, length int32,
) {
	var (
		k, f_Q16, c_Q16 int32
		S0_Q16, S1_Q16  int32
		px32            int32
	)

	k = (length >> 2) - 4
	f_Q16 = int32(freq_table_Q16[k])

	c_Q16 = SMULWB(f_Q16, -f_Q16)

	if win_type == 1 {
		S0_Q16 = 0
		S1_Q16 = f_Q16 + RSHIFT(length, 3)
	} else {
		S0_Q16 = 1 << 16
		S1_Q16 = (1 << 16) + RSHIFT(c_Q16, 1) + RSHIFT(length, 4)
	}

	px32slice := slice2[int32](px)
	for k = 0; k < length; k += 4 {
		px32 = px32slice.idx(int(k / 2))
		*px_win.ptr(int(k)) = int16(SMULWB(RSHIFT(S0_Q16+S1_Q16, 1), px32))
		*px_win.ptr(int(k + 1)) = int16(SMULWT(S1_Q16, px32))
		S0_Q16 = SMULWB(S1_Q16, c_Q16) + LSHIFT(S1_Q16, 1) - S0_Q16 + 1
		S0_Q16 = min(S0_Q16, 1<<16)

		px32 = px32slice.idx(int((k / 2) + 1))
		*px_win.ptr(int(k + 2)) = int16(SMULWB(RSHIFT(S0_Q16+S1_Q16, 1), px32))
		*px_win.ptr(int(k + 3)) = int16(SMULWT(S0_Q16, px32))
		S1_Q16 = SMULWB(S0_Q16, c_Q16) + LSHIFT(S0_Q16, 1) - S1_Q16
		S1_Q16 = min(S1_Q16, 1<<16)
	}
}

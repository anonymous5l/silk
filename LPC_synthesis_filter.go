package silk

func LPC_synthesis_filter(in *slice[int16], A_Q12 *slice[int16],
	Gain_Q26 int32, S *slice[int32], out *slice[int16], len int32, Order int32) {
	var (
		k, j, idx                      int32
		Order_half                     = RSHIFT(Order, 1)
		SA, SB, out32_Q10, out32, Atmp int32
		A_align_Q12                    [MAX_ORDER_LPC >> 1]int32
	)

	for k = 0; k < Order_half; k++ {
		idx = SMULBB(2, k)
		A_align_Q12[k] = (int32(A_Q12.idx(int(idx))) & 0x0000ffff) | LSHIFT(int32(A_Q12.idx(int(idx+1))), 16)
	}

	for k = 0; k < len; k++ {
		SA = S.idx(int(Order - 1))
		out32_Q10 = 0
		for j = 0; j < (Order_half - 1); j++ {
			idx = SMULBB(2, j) + 1

			Atmp = A_align_Q12[j]
			SB = S.idx(int(Order - 1 - idx))
			*S.ptr(int(Order - 1 - idx)) = SA
			out32_Q10 = SMLAWB(out32_Q10, SA, Atmp)
			out32_Q10 = SMLAWT(out32_Q10, SB, Atmp)
			SA = S.idx(int(Order - 2 - idx))
			*S.ptr(int(Order - 2 - idx)) = SB

		}

		Atmp = A_align_Q12[Order_half-1]
		SB = S.idx(0)
		*S.ptr(0) = SA
		out32_Q10 = SMLAWB(out32_Q10, SA, Atmp)
		out32_Q10 = SMLAWT(out32_Q10, SB, Atmp)

		out32_Q10 = ADD_SAT32(out32_Q10, SMULWB(Gain_Q26, int32(in.idx(int(k)))))

		out32 = RSHIFT_ROUND(out32_Q10, 10)

		*out.ptr(int(k)) = int16(SAT16(out32))

		*S.ptr(int(Order - 1)) = LSHIFT_SAT32(out32_Q10, 4)
	}
}

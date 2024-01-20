package silk

func MA_Prediction(in, B *slice[int16], S *slice[int32], out *slice[int16], length, order int32) {
	var (
		k, d, in16 int32
		out32      int32
	)
	for k = 0; k < length; k++ {
		in16 = int32(in.idx(int(k)))
		out32 = LSHIFT(in16, 12) - S.idx(0)
		out32 = RSHIFT_ROUND(out32, 12)

		for d = 0; d < order-1; d++ {
			*S.ptr(int(d)) = SMLABB_ovflw(S.idx(int(d+1)), in16, int32(B.idx(int(d))))
		}
		*S.ptr(int(order - 1)) = SMULBB(in16, int32(B.idx(int(order-1))))

		*out.ptr(int(k)) = int16(SAT16(out32))
	}
}

func LPC_analysis_filter(
	in, B, S, out *slice[int16],
	length, Order int32) {
	var (
		k, j, idx        int32
		Order_half       = RSHIFT(Order, 1)
		out32_Q12, out32 int32
		SA, SB           int16
	)

	for k = 0; k < length; k++ {
		SA = S.idx(0)
		out32_Q12 = 0
		for j = 0; j < (Order_half - 1); j++ {
			idx = SMULBB(2, j) + 1
			SB = S.idx(int(idx))
			*S.ptr(int(idx)) = SA
			out32_Q12 = SMLABB(out32_Q12, int32(SA), int32(B.idx(int(idx-1))))
			out32_Q12 = SMLABB(out32_Q12, int32(SB), int32(B.idx(int(idx))))
			SA = S.idx(int(idx + 1))
			*S.ptr(int(idx + 1)) = SB
		}

		SB = S.idx(int(Order - 1))
		*S.ptr(int(Order - 1)) = SA
		out32_Q12 = SMLABB(out32_Q12, int32(SA), int32(B.idx(int(Order-2))))
		out32_Q12 = SMLABB(out32_Q12, int32(SB), int32(B.idx(int(Order-1))))

		out32_Q12 = SUB_SAT32(LSHIFT(int32(in.idx(int(k))), 12), out32_Q12)

		out32 = RSHIFT_ROUND(out32_Q12, 12)

		*out.ptr(int(k)) = int16(SAT16(out32))

		*S.ptr(0) = in.idx(int(k))
	}
}

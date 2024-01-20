package silk

func LPC_synthesis_order16(in *slice[int16], A_Q12 *slice[int16], Gain_Q26 int32,
	S *slice[int32], out *slice[int16], len int32) {

	var (
		k                        int32
		SA, SB, out32_Q10, out32 int32

		Atmp        int32
		A_align_Q12 [8]int32
	)

	for k = 0; k < 8; k++ {
		A_align_Q12[k] = (int32(A_Q12.idx(int(2*k))) & 0x0000ffff) | LSHIFT(int32(A_Q12.idx(int(2*k+1))), 16)
	}

	for k = 0; k < len; k++ {
		SA = S.idx(15)
		Atmp = A_align_Q12[0]
		SB = S.idx(14)
		*S.ptr(14) = SA
		out32_Q10 = SMULWB(SA, Atmp)
		out32_Q10 = SMLAWT_ovflw(out32_Q10, SB, Atmp)
		SA = S.idx(13)
		*S.ptr(13) = SB

		Atmp = A_align_Q12[1]
		SB = S.idx(12)
		*S.ptr(12) = SA
		out32_Q10 = SMLAWB_ovflw(out32_Q10, SA, Atmp)
		out32_Q10 = SMLAWT_ovflw(out32_Q10, SB, Atmp)
		SA = S.idx(11)
		*S.ptr(11) = SB

		Atmp = A_align_Q12[2]
		SB = S.idx(10)
		*S.ptr(10) = SA
		out32_Q10 = SMLAWB_ovflw(out32_Q10, SA, Atmp)
		out32_Q10 = SMLAWT_ovflw(out32_Q10, SB, Atmp)
		SA = S.idx(9)
		*S.ptr(9) = SB

		Atmp = A_align_Q12[3]
		SB = S.idx(8)
		*S.ptr(8) = SA
		out32_Q10 = SMLAWB_ovflw(out32_Q10, SA, Atmp)
		out32_Q10 = SMLAWT_ovflw(out32_Q10, SB, Atmp)
		SA = S.idx(7)
		*S.ptr(7) = SB

		Atmp = A_align_Q12[4]
		SB = S.idx(6)
		*S.ptr(6) = SA
		out32_Q10 = SMLAWB_ovflw(out32_Q10, SA, Atmp)
		out32_Q10 = SMLAWT_ovflw(out32_Q10, SB, Atmp)
		SA = S.idx(5)
		*S.ptr(5) = SB

		Atmp = A_align_Q12[5]
		SB = S.idx(4)
		*S.ptr(4) = SA
		out32_Q10 = SMLAWB_ovflw(out32_Q10, SA, Atmp)
		out32_Q10 = SMLAWT_ovflw(out32_Q10, SB, Atmp)
		SA = S.idx(3)
		*S.ptr(3) = SB

		Atmp = A_align_Q12[6]
		SB = S.idx(2)
		*S.ptr(2) = SA
		out32_Q10 = SMLAWB_ovflw(out32_Q10, SA, Atmp)
		out32_Q10 = SMLAWT_ovflw(out32_Q10, SB, Atmp)
		SA = S.idx(1)
		*S.ptr(1) = SB

		Atmp = A_align_Q12[7]
		SB = S.idx(0)
		*S.ptr(0) = SA
		out32_Q10 = SMLAWB_ovflw(out32_Q10, SA, Atmp)
		out32_Q10 = SMLAWT_ovflw(out32_Q10, SB, Atmp)

		out32_Q10 = ADD_SAT32(out32_Q10, SMULWB(Gain_Q26, int32(in.idx(int(k)))))

		out32 = RSHIFT_ROUND(out32_Q10, 10)

		*out.ptr(int(k)) = int16(SAT16(out32))

		*S.ptr(15) = LSHIFT_SAT32(out32_Q10, 4)
	}
}

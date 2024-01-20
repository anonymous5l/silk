package silk

var (
	A_fb1_20 = []int16{5394 << 1}
	A_fb1_21 = []int16{-24290}
)

func ana_filt_bank_1(
	in *slice[int16], S *slice[int32], outL, outH *slice[int16], scratch *int32, N int32) {
	var (
		k                        int32
		N2                       = RSHIFT(N, 1)
		in32, X, Y, out_1, out_2 int32
	)

	for k = 0; k < N2; k++ {
		in32 = LSHIFT(int32(in.idx(int(2*k))), 10)

		Y = SUB32(in32, S.idx(0))
		X = SMLAWB(Y, Y, int32(A_fb1_21[0]))
		out_1 = ADD32(S.idx(0), X)
		*S.ptr(0) = ADD32(in32, X)

		in32 = LSHIFT(int32(in.idx(int(2*k+1))), 10)

		Y = SUB32(in32, S.idx(1))
		X = SMULWB(Y, int32(A_fb1_20[0]))
		out_2 = ADD32(S.idx(1), X)
		*S.ptr(1) = ADD32(in32, X)

		*outL.ptr(int(k)) = int16(SAT16(RSHIFT_ROUND(ADD32(out_2, out_1), 11)))
		*outH.ptr(int(k)) = int16(SAT16(RSHIFT_ROUND(SUB32(out_2, out_1), 11)))
	}
}

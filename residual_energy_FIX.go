package silk

func residual_energy_FIX(
	nrgs *slice[int32],
	nrgsQ *slice[int32],
	x *slice[int16],
	a_Q12 [2]*slice[int16],
	gains *slice[int32],
	subfr_length, LPC_order int32) {
	var (
		offset, i, j, rshift, lz1, lz2 int32
		LPC_res_ptr                    *slice[int16]
		LPC_res                        = alloc[int16]((MAX_FRAME_LENGTH + NB_SUBFR*MAX_LPC_ORDER) / 2)
		x_ptr                          *slice[int16]
		S                              = alloc[int16](MAX_LPC_ORDER)
		tmp32                          int32
	)

	x_ptr = x
	offset = LPC_order + subfr_length

	for i = 0; i < 2; i++ {
		memset(S, 0, int(LPC_order))
		LPC_analysis_filter(x_ptr, a_Q12[i], S, LPC_res, (NB_SUBFR>>1)*offset, LPC_order)

		LPC_res_ptr = LPC_res.off(int(LPC_order))
		for j = 0; j < (NB_SUBFR >> 1); j++ {
			sum_sqr_shift(nrgs.ptr(int(i*(NB_SUBFR>>1)+j)), &rshift, LPC_res_ptr, subfr_length)

			*nrgsQ.ptr(int(i*(NB_SUBFR>>1) + j)) = -rshift

			LPC_res_ptr = LPC_res_ptr.off(int(offset))
		}
		x_ptr = x_ptr.off(int((NB_SUBFR >> 1) * offset))
	}

	for i = 0; i < NB_SUBFR; i++ {
		lz1 = CLZ32(nrgs.idx(int(i))) - 1
		lz2 = CLZ32(gains.idx(int(i))) - 1

		tmp32 = LSHIFT32(gains.idx(int(i)), lz2)

		tmp32 = SMMUL(tmp32, tmp32)

		*nrgs.ptr(int(i)) = SMMUL(tmp32, LSHIFT32(nrgs.idx(int(i)), lz1))
		*nrgsQ.ptr(int(i)) += lz1 + 2*lz2 - 32 - 32
	}
}

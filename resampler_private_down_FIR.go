package silk

const (
	RESAMPLER_DOWN_ORDER_FIR = 12
	RESAMPLER_ORDER_FIR_144  = 6
)

func resampler_private_down_FIR_INTERPOL0(out *slice[int16], buf2 *slice[int32],
	FIR_Coefs *slice[int16], max_index_Q16, index_increment_Q16 int32) *slice[int16] {

	var (
		index_Q16, res_Q6 int32
		buf_ptr           *slice[int32]
	)

	for index_Q16 = 0; index_Q16 < max_index_Q16; index_Q16 += index_increment_Q16 {
		buf_ptr = buf2.off(int(RSHIFT(index_Q16, 16)))

		res_Q6 = SMULWB(ADD32(buf_ptr.idx(0), buf_ptr.idx(1)), int32(FIR_Coefs.idx(0)))
		res_Q6 = SMLAWB(res_Q6, ADD32(buf_ptr.idx(1), buf_ptr.idx(10)), int32(FIR_Coefs.idx(1)))
		res_Q6 = SMLAWB(res_Q6, ADD32(buf_ptr.idx(2), buf_ptr.idx(9)), int32(FIR_Coefs.idx(2)))
		res_Q6 = SMLAWB(res_Q6, ADD32(buf_ptr.idx(3), buf_ptr.idx(8)), int32(FIR_Coefs.idx(3)))
		res_Q6 = SMLAWB(res_Q6, ADD32(buf_ptr.idx(4), buf_ptr.idx(7)), int32(FIR_Coefs.idx(4)))
		res_Q6 = SMLAWB(res_Q6, ADD32(buf_ptr.idx(5), buf_ptr.idx(6)), int32(FIR_Coefs.idx(5)))

		*out.ptr(0) = int16(SAT16(RSHIFT_ROUND(res_Q6, 6)))
		out = out.off(1)
	}
	return out
}

func resampler_private_down_FIR_INTERPOL1(out *slice[int16], buf2 *slice[int32],
	FIR_Coefs *slice[int16], max_index_Q16, index_increment_Q16, FIR_Fracs int32) *slice[int16] {

	var (
		index_Q16, res_Q6 int32
		buf_ptr           *slice[int32]
		interpol_ind      int32
		interpol_ptr      *slice[int16]
	)

	out = out.off(0)

	for index_Q16 = 0; index_Q16 < max_index_Q16; index_Q16 += index_increment_Q16 {
		buf_ptr = buf2.off(int(RSHIFT(index_Q16, 16)))

		interpol_ind = SMULWB(index_Q16&0xFFFF, FIR_Fracs)

		interpol_ptr = FIR_Coefs.off(int(RESAMPLER_DOWN_ORDER_FIR / 2 * interpol_ind))
		res_Q6 = SMULWB(buf_ptr.idx(0), int32(interpol_ptr.idx(0)))
		res_Q6 = SMLAWB(res_Q6, buf_ptr.idx(1), int32(interpol_ptr.idx(1)))
		res_Q6 = SMLAWB(res_Q6, buf_ptr.idx(2), int32(interpol_ptr.idx(2)))
		res_Q6 = SMLAWB(res_Q6, buf_ptr.idx(3), int32(interpol_ptr.idx(3)))
		res_Q6 = SMLAWB(res_Q6, buf_ptr.idx(4), int32(interpol_ptr.idx(4)))
		res_Q6 = SMLAWB(res_Q6, buf_ptr.idx(5), int32(interpol_ptr.idx(5)))
		interpol_ptr = FIR_Coefs.off(int(RESAMPLER_DOWN_ORDER_FIR / 2 * (FIR_Fracs - 1 - interpol_ind)))
		res_Q6 = SMLAWB(res_Q6, buf_ptr.idx(11), int32(interpol_ptr.idx(0)))
		res_Q6 = SMLAWB(res_Q6, buf_ptr.idx(10), int32(interpol_ptr.idx(1)))
		res_Q6 = SMLAWB(res_Q6, buf_ptr.idx(9), int32(interpol_ptr.idx(2)))
		res_Q6 = SMLAWB(res_Q6, buf_ptr.idx(8), int32(interpol_ptr.idx(3)))
		res_Q6 = SMLAWB(res_Q6, buf_ptr.idx(7), int32(interpol_ptr.idx(4)))
		res_Q6 = SMLAWB(res_Q6, buf_ptr.idx(6), int32(interpol_ptr.idx(5)))

		*out.ptr(0) = int16(SAT16(RSHIFT_ROUND(res_Q6, 6)))
		out = out.off(1)
	}
	return out
}

func resampler_private_down_FIR(S *resampler_state_struct, out, in *slice[int16], inLen int32) {
	var (
		nSamplesIn, max_index_Q16, index_increment_Q16 int32
		buf1                                           = alloc[int16](RESAMPLER_MAX_BATCH_SIZE_IN / 2)
		buf2                                           = alloc[int32](RESAMPLER_MAX_BATCH_SIZE_IN + RESAMPLER_DOWN_ORDER_FIR)
		FIR_Coefs                                      *slice[int16]
	)

	S.sFIR.copy(buf2, RESAMPLER_DOWN_ORDER_FIR)

	FIR_Coefs = S.Coefs.off(2)

	index_increment_Q16 = S.invRatio_Q16
	for {
		nSamplesIn = min(inLen, S.batchSize)

		if S.input2x == 1 {
			resampler_down2(S.sDown2, buf1, in, nSamplesIn)

			nSamplesIn = RSHIFT32(nSamplesIn, 1)

			resampler_private_AR2(S.sIIR, buf2.off(RESAMPLER_DOWN_ORDER_FIR), buf1, S.Coefs, nSamplesIn)
		} else {
			resampler_private_AR2(S.sIIR, buf2.off(RESAMPLER_DOWN_ORDER_FIR), in, S.Coefs, nSamplesIn)
		}

		max_index_Q16 = LSHIFT32(nSamplesIn, 16)

		if S.FIR_Fracs == 1 {
			out = resampler_private_down_FIR_INTERPOL0(out, buf2, FIR_Coefs, max_index_Q16, index_increment_Q16)
		} else {
			out = resampler_private_down_FIR_INTERPOL1(out, buf2, FIR_Coefs, max_index_Q16, index_increment_Q16, S.FIR_Fracs)
		}

		in = in.off(int(nSamplesIn << S.input2x))
		inLen -= nSamplesIn << S.input2x

		if inLen > S.input2x {
			buf2.off(int(nSamplesIn)).copy(buf2, RESAMPLER_DOWN_ORDER_FIR)
		} else {
			break
		}
	}

	buf2.off(int(nSamplesIn)).copy(S.sFIR, RESAMPLER_DOWN_ORDER_FIR)
}

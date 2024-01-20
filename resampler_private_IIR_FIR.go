package silk

func resampler_private_IIR_FIR_INTERPOL(out, buf *slice[int16], max_index_Q16, index_increment_Q16 int32) *slice[int16] {
	var (
		index_Q16, res_Q15 int32
		buf_ptr            *slice[int16]
		table_index        int32
	)

	for index_Q16 = 0; index_Q16 < max_index_Q16; index_Q16 += index_increment_Q16 {
		table_index = SMULWB(index_Q16&0xFFFF, 144)
		buf_ptr = buf.off(int(index_Q16 >> 16))

		res_Q15 = SMULBB(int32(buf_ptr.idx(0)), int32(resampler_frac_FIR_144[table_index][0]))
		res_Q15 = SMLABB(res_Q15, int32(buf_ptr.idx(1)), int32(resampler_frac_FIR_144[table_index][1]))
		res_Q15 = SMLABB(res_Q15, int32(buf_ptr.idx(2)), int32(resampler_frac_FIR_144[table_index][2]))
		res_Q15 = SMLABB(res_Q15, int32(buf_ptr.idx(3)), int32(resampler_frac_FIR_144[143-table_index][2]))
		res_Q15 = SMLABB(res_Q15, int32(buf_ptr.idx(4)), int32(resampler_frac_FIR_144[143-table_index][1]))
		res_Q15 = SMLABB(res_Q15, int32(buf_ptr.idx(5)), int32(resampler_frac_FIR_144[143-table_index][0]))
		*out.ptr(0) = int16(SAT16(RSHIFT_ROUND(res_Q15, 15)))
		out = out.off(1)
	}
	return out
}

func resampler_private_IIR_FIR(S *resampler_state_struct, out, in *slice[int16], inLen int32) {
	var (
		nSamplesIn, max_index_Q16, index_increment_Q16 int32
		buf                                            = alloc[int16](2*RESAMPLER_MAX_BATCH_SIZE_IN + RESAMPLER_ORDER_FIR_144)
	)

	sFIR := slice2[int16](S.sFIR)
	sFIR.copy(buf, RESAMPLER_ORDER_FIR_144*2)

	index_increment_Q16 = S.invRatio_Q16
	for {
		nSamplesIn = min(inLen, S.batchSize)

		if S.input2x == 1 {
			S.up2_function(S.sIIR, buf.off(RESAMPLER_ORDER_FIR_144), in, nSamplesIn)
		} else {
			resampler_private_ARMA4(S.sIIR, buf.off(RESAMPLER_ORDER_FIR_144), in, S.Coefs, nSamplesIn)
		}

		max_index_Q16 = LSHIFT32(nSamplesIn, 16+S.input2x)
		out = resampler_private_IIR_FIR_INTERPOL(out, buf, max_index_Q16, index_increment_Q16)
		in = in.off(int(nSamplesIn))
		inLen -= nSamplesIn

		if inLen > 0 {
			buf.off(int(nSamplesIn<<S.input2x)).copy(buf, RESAMPLER_ORDER_FIR_144*2)
		} else {
			break
		}
	}

	buf.off(int(nSamplesIn<<S.input2x)).copy(sFIR, RESAMPLER_ORDER_FIR_144*2)
}

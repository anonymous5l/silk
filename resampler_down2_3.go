package silk

const DOWN23_ORDER_FIR = 4

func resampler_down2_3(S *slice[int32], out, in *slice[int16], inLen int32) {
	var (
		nSamplesIn, counter, res_Q6 int32
		buf                         = alloc[int32](RESAMPLER_MAX_BATCH_SIZE_IN + DOWN23_ORDER_FIR)
		buf_ptr                     *slice[int32]
	)

	S.copy(buf, DOWN23_ORDER_FIR)

	coefs := mem2Slice[int16](Resampler_2_3_COEFS_LQ)

	for {
		nSamplesIn = min(inLen, RESAMPLER_MAX_BATCH_SIZE_IN)

		resampler_private_AR2(S.off(DOWN23_ORDER_FIR), buf.off(DOWN23_ORDER_FIR), in,
			coefs.off(0), nSamplesIn)

		buf_ptr = buf.off(0)
		counter = nSamplesIn
		for counter > 2 {
			res_Q6 = SMULWB(buf_ptr.idx(0), int32(coefs.idx(2)))
			res_Q6 = SMLAWB(res_Q6, buf_ptr.idx(1), int32(coefs.idx(3)))
			res_Q6 = SMLAWB(res_Q6, buf_ptr.idx(2), int32(coefs.idx(5)))
			res_Q6 = SMLAWB(res_Q6, buf_ptr.idx(3), int32(coefs.idx(4)))

			*out.ptr(0) = int16(SAT16(RSHIFT_ROUND(res_Q6, 6)))
			out = out.off(1)

			res_Q6 = SMULWB(buf_ptr.idx(1), int32(coefs.idx(4)))
			res_Q6 = SMLAWB(res_Q6, buf_ptr.idx(2), int32(coefs.idx(5)))
			res_Q6 = SMLAWB(res_Q6, buf_ptr.idx(3), int32(coefs.idx(3)))
			res_Q6 = SMLAWB(res_Q6, buf_ptr.idx(4), int32(coefs.idx(2)))

			*out.ptr(0) = int16(SAT16(RSHIFT_ROUND(res_Q6, 6)))
			out = out.off(1)

			buf_ptr = buf_ptr.off(3)
			counter -= 3
		}

		in = in.off(int(nSamplesIn))
		inLen -= nSamplesIn

		if inLen > 0 {
			buf.off(int(nSamplesIn)).copy(buf, DOWN23_ORDER_FIR)
		} else {
			break
		}
	}

	buf.off(int(nSamplesIn)).copy(S, DOWN23_ORDER_FIR)
}

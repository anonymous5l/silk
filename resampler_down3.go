package silk

const ORDER_FIR = 6

func resampler_down3(S *slice[int32], out, in *slice[int16], inLen int32) {
	var (
		nSamplesIn, counter, res_Q6 int32

		buf     = alloc[int32](RESAMPLER_MAX_BATCH_SIZE_IN + ORDER_FIR)
		buf_ptr *slice[int32]
	)

	S.copy(buf, ORDER_FIR)

	coefs := mem2Slice[int16](Resampler_1_3_COEFS_LQ)

	for {
		nSamplesIn = min(inLen, RESAMPLER_MAX_BATCH_SIZE_IN)

		resampler_private_AR2(S.off(ORDER_FIR), buf.off(ORDER_FIR), in, coefs, nSamplesIn)

		buf_ptr = buf.off(0)
		counter = nSamplesIn
		for counter > 2 {
			res_Q6 = SMULWB(ADD32(buf_ptr.idx(0), buf_ptr.idx(5)), int32(coefs.idx(2)))
			res_Q6 = SMLAWB(res_Q6, ADD32(buf_ptr.idx(1), buf_ptr.idx(4)), int32(coefs.idx(3)))
			res_Q6 = SMLAWB(res_Q6, ADD32(buf_ptr.idx(2), buf_ptr.idx(3)), int32(coefs.idx(4)))

			*out.ptr(0) = int16(SAT16(RSHIFT_ROUND(res_Q6, 6)))
			out = out.off(1)

			buf_ptr = buf_ptr.off(3)
			counter -= 3
		}

		in = in.off(int(nSamplesIn))
		inLen -= nSamplesIn

		if inLen > 0 {
			buf.off(int(nSamplesIn)).copy(buf, ORDER_FIR)
		} else {
			break
		}
	}

	buf.off(int(nSamplesIn)).copy(S, ORDER_FIR)
}

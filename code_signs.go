package silk

func enc_map(a int32) int32 {
	return RSHIFT(a, 15) + 1
}

func dec_map(a int32) int32 {
	return LSHIFT(a, 1) - 1
}

func encode_signs(sRC *range_coder_state, q *slice[int8], length int32, sigtype int32, QuantOffsetType int32, RateLevelIndex int32) {
	var (
		i, inData int32
		cdf       [3]uint16
	)

	i = SMULBB(N_RATE_LEVELS-1, LSHIFT(sigtype, 1)+QuantOffsetType) + RateLevelIndex

	cdf[1] = sign_CDF[i]
	cdf[2] = 65535

	for i = 0; i < length; i++ {
		if q.idx(int(i)) != 0 {
			inData = enc_map(int32(q.idx(int(i))))
			range_encoder(sRC, inData, cdf[:])
		}
	}
}

func decode_signs(sRC *range_coder_state, q *slice[int32], length int32, sigtype int32, QuantOffsetType int32, RateLevelIndex int32) {
	var (
		i, data int32
		cdf     [3]uint16
	)

	i = SMULBB(N_RATE_LEVELS-1, LSHIFT(sigtype, 1)+QuantOffsetType) + RateLevelIndex
	cdf[1] = sign_CDF[i]
	cdf[2] = 65535

	for i = 0; i < length; i++ {
		if q.idx(int(i)) > 0 {
			range_decoder(&data, sRC, cdf[:], 1)
			*q.ptr(int(i)) *= dec_map(data)
		}
	}
}

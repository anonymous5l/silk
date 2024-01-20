package silk

func decode_pulses(psRC *range_coder_state, psDecCtrl *decoder_control, q *slice[int32], frame_length int32) {

	var (
		i, j, k, iter, abs_q, nLS, bit int32
		sum_pulses                     [MAX_NB_SHELL_BLOCKS]int32
		nLshifts                       [MAX_NB_SHELL_BLOCKS]int32

		pulses_ptr *slice[int32]
		cdf_ptr    []uint16
	)

	range_decoder(&psDecCtrl.RateLevelIndex, psRC,
		rate_levels_CDF[psDecCtrl.sigtype], rate_levels_CDF_offset)

	iter = frame_length / SHELL_CODEC_FRAME_LENGTH

	cdf_ptr = pulses_per_block_CDF[psDecCtrl.RateLevelIndex]
	for i = 0; i < iter; i++ {
		nLshifts[i] = 0
		range_decoder(&sum_pulses[i], psRC, cdf_ptr, pulses_per_block_CDF_offset)

		for sum_pulses[i] == (MAX_PULSES + 1) {
			nLshifts[i]++
			range_decoder(&sum_pulses[i], psRC,
				pulses_per_block_CDF[N_RATE_LEVELS-1], pulses_per_block_CDF_offset)
		}
	}

	for i = 0; i < iter; i++ {
		if sum_pulses[i] > 0 {
			shell_decoder(q.off(int(SMULBB(i, SHELL_CODEC_FRAME_LENGTH))), psRC, sum_pulses[i])
		} else {
			memset(q.off(int(SMULBB(i, SHELL_CODEC_FRAME_LENGTH))), 0, SHELL_CODEC_FRAME_LENGTH)
		}
	}

	for i = 0; i < iter; i++ {
		if nLshifts[i] > 0 {
			nLS = nLshifts[i]
			pulses_ptr = q.off(int(SMULBB(i, SHELL_CODEC_FRAME_LENGTH)))
			for k = 0; k < SHELL_CODEC_FRAME_LENGTH; k++ {
				abs_q = pulses_ptr.idx(int(k))
				for j = 0; j < nLS; j++ {
					abs_q = LSHIFT(abs_q, 1)
					range_decoder(&bit, psRC, lsb_CDF, 1)
					abs_q += bit
				}
				*pulses_ptr.ptr(int(k)) = abs_q
			}
		}
	}

	decode_signs(psRC, q, frame_length, psDecCtrl.sigtype,
		psDecCtrl.QuantOffsetType, psDecCtrl.RateLevelIndex)
}

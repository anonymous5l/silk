package silk

import (
	"math"
)

func combine_and_check(
	pulses_comb, pulses_in *slice[int32],
	max_pulses, len int32) int32 {
	var k, sum int32

	for k = 0; k < len; k++ {
		sum = pulses_in.idx(int(2*k)) + pulses_in.idx(int(2*k+1))
		if sum > max_pulses {
			return 1
		}
		*pulses_comb.ptr(int(k)) = sum
	}

	return 0
}

func encode_pulses(psRC *range_coder_state, sigtype, QuantOffsetType int32, q *slice[int8], frame_length int32) {
	var (
		i, k, j, iter, bit, nLS, scale_down, RateLevelIndex int32
		abs_q, minSumBits_Q6, sumBits_Q6                    int32
		abs_pulses                                          = alloc[int32](MAX_FRAME_LENGTH)
		sum_pulses                                          = alloc[int32](MAX_NB_SHELL_BLOCKS)
		nRshifts                                            = alloc[int32](MAX_NB_SHELL_BLOCKS)
		pulses_comb                                         = alloc[int32](8)
		abs_pulses_ptr                                      *slice[int32]
		pulses_ptr                                          *slice[int8]
		cdf_ptr                                             []uint16
		nBits_ptr                                           *slice[int16]
	)

	iter = frame_length / SHELL_CODEC_FRAME_LENGTH

	for i = 0; i < frame_length; i += 4 {
		*abs_pulses.ptr(int(i + 0)) = abs(int32(q.idx(int(i + 0))))
		*abs_pulses.ptr(int(i + 1)) = abs(int32(q.idx(int(i + 1))))
		*abs_pulses.ptr(int(i + 2)) = abs(int32(q.idx(int(i + 2))))
		*abs_pulses.ptr(int(i + 3)) = abs(int32(q.idx(int(i + 3))))
	}

	abs_pulses_ptr = abs_pulses
	for i = 0; i < iter; i++ {
		*nRshifts.ptr(int(i)) = 0

		for {
			scale_down = combine_and_check(pulses_comb, abs_pulses_ptr, max_pulses_table[0], 8)

			scale_down += combine_and_check(pulses_comb, pulses_comb, max_pulses_table[1], 4)

			scale_down += combine_and_check(pulses_comb, pulses_comb, max_pulses_table[2], 2)

			*sum_pulses.ptr(int(i)) = pulses_comb.idx(0) + pulses_comb.idx(1)
			if sum_pulses.idx(int(i)) > max_pulses_table[3] {
				scale_down++
			}

			if scale_down != 0 {
				*nRshifts.ptr(int(i))++
				for k = 0; k < SHELL_CODEC_FRAME_LENGTH; k++ {
					*abs_pulses_ptr.ptr(int(k)) = RSHIFT(abs_pulses_ptr.idx(int(k)), 1)
				}
			} else {
				break
			}
		}
		abs_pulses_ptr = abs_pulses_ptr.off(SHELL_CODEC_FRAME_LENGTH)
	}

	minSumBits_Q6 = math.MaxInt32
	for k = 0; k < N_RATE_LEVELS-1; k++ {
		nBits_ptr = mem2Slice[int16](pulses_per_block_BITS_Q6[k])
		sumBits_Q6 = int32(rate_levels_BITS_Q6[sigtype][k])
		for i = 0; i < iter; i++ {
			if nRshifts.idx(int(i)) > 0 {
				sumBits_Q6 += int32(nBits_ptr.idx(MAX_PULSES + 1))
			} else {
				sumBits_Q6 += int32(nBits_ptr.idx(int(sum_pulses.idx(int(i)))))
			}
		}
		if sumBits_Q6 < minSumBits_Q6 {
			minSumBits_Q6 = sumBits_Q6
			RateLevelIndex = k
		}
	}

	range_encoder(psRC, RateLevelIndex, rate_levels_CDF[sigtype])

	cdf_ptr = pulses_per_block_CDF[RateLevelIndex]
	for i = 0; i < iter; i++ {
		if nRshifts.idx(int(i)) == 0 {
			range_encoder(psRC, sum_pulses.idx(int(i)), cdf_ptr)
		} else {
			range_encoder(psRC, MAX_PULSES+1, cdf_ptr)
			for k = 0; k < nRshifts.idx(int(i))-1; k++ {
				range_encoder(psRC, MAX_PULSES+1, pulses_per_block_CDF[N_RATE_LEVELS-1])
			}
			range_encoder(psRC, sum_pulses.idx(int(i)), pulses_per_block_CDF[N_RATE_LEVELS-1])
		}
	}

	for i = 0; i < iter; i++ {
		if sum_pulses.idx(int(i)) > 0 {
			shell_encoder(psRC, abs_pulses.off(int(i*SHELL_CODEC_FRAME_LENGTH)))
		}
	}

	for i = 0; i < iter; i++ {
		if nRshifts.idx(int(i)) > 0 {
			pulses_ptr = q.off(int(i * SHELL_CODEC_FRAME_LENGTH))
			nLS = nRshifts.idx(int(i)) - 1
			for k = 0; k < SHELL_CODEC_FRAME_LENGTH; k++ {
				abs_q = abs(int32(pulses_ptr.idx(int(k))))
				for j = nLS; j > 0; j-- {
					bit = RSHIFT(abs_q, j) & 1
					range_encoder(psRC, bit, lsb_CDF)
				}
				bit = abs_q & 1
				range_encoder(psRC, bit, lsb_CDF)
			}
		}
	}

	encode_signs(psRC, q, frame_length, sigtype, QuantOffsetType, RateLevelIndex)
}

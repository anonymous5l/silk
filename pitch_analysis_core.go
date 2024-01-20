package silk

import (
	"math"
)

const (
	SCRATCH_SIZE                   = 22
	PITCH_EST_SHORTLAG_BIAS_Q15    = 6554
	PITCH_EST_PREVLAG_BIAS_Q15     = 6554
	PITCH_EST_FLATCONTOUR_BIAS_Q20 = 52429
)

func pitch_analysis_core(signal *slice[int16], pitch_out *slice[int32],
	lagIndex, contourIndex, LTPCorr_Q15 *int32,
	prevLag, search_thres1_Q16, search_thres2_Q15, Fs_kHz, complexity int32, forLJC int32) int32 {

	var (
		signal_8kHz                                                                            = alloc[int16](PITCH_EST_MAX_FRAME_LENGTH_ST_2)
		signal_4kHz                                                                            = alloc[int16](PITCH_EST_MAX_FRAME_LENGTH_ST_1)
		scratch_mem                                                                            = alloc[int16](3 * PITCH_EST_MAX_FRAME_LENGTH * 2)
		input_signal_ptr                                                                       *slice[int16]
		filt_state                                                                             = alloc[int32](PITCH_EST_MAX_DECIMATE_STATE_LENGTH)
		i, k, d, j                                                                             int32
		C                                                                                      [PITCH_EST_NB_SUBFR]*slice[int16]
		target_ptr, basis_ptr                                                                  *slice[int16]
		cross_corr, normalizer, energy, shift, energy_basis, energy_target                     int32
		d_srch                                                                                 = alloc[int32](PITCH_EST_D_SRCH_LENGTH)
		d_comp                                                                                 = alloc[int16]((PITCH_EST_MAX_LAG >> 1) + 5)
		Cmax, length_d_srch, length_d_comp                                                     int32
		sum, threshold, temp32                                                                 int32
		CBimax, CBimax_new, CBimax_old, lag, start_lag, end_lag, lag_new                       int32
		CC                                                                                     [PITCH_EST_NB_CBKS_STAGE2_EXT]int32
		CCmax, CCmax_b, CCmax_new_b, CCmax_new                                                 int32
		energies_st3                                                                           [PITCH_EST_NB_SUBFR][PITCH_EST_NB_CBKS_STAGE3_MAX][PITCH_EST_NB_STAGE3_LAGS]int32
		crosscorr_st3                                                                          [PITCH_EST_NB_SUBFR][PITCH_EST_NB_CBKS_STAGE3_MAX][PITCH_EST_NB_STAGE3_LAGS]int32
		lag_counter                                                                            int32
		frame_length, frame_length_8kHz, frame_length_4kHz, max_sum_sq_length                  int32
		sf_length, sf_length_8kHz                                                              int32
		min_lag, min_lag_8kHz, min_lag_4kHz                                                    int32
		max_lag, max_lag_8kHz, max_lag_4kHz                                                    int32
		contour_bias, diff                                                                     int32
		lz, lshift                                                                             int32
		cbk_offset, cbk_size, nb_cbks_stage2                                                   int32
		delta_lag_log2_sqr_Q7, lag_log2_Q7, prevLag_log2_Q7, prev_lag_bias_Q15, corr_thres_Q15 int32
	)

	resetC := func() {
		for q := 0; q < len(C); q++ {
			C[q] = alloc[int16]((PITCH_EST_MAX_LAG >> 1) + 5)
		}
	}
	resetC()

	frame_length = PITCH_EST_FRAME_LENGTH_MS * Fs_kHz
	frame_length_4kHz = PITCH_EST_FRAME_LENGTH_MS * 4
	frame_length_8kHz = PITCH_EST_FRAME_LENGTH_MS * 8
	sf_length = RSHIFT(frame_length, 3)
	sf_length_8kHz = RSHIFT(frame_length_8kHz, 3)
	min_lag = PITCH_EST_MIN_LAG_MS * Fs_kHz
	min_lag_4kHz = PITCH_EST_MIN_LAG_MS * 4
	min_lag_8kHz = PITCH_EST_MIN_LAG_MS * 8
	max_lag = PITCH_EST_MAX_LAG_MS * Fs_kHz
	max_lag_4kHz = PITCH_EST_MAX_LAG_MS * 4
	max_lag_8kHz = PITCH_EST_MAX_LAG_MS * 8

	if Fs_kHz == 16 {
		resampler_down2(filt_state, signal_8kHz, signal, frame_length)
	} else if Fs_kHz == 12 {
		R23 := alloc[int32](6)
		resampler_down2_3(R23, signal_8kHz, signal, PITCH_EST_FRAME_LENGTH_MS*12)
	} else if Fs_kHz == 24 {
		filt_state_fix := alloc[int32](8)
		resampler_down3(filt_state_fix, signal_8kHz, signal, 24*PITCH_EST_FRAME_LENGTH_MS)
	} else {
		signal.copy(signal_8kHz, int(frame_length_8kHz))
	}
	memset(filt_state, 0, 2)
	resampler_down2(filt_state, signal_4kHz, signal_8kHz, frame_length_8kHz)

	for i = frame_length_4kHz - 1; i > 0; i-- {
		*signal_4kHz.ptr(int(i)) = ADD_SAT16(signal_4kHz.idx(int(i)), signal_4kHz.idx(int(i-1)))
	}

	max_sum_sq_length = max(sf_length_8kHz, RSHIFT(frame_length_4kHz, 1))
	shift = FIX_P_Ana_find_scaling(signal_4kHz, frame_length_4kHz, max_sum_sq_length)
	if shift > 0 {
		for i = 0; i < frame_length_4kHz; i++ {
			*signal_4kHz.ptr(int(i)) = int16(RSHIFT(int32(signal_4kHz.idx(int(i))), shift))
		}
	}

	target_ptr = signal_4kHz.off(int(RSHIFT(frame_length_4kHz, 1)))
	for k = 0; k < 2; k++ {

		basis_ptr = target_ptr.off(int(-min_lag_4kHz))

		normalizer = 0
		cross_corr = 0
		cross_corr = inner_prod_aligned(target_ptr, basis_ptr, sf_length_8kHz)
		normalizer = inner_prod_aligned(basis_ptr, basis_ptr, sf_length_8kHz)
		normalizer = ADD_SAT32(normalizer, SMULBB(sf_length_8kHz, 4000))

		temp32 = DIV32(cross_corr, SQRT_APPROX(normalizer)+1)
		*C[k].ptr(int(min_lag_4kHz)) = int16(SAT16(temp32))

		for d = min_lag_4kHz + 1; d <= max_lag_4kHz; d++ {
			basis_ptr = basis_ptr.off(-1)

			cross_corr = inner_prod_aligned(target_ptr, basis_ptr, sf_length_8kHz)

			normalizer +=
				SMULBB(int32(basis_ptr.idx(0)), int32(basis_ptr.idx(0))) -
					SMULBB(int32(basis_ptr.idx(int(sf_length_8kHz))), int32(basis_ptr.idx(int(sf_length_8kHz))))

			temp32 = DIV32(cross_corr, SQRT_APPROX(normalizer)+1)
			*C[k].ptr(int(d)) = int16(SAT16(temp32))
		}
		target_ptr = target_ptr.off(int(sf_length_8kHz))
	}

	for i = max_lag_4kHz; i >= min_lag_4kHz; i-- {
		sum = int32(C[0].idx(int(i))) + int32(C[1].idx(int(i)))
		sum = RSHIFT(sum, 1)
		sum = SMLAWB(sum, sum, LSHIFT(-i, 4))
		*C[0].ptr(int(i)) = int16(sum)
	}

	length_d_srch = 4 + 2*complexity
	insertion_sort_decreasing_int16(C[0].off(int(min_lag_4kHz)), d_srch, max_lag_4kHz-min_lag_4kHz+1, length_d_srch)

	target_ptr = signal_4kHz.off(int(RSHIFT(frame_length_4kHz, 1)))
	energy = inner_prod_aligned(target_ptr, target_ptr, RSHIFT(frame_length_4kHz, 1))
	energy = ADD_POS_SAT32(energy, 1000)
	Cmax = int32(C[0].idx(int(min_lag_4kHz)))
	threshold = SMULBB(Cmax, Cmax)
	if RSHIFT(energy, 4+2) > threshold {
		memset(pitch_out, 0, PITCH_EST_NB_SUBFR)
		*LTPCorr_Q15 = 0
		*lagIndex = 0
		*contourIndex = 0
		return 1
	}

	threshold = SMULWB(search_thres1_Q16, Cmax)
	for i = 0; i < length_d_srch; i++ {
		if int32(C[0].idx(int(min_lag_4kHz+i))) > threshold {
			*d_srch.ptr(int(i)) = (d_srch.idx(int(i)) + min_lag_4kHz) << 1
		} else {
			length_d_srch = i
			break
		}
	}

	for i = min_lag_8kHz - 5; i < max_lag_8kHz+5; i++ {
		*d_comp.ptr(int(i)) = 0
	}
	for i = 0; i < length_d_srch; i++ {
		*d_comp.ptr(int(d_srch.idx(int(i)))) = 1
	}

	for i = max_lag_8kHz + 3; i >= min_lag_8kHz; i-- {
		*d_comp.ptr(int(i)) += d_comp.idx(int(i-1)) + d_comp.idx(int(i-2))
	}

	length_d_srch = 0
	for i = min_lag_8kHz; i < max_lag_8kHz+1; i++ {
		if d_comp.idx(int(i+1)) > 0 {
			*d_srch.ptr(int(length_d_srch)) = i
			length_d_srch++
		}
	}

	for i = max_lag_8kHz + 3; i >= min_lag_8kHz; i-- {
		*d_comp.ptr(int(i)) += d_comp.idx(int(i-1)) + d_comp.idx(int(i-2)) + d_comp.idx(int(i-3))
	}

	length_d_comp = 0
	for i = min_lag_8kHz; i < max_lag_8kHz+4; i++ {
		if d_comp.idx(int(i)) > 0 {
			*d_comp.ptr(int(length_d_comp)) = int16(i - 2)
			length_d_comp++
		}
	}

	shift = FIX_P_Ana_find_scaling(signal_8kHz, frame_length_8kHz, sf_length_8kHz)
	if shift > 0 {
		for i = 0; i < frame_length_8kHz; i++ {
			*signal_8kHz.ptr(int(i)) = int16(RSHIFT(int32(signal_8kHz.idx(int(i))), shift))
		}
	}

	resetC()

	target_ptr = signal_8kHz.off(int(frame_length_4kHz))
	for k = 0; k < PITCH_EST_NB_SUBFR; k++ {

		energy_target = inner_prod_aligned(target_ptr, target_ptr, sf_length_8kHz)

		for j = 0; j < length_d_comp; j++ {
			d = int32(d_comp.idx(int(j)))
			basis_ptr = target_ptr.off(int(-d))

			cross_corr = inner_prod_aligned(target_ptr, basis_ptr, sf_length_8kHz)
			energy_basis = inner_prod_aligned(basis_ptr, basis_ptr, sf_length_8kHz)
			if cross_corr > 0 {
				energy = max(energy_target, energy_basis)
				lz = CLZ32(cross_corr)
				lshift = LIMIT_32(lz-1, 0, 15)
				temp32 = DIV32(LSHIFT(cross_corr, lshift), RSHIFT(energy, 15-lshift)+1)
				temp32 = SMULWB(cross_corr, temp32)
				temp32 = ADD_SAT32(temp32, temp32)
				lz = CLZ32(temp32)
				lshift = LIMIT_32(lz-1, 0, 15)
				energy = min(energy_target, energy_basis)
				*C[k].ptr(int(d)) = int16(DIV32(LSHIFT(temp32, lshift), RSHIFT(energy, 15-lshift)+1))
			} else {
				*C[k].ptr(int(d)) = 0
			}
		}
		target_ptr = target_ptr.off(int(sf_length_8kHz))
	}

	CCmax = math.MinInt32
	CCmax_b = math.MinInt32

	CBimax = 0
	lag = -1

	if prevLag > 0 {
		if Fs_kHz == 12 {
			prevLag = DIV32_16(LSHIFT(prevLag, 1), 3)
		} else if Fs_kHz == 16 {
			prevLag = RSHIFT(prevLag, 1)
		} else if Fs_kHz == 24 {
			prevLag = DIV32_16(prevLag, 3)
		}
		prevLag_log2_Q7 = lin2log(prevLag)
	} else {
		prevLag_log2_Q7 = 0
	}
	corr_thres_Q15 = RSHIFT(SMULBB(search_thres2_Q15, search_thres2_Q15), 13)

	if Fs_kHz == 8 && complexity > PITCH_EST_MIN_COMPLEX {
		nb_cbks_stage2 = PITCH_EST_NB_CBKS_STAGE2_EXT
	} else {
		nb_cbks_stage2 = PITCH_EST_NB_CBKS_STAGE2
	}

	for k = 0; k < length_d_srch; k++ {
		d = d_srch.idx(int(k))
		for j = 0; j < nb_cbks_stage2; j++ {
			CC[j] = 0
			for i = 0; i < PITCH_EST_NB_SUBFR; i++ {
				CC[j] = CC[j] + int32(C[i].idx(int(d+int32(CB_lags_stage2[i][j]))))
			}
		}
		CCmax_new = math.MinInt32
		CBimax_new = 0
		for i = 0; i < nb_cbks_stage2; i++ {
			if CC[i] > CCmax_new {
				CCmax_new = CC[i]
				CBimax_new = i
			}
		}

		lag_log2_Q7 = lin2log(d)

		if forLJC != 0 {
			CCmax_new_b = CCmax_new
		} else {
			CCmax_new_b = CCmax_new - RSHIFT(SMULBB(PITCH_EST_NB_SUBFR*PITCH_EST_SHORTLAG_BIAS_Q15, lag_log2_Q7), 7)
		}

		if prevLag > 0 {
			delta_lag_log2_sqr_Q7 = lag_log2_Q7 - prevLag_log2_Q7
			delta_lag_log2_sqr_Q7 = RSHIFT(SMULBB(delta_lag_log2_sqr_Q7, delta_lag_log2_sqr_Q7), 7)
			prev_lag_bias_Q15 = RSHIFT(SMULBB(PITCH_EST_NB_SUBFR*PITCH_EST_PREVLAG_BIAS_Q15, *LTPCorr_Q15), 15)
			prev_lag_bias_Q15 = DIV32(MUL(prev_lag_bias_Q15, delta_lag_log2_sqr_Q7), delta_lag_log2_sqr_Q7+(1<<6))
			CCmax_new_b -= prev_lag_bias_Q15
		}

		if CCmax_new_b > CCmax_b &&
			CCmax_new > corr_thres_Q15 &&
			int32(CB_lags_stage2[0][CBimax_new]) <= min_lag_8kHz {
			CCmax_b = CCmax_new_b
			CCmax = CCmax_new
			lag = d
			CBimax = CBimax_new
		}
	}

	if lag == -1 {
		memset(pitch_out, 0, PITCH_EST_NB_SUBFR)
		*LTPCorr_Q15 = 0
		*lagIndex = 0
		*contourIndex = 0
		return 1
	}

	if Fs_kHz > 8 {

		shift = FIX_P_Ana_find_scaling(signal, frame_length, sf_length)
		if shift > 0 {
			input_signal_ptr = scratch_mem
			for i = 0; i < frame_length; i++ {
				*input_signal_ptr.ptr(int(i)) = int16(RSHIFT(int32(signal.idx(int(i))), shift))
			}
		} else {
			input_signal_ptr = signal.off(0)
		}

		CBimax_old = CBimax
		if Fs_kHz == 12 {
			lag = RSHIFT(SMULBB(lag, 3), 1)
		} else if Fs_kHz == 16 {
			lag = LSHIFT(lag, 1)
		} else {
			lag = SMULBB(lag, 3)
		}

		lag = LIMIT(lag, min_lag, max_lag)
		start_lag = max(lag-2, min_lag)
		end_lag = min(lag+2, max_lag)
		lag_new = lag
		CBimax = 0
		*LTPCorr_Q15 = SQRT_APPROX(LSHIFT(CCmax, 13))

		CCmax = math.MinInt32
		for k = 0; k < PITCH_EST_NB_SUBFR; k++ {
			*pitch_out.ptr(int(k)) = lag + 2*int32(CB_lags_stage2[k][CBimax_old])
		}
		FIX_P_Ana_calc_corr_st3(&crosscorr_st3, input_signal_ptr, start_lag, sf_length, complexity)
		FIX_P_Ana_calc_energy_st3(&energies_st3, input_signal_ptr, start_lag, sf_length, complexity)

		lag_counter = 0
		contour_bias = DIV32_16(PITCH_EST_FLATCONTOUR_BIAS_Q20, int16(lag))

		cbk_size = int32(cbk_sizes_stage3[complexity])
		cbk_offset = int32(cbk_offsets_stage3[complexity])

		for d = start_lag; d <= end_lag; d++ {
			for j = cbk_offset; j < (cbk_offset + cbk_size); j++ {
				cross_corr = 0
				energy = 0
				for k = 0; k < PITCH_EST_NB_SUBFR; k++ {
					energy += RSHIFT(energies_st3[k][j][lag_counter], 2)
					cross_corr += RSHIFT(crosscorr_st3[k][j][lag_counter], 2)
				}

				if cross_corr > 0 {
					lz = CLZ32(cross_corr)
					lshift = LIMIT_32(lz-1, 0, 13)
					CCmax_new = DIV32(LSHIFT(cross_corr, lshift), RSHIFT(energy, 13-lshift)+1)
					CCmax_new = SAT16(CCmax_new)
					CCmax_new = SMULWB(cross_corr, CCmax_new)
					if CCmax_new > RSHIFT(math.MaxInt32, 3) {
						CCmax_new = math.MaxInt32
					} else {
						CCmax_new = LSHIFT(CCmax_new, 3)
					}
					diff = j - RSHIFT(PITCH_EST_NB_CBKS_STAGE3_MAX, 1)
					diff = MUL(diff, diff)
					diff = math.MaxInt16 - RSHIFT(MUL(contour_bias, diff), 5)
					CCmax_new = LSHIFT(SMULWB(CCmax_new, diff), 1)
				} else {
					CCmax_new = 0
				}

				if CCmax_new > CCmax && d+int32(CB_lags_stage3[0][j]) <= max_lag {
					CCmax = CCmax_new
					lag_new = d
					CBimax = j
				}
			}
			lag_counter++
		}

		for k = 0; k < PITCH_EST_NB_SUBFR; k++ {
			*pitch_out.ptr(int(k)) = lag_new + int32(CB_lags_stage3[k][CBimax])
		}
		*lagIndex = lag_new - min_lag
		*contourIndex = CBimax
	} else {
		CCmax = max(CCmax, 0)
		*LTPCorr_Q15 = SQRT_APPROX(LSHIFT(CCmax, 13))
		for k = 0; k < PITCH_EST_NB_SUBFR; k++ {
			*pitch_out.ptr(int(k)) = lag + int32(CB_lags_stage2[k][CBimax])
		}
		*lagIndex = lag - min_lag_8kHz
		*contourIndex = CBimax
	}
	return 0
}

func FIX_P_Ana_calc_corr_st3(
	cross_corr_st3 *[PITCH_EST_NB_SUBFR][PITCH_EST_NB_CBKS_STAGE3_MAX][PITCH_EST_NB_STAGE3_LAGS]int32,
	signal *slice[int16], start_lag, sf_length, complexity int32) {
	var (
		target_ptr, basis_ptr            *slice[int16]
		cross_corr                       int32
		i, j, k, lag_counter             int32
		cbk_offset, cbk_size, delta, idx int32
		scratch_mem                      = alloc[int32](SCRATCH_SIZE)
	)

	cbk_offset = int32(cbk_offsets_stage3[complexity])
	cbk_size = int32(cbk_sizes_stage3[complexity])

	target_ptr = signal.off(int(LSHIFT(sf_length, 2)))
	for k = 0; k < PITCH_EST_NB_SUBFR; k++ {
		lag_counter = 0

		for j = int32(Lag_range_stage3[complexity][k][0]); j <= int32(Lag_range_stage3[complexity][k][1]); j++ {
			basis_ptr = target_ptr.off(int(-(start_lag + j)))
			cross_corr = inner_prod_aligned(target_ptr.off(0), basis_ptr.off(0), sf_length)
			*scratch_mem.ptr(int(lag_counter)) = cross_corr
			lag_counter++
		}

		delta = int32(Lag_range_stage3[complexity][k][0])
		for i = cbk_offset; i < (cbk_offset + cbk_size); i++ {
			idx = int32(CB_lags_stage3[k][i]) - delta
			for j = 0; j < PITCH_EST_NB_STAGE3_LAGS; j++ {
				cross_corr_st3[k][i][j] = scratch_mem.idx(int(idx + j))
			}
		}
		target_ptr = target_ptr.off(int(sf_length))
	}
}

func FIX_P_Ana_calc_energy_st3(
	energies_st3 *[PITCH_EST_NB_SUBFR][PITCH_EST_NB_CBKS_STAGE3_MAX][PITCH_EST_NB_STAGE3_LAGS]int32,
	signal *slice[int16], start_lag, sf_length, complexity int32) {
	var (
		target_ptr, basis_ptr            *slice[int16]
		energy                           int32
		k, i, j, lag_counter             int32
		cbk_offset, cbk_size, delta, idx int32
		scratch_mem                      [SCRATCH_SIZE]int32
	)

	cbk_offset = int32(cbk_offsets_stage3[complexity])
	cbk_size = int32(cbk_sizes_stage3[complexity])

	target_ptr = signal.off(int(LSHIFT(sf_length, 2)))
	for k = 0; k < PITCH_EST_NB_SUBFR; k++ {
		lag_counter = 0

		basis_ptr = target_ptr.off(-int(start_lag + int32(Lag_range_stage3[complexity][k][0])))
		energy = inner_prod_aligned(basis_ptr, basis_ptr, sf_length)
		scratch_mem[lag_counter] = energy
		lag_counter++

		for i = 1; i < int32(Lag_range_stage3[complexity][k][1]-Lag_range_stage3[complexity][k][0]+1); i++ {
			energy -= SMULBB(int32(basis_ptr.idx(int(sf_length-i))), int32(basis_ptr.idx(int(sf_length-i))))

			energy = ADD_SAT32(energy, SMULBB(int32(basis_ptr.idx(-int(i))), int32(basis_ptr.idx(-int(i)))))
			scratch_mem[lag_counter] = energy
			lag_counter++
		}

		delta = int32(Lag_range_stage3[complexity][k][0])
		for i = cbk_offset; i < (cbk_offset + cbk_size); i++ {
			idx = int32(CB_lags_stage3[k][i]) - delta
			for j = 0; j < PITCH_EST_NB_STAGE3_LAGS; j++ {
				energies_st3[k][i][j] = scratch_mem[idx+j]
			}
		}
		target_ptr = target_ptr.off(int(sf_length))
	}
}

func FIX_P_Ana_find_scaling(signal *slice[int16], signal_length, sum_sqr_len int32) int32 {
	var nbits, x_max int32

	x_max = int32(int16_array_maxabs(signal, signal_length))

	if x_max < math.MaxInt16 {
		nbits = 32 - CLZ32(SMULBB(x_max, x_max))
	} else {
		nbits = 30
	}
	nbits += 17 - int32(CLZ16(uint16(sum_sqr_len)))

	if nbits < 31 {
		return 0
	} else {
		return nbits - 30
	}
}

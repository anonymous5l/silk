package silk

const (
	OFFSET        = (MIN_QGAIN_DB*128)/6 + 16*128
	SCALE_Q16     = (65536 * (N_LEVELS_QGAIN - 1)) / (((MAX_QGAIN_DB - MIN_QGAIN_DB) * 128) / 6)
	INV_SCALE_Q16 = (65536 * (((MAX_QGAIN_DB - MIN_QGAIN_DB) * 128) / 6)) / (N_LEVELS_QGAIN - 1)
)

func gains_quant(ind *slice[int32], gain_Q16 *slice[int32], prev_ind *int32, conditional int32) {
	var k int32

	for k = 0; k < NB_SUBFR; k++ {
		*ind.ptr(int(k)) = SMULWB(SCALE_Q16, lin2log(gain_Q16.idx(int(k)))-OFFSET)

		if ind.idx(int(k)) < *prev_ind {
			*ind.ptr(int(k)) = ind.idx(int(k)) + 1
		}

		if k == 0 && conditional == 0 {
			*ind.ptr(int(k)) = LIMIT(ind.idx(int(k)), 0, N_LEVELS_QGAIN-1)
			*ind.ptr(int(k)) = max(ind.idx(int(k)), *prev_ind+MIN_DELTA_GAIN_QUANT)
			*prev_ind = ind.idx(int(k))
		} else {
			*ind.ptr(int(k)) = LIMIT(ind.idx(int(k))-*prev_ind, MIN_DELTA_GAIN_QUANT, MAX_DELTA_GAIN_QUANT)
			*prev_ind += ind.idx(int(k))
			*ind.ptr(int(k)) -= MIN_DELTA_GAIN_QUANT
		}

		*gain_Q16.ptr(int(k)) = log2lin(min(SMULWB(INV_SCALE_Q16, *prev_ind)+OFFSET, 3967))
	}
}

func gains_dequant(gain_Q16, ind *slice[int32],
	prev_ind *int32,
	conditional int32) {
	var k int32

	for k = 0; k < NB_SUBFR; k++ {
		if k == 0 && conditional == 0 {
			*prev_ind = ind.idx(int(k))
		} else {
			*prev_ind += ind.idx(int(k)) + MIN_DELTA_GAIN_QUANT
		}

		*gain_Q16.ptr(int(k)) = log2lin(min(SMULWB(INV_SCALE_Q16, *prev_ind)+OFFSET, 3967))
	}
}

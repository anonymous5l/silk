package silk

func corrVector_FIX(x, t *slice[int16], L, order int32, Xt *slice[int32], rshifts int32) {
	var (
		lag, i     int32
		ptr1, ptr2 *slice[int16]
		inner_prod int32
	)
	ptr1 = x.off(int(order - 1))
	ptr2 = t.off(0)
	if rshifts > 0 {
		for lag = 0; lag < order; lag++ {
			inner_prod = 0
			for i = 0; i < L; i++ {
				inner_prod += RSHIFT32(SMULBB(int32(ptr1.idx(int(i))), int32(ptr2.idx(int(i)))), rshifts)
			}
			*Xt.ptr(int(lag)) = inner_prod
			ptr1 = ptr1.off(-1)
		}
	} else {
		for lag = 0; lag < order; lag++ {
			*Xt.ptr(int(lag)) = inner_prod_aligned(ptr1, ptr2, L)
			ptr1 = ptr1.off(-1)
		}
	}
}

func corrMatrix_FIX(x *slice[int16], L int32, order int32, head_room int32, XX *slice[int32], rshifts *int32) {
	var (
		i, j, lag, rshifts_local, head_room_rshifts int32
		energy                                      int32
		ptr1, ptr2                                  *slice[int16]
	)

	sum_sqr_shift(&energy, &rshifts_local, x, L+order-1)

	head_room_rshifts = max(head_room-CLZ32(energy), 0)

	energy = RSHIFT32(energy, head_room_rshifts)
	rshifts_local += head_room_rshifts

	for i = 0; i < order-1; i++ {
		energy -= RSHIFT32(SMULBB(int32(x.idx(int(i))), int32(x.idx(int(i)))), rshifts_local)
	}
	if rshifts_local < *rshifts {
		energy = RSHIFT32(energy, *rshifts-rshifts_local)
		rshifts_local = *rshifts
	}

	*matrix_ptr(XX, 0, 0, order) = energy
	ptr1 = x.off(int(order - 1))
	for j = 1; j < order; j++ {
		energy = SUB32(energy, RSHIFT32(SMULBB(int32(ptr1.idx(int(L-j))), int32(ptr1.idx(int(L-j)))), rshifts_local))
		energy = ADD32(energy, RSHIFT32(SMULBB(int32(ptr1.idx(int(-j))), int32(ptr1.idx(int(-j)))), rshifts_local))
		*matrix_ptr(XX, j, j, order) = energy
	}

	ptr2 = x.off(int(order - 2))
	if rshifts_local > 0 {
		for lag = 1; lag < order; lag++ {
			energy = 0
			for i = 0; i < L; i++ {
				energy += RSHIFT32(SMULBB(int32(ptr1.idx(int(i))), int32(ptr2.idx(int(i)))), rshifts_local)
			}
			*matrix_ptr(XX, lag, 0, order) = energy
			*matrix_ptr(XX, 0, lag, order) = energy
			for j = 1; j < (order - lag); j++ {
				energy = SUB32(energy, RSHIFT32(SMULBB(int32(ptr1.idx(int(L-j))), int32(ptr2.idx(int(L-j)))), rshifts_local))
				energy = ADD32(energy, RSHIFT32(SMULBB(int32(ptr1.idx(int(-j))), int32(ptr2.idx(int(-j)))), rshifts_local))
				*matrix_ptr(XX, lag+j, j, order) = energy
				*matrix_ptr(XX, j, lag+j, order) = energy
			}
			ptr2 = ptr2.off(-1)
		}
	} else {
		for lag = 1; lag < order; lag++ {
			energy = inner_prod_aligned(ptr1, ptr2, L)
			*matrix_ptr(XX, lag, 0, order) = energy
			*matrix_ptr(XX, 0, lag, order) = energy
			for j = 1; j < (order - lag); j++ {
				energy = SUB32(energy, SMULBB(int32(ptr1.idx(int(L-j))), int32(ptr2.idx(int(L-j)))))
				energy = SMLABB(energy, int32(ptr1.idx(int(-j))), int32(ptr2.idx(int(-j))))
				*matrix_ptr(XX, lag+j, j, order) = energy
				*matrix_ptr(XX, j, lag+j, order) = energy
			}
			ptr2 = ptr2.off(-1)
		}
	}
	*rshifts = rshifts_local
}

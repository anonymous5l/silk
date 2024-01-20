package silk

const (
	MAX_FRAME_SIZE = 544
	MAX_NB_SUBFR   = 4

	BURG_MODIFIED_QA = 25
	N_BITS_HEAD_ROOM = 2
	MIN_RSHIFTS      = -16
	MAX_RSHIFTS      = 32 - BURG_MODIFIED_QA
)

func burg_modified(res_nrg, res_nrg_Q *int32,
	A_Q16 *slice[int32], x *slice[int16],
	subfr_length, nb_subfr, WhiteNoiseFrac_Q32, D int32) {
	var (
		k, n, s, lz, rshifts, rshifts_extra                      int32
		C0, num, nrg, rc_Q31, Atmp_QA, Atmp1, tmp1, tmp2, x1, x2 int32
		x_ptr                                                    *slice[int16]
		C_first_row                                              = alloc[int32](MAX_ORDER_LPC)
		C_last_row                                               = alloc[int32](MAX_ORDER_LPC)
		Af_QA                                                    = alloc[int32](MAX_ORDER_LPC)
		CAf                                                      = alloc[int32](MAX_ORDER_LPC + 1)
		CAb                                                      = alloc[int32](MAX_ORDER_LPC + 1)
	)

	sum_sqr_shift(&C0, &rshifts, x, nb_subfr*subfr_length)
	if rshifts > MAX_RSHIFTS {
		C0 = LSHIFT32(C0, rshifts-MAX_RSHIFTS)
		rshifts = MAX_RSHIFTS
	} else {
		lz = CLZ32(C0) - 1
		rshifts_extra = N_BITS_HEAD_ROOM - lz
		if rshifts_extra > 0 {
			rshifts_extra = min(rshifts_extra, MAX_RSHIFTS-rshifts)
			C0 = RSHIFT32(C0, rshifts_extra)
		} else {
			rshifts_extra = max(rshifts_extra, MIN_RSHIFTS-rshifts)
			C0 = LSHIFT32(C0, -rshifts_extra)
		}
		rshifts += rshifts_extra
	}

	if rshifts > 0 {
		for s = 0; s < nb_subfr; s++ {
			x_ptr = x.off(int(s * subfr_length))
			for n = 1; n < D+1; n++ {
				*C_first_row.ptr(int(n - 1)) += int32(RSHIFT64(
					inner_prod16_aligned_64(x_ptr, x_ptr.off(int(n)), subfr_length-n), rshifts))
			}
		}
	} else {
		for s = 0; s < nb_subfr; s++ {
			x_ptr = x.off(int(s * subfr_length))
			for n = 1; n < D+1; n++ {
				*C_first_row.ptr(int(n - 1)) += LSHIFT32(
					inner_prod_aligned(x_ptr, x_ptr.off(int(n)), subfr_length-n), -rshifts)
			}
		}
	}
	C_first_row.copy(C_last_row, MAX_ORDER_LPC)

	*CAb.ptr(0) = C0 + SMMUL(WhiteNoiseFrac_Q32, C0) + 1
	*CAf.ptr(0) = CAb.idx(0)

	for n = 0; n < D; n++ {
		if rshifts > -2 {
			for s = 0; s < nb_subfr; s++ {
				x_ptr = x.off(int(s * subfr_length))
				x1 = -LSHIFT32(int32(x_ptr.idx(int(n))), 16-rshifts)
				x2 = -LSHIFT32(int32(x_ptr.idx(int(subfr_length-n-1))), 16-rshifts)
				tmp1 = LSHIFT32(int32(x_ptr.idx(int(n))), BURG_MODIFIED_QA-16)
				tmp2 = LSHIFT32(int32(x_ptr.idx(int(subfr_length-n-1))), BURG_MODIFIED_QA-16)
				for k = 0; k < n; k++ {
					*C_first_row.ptr(int(k)) = SMLAWB(C_first_row.idx(int(k)), x1, int32(x_ptr.idx(int(n-k-1))))
					*C_last_row.ptr(int(k)) = SMLAWB(C_last_row.idx(int(k)), x2, int32(x_ptr.idx(int(subfr_length-n+k))))
					Atmp_QA = Af_QA.idx(int(k))
					tmp1 = SMLAWB(tmp1, Atmp_QA, int32(x_ptr.idx(int(n-k-1))))
					tmp2 = SMLAWB(tmp2, Atmp_QA, int32(x_ptr.idx(int(subfr_length-n+k))))
				}
				tmp1 = LSHIFT32(-tmp1, 32-BURG_MODIFIED_QA-rshifts)
				tmp2 = LSHIFT32(-tmp2, 32-BURG_MODIFIED_QA-rshifts)
				for k = 0; k <= n; k++ {
					*CAf.ptr(int(k)) = SMLAWB(CAf.idx(int(k)), tmp1, int32(x_ptr.idx(int(n-k))))
					*CAb.ptr(int(k)) = SMLAWB(CAb.idx(int(k)), tmp2, int32(x_ptr.idx(int(subfr_length-n+k-1))))
				}
			}
		} else {
			for s = 0; s < nb_subfr; s++ {
				x_ptr = x.off(int(s * subfr_length))
				x1 = -LSHIFT32(int32(x_ptr.idx(int(n))), -rshifts)
				x2 = -LSHIFT32(int32(x_ptr.idx(int(subfr_length-n-1))), -rshifts)
				tmp1 = LSHIFT32(int32(x_ptr.idx(int(n))), 17)
				tmp2 = LSHIFT32(int32(x_ptr.idx(int(subfr_length-n-1))), 17)
				for k = 0; k < n; k++ {
					*C_first_row.ptr(int(k)) = MLA(C_first_row.idx(int(k)), x1, int32(x_ptr.idx(int(n-k-1))))
					*C_last_row.ptr(int(k)) = MLA(C_last_row.idx(int(k)), x2, int32(x_ptr.idx(int(subfr_length-n+k))))
					Atmp1 = RSHIFT_ROUND(Af_QA.idx(int(k)), BURG_MODIFIED_QA-17)
					tmp1 = MLA(tmp1, int32(x_ptr.idx(int(n-k-1))), Atmp1)
					tmp2 = MLA(tmp2, int32(x_ptr.idx(int(subfr_length-n+k))), Atmp1)
				}
				tmp1 = -tmp1
				tmp2 = -tmp2
				for k = 0; k <= n; k++ {
					*CAf.ptr(int(k)) = SMLAWW(CAf.idx(int(k)), tmp1,
						LSHIFT32(int32(x_ptr.idx(int(n-k))), -rshifts-1))
					*CAb.ptr(int(k)) = SMLAWW(CAb.idx(int(k)), tmp2,
						LSHIFT32(int32(x_ptr.idx(int(subfr_length-n+k-1))), -rshifts-1))
				}
			}
		}

		tmp1 = C_first_row.idx(int(n))
		tmp2 = C_last_row.idx(int(n))
		num = 0
		nrg = ADD32(CAb.idx(0), CAf.idx(0))
		for k = 0; k < n; k++ {
			Atmp_QA = Af_QA.idx(int(k))
			lz = CLZ32(abs(Atmp_QA)) - 1
			lz = min(32-BURG_MODIFIED_QA, lz)
			Atmp1 = LSHIFT32(Atmp_QA, lz)

			tmp1 = ADD_LSHIFT32(tmp1, SMMUL(C_last_row.idx(int(n-k-1)), Atmp1), 32-BURG_MODIFIED_QA-lz)
			tmp2 = ADD_LSHIFT32(tmp2, SMMUL(C_first_row.idx(int(n-k-1)), Atmp1), 32-BURG_MODIFIED_QA-lz)
			num = ADD_LSHIFT32(num, SMMUL(CAb.idx(int(n-k)), Atmp1), 32-BURG_MODIFIED_QA-lz)
			nrg = ADD_LSHIFT32(nrg, SMMUL(ADD32(CAb.idx(int(k+1)), CAf.idx(int(k+1))),
				Atmp1), 32-BURG_MODIFIED_QA-lz)
		}
		*CAf.ptr(int(n + 1)) = tmp1
		*CAb.ptr(int(n + 1)) = tmp2
		num = ADD32(num, tmp2)
		num = LSHIFT32(-num, 1)

		if abs(num) < nrg {
			rc_Q31 = DIV32_varQ(num, nrg, 31)
		} else {
			memset(Af_QA.off(int(n)), 0, int(D-n))
			break
		}

		for k = 0; k < (n+1)>>1; k++ {
			tmp1 = Af_QA.idx(int(k))
			tmp2 = Af_QA.idx(int(n - k - 1))
			*Af_QA.ptr(int(k)) = ADD_LSHIFT32(tmp1, SMMUL(tmp2, rc_Q31), 1)
			*Af_QA.ptr(int(n - k - 1)) = ADD_LSHIFT32(tmp2, SMMUL(tmp1, rc_Q31), 1)
		}
		*Af_QA.ptr(int(n)) = RSHIFT32(rc_Q31, 31-BURG_MODIFIED_QA)

		for k = 0; k <= n+1; k++ {
			tmp1 = CAf.idx(int(k))
			tmp2 = CAb.idx(int(n - k + 1))
			*CAf.ptr(int(k)) = ADD_LSHIFT32(tmp1, SMMUL(tmp2, rc_Q31), 1)
			*CAb.ptr(int(n - k + 1)) = ADD_LSHIFT32(tmp2, SMMUL(tmp1, rc_Q31), 1)
		}
	}

	nrg = CAf.idx(0)
	tmp1 = 1 << 16
	for k = 0; k < D; k++ {
		Atmp1 = RSHIFT_ROUND(Af_QA.idx(int(k)), BURG_MODIFIED_QA-16)
		nrg = SMLAWW(nrg, CAf.idx(int(k+1)), Atmp1)
		tmp1 = SMLAWW(tmp1, Atmp1, Atmp1)
		*A_Q16.ptr(int(k)) = -Atmp1
	}
	*res_nrg = SMLAWW(nrg, SMMUL(WhiteNoiseFrac_Q32, C0), -tmp1)
	*res_nrg_Q = -rshifts
}

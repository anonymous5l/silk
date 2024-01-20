package silk

import "math"

func NLSF2A_find_poly(out *slice[int32], cLSF *slice[int32], dd int32) {
	var (
		k, n, ftmp int32
	)

	*out.ptr(0) = LSHIFT(1, 20)
	*out.ptr(1) = -cLSF.idx(0)
	for k = 1; k < dd; k++ {
		ftmp = cLSF.idx(2 * int(k))
		*out.ptr(int(k) + 1) = LSHIFT(out.idx(int(k)-1), 1) -
			int32(RSHIFT_ROUND64(SMULL(ftmp, out.idx(int(k))), 20))
		for n = k; n > 1; n-- {
			*out.ptr(int(n)) += out.idx(int(n)-2) -
				int32(RSHIFT_ROUND64(SMULL(ftmp, out.idx(int(n)-1)), 20))
		}
		*out.ptr(1) -= ftmp
	}
}

func NLSF2A(a *slice[int16], NLSF *slice[int32], d int32) {
	var (
		k, i, dd                    int32
		cos_LSF_Q20                 = alloc[int32](MAX_ORDER_LPC)
		P                           = alloc[int32](MAX_ORDER_LPC/2 + 1)
		Q                           = alloc[int32](MAX_ORDER_LPC/2 + 1)
		Ptmp, Qtmp                  int32
		f_int                       int32
		f_frac                      int32
		cos_val, delta              int32
		a_int32                     = alloc[int32](MAX_ORDER_LPC)
		maxabs, absval, idx, sc_Q16 int32
	)

	for k = 0; k < d; k++ {

		f_int = RSHIFT(NLSF.idx(int(k)), 15-7)

		f_frac = NLSF.idx(int(k)) - LSHIFT(f_int, 15-7)

		cos_val = LSFCosTab_FIX_Q12[f_int]
		delta = LSFCosTab_FIX_Q12[f_int+1] - cos_val

		*cos_LSF_Q20.ptr(int(k)) = LSHIFT(cos_val, 8) + MUL(delta, f_frac)
	}

	dd = RSHIFT(d, 1)

	NLSF2A_find_poly(P, cos_LSF_Q20, dd)
	NLSF2A_find_poly(Q, cos_LSF_Q20.off(1), dd)

	for k = 0; k < dd; k++ {
		Ptmp = P.idx(int(k+1)) + P.idx(int(k))
		Qtmp = Q.idx(int(k+1)) - Q.idx(int(k))

		*a_int32.ptr(int(k)) = -RSHIFT_ROUND(Ptmp+Qtmp, 9)
		*a_int32.ptr(int(d - k - 1)) = RSHIFT_ROUND(Qtmp-Ptmp, 9)
	}

	for i = 0; i < 10; i++ {

		maxabs = 0
		for k = 0; k < d; k++ {
			absval = abs(a_int32.idx(int(k)))
			if absval > maxabs {
				maxabs = absval
				idx = k
			}
		}

		if maxabs > math.MaxInt16 {

			maxabs = min(maxabs, 98369)
			sc_Q16 = 65470 - DIV32(MUL(65470>>2, maxabs-math.MaxInt16),
				RSHIFT32(MUL(maxabs, idx+1), 2))
			bwexpander_32(a_int32, d, sc_Q16)
		} else {
			break
		}
	}

	if i == 10 {
		for k = 0; k < d; k++ {
			*a_int32.ptr(int(k)) = SAT16(a_int32.idx(int(k)))
		}
	}

	for k = 0; k < d; k++ {
		*a.ptr(int(k)) = int16(a_int32.idx(int(k)))
	}
}

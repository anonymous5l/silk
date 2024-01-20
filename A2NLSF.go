package silk

import "math"

const (
	BIN_DIV_STEPS_A2NLSF_FIX  = 3
	QPoly                     = 16
	MAX_ITERATIONS_A2NLSF_FIX = 30
	OVERSAMPLE_COSINE_TABLE   = 0
)

func A2NLSF_trans_poly(p *slice[int32], dd int32) {
	var k, n int32
	for k = 2; k <= dd; k++ {
		for n = dd; n > k; n-- {
			*p.ptr(int(n - 2)) -= p.idx(int(n))
		}
		*p.ptr(int(k - 2)) -= LSHIFT(p.idx(int(k)), 1)
	}
}

func A2NLSF_eval_poly(p *slice[int32], x, dd int32) int32 {
	var (
		n, x_Q16, y32 int32
	)
	y32 = p.idx(int(dd))
	x_Q16 = LSHIFT(x, 4)
	for n = dd - 1; n >= 0; n-- {
		y32 = SMLAWW(p.idx(int(n)), y32, x_Q16)
	}
	return y32
}

func A2NLSF_init(a_Q16 *slice[int32], P, Q *slice[int32], dd int32) {
	var k int32

	*P.ptr(int(dd)) = LSHIFT(1, QPoly)
	*Q.ptr(int(dd)) = LSHIFT(1, QPoly)
	for k = 0; k < dd; k++ {
		*P.ptr(int(k)) = -a_Q16.idx(int(dd-k-1)) - a_Q16.idx(int(dd+k))
		*Q.ptr(int(k)) = -a_Q16.idx(int(dd-k-1)) + a_Q16.idx(int(dd+k))
	}

	for k = dd; k > 0; k-- {
		*P.ptr(int(k - 1)) -= P.idx(int(k))
		*Q.ptr(int(k - 1)) += Q.idx(int(k))
	}

	A2NLSF_trans_poly(P, dd)
	A2NLSF_trans_poly(Q, dd)
}

func A2NLSF(NLSF, a_Q16 *slice[int32], d int32) {
	var (
		i, k, m, dd, root_ix, ffrac int32
		xlo, xhi, xmid              int32
		ylo, yhi, ymid              int32
		nom, den                    int32
		P                           = alloc[int32](MAX_ORDER_LPC/2 + 1)
		Q                           = alloc[int32](MAX_ORDER_LPC/2 + 1)
		PQ                          [2]*slice[int32]
		p                           *slice[int32]
	)

	PQ[0] = P
	PQ[1] = Q

	dd = RSHIFT(d, 1)

	A2NLSF_init(a_Q16, P, Q, dd)

	p = P

	xlo = LSFCosTab_FIX_Q12[0]
	ylo = A2NLSF_eval_poly(p, xlo, dd)

	if ylo < 0 {
		*NLSF.ptr(0) = 0
		p = Q
		ylo = A2NLSF_eval_poly(p, xlo, dd)
		root_ix = 1
	} else {
		root_ix = 0
	}
	k = 1
	i = 0
	for {
		xhi = LSFCosTab_FIX_Q12[k]
		yhi = A2NLSF_eval_poly(p, xhi, dd)

		if (ylo <= 0 && yhi >= 0) || (ylo >= 0 && yhi <= 0) {
			ffrac = -256

			for m = 0; m < BIN_DIV_STEPS_A2NLSF_FIX; m++ {

				xmid = RSHIFT_ROUND(xlo+xhi, 1)
				ymid = A2NLSF_eval_poly(p, xmid, dd)

				if (ylo <= 0 && ymid >= 0) || (ylo >= 0 && ymid <= 0) {
					xhi = xmid
					yhi = ymid
				} else {
					xlo = xmid
					ylo = ymid
					ffrac = ADD_RSHIFT(ffrac, 128, m)
				}
			}

			if abs(ylo) < 65536 {
				den = ylo - yhi
				nom = LSHIFT(ylo, 8-BIN_DIV_STEPS_A2NLSF_FIX) + RSHIFT(den, 1)
				if den != 0 {
					ffrac += DIV32(nom, den)
				}
			} else {
				ffrac += DIV32(ylo, RSHIFT(ylo-yhi, 8-BIN_DIV_STEPS_A2NLSF_FIX))
			}

			*NLSF.ptr(int(root_ix)) = min(LSHIFT(k, 8)+ffrac, math.MaxInt16)

			root_ix++
			if root_ix >= d {
				break
			}
			p = PQ[root_ix&1]

			xlo = LSFCosTab_FIX_Q12[k-1]
			ylo = LSHIFT(1-(root_ix&2), 12)
		} else {
			k++
			xlo = xhi
			ylo = yhi
			if k > LSF_COS_TAB_SZ_FIX {
				i++
				if i > MAX_ITERATIONS_A2NLSF_FIX {
					*NLSF.ptr(0) = DIV32_16(1<<15, int16(d+1))
					for k = 1; k < d; k++ {
						*NLSF.ptr(int(k)) = SMULBB(k+1, NLSF.idx(0))
					}
					return
				}

				bwexpander_32(a_Q16, d, 65536-SMULBB(10+i, i))

				A2NLSF_init(a_Q16, P, Q, dd)
				p = P
				xlo = LSFCosTab_FIX_Q12[0]
				ylo = A2NLSF_eval_poly(p, xlo, dd)
				if ylo < 0 {
					*NLSF.ptr(0) = 0
					p = Q
					ylo = A2NLSF_eval_poly(p, xlo, dd)
					root_ix = 1
				} else {
					root_ix = 0
				}
				k = 1
			}
		}
	}
}

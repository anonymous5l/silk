package silk

import "math"

func residual_energy16_covar_FIX(c *slice[int16], wXX, wXx *slice[int32], wxx, D, cQ int32) int32 {
	var (
		i, j, lshifts, Qxtra         int32
		c_max, w_max, tmp, tmp2, nrg int32
		cn                           = alloc[int32](MAX_MATRIX_SIZE)
		pRow                         *slice[int32]
	)

	lshifts = 16 - cQ
	Qxtra = lshifts

	c_max = 0
	for i = 0; i < D; i++ {
		c_max = max(c_max, abs(int32(c.idx(int(i)))))
	}
	Qxtra = min(Qxtra, CLZ32(c_max)-17)

	w_max = max(wXX.idx(0), wXX.idx(int(D*D-1)))
	Qxtra = min(Qxtra, CLZ32(MUL(D, RSHIFT(SMULWB(w_max, c_max), 4)))-5)
	Qxtra = max(Qxtra, 0)
	for i = 0; i < D; i++ {
		*cn.ptr(int(i)) = LSHIFT(int32(c.idx(int(i))), Qxtra)
	}
	lshifts -= Qxtra

	tmp = 0
	for i = 0; i < D; i++ {
		tmp = SMLAWB(tmp, wXx.idx(int(i)), cn.idx(int(i)))
	}
	nrg = RSHIFT(wxx, 1+lshifts) - tmp

	tmp2 = 0
	for i = 0; i < D; i++ {
		tmp = 0
		pRow = wXX.off(int(i * D))
		for j = i + 1; j < D; j++ {
			tmp = SMLAWB(tmp, pRow.idx(int(j)), cn.idx(int(j)))
		}
		tmp = SMLAWB(tmp, RSHIFT(pRow.idx(int(i)), 1), cn.idx(int(i)))
		tmp2 = SMLAWB(tmp2, tmp, cn.idx(int(i)))
	}
	nrg = ADD_LSHIFT32(nrg, tmp2, lshifts)

	if nrg < 1 {
		nrg = 1
	} else if nrg > RSHIFT(math.MaxInt32, lshifts+2) {
		nrg = math.MaxInt32 >> 1
	} else {
		nrg = LSHIFT(nrg, lshifts+1)
	}
	return nrg
}

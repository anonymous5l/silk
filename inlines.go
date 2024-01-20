package silk

import "math"

func CLZ64(in int64) int32 {
	var in_upper int32

	in_upper = int32(RSHIFT64(in, 32))
	if in_upper == 0 {
		return 32 + CLZ32(int32(in))
	} else {
		return CLZ32(in_upper)
	}
}

func CLZ_FRAC(in int32, lz *int32, frac_Q7 *int32) {
	lzeros := CLZ32(in)

	*lz = lzeros
	*frac_Q7 = ROR32(in, 24-lzeros) & 0x7f
}

func SQRT_APPROX(x int32) int32 {
	var y, lz, frac_Q7 int32

	if x <= 0 {
		return 0
	}

	CLZ_FRAC(x, &lz, &frac_Q7)

	if lz&1 != 0 {
		y = 32768
	} else {
		y = 46214
	}

	y >>= RSHIFT(lz, 1)

	y = SMLAWB(y, y, SMULBB(213, frac_Q7))

	return y
}

func DIV32_varQ(a32, b32, Qres int32) int32 {
	var (
		a_headrm, b_headrm, lshift        int32
		b32_inv, a32_nrm, b32_nrm, result int32
	)

	a_headrm = CLZ32(abs(a32)) - 1
	a32_nrm = LSHIFT(a32, a_headrm)
	b_headrm = CLZ32(abs(b32)) - 1
	b32_nrm = LSHIFT(b32, b_headrm)

	b32_inv = DIV32_16(math.MaxInt32>>2, int16(RSHIFT(b32_nrm, 16)))

	result = SMULWB(a32_nrm, b32_inv)

	a32_nrm -= int32(LSHIFT_ovflw(uint32(SMMUL(b32_nrm, result)), 3))

	result = SMLAWB(result, a32_nrm, b32_inv)

	lshift = 29 + a_headrm - b_headrm - Qres
	if lshift <= 0 {
		return LSHIFT_SAT32(result, -lshift)
	} else {
		if lshift < 32 {
			return RSHIFT(result, lshift)
		} else {
			return 0
		}
	}
}

func INVERSE32_varQ(b32, Qres int32) int32 {
	var (
		b_headrm, lshift                  int32
		b32_inv, b32_nrm, err_Q32, result int32
	)

	b_headrm = CLZ32(abs(b32)) - 1
	b32_nrm = LSHIFT(b32, b_headrm)

	b32_inv = DIV32_16(math.MaxInt32>>2, int16(RSHIFT(b32_nrm, 16)))

	result = LSHIFT(b32_inv, 16)

	err_Q32 = int32(LSHIFT_ovflw(uint32(-SMULWB(b32_nrm, b32_inv)), 3))

	result = SMLAWW(result, err_Q32, b32_inv)

	lshift = 61 - b_headrm - Qres
	if lshift <= 0 {
		return LSHIFT_SAT32(result, -lshift)
	} else {
		if lshift < 32 {
			return RSHIFT(result, lshift)
		} else {
			return 0
		}
	}
}

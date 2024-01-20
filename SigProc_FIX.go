package silk

import "math"

const (
	MAX_ORDER_LPC         = 16
	LSF_COS_TAB_SZ_FIX    = 128
	PITCH_EST_MIN_COMPLEX = 0
	PITCH_EST_MID_COMPLEX = 1
	PITCH_EST_MAX_COMPLEX = 2
)

func ROR32(a32, rot int32) int32 {
	var x, r, m uint32
	x = uint32(a32)
	r = uint32(rot)
	m = uint32(-rot)
	if rot <= 0 {
		return int32((x << m) | (x >> (32 - m)))
	} else {
		return int32((x << (32 - r)) | (x >> r))
	}
}

func MUL(a32, b32 int32) int32 {
	return a32 * b32
}

func MUL_uint(a32, b32 uint32) uint32 {
	return a32 * b32
}

func MLA(a32, b32, c32 int32) int32 {
	return ADD32(a32, b32*c32)
}

func SMLABB_ovflw(a32, b32, c32 int32) int32 {
	return ADD32_ovflw(a32, uint32(SMULBB((b32), (c32))))
}

func SMULTT(a32, b32 int32) int32 {
	return (a32 >> 16) * (b32 >> 16)
}

func SMLALBB(a64 int64, b16, c16 int16) int64 {
	return ADD64(a64, int64(int32(b16)*int32(c16)))
}

func SMULL(a32, b32 int32) int64 {
	return int64(a32) * int64(b32)
}

func ADD32_ovflw(a int32, b uint32) int32 {
	return int32(uint32(a) + b)
}

func MLA_ovflw(a32, b32, c32 int32) int32 {
	return ADD32_ovflw(a32, uint32(b32)*uint32(c32))
}

func SMLATT_ovflw(a32, b32, c32 int32) int32 {
	return ADD32_ovflw(a32, uint32(SMULTT(b32, c32)))
}

func SMLAWB_ovflw(a32, b32, c32 int32) int32 {
	return ADD32_ovflw(a32, uint32(SMULWB(b32, c32)))
}

func SMLAWT_ovflw(a32, b32, c32 int32) int32 {
	return ADD32_ovflw(a32, uint32(SMULWT(b32, c32)))
}

func DIV32_16(a32 int32, b16 int16) int32 {
	return DIV32(a32, int32(b16))
}

func DIV32(a32, b32 int32) int32 {
	return a32 / b32
}

func ADD32(a, b int32) int32 {
	return a + b
}

func ADD64(a, b int64) int64 {
	return a + b
}

func SUB32(a, b int32) int32 {
	return a - b
}

func SAT16(a int32) int32 {
	if a > math.MaxInt16 {
		return math.MaxInt16
	} else if a < math.MinInt16 {
		return math.MinInt16
	}
	return a
}

func ADD_SAT16(a, b int16) int16 {
	return int16(SAT16(ADD32(int32(a), int32(b))))
}

func ADD_POS_SAT32(a, b int32) int32 {
	ures := a + b
	if uint(ures)&0x80000000 != 0 {
		return math.MaxInt32
	}
	return ures
}

func LSHIFT32(a, shift int32) int32 {
	return a << shift
}

func LSHIFT64(a int64, shift int32) int64 {
	return a << shift
}

func LSHIFT(a, shift int32) int32 {
	return LSHIFT32(a, shift)
}

func RSHIFT32(a, shift int32) int32 {
	return a >> shift
}

func RSHIFT64(a int64, shift int32) int64 {
	return a >> shift
}

func RSHIFT(a, shift int32) int32 {
	return RSHIFT32(a, shift)
}

func LSHIFT_SAT32(a, shift int32) int32 {
	return LSHIFT32(LIMIT_32(a, RSHIFT32(math.MinInt32, shift),
		RSHIFT32(math.MaxInt32, shift)), shift)
}

func LSHIFT_ovflw(a uint32, shift int32) uint32 {
	return a << shift
}

func LSHIFT_uint(a uint32, shift int32) uint32 {
	return a << shift
}

func RSHIFT_uint(a uint32, shift int32) uint32 {
	return a >> shift
}

func ADD_LSHIFT(a, b, shift int32) int32 {
	return a + LSHIFT(b, shift)
}

func ADD_LSHIFT32(a, b, shift int32) int32 {
	return ADD32(a, LSHIFT32(b, shift))
}

func ADD_RSHIFT(a, b, shift int32) int32 {
	return a + RSHIFT(b, shift)
}

func ADD_RSHIFT32(a, b, shift int32) int32 {
	return ADD32(a, RSHIFT32(b, shift))
}

func SUB_LSHIFT32(a, b, shift int32) int32 {
	return SUB32(a, LSHIFT32(b, shift))
}
func SUB_RSHIFT32(a, b, shift int32) int32 {
	return SUB32(a, RSHIFT32(b, shift))
}

func ADD_RSHIFT_uint(a, b uint32, shift int32) uint32 {
	return (a) + RSHIFT_uint(b, shift)
}

func SUB_LSHIFT(a, b, shift int32) int32 {
	return a - LSHIFT(b, shift)
}

func SUB_RSHIFT(a, b, shift int32) int32 {
	return a - RSHIFT(b, shift)
}

func RSHIFT_ROUND(a, shift int32) int32 {
	if shift == 1 {
		return (a >> 1) + (a & 1)
	}
	return ((a >> (shift - 1)) + 1) >> 1
}

func RSHIFT_ROUND64(a int64, shift int32) int64 {
	if shift == 1 {
		return (a >> 1) + (a & 1)
	}
	return ((a >> (shift - 1)) + 1) >> 1
}

func FIX_CONST(C float64, Q int64) int32 {
	return int32((C * float64(int64(1<<Q))) + 0.5)
}

func LIMIT(a, limit1, limit2 int32) int32 {
	if limit1 > limit2 {
		if a > limit1 {
			return limit1
		} else if a < limit2 {
			return limit2
		}
	} else if a > limit2 {
		return limit2
	} else if a < limit1 {
		return limit1
	}
	return a
}

func LIMIT_32(a, limit1, limit2 int32) int32 {
	return LIMIT(a, limit1, limit2)
}

func abs(a int32) int32 {
	if a > 0 {
		return a
	}
	return -a
}

func abs_int32(a int32) int32 {
	return (a ^ (a >> 31)) - (a >> 31)
}

func RAND(seed int32) int32 {
	return MLA_ovflw(907633515, seed, 196314165)
}

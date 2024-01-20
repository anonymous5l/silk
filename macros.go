package silk

import "math"

func SMULWB(a32, b32 int32) int32 {
	b32 = int32(int16(b32))
	return ((a32 >> 16) * b32) + (((a32 & 0xffff) * b32) >> 16)
}

func SMLAWB(a32, b32, c32 int32) int32 {
	c32 = int32(int16(c32))
	return a32 + ((b32 >> 16) * c32) + (((b32 & 0xffff) * c32) >> 16)
}

func SMULWT(a32, b32 int32) int32 {
	return (a32>>16)*(b32>>16) + (((a32 & 0x0000FFFF) * (b32 >> 16)) >> 16)
}

func SMLAWT(a32, b32, c32 int32) int32 {
	return a32 + ((b32 >> 16) * (c32 >> 16)) + (((b32 & 0xffff) * (c32 >> 16)) >> 16)
}

func SMULBB(a32, b32 int32) int32 {
	return int32(int16(a32)) * int32(int16(b32))
}

func SMLABB(a32, b32, c32 int32) int32 {
	b32, c32 = int32(int16(b32)), int32(int16(c32))
	return a32 + (b32 * c32)
}

func SMLABT(a32, b32, c32 int32) int32 {
	return a32 + int32(int16(b32))*(c32>>16)
}

func SMULWW(a32, b32 int32) int32 {
	return MLA(SMULWB(a32, b32), a32, RSHIFT_ROUND(b32, 16))
}

func SMLAWW(a32, b32, c32 int32) int32 {
	return MLA(SMLAWB(a32, b32, c32), b32, RSHIFT_ROUND(c32, 16))
}

func SMMUL(a32, b32 int32) int32 {
	return int32(RSHIFT64(SMULL(a32, b32), 32))
}

func ADD_SAT32(a, b int32) int32 {
	if (int(a+b) & 0x80000000) == 0 {
		if (int(a&b) & 0x80000000) != 0 {
			return math.MinInt32
		}
	} else if (int(a|b) & 0x80000000) == 0 {
		return math.MaxInt32
	}
	return a + b
}

func SUB_SAT32(a, b int32) int32 {
	if int(a-b)&0x80000000 == 0 {
		if int(a)&(int(b)^0x80000000)&0x80000000 != 0 {
			return math.MinInt32
		}
	} else if ((int(a) ^ 0x80000000) & int(b) & 0x80000000) != 0 {
		return math.MaxInt32
	}
	return a - b
}

func CLZ16(in16 uint16) uint16 {
	var out32 uint16

	if in16 == 0 {
		return 16
	}

	if uint(in16)&0xFF00 != 0 {
		if uint(in16)&0xF000 != 0 {
			in16 >>= 12
		} else {
			out32 += 4
			in16 >>= 8
		}
	} else {
		if uint(in16)&0xFFF0 != 0 {
			out32 += 8
			in16 >>= 4
		} else {
			out32 += 12
		}
	}

	if uint(in16)&0xC != 0 {
		if uint(in16)&0x8 != 0 {
			return out32
		} else {
			return out32 + 1
		}
	} else {
		if uint(in16)&0xE != 0 {
			return out32 + 2
		} else {
			return out32 + 3
		}
	}
}

func CLZ32(in32 int32) int32 {
	if uint(in32)&0xFFFF0000 != 0 {
		return int32(CLZ16(uint16(in32 >> 16)))
	} else {
		return int32(CLZ16(uint16(in32)) + 16)
	}
}

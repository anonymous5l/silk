package silk

import (
	"math"
)

func smlawb(a32, b32, c32 int32) int32 {
	c32 = int32(int16(c32))
	return a32 + ((b32 >> 16) * c32) + (((b32 & 0xffff) * c32) >> 16)
}

func smulwb(a32, b32 int32) int32 {
	b32 = int32(int16(b32))
	return ((a32 >> 16) * b32) + (((a32 & 0xffff) * b32) >> 16)
}

func smlawt(a32, b32, c32 int32) int32 {
	return a32 + ((b32 >> 16) * (c32 >> 16)) + (((b32 & 0xffff) * (c32 >> 16)) >> 16)
}

func smlabb(a32, b32, c32 int32) int32 {
	b32, c32 = int32(int16(b32)), int32(int16(c32))
	return a32 + (b32 * c32)
}

func mla(a32, b32, c32 int32) int32 {
	return a32 + (b32 * c32)
}

func smulww(a32, b32 int32) int32 {
	return mla(smulwb(a32, b32), a32, rrshift(b32, 16))
}

func smmul(a32, b32 int32) int32 {
	return int32((int64(a32) * int64(b32)) >> 32)
}

func smlaww(a32, b32, c32 int32) int32 {
	return mla(smlawb(a32, b32, c32), b32, rrshift(c32, 16))
}

func clz16(a uint16) uint16 {
	if a == 0 {
		return 0x10
	}

	v2, v3 := uint16(0), uint16(0)
	if a&0xff00 == 0 {
		if a&0xfff0 == 0 {
			v2 = 12
			v3 = a
		} else {
			v2 = 8
			v3 = a >> 4
		}
	} else if a&0xf000 == 0 {
		v2 = 4
		v3 = a >> 8
	} else {
		v3 = a >> 12
	}

	if v3&12 == 0 {
		if v3&14 == 0 {
			return v2 + 3
		} else {
			return v2 + 2
		}
	} else if (v3 >> 3 & 1) == 0 {
		return v2 + 1
	}
	return v2
}

func addOverflow(a32 int32, b32 uint32) int32 {
	return int32(uint32(a32) + b32)
}

func _MLAOverflow(a32, b32, c32 int32) int32 {
	return addOverflow(a32, uint32(b32)*uint32(c32))
}

func rand(seed int32) int32 {
	return _MLAOverflow(907633515, seed, 196314165)
}

func clz32(a int32) int32 {
	a1 := int(a)
	if a1&0xffff0000 == 0 {
		return int32(clz16(uint16(a))) + 0x10
	}
	return int32(clz16(uint16(a >> 0x10)))
}

func lshiftSAT32(a, shift int32) int32 {
	return lshift(limit(a, rshift(math.MinInt32, shift), rshift(math.MaxInt32, shift)), shift)
}

func inverse32varQ(b32 int32, QRes int32) int32 {
	assert(b32 != 0)
	assert(b32 != math.MinInt32)
	assert(QRes > 0)

	bHeadRM := clz32(i32abs(b32)) - 1
	b32nRM := lshift(b32, bHeadRM)

	b32Inv := (math.MaxInt32 >> 2) / rshift(b32nRM, 16)
	result := lshift(b32Inv, 16)
	errQ32 := lshift(-smulwb(b32nRM, b32Inv), 3)
	result = smlaww(result, errQ32, b32Inv)

	ls := 61 - bHeadRM - QRes
	if ls <= 0 {
		return lshiftSAT32(result, -ls)
	} else {
		if ls < 32 {
			return rshift(result, ls)
		}
		return 0
	}
}

func sat16(a int32) int16 {
	if a > math.MaxInt16 {
		return math.MaxInt16
	} else if a < math.MinInt16 {
		return math.MinInt16
	}
	return int16(a)
}

func rrshift(a, shift int32) int32 {
	if shift == 1 {
		return (a >> 1) + (a & 1)
	}
	return ((a >> (shift - 1)) + 1) >> 1
}

func i64rrshift(a int64, shift int32) int64 {
	if shift == 1 {
		return a>>1 + (a & 1)
	}
	return ((a >> (shift - 1)) + 1) >> 1
}

func limit(a, limit1, limit2 int32) int32 {
	if limit1 > limit2 {
		if a > limit1 {
			return limit1
		} else if a < limit2 {
			return limit2
		} else {
			return a
		}
	}
	if a > limit2 {
		return limit2
	} else if a < limit1 {
		return limit1
	}
	return a
}

func i32min(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}

func i32max(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

func i32abs(a int32) int32 {
	if a > 0 {
		return a
	}
	return -a
}

func log2lin(inLogQ7 int32) int32 {
	out, fracQ7 := int32(0), int32(0)

	if inLogQ7 < 0 {
		return 0
	} else if inLogQ7 >= (31 << 7) {
		return math.MaxInt32
	}

	out = lshift(1, rshift(inLogQ7, 7))
	fracQ7 = inLogQ7 & 0x7F

	if inLogQ7 < 2048 {
		out = out + rshift(mul(out, smlawb(fracQ7, mul(fracQ7, 128-fracQ7), -174)), 7)
	} else {
		out = mla(out, rshift(out, 7), smlawb(fracQ7, mul(fracQ7, 128-fracQ7), -174))
	}

	return out
}

func insertionSortIncreasingAllValues(a []int32, L int32) {
	var i, j int
	for i = 1; i < int(L); i++ {
		value := a[i]
		for j = i - 1; j >= 0 && (value < a[j]); j-- {
			a[j+1] = a[j]
		}
		a[j+1] = value
	}
}

func div32varQ(a32, b32, Qres int32) int32 {
	assert(b32 != 0)
	assert(Qres >= 0)

	aHeadrm := clz32(i32abs(a32)) - 1
	a32nrm := lshift(a32, aHeadrm)
	bHeadrm := clz32(i32abs(b32)) - 1
	b32nrm := lshift(b32, bHeadrm)

	b32inv := (math.MaxInt32 >> 2) / rshift(b32nrm, 16)

	result := smulwb(a32nrm, b32inv)

	a32nrm -= lshift(smmul(b32nrm, result), 3)

	result = smlawb(result, a32nrm, b32inv)

	ls := 29 + aHeadrm - bHeadrm - Qres
	if ls <= 0 {
		return lshiftSAT32(result, -ls)
	}

	if ls < 32 {
		return rshift(result, ls)
	}

	return 0
}

func _MAPrediction(in []int16, B []int16, S []int32, out []int16, length int32, order int32) {
	var k, d, in16, out32 int32
	for k = int32(0); k < length; k++ {
		in16 = int32(in[k])
		out32 = lshift(in16, 12) - S[0]
		out32 = rrshift(out32, 12)

		for d = 0; d < order-1; d++ {
			S[d] = smlabb(S[d+1], in16, int32(B[d]))
		}
		S[order-1] = smulbb(in16, int32(B[order-1]))

		out[k] = sat16(out32)
	}
}

func sumSqrShift(energy, shift *int32, x []int16, len int32) {
	nrg := int32(x[0] * x[0])
	i := int32(1)
	shft := int32(0)
	len--

	for ; i < len; i += 2 {
		nrg = nrg + (int32(x[i+1]) * int32(x[i+1]))
		nrg = nrg + (int32(x[i]) * int32(x[i]))
		if nrg < 0 {
			nrg = nrg >> 2
			shft = 2
			break
		}
	}

	var nrgTmp int32
	for ; i < len; i += 2 {
		nrgTmp = int32(x[i+1]) * int32(x[i+1])
		nrgTmp = nrgTmp + (int32(x[i]) * int32(x[i]))
		nrg += nrgTmp >> shft
		if nrg < 0 {
			nrg = nrg >> 2
			shft += 2
		}
	}

	if i == len {
		nrgTmp = int32(x[i]) * int32(x[i])
		nrg += nrgTmp >> 2
	}

	if int(nrg)&0xc0000000 > 0 {
		nrg = nrg >> 2
		shft += 2
	}

	*shift = shft
	*energy = nrg
}

func ror32(a32, rot int32) int32 {
	x := uint32(a32)
	r := uint32(rot)
	m := uint32(-rot)
	if rot <= 0 {
		return int32((x << m) | (x >> (32 - m)))
	}
	return int32((x << (32 - r)) | (x >> r))
}

func clzFrac(in int32, lz, fracQ7 *int32) {
	lzeros := clz32(in)
	*lz = lzeros
	*fracQ7 = ror32(in, 24-lzeros) & 0x7f
}

func sqrtApprox(x int32) int32 {
	if x <= 0 {
		return 0
	}

	var y, _lz, fracQ7 int32

	clzFrac(x, &_lz, &fracQ7)

	if _lz&1 > 0 {
		y = 32768
	} else {
		y = 46214
	}

	y >>= rshift(_lz, 1)
	y = smlawb(y, y, smulbb(213, fracQ7))

	return y
}

func addSAT32(a, b int32) int32 {
	if (int(a+b) & 0x80000000) == 0 {
		if (int(a&b) & 0x80000000) != 0 {
			return math.MinInt32
		} else {
			return a + b
		}
	} else if (int(a|b) & 0x80000000) == 0 {
		return math.MaxInt32
	}
	return a + b
}

func rshift(a, b int32) int32 {
	return a >> b
}

func u32rshift(a, b uint32) uint32 {
	return a >> b
}

func lshift(a, b int32) int32 {
	return a << b
}

func u32lshift(a, b uint32) uint32 {
	return a << b
}

func mul(a, b int32) int32 {
	return a * b
}

func u32mul(a, b uint32) uint32 {
	return a * b
}

func smulbb(a, b int32) int32 {
	a, b = int32(int16(a)), int32(int16(b))
	return a * b
}

func smull(a, b int32) int64 {
	return int64(a) * int64(b)
}

func div(a, b int32) int32 {
	return a / b
}

func ua2i32(a []int16) int32 {
	return (int32(a[1]) << 16) | (int32(a[0]) & 0xffff)
}

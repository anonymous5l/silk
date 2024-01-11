package silk

import (
	"math"
)

func smlawb(a32, b32, c32 int32) int32 {
	c32 = int32(int16(c32))
	return a32 + ((b32 >> 16) * c32) + (((b32 & 0xffff) * c32) >> 16)
}

func smlalbb(a64 int64, b16, c16 int16) int64 {
	return a64 + int64(int32(b16)*int32(c16))
}

func smulwb(a32, b32 int32) int32 {
	b32 = int32(int16(b32))
	return ((a32 >> 16) * b32) + (((a32 & 0xffff) * b32) >> 16)
}

func smulwt(a32, b32 int32) int32 {
	return (a32>>16)*(b32>>16) + (((a32 & 0x0000FFFF) * (b32 >> 16)) >> 16)
}

func smlawt(a32, b32, c32 int32) int32 {
	return a32 + ((b32 >> 16) * (c32 >> 16)) + (((b32 & 0xffff) * (c32 >> 16)) >> 16)
}

func smlabt(a32, b32, c32 int32) int32 {
	return a32 + int32(int16(b32))*(c32>>16)
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

	bHeadRM := clz32(abs(b32)) - 1
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

func abs(a int32) int32 {
	if a < 0 {
		return -a
	}
	return a
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

func lin2log(inLin int32) int32 {
	var lz, fracQ7 int32

	clzFrac(inLin, &lz, &fracQ7)

	return lshift(31-lz, 7) + smlawb(fracQ7, mul(fracQ7, 128-fracQ7), 179)
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

	aHeadrm := clz32(abs(a32)) - 1
	a32nrm := lshift(a32, aHeadrm)
	bHeadrm := clz32(abs(b32)) - 1
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

		if out != nil {
			out[k] = sat16(out32)
		}
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
	return int32(int16(a)) * int32(int16(b))
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

func gcd(a, b int32) int32 {
	tmp := int32(0)
	for b > 0 {
		tmp = a - b*div(a, b)
		a, b = b, tmp
	}
	return a
}

func fixConst(C float32, Q int32) int32 {
	return int32((C * float32(1<<Q)) + 0.5)
}

func addPosSAT32(a, b int32) int32 {
	ret := a + b
	if int(ret)&0x80000000 > 0 {
		return math.MaxInt32
	}
	return ret
}

func clz64(in int64) int32 {
	var inUpper int32

	inUpper = int32(in >> 32)
	if inUpper == 0 {
		return 32 + clz32(int32(in))
	}

	return clz32(inUpper)
}

func int16ArrayMaxABS(vec []int16) int16 {
	var _max, i, lvl, ind int32

	length := int32(len(vec))
	if length == 0 {
		return 0
	}

	ind = length - 1
	_max = smulbb(int32(vec[ind]), int32(vec[ind]))

	for i = length - 2; i >= 0; i-- {
		lvl = smulbb(int32(vec[i]), int32(vec[i]))
		if lvl > _max {
			_max = lvl
			ind = i
		}
	}

	if _max >= 1073676289 {
		return math.MaxInt16
	} else {
		if vec[ind] < 0 {
			return -vec[ind]
		} else {
			return vec[ind]
		}
	}
}

func _PAnaFindScaling(signal []int16, signalLength, sumSqrLen int32) int32 {
	var nbits, xMax int32

	xMax = int32(int16ArrayMaxABS(signal[:signalLength]))

	if xMax < math.MaxInt16 {
		nbits = 32 - clz32(smulbb(xMax, xMax))
	} else {
		nbits = 30
	}
	nbits += int32(17 - clz16(uint16(sumSqrLen)))

	if nbits < 31 {
		return 0
	}
	return nbits - 30
}

const ScratchSize = 22

func _PAnaCalcCorrSt3(crossCorrSt3 [][][]int32, signal []int16, startLag, sfLength, complexity int32) {
	var (
		targetPtr, basisPtr            []int16
		crossCorr                      int32
		i, j, k, lagCounter            int32
		cbkOffset, cbkSize, delta, idx int32
		scratchMem                     [ScratchSize]int32
	)

	assert(complexity >= PitchESTMinComplex)
	assert(complexity <= PitchESTMaxComplex)

	cbkOffset = int32(cbkOffsetsStage3[complexity])
	cbkSize = int32(cbkSizesStage3[complexity])

	off := lshift(sfLength, 2)

	targetPtr = signal[off:]
	for k = 0; k < PitchESTNBSubFR; k++ {
		lagCounter = 0

		for j = int32(LagRangeStage3[complexity][k][0]); j <= int32(LagRangeStage3[complexity][k][1]); j++ {
			basisPtr = signal[off-(startLag+j)+(sfLength*k):]
			crossCorr = innerProdAligned(targetPtr, basisPtr, sfLength)
			assert(lagCounter < ScratchSize)
			scratchMem[lagCounter] = crossCorr
			lagCounter++
		}

		delta = int32(LagRangeStage3[complexity][k][0])
		for i = cbkOffset; i < (cbkOffset + cbkSize); i++ {
			idx = int32(CBLagsStage3[k][i]) - delta
			for j = 0; j < PitchESTNBStage3Lags; j++ {
				assert(idx+j < ScratchSize)
				assert(idx+j < lagCounter)

				crossCorrSt3[k][i][j] = scratchMem[idx+j]
			}
		}
		targetPtr = targetPtr[sfLength:]
	}
}

func _PAnaCalcEnergySt3(energiesSt3 [][][]int32, signal []int16, startLag, sfLength, complexity int32) {
	var (
		energy                         int32
		targetPtr, basisPtr            []int16
		k, i, j, lagCounter            int32
		cbkOffset, cbkSize, delta, idx int32
		scratchMem                     [ScratchSize]int32
	)

	cbkOffset = int32(cbkOffsetsStage3[complexity])
	cbkSize = int32(cbkSizesStage3[complexity])

	off := lshift(sfLength, 2)
	targetPtr = signal[off:]
	for k = 0; k < PitchESTNBSubFR; k++ {
		lagCounter = 0

		basisOff := off - (startLag + int32(LagRangeStage3[complexity][k][0])) + (sfLength * k)
		basisPtr = signal[basisOff:]
		energy = innerProdAligned(basisPtr, basisPtr, sfLength)
		assert(energy >= 0)
		scratchMem[lagCounter] = energy
		lagCounter++

		for i = 1; i < int32(LagRangeStage3[complexity][k][1])-int32(LagRangeStage3[complexity][k][0])+1; i++ {
			energy -= smulbb(int32(basisPtr[sfLength-i]), int32(basisPtr[sfLength-i]))
			assert(energy >= 0)

			energy = addSAT32(energy, smulbb(int32(signal[basisOff-i]), int32(signal[basisOff-i])))
			assert(energy >= 0)
			assert(lagCounter < ScratchSize)

			scratchMem[lagCounter] = energy
			lagCounter++
		}

		delta = int32(LagRangeStage3[complexity][k][0])
		for i = cbkOffset; i < cbkOffset+cbkSize; i++ {
			idx = int32(CBLagsStage3[k][i]) - delta
			for j = 0; j < PitchESTNBStage3Lags; j++ {
				assert(idx+j < ScratchSize)
				assert(idx+j < lagCounter)
				energiesSt3[k][i][j] = scratchMem[idx+j]
				assert(energiesSt3[k][i][j] >= 0)
			}
		}

		targetPtr = targetPtr[sfLength:]
	}
}

func insertionSortDecreasingInt16(a []int16, index []int32, L, K int32) {
	var i, j, value int32
	assert(K > 0)
	assert(L > 0)
	assert(L >= K)

	for i = 0; i < K; i++ {
		index[i] = i
	}

	for i = 1; i < K; i++ {
		value = int32(a[i])
		for j = i - 1; (j >= 0) && (value > int32(a[j])); j-- {
			a[j+1] = a[j]
			index[j+1] = index[j]
		}
		a[j+1] = int16(value)
		index[j+1] = i
	}

	for i = K; i < L; i++ {
		value = int32(a[i])
		if value > int32(a[K-1]) {
			for j = K - 2; (j >= 0) && (value > int32(a[j])); j-- {
				a[j+1] = a[j]
				index[j+1] = index[j]
			}
			a[j+1] = int16(value)
			index[j+1] = i
		}
	}
}

func warpedGain(coefsQ24 []int32, lambdaQ16, order int32) int32 {
	var i, gainQ24 int32

	lambdaQ16 = -lambdaQ16
	gainQ24 = coefsQ24[order-1]
	for i = order - 2; i >= 0; i-- {
		gainQ24 = smlawb(coefsQ24[i], gainQ24, lambdaQ16)
	}
	gainQ24 = smlawb(fixConst(1.0, 24), gainQ24, -lambdaQ16)
	return inverse32varQ(gainQ24, 40)
}

func matrixPtr(row, column, N int32) int32 {
	return row*N + column
}

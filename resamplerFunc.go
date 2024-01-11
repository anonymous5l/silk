package silk

func resamplePrivateDown4(S []int32, out, in []int16) {
	var (
		k                 int32
		in32, out32, Y, X int32
	)

	inLen := int32(len(in))
	len4 := rshift(inLen, 2)

	for k = 0; k < len4; k++ {
		in32 = lshift(int32(in[4*k])+int32(in[4*k+1]), 9)

		Y = in32 - S[0]
		X = smlawb(Y, Y, ResamplerDown21)
		out32, S[0] = S[0]+X, in32+X

		in32 = lshift(int32(in[4*k+2])+int32(in[4*k+3]), 9)

		Y = in32 - S[1]
		X = smulwb(Y, ResamplerDown20)

		out32 = out32 + S[1]
		out32, S[1] = out32+X, in32+X

		out[k] = sat16(rrshift(out32, 11))
	}
}

func resamplePrivateUp4(S []int32, out, in []int16) {
	var k, in32, out32, Y, X int32
	var out16 int16

	length := int32(len(in))

	for k = 0; k < length; k++ {
		in32 = lshift(int32(in[k]), 10)

		Y = in32 - S[0]
		X = smulwb(Y, ResamplerUp2LQ0)
		out32, S[0] = S[0]+X, in32+X

		out16 = sat16(rrshift(out32, 10))
		out[4*k], out[4*k+1] = out16, out16

		Y = in32 - S[1]
		X = smlawb(Y, Y, ResamplerUp2LQ1)
		out32, S[1] = S[1]+X, in32+X

		out16 = sat16(rrshift(out32, 10))
		out[4*k+2], out[4*k+3] = out16, out16
	}
}

func resampleDown2(S []int32, out, in []int16) {
	var k, in32, out32, Y, X int32

	len2 := rshift(int32(len(in)), 1)

	for k = 0; k < len2; k++ {
		in32 = lshift(int32(in[2*k]), 10)

		Y = in32 - S[0]
		X = smlawb(Y, Y, ResamplerDown21)
		out32, S[0] = S[0]+X, in32+X

		in32 = lshift(int32(in[2*k+1]), 10)

		Y = in32 - S[1]
		X = smulwb(Y, ResamplerDown20)
		out32 = out32 + S[1]
		out32 = out32 + X
		S[1] = in32 + X

		out[k] = sat16(rrshift(out32, 11))
	}
}

func resampleUp2(S []int32, out, in []int16) {
	var k, in32, out32, Y, X int32

	length := int32(len(in))

	for k = 0; k < length; k++ {
		in32 = lshift(int32(in[k]), 10)

		Y = in32 - S[0]
		X = smulwb(Y, ResamplerUp2LQ0)
		out32, S[0] = S[0]+X, in32+X

		out[2*k] = sat16(rrshift(out32, 10))

		Y = in32 - S[1]
		X = smlawb(Y, Y, ResamplerUp2LQ1)
		out32, S[1] = S[1]+X, in32+X

		out[2*k+1] = sat16(rrshift(out32, 10))
	}
}

func resamplePrivateUp2HQ(S []int32, out, in []int16) {
	var k, in32, out321, out322, Y, X int32

	length := int32(len(in))

	for k = 0; k < length; k++ {
		in32 = lshift(int32(in[k]), 10)

		Y = in32 - S[0]
		X = smulwb(Y, ResamplerUp2HQ0[0])
		out321, S[0] = S[0]+X, in32+X

		Y = out321 - S[1]
		X = smlawb(Y, Y, ResamplerUp2HQ0[1])
		out322, S[1] = S[1]+X, out321+X

		out322 = smlawb(out322, S[5], ResamplerUp2HQNotch[2])
		out322 = smlawb(out322, S[4], ResamplerUp2HQNotch[1])
		out321 = smlawb(out322, S[4], ResamplerUp2HQNotch[0])
		S[5] = out322 - S[5]

		out[2*k] = sat16(rshift(smlawb(256, out321, ResamplerUp2HQNotch[3]), 9))

		Y = in32 - S[2]
		X = smulwb(Y, ResamplerUp2HQ1[0])
		out321, S[2] = S[2]+X, in32+X

		Y = out321 - S[3]
		X = smlawb(Y, Y, ResamplerUp2HQ1[1])
		out322, S[3] = S[3]+X, out321+X

		out322 = smlawb(out322, S[4], ResamplerUp2HQNotch[2])
		out322 = smlawb(out322, S[5], ResamplerUp2HQNotch[1])
		out321 = smlawb(out322, S[5], ResamplerUp2HQNotch[0])
		S[4] = out322 - S[4]

		out[2*k+1] = sat16(rshift(smlawb(256, out321, ResamplerUp2HQNotch[3]), 9))
	}
}

func resamplePrivateIIRFIRInterpol(out, buf []int16, maxIndexQ16, indexIncrementQ16 int32) []int16 {
	var indexQ16, resQ15, tableIndex int32

	var bufPtr []int16
	for indexQ16 = 0; indexQ16 < maxIndexQ16; indexQ16 += indexIncrementQ16 {
		tableIndex = smulwb(indexQ16&0xFFFF, 144)
		bufPtr = buf[indexQ16>>16:]

		resQ15 = smulbb(int32(bufPtr[0]), int32(ResamplerFracFIR144[tableIndex][0]))
		resQ15 = smlabb(resQ15, int32(bufPtr[1]), int32(ResamplerFracFIR144[tableIndex][1]))
		resQ15 = smlabb(resQ15, int32(bufPtr[2]), int32(ResamplerFracFIR144[tableIndex][2]))
		resQ15 = smlabb(resQ15, int32(bufPtr[3]), int32(ResamplerFracFIR144[143-tableIndex][2]))
		resQ15 = smlabb(resQ15, int32(bufPtr[4]), int32(ResamplerFracFIR144[143-tableIndex][1]))
		resQ15 = smlabb(resQ15, int32(bufPtr[5]), int32(ResamplerFracFIR144[143-tableIndex][0]))

		out[0] = sat16(rrshift(resQ15, 15))
		out = out[1:]
	}
	return out
}

func resamplePrivateAR2(S []int32, outQ8 []int32, in, AQ14 []int16) {
	var k, out32 int32
	for k = 0; k < int32(len(in)); k++ {
		out32 = S[0] + lshift(int32(in[k]), 8)
		outQ8[k] = out32

		out32 = lshift(out32, 2)

		S[0] = smlawb(S[1], out32, int32(AQ14[0]))
		S[1] = smulwb(out32, int32(AQ14[1]))
	}
}

func resamplePrivateDownFIRInterrpol0(out []int16, buf2 []int32, FIRCoefs []int16,
	maxIndexQ16, indexIncrementQ16 int32) []int16 {
	var indexQ16, resQ6 int32
	var bufPtr []int32
	for indexQ16 = 0; indexQ16 < maxIndexQ16; indexQ16 += indexIncrementQ16 {
		bufPtr = buf2[rshift(indexQ16, 16):]

		resQ6 = smulwb(bufPtr[0]+bufPtr[11], int32(FIRCoefs[0]))
		resQ6 = smlawb(resQ6, bufPtr[1]+bufPtr[10], int32(FIRCoefs[1]))
		resQ6 = smlawb(resQ6, bufPtr[2]+bufPtr[9], int32(FIRCoefs[2]))
		resQ6 = smlawb(resQ6, bufPtr[3]+bufPtr[8], int32(FIRCoefs[3]))
		resQ6 = smlawb(resQ6, bufPtr[4]+bufPtr[7], int32(FIRCoefs[4]))
		resQ6 = smlawb(resQ6, bufPtr[5]+bufPtr[6], int32(FIRCoefs[5]))

		out[0] = sat16(rrshift(resQ6, 6))
		out = out[1:]
	}
	return out
}

func resamplePrivateDownFIRInterrpol1(out []int16, buf2 []int32, FIRCoefs []int16,
	maxIndexQ16, indexIncrementQ16, FIRFracs int32) []int16 {
	var indexQ16, resQ6, interpolInd int32
	var bufPtr []int32
	var interpolPtr []int16

	for indexQ16 = 0; indexQ16 < maxIndexQ16; indexQ16 += indexIncrementQ16 {
		bufPtr = buf2[rshift(indexQ16, 16):]

		interpolInd = smulwb(indexQ16&0xffff, FIRFracs)

		interpolPtr = FIRCoefs[ResamplerDownOrderFIR/2*interpolInd:]
		resQ6 = smulwb(bufPtr[0], int32(interpolPtr[0]))
		resQ6 = smlawb(resQ6, bufPtr[1], int32(interpolPtr[1]))
		resQ6 = smlawb(resQ6, bufPtr[2], int32(interpolPtr[2]))
		resQ6 = smlawb(resQ6, bufPtr[3], int32(interpolPtr[3]))
		resQ6 = smlawb(resQ6, bufPtr[4], int32(interpolPtr[4]))
		resQ6 = smlawb(resQ6, bufPtr[5], int32(interpolPtr[5]))

		interpolPtr = FIRCoefs[ResamplerDownOrderFIR/2*(FIRFracs-1-interpolInd):]
		resQ6 = smlawb(resQ6, bufPtr[11], int32(interpolPtr[0]))
		resQ6 = smlawb(resQ6, bufPtr[10], int32(interpolPtr[1]))
		resQ6 = smlawb(resQ6, bufPtr[9], int32(interpolPtr[2]))
		resQ6 = smlawb(resQ6, bufPtr[8], int32(interpolPtr[3]))
		resQ6 = smlawb(resQ6, bufPtr[7], int32(interpolPtr[4]))
		resQ6 = smlawb(resQ6, bufPtr[6], int32(interpolPtr[5]))

		out[0] = sat16(rrshift(resQ6, 6))
		out = out[1:]
	}
	return out
}

func resamplePrivateCopy(out, in []int16) {
	memcpy(out, in, len(in))
}

const (
	down23OrderFIR = 4
	down3OrderFIR  = 6
)

func resampleDown23(S []int32, out, in []int16) {
	var nSamplesIn, counter, resQ6 int32
	var buf [ResamplerMaxBatchSizeIn + down23OrderFIR]int32
	var bufPtr []int32

	memcpy(buf[:], S, down23OrderFIR)

	for {
		nSamplesIn = min(int32(len(in)), ResamplerMaxBatchSizeIn)

		resamplePrivateAR2(S[down23OrderFIR:], buf[down23OrderFIR:], in[:nSamplesIn],
			Resampler23CoefsLQ)

		bufPtr = buf[:]
		counter = nSamplesIn
		for counter > 2 {
			resQ6 = smulwb(bufPtr[0], int32(Resampler23CoefsLQ[2]))
			resQ6 = smlawb(resQ6, bufPtr[1], int32(Resampler23CoefsLQ[3]))
			resQ6 = smlawb(resQ6, bufPtr[2], int32(Resampler23CoefsLQ[5]))
			resQ6 = smlawb(resQ6, bufPtr[3], int32(Resampler23CoefsLQ[4]))

			out[0] = sat16(rrshift(resQ6, 6))
			out = out[1:]

			resQ6 = smulwb(bufPtr[1], int32(Resampler23CoefsLQ[4]))
			resQ6 = smlawb(resQ6, bufPtr[2], int32(Resampler23CoefsLQ[5]))
			resQ6 = smlawb(resQ6, bufPtr[3], int32(Resampler23CoefsLQ[3]))
			resQ6 = smlawb(resQ6, bufPtr[4], int32(Resampler23CoefsLQ[2]))

			out[0] = sat16(rrshift(resQ6, 6))
			out = out[1:]

			bufPtr = bufPtr[3:]
			counter -= 3
		}

		in = in[nSamplesIn:]
		if len(in) > 0 {
			memcpy(buf[:], buf[nSamplesIn:], down23OrderFIR)
		} else {
			break
		}
	}

	memcpy(S, buf[nSamplesIn:], down23OrderFIR)
}

func resampleDown3(S []int32, out, in []int16) {
	var nSamplesIn, counter, resQ6 int32
	var buf [ResamplerMaxBatchSizeIn + down3OrderFIR]int32
	var bufPtr []int32

	memcpy(buf[:], S, down3OrderFIR)

	for {
		nSamplesIn = min(int32(len(in)), ResamplerMaxBatchSizeIn)

		resamplePrivateAR2(S[down3OrderFIR:], buf[down3OrderFIR:], in[:nSamplesIn],
			Resampler13CoefsLQ)

		bufPtr = buf[:]
		counter = nSamplesIn
		for counter > 2 {
			resQ6 = smulwb(bufPtr[0]+bufPtr[5], int32(Resampler13CoefsLQ[2]))
			resQ6 = smlawb(resQ6, bufPtr[1]+bufPtr[4], int32(Resampler13CoefsLQ[3]))
			resQ6 = smlawb(resQ6, bufPtr[2]+bufPtr[3], int32(Resampler13CoefsLQ[4]))

			out[0] = sat16(rrshift(resQ6, 6))
			out = out[1:]

			bufPtr = bufPtr[3:]
			counter -= 3
		}

		in = in[nSamplesIn:]

		if len(in) > 0 {
			memcpy(buf[:], buf[nSamplesIn:], down3OrderFIR)
		} else {
			break
		}
	}

	memcpy(S[:], buf[nSamplesIn:], down3OrderFIR)
}

func (r *resampler) resamplePrivateIIRFIR(out, in []int16) {
	var nSamplesIn, maxIndexQ16, indexIncrementQ16 int32

	bufSize := 2*ResamplerMaxBatchSizeIn + ResamplerOrderFIR144
	buf := make([]int16, bufSize, bufSize)

	for i := 0; i < ResamplerOrderFIR144; i++ {
		buf[(i * 2)] = int16(r.sFIR[i] & 0xffff)
		buf[(i*2)+1] = int16((r.sFIR[i+1] >> 16) & 0xffff)
	}

	inLen := int32(len(in))

	indexIncrementQ16 = r.invRatioQ16
	for {
		nSamplesIn = min(inLen, r.batchSize)

		if r.input2x == 1 {
			r.up2Func(r.sIIR[:], buf[ResamplerOrderFIR144:], in[:nSamplesIn])
		} else {
			r.resamplePrivateARMA4(r.sIIR[:], buf[ResamplerOrderFIR144:], in[:nSamplesIn], r.Coefs)
		}

		maxIndexQ16 = lshift(nSamplesIn, 16+r.input2x)
		out = resamplePrivateIIRFIRInterpol(out, buf, maxIndexQ16, indexIncrementQ16)
		in = in[nSamplesIn:]

		if len(in) > 0 {
			memcpy(buf, buf[nSamplesIn<<r.input2x:], ResamplerOrderFIR144*2)
		} else {
			break
		}
	}

	for i := 0; i < ResamplerOrderFIR144; i++ {
		r.sFIR[i] = ua2i32(buf[(nSamplesIn<<r.input2x)+(i*2):])
	}
}

func (r *resampler) resamplePrivateUp2HQWrapper(out, in []int16) {
	resamplePrivateUp2HQ(r.sIIR[:], out, in)
}

func (r *resampler) resamplePrivateARMA4(S []int32, out, in []int16, coef []int16) {
	var k, inQ8, out1Q8, out2Q8, X int32
	length := int32(len(in))

	for k = 0; k < length; k++ {
		inQ8 = lshift(int32(in[k]), 8)

		out1Q8 = inQ8 + lshift(S[0], 2)
		out2Q8 = out1Q8 + lshift(S[2], 2)

		X = smlawb(S[1], inQ8, int32(coef[0]))
		S[0] = smlawb(X, out1Q8, int32(coef[2]))

		X = smlawb(S[3], out1Q8, int32(coef[1]))
		S[2] = smlawb(X, out2Q8, int32(coef[4]))

		S[1] = smlawb(rshift(inQ8, 2), out1Q8, int32(coef[3]))
		S[3] = smlawb(rshift(out1Q8, 2), out2Q8, int32(coef[5]))

		out[k] = sat16(rshift(smlawb(128, out2Q8, int32(coef[6])), 8))
	}
}

func (r *resampler) resamplePrivateDownFIR(out, in []int16) {
	var nSamplesIn, maxIndexQ16, indexIncrementQ16 int32

	buf1 := make([]int16, ResamplerMaxBatchSizeIn/2, ResamplerMaxBatchSizeIn/2)
	buf2 := make([]int32, ResamplerMaxBatchSizeIn+ResamplerDownOrderFIR)

	var FIRCoefs []int16

	memcpy(buf2, r.sFIR[:], ResamplerDownOrderFIR)

	FIRCoefs = r.Coefs[2:]

	indexIncrementQ16 = r.invRatioQ16
	for {
		nSamplesIn = min(int32(len(in)), r.batchSize)

		if r.input2x == 1 {
			resampleDown2(r.sDown2[:], buf1, in[:nSamplesIn])
			nSamplesIn = rshift(nSamplesIn, 1)
			resamplePrivateAR2(r.sIIR[:], buf2[ResamplerDownOrderFIR:], buf1[:nSamplesIn], r.Coefs)
		} else {
			resamplePrivateAR2(r.sIIR[:], buf2[ResamplerDownOrderFIR:], in[:nSamplesIn], r.Coefs)
		}

		maxIndexQ16 = lshift(nSamplesIn, 16)

		if r.FIRFracs == 1 {
			out = resamplePrivateDownFIRInterrpol0(out, buf2, FIRCoefs, maxIndexQ16, indexIncrementQ16)
		} else {
			out = resamplePrivateDownFIRInterrpol1(out, buf2, FIRCoefs, maxIndexQ16, indexIncrementQ16, r.FIRFracs)
		}

		in = in[nSamplesIn<<r.input2x:]

		if int32(len(in)) > r.input2x {
			memcpy(buf2, buf2[nSamplesIn:], ResamplerDownOrderFIR)
		} else {
			break
		}
	}

	memcpy(r.sFIR[:], buf2[nSamplesIn:], ResamplerDownOrderFIR)
}

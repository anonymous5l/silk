package silk

import "math"

var (
	LTPGainBITSQ60 = []int16{
		26, 236, 321, 325, 339, 344, 362, 379,
		412, 418,
	}

	LTPGainBITSQ61 = []int16{
		88, 231, 237, 244, 300, 309, 313, 324,
		325, 341, 346, 351, 352, 352, 354, 356,
		367, 393, 396, 406,
	}

	LTPGainBITSQ62 = []int16{
		238, 248, 255, 257, 258, 274, 284, 311,
		317, 326, 326, 327, 339, 349, 350, 351,
		352, 355, 358, 366, 371, 379, 383, 387,
		388, 393, 394, 394, 407, 409, 412, 412,
		413, 422, 426, 432, 434, 449, 454, 455,
	}

	LTPGainBITSQ6Ptrs = [][]int16{
		LTPGainBITSQ60,
		LTPGainBITSQ61,
		LTPGainBITSQ62,
	}

	LTPScaleThresholdsQ15 = []int16{
		31129, 26214, 16384, 13107, 9830, 6554,
		4915, 3276, 2621, 2458, 0,
	}
)

func corrMatrix(x []int16, L, order, headRoom int32, XX []int32, rshifts *int32) {
	var (
		i, j, lag, rshiftsLocal, headRoomRshifts int32
		energy                                   int32
		ptr1, ptr2                               int32
	)

	sumSqrShift(&energy, &rshiftsLocal, x, L+order-1)

	headRoomRshifts = max(headRoom-clz32(energy), 0)
	energy = rshift(energy, headRoomRshifts)
	rshiftsLocal += headRoomRshifts

	for i = 0; i < order-1; i++ {
		energy -= rshift(smulbb(int32(x[i]), int32(x[i])), rshiftsLocal)
	}
	if rshiftsLocal < *rshifts {
		energy = rshift(energy, *rshifts-rshiftsLocal)
		rshiftsLocal = *rshifts
	}

	XX[matrixPtr(0, 0, order)] = energy
	ptr1 = order - 1
	for j = 1; j < order; j++ {
		energy = energy - rshift(smulbb(int32(x[ptr1+(L-j)]), int32(x[ptr1+(L-j)])), rshiftsLocal)
		energy = energy + rshift(smulbb(int32(x[ptr1-j]), int32(x[ptr1-j])), rshiftsLocal)
		XX[matrixPtr(j, j, order)] = energy
	}

	ptr2 = order - 2
	if rshiftsLocal > 0 {
		for lag = 1; lag < order; lag++ {
			energy = 0
			for i = 0; i < L; i++ {
				energy += rshift(smulbb(int32(x[ptr1+i]), int32(x[ptr2+i])), rshiftsLocal)
			}

			XX[matrixPtr(lag, 0, order)] = energy
			XX[matrixPtr(0, lag, order)] = energy
			for j = 1; j < (order - lag); j++ {
				energy -= rshift(smulbb(int32(x[ptr1+L-j]), int32(x[ptr2+L-j])), rshiftsLocal)
				energy += rshift(smulbb(int32(x[ptr1-j]), int32(x[ptr2-j])), rshiftsLocal)
				XX[matrixPtr(lag+j, j, order)] = energy
				XX[matrixPtr(j, lag+j, order)] = energy
			}
			ptr2--
		}
	} else {
		for lag = 1; lag < order; lag++ {
			energy = innerProdAligned(x[ptr1:], x[ptr2:], L)
			XX[matrixPtr(lag, 0, order)] = energy
			XX[matrixPtr(0, lag, order)] = energy

			for j = 1; j < order-lag; j++ {
				energy = energy - smulbb(int32(x[ptr1+L-j]), int32(x[ptr2+L-j]))
				energy = smlabb(energy, int32(x[ptr1-j]), int32(x[ptr2-j]))
				XX[matrixPtr(lag+j, j, order)] = energy
				XX[matrixPtr(j, lag+j, order)] = energy
			}
			ptr2--
		}
	}
	*rshifts = rshiftsLocal
}

func corrVector(x, t []int16, L, order int32, Xt []int32, rshifts int32) {
	var (
		lag, i, innerProd int32
		ptr1, ptr2        int32
	)

	ptr1 = order - 1
	if rshifts > 0 {
		for lag = 0; lag < order; lag++ {
			innerProd = 0
			for i = 0; i < L; i++ {
				innerProd += rshift(smulbb(int32(x[ptr1+i]), int32(t[ptr2+i])), rshifts)
			}
			Xt[lag] = innerProd
			ptr1--
		}
	} else {
		assert(rshifts == 0)
		for lag = 0; lag < order; lag++ {
			Xt[lag] = innerProdAligned(x[ptr1:], t[ptr2:], L)
			ptr1--
		}
	}
}

func regularizeCorrelations(XX, xx []int32, noise, D int32) {
	for i := int32(0); i < D; i++ {
		XX[matrixPtr(i, i, D)] = XX[matrixPtr(i, i, D)] + noise
	}
	xx[0] += noise
}

func scaleVector32Q26lshift18(data1 []int32, gainQ26, dataSize int32) {
	for i := int32(0); i < dataSize; i++ {
		data1[i] = rshift(int32(smull(data1[i], gainQ26)), 8)
	}
}

func residualEnergy16Covar(c []int16, wXX, wXx []int32, wxx, D, cQ int32) int32 {
	var (
		i, j, lshifts, Qxtra       int32
		cMax, wMax, tmp, tmp2, nrg int32
		cn                         [MaxLPCOrder]int32
		pRow                       []int32
	)

	lshifts = 16 - cQ
	Qxtra = lshifts

	cMax = 0
	for i = 0; i < D; i++ {
		cMax = max(cMax, abs(int32(c[i])))
	}
	Qxtra = min(Qxtra, clz32(cMax)-17)

	wMax = max(wXX[0], wXX[D*D-1])
	Qxtra = min(Qxtra, clz32(mul(D, rshift(smulwb(wMax, cMax), 4)))-5)
	Qxtra = max(Qxtra, 0)
	for i = 0; i < D; i++ {
		cn[i] = lshift(int32(c[i]), Qxtra)
		assert(abs(cn[i]) <= math.MaxInt16+1)
	}
	lshifts -= Qxtra

	tmp = 0
	for i = 0; i < D; i++ {
		tmp = smlawb(tmp, wXx[i], cn[i])
	}
	nrg = rshift(wxx, 1+lshifts) - tmp

	tmp2 = 0
	for i = 0; i < D; i++ {
		tmp = 0
		pRow = wXX[i*D:]
		for j = i + 1; j < D; j++ {
			tmp = smlawb(tmp, pRow[j], cn[j])
		}
		tmp = smlawb(tmp, rshift(pRow[i], 1), cn[i])
		tmp2 = smlawb(tmp2, tmp, cn[i])
	}
	nrg = nrg + lshift(tmp2, lshifts)

	if nrg < 1 {
		nrg = 1
	} else if nrg > rshift(math.MaxInt32, lshifts+2) {
		nrg = math.MaxInt32 >> 1
	} else {
		nrg = lshift(nrg, lshifts+1)
	}
	return nrg
}

func fitLTP(LTPcoefsQ16 []int32, LTPcoefsQ14 []int16) {
	for i := int32(0); i < LTPOrder; i++ {
		LTPcoefsQ14[i] = sat16(rrshift(LTPcoefsQ16[i], 2))
	}
}

type invDt struct {
	Q36Part int32
	Q48Part int32
}

func _LSSoleLast(LQ16 []int32, M int32, b, xQ16 []int32) {
	var (
		i, j  int32
		ptr32 []int32
		tmp32 int32
	)
	for i = M - 1; i >= 0; i-- {
		ptr32 = LQ16[matrixPtr(0, i, M):]
		tmp32 = 0
		for j = M - 1; j > i; j-- {
			tmp32 = smlaww(tmp32, ptr32[smulbb(j, M)], xQ16[j])
		}
		xQ16[i] = b[i] - tmp32
	}
}

func _LSSolveFirst(LQ16 []int32, M int32, b []int32, xQ16 []int32) {
	var (
		i, j, tmp32 int32
		ptr32       []int32
	)

	for i = 0; i < M; i++ {
		ptr32 = LQ16[matrixPtr(i, 0, M):]
		tmp32 = 0
		for j = 0; j < i; j++ {
			tmp32 = smlaww(tmp32, ptr32[j], xQ16[j])
		}
		xQ16[i] = b[i] - tmp32
	}

}

func _LSDivideQ16(T []int32, invD []invDt, M int32) {
	var i, tmp32, oneDivDiagQ36, oneDivDiagQ48 int32
	for i = 0; i < M; i++ {
		oneDivDiagQ36 = invD[i].Q36Part
		oneDivDiagQ48 = invD[i].Q48Part

		tmp32 = T[i]
		T[i] = smmul(tmp32, oneDivDiagQ48) + rshift(smulww(tmp32, oneDivDiagQ36), 4)
	}
}

func _LDLFactorize(A []int32, M int32, LQ16 []int32, invD []invDt) {
	var (
		i, j, k, status, loopCount                  int32
		ptr1, ptr2                                  int32
		diagMinValue, tmp32, err                    int32
		vQ0                                         [MaxLPCOrder]int32
		DQ0                                         [MaxLPCOrder]int32
		oneDivDiagQ36, oneDivDiagQ40, oneDivDiagQ48 int32
	)

	assert(M <= MaxLPCOrder)

	status = 1
	diagMinValue = max(smmul(addSAT32(A[0], A[smulbb(M, M)-1]),
		fixConst(FindLTPCondFac, 32)), 1<<9)
	for loopCount = 0; loopCount < M && status == 1; loopCount++ {
		status = 0
		for j = 0; j < M; j++ {
			ptr1 = matrixPtr(j, 0, M)
			tmp32 = 0
			for i = 0; i < j; i++ {
				vQ0[i] = smulww(DQ0[i], LQ16[ptr1+i])
				tmp32 = smlaww(tmp32, vQ0[i], LQ16[ptr1+i])
			}
			tmp32 = A[matrixPtr(j, j, M)] - tmp32

			if tmp32 < diagMinValue {
				tmp32 = smulbb(loopCount+1, diagMinValue) - tmp32
				for i = 0; i < M; i++ {
					A[matrixPtr(i, i, M)] = A[matrixPtr(i, i, M)] + tmp32
				}
				status = 1
				break
			}
			DQ0[j] = tmp32

			oneDivDiagQ36 = inverse32varQ(tmp32, 36)
			oneDivDiagQ40 = lshift(oneDivDiagQ36, 4)
			err = (1 << 24) - smulww(tmp32, oneDivDiagQ40)
			oneDivDiagQ48 = smulww(err, oneDivDiagQ40)

			invD[j].Q36Part = oneDivDiagQ36
			invD[j].Q48Part = oneDivDiagQ48

			LQ16[matrixPtr(j, j, M)] = 65536
			ptr1 = matrixPtr(j, 0, M)
			ptr2 = matrixPtr(j+1, 0, M)
			for i = j + 1; i < M; i++ {
				tmp32 = 0
				for k = 0; k < j; k++ {
					tmp32 = smlaww(tmp32, vQ0[k], LQ16[ptr2+k])
				}
				tmp32 = A[ptr1+i] - tmp32

				LQ16[matrixPtr(i, j, M)] = smmul(tmp32, oneDivDiagQ48) + rshift(smulww(tmp32, oneDivDiagQ36), 4)

				ptr2 += M
			}
		}
	}

	assert(status == 0)
}

func solveLDL(A []int32, M int32, b []int32, xQ16 []int32) {
	var (
		LQ16 [MaxLPCOrder * MaxLPCOrder]int32
		Y    [MaxLPCOrder]int32
		invD [MaxLPCOrder]invDt
	)

	assert(M <= MaxLPCOrder)

	_LDLFactorize(A, M, LQ16[:], invD[:])
	_LSSolveFirst(LQ16[:], M, b, Y[:])
	_LSDivideQ16(Y[:], invD[:], M)
	_LSSoleLast(LQ16[:], M, Y[:], xQ16)
}

func findLTP(
	bQ14 []int16,
	WLTP []int32,
	LTPredCodGainQ7 *int32,
	rFirst, rLast []int16,
	lag, WghtQ15 []int32,
	subfrLength, memOffset int32,
	corrRShifts []int32) {

	var (
		i, k, ls     int32
		rPtr, lagPtr []int16
		bQ14Ptr      []int16

		regu                                        int32
		WLTPPtr                                     []int32
		bQ16                                        [LTPOrder]int32
		deltabQ14                                   [LTPOrder]int32
		dQ14                                        [NBSubFR]int32
		nrg                                         [NBSubFR]int32
		gQ26                                        int32
		w                                           [NBSubFR]int32
		WLTPmax, maxAbsdQ14, maxwbits               int32
		temp32, denom32                             int32
		extraShifts                                 int32
		rrShifts, maxRshifsts, maxRshiftswxtra, LZs int32
		LPCresnrg, LPCLTPresnrg, divQ16             int32
		Rr, rr                                      [LTPOrder]int32
		wd, mQ12                                    int32
	)

	bQ14Ptr = bQ14
	WLTPPtr = WLTP
	rPtr = rFirst
	for k = 0; k < NBSubFR; k++ {
		if k == (NBSubFR >> 1) {
			rPtr = rLast
		}
		lagPtr = rPtr[memOffset-(lag[k]+LTPOrder/2):]
		rPtr = rPtr[memOffset:]

		sumSqrShift(&rr[k], &rrShifts, rPtr, subfrLength)

		LZs = clz32(rr[k])
		if LZs < LTPCorrsHeadRoom {
			rr[k] = rrshift(rr[k], LTPCorrsHeadRoom-LZs)
			rrShifts += LTPCorrsHeadRoom - LZs
		}

		corrRShifts[k] = rrShifts
		corrMatrix(lagPtr, subfrLength, LTPOrder, LTPCorrsHeadRoom, WLTPPtr, &corrRShifts[k])
		corrVector(lagPtr, rPtr, subfrLength, LTPOrder, Rr[:], corrRShifts[k])

		if corrRShifts[k] > rrShifts {
			rr[k] = rshift(rr[k], corrRShifts[k]-rrShifts)
		}

		assert(rr[k] >= 0)

		regu = 1
		regu = smlawb(regu, rr[k], fixConst(LTPDAMping/3, 16))
		regu = smlawb(regu, WLTPPtr[matrixPtr(0, 0, LTPOrder)], fixConst(LTPDAMping/3, 16))
		regu = smlawb(regu, WLTPPtr[matrixPtr(LTPOrder-1, LTPOrder-1, LTPOrder)], fixConst(LTPDAMping/3, 16))

		regularizeCorrelations(WLTPPtr, rr[k:], regu, LTPOrder)

		solveLDL(WLTPPtr, LTPOrder, Rr[:], bQ16[:])

		fitLTP(bQ16[:], bQ14Ptr)

		nrg[k] = residualEnergy16Covar(bQ14Ptr, WLTPPtr, Rr[:], rr[k], LTPOrder, 14)

		extraShifts = min(corrRShifts[k], LTPCorrsHeadRoom)
		denom32 = lshiftSAT32(smulwb(nrg[k], WghtQ15[k]), 1+extraShifts) +
			rshift(smulwb(subfrLength, 655), corrRShifts[k]-extraShifts)
		denom32 = max(denom32, 1)
		assert(WghtQ15[k]<<16 < math.MaxInt32)

		temp32 = div(lshift(WghtQ15[k], 16), denom32)
		temp32 = rshift(temp32, 31+corrRShifts[k]-extraShifts-26)

		WLTPmax = 0
		for i = 0; i < LTPOrder*LTPOrder; i++ {
			WLTPmax = max(WLTPPtr[i], WLTPmax)
		}

		ls = clz32(WLTPmax) - 1 - 3
		assert(26-18+ls >= 0)
		if 26-18+ls < 31 {
			temp32 = min(temp32, lshift(1, 26-18+ls))
		}

		scaleVector32Q26lshift18(WLTPPtr, temp32, LTPOrder*LTPOrder)

		w[k] = WLTPPtr[matrixPtr(LTPOrder>>1, LTPOrder>>1, LTPOrder)]
		assert(w[k] >= 0)

		rPtr = rPtr[subfrLength:]
		bQ14Ptr = bQ14Ptr[LTPOrder:]
		WLTPPtr = WLTPPtr[LTPOrder*LTPOrder:]
	}

	maxRshifsts = 0
	for k = 0; k < NBSubFR; k++ {
		maxRshifsts = max(corrRShifts[k], maxRshifsts)
	}

	if LTPredCodGainQ7 != nil {
		LPCLTPresnrg = 0
		LPCresnrg = 0
		for k = 0; k < NBSubFR; k++ {
			LPCresnrg = LPCresnrg + rshift(smulwb(rr[k], WghtQ15[k])+1, 1+(maxRshifsts-corrRShifts[k]))
			LPCLTPresnrg = LPCLTPresnrg + rshift(smulwb(nrg[k], WghtQ15[k])+1, 1+(maxRshifsts-corrRShifts[k]))
		}
		LPCLTPresnrg = max(LPCLTPresnrg, 1)

		divQ16 = div32varQ(LPCresnrg, LPCLTPresnrg, 16)

		*LTPredCodGainQ7 = smulbb(3, lin2log(divQ16)-(16<<7))

		assert(*LTPredCodGainQ7 == int32(sat16(mul(3, lin2log(divQ16)-(16<<7)))))
	}

	bQ14Ptr = bQ14
	for k = 0; k < NBSubFR; k++ {
		dQ14[k] = 0
		for i = 0; i < LTPOrder; i++ {
			dQ14[k] += int32(bQ14Ptr[i])
		}
		bQ14Ptr = bQ14Ptr[LTPOrder:]
	}

	maxAbsdQ14, maxwbits = 0, 0
	for k = 0; k < NBSubFR; k++ {
		maxAbsdQ14 = max(maxAbsdQ14, abs(dQ14[k]))
		maxwbits = max(maxwbits, 32-clz32(w[k])+corrRShifts[k]-maxRshifsts)
	}

	assert(maxAbsdQ14 <= (5 << 15))

	extraShifts = maxwbits + 32 - clz32(maxAbsdQ14) - 14

	extraShifts -= 32 - 1 - 2 + maxRshifsts
	extraShifts = max(extraShifts, 0)

	maxRshiftswxtra = maxRshifsts + extraShifts

	temp32 = rshift(262, maxRshifsts+extraShifts) + 1
	wd = 0
	for k = 0; k < NBSubFR; k++ {
		temp32 = temp32 + rshift(w[k], maxRshiftswxtra-corrRShifts[k])
		wd = wd + lshift(smulww(rshift(w[k], maxRshiftswxtra-corrRShifts[k]), dQ14[k]), 2)
	}
	mQ12 = div32varQ(wd, temp32, 12)

	bQ14Ptr = bQ14
	for k = 0; k < NBSubFR; k++ {
		if 2-corrRShifts[k] > 0 {
			temp32 = rshift(w[k], 2-corrRShifts[k])
		} else {
			temp32 = lshiftSAT32(w[k], corrRShifts[k]-2)
		}

		gQ26 = mul(
			div(
				fixConst(LTPSmoothing, 26),
				rshift(fixConst(LTPSmoothing, 26), 10)+temp32),
			lshiftSAT32(mQ12-rshift(dQ14[k], 2), 4))

		temp32 = 0
		for i = 0; i < LTPOrder; i++ {
			deltabQ14[i] = int32(max(bQ14Ptr[i], 1638))
			temp32 += deltabQ14[i]
		}

		temp32 = div(gQ26, temp32)
		for i = 0; i < LTPOrder; i++ {
			bQ14Ptr[i] = int16(limit(int32(bQ14Ptr[i])+smulwb(lshiftSAT32(temp32, 4), deltabQ14[i]),
				-16000, 28000))
		}
		bQ14Ptr = bQ14Ptr[LTPOrder:]
	}
}

func _VQWMatEC(ind, rateDistQ14 *int32, inQ14 []int16, WQ18 []int32, cbQ14, clQ6 []int16, muQ8, L int32) {
	var (
		k                                                int32
		cbRowQ14                                         []int16
		sum1Q14, sum2Q16, diffQ1401, diffQ1423, diffQ144 int32
	)

	*rateDistQ14 = math.MaxInt32
	cbRowQ14 = cbQ14
	for k = 0; k < L; k++ {
		diffQ1401 = int32(uint16(int32(inQ14[0]-cbRowQ14[0]) | lshift(int32(inQ14[1]-cbRowQ14[1]), 16)))
		diffQ1423 = int32(uint16(int32(inQ14[2]-cbRowQ14[2]) | lshift(int32(inQ14[3]-cbRowQ14[3]), 16)))
		diffQ144 = int32(inQ14[4] - cbRowQ14[4])

		sum1Q14 = smulbb(muQ8, int32(clQ6[k]))
		assert(sum1Q14 >= 0)

		sum2Q16 = smulwt(WQ18[1], diffQ1401)
		sum2Q16 = smlawb(sum2Q16, WQ18[2], diffQ1423)
		sum2Q16 = smlawt(sum2Q16, WQ18[3], diffQ1423)
		sum2Q16 = smlawb(sum2Q16, WQ18[4], diffQ144)
		sum2Q16 = lshift(sum2Q16, 1)
		sum2Q16 = smlawb(sum2Q16, WQ18[0], diffQ1401)
		sum1Q14 = smlawb(sum1Q14, sum2Q16, diffQ1401)

		sum2Q16 = smulwb(WQ18[7], diffQ1423)
		sum2Q16 = smlawt(sum2Q16, WQ18[8], diffQ1423)
		sum2Q16 = smlawb(sum2Q16, WQ18[9], diffQ144)
		sum2Q16 = lshift(sum2Q16, 1)
		sum2Q16 = smlawt(sum2Q16, WQ18[6], diffQ1401)
		sum1Q14 = smlawt(sum1Q14, sum2Q16, diffQ1401)

		sum2Q16 = smulwt(WQ18[13], diffQ1423)
		sum2Q16 = smlawb(sum2Q16, WQ18[14], diffQ144)
		sum2Q16 = lshift(sum2Q16, 1)
		sum2Q16 = smlawb(sum2Q16, WQ18[12], diffQ1423)
		sum1Q14 = smlawb(sum1Q14, sum2Q16, diffQ1423)

		sum2Q16 = smulwb(WQ18[19], diffQ144)
		sum2Q16 = lshift(sum2Q16, 1)
		sum2Q16 = smlawt(sum2Q16, WQ18[18], diffQ1423)
		sum1Q14 = smlawt(sum1Q14, sum2Q16, diffQ1423)

		sum2Q16 = smulwb(WQ18[24], diffQ144)
		sum1Q14 = smlawb(sum1Q14, sum2Q16, diffQ144)

		assert(sum1Q14 >= 0)

		if sum1Q14 < *rateDistQ14 {
			*rateDistQ14 = sum1Q14
			*ind = k
		}

		cbRowQ14 = cbRowQ14[LTPOrder:]
	}
}

func quantLTPGains(BQ14 []int16, cbkIndex []int32, periodicityIndex *int32,
	WQ18 []int32, muQ8, lowComplexity int32) {
	var (
		j, k                                 int32
		tempIdx                              [NBSubFR]int32
		cbkSize                              int32
		clPtr, cbkPtrQ14, bQ14Ptr            []int16
		WQ18Ptr                              []int32
		rateDistSubFR, rateDist, minRateDist int32
	)

	minRateDist = math.MaxInt32
	for k = 0; k < 3; k++ {
		clPtr = LTPGainBITSQ6Ptrs[k]
		cbkPtrQ14 = LTPVqPtrsQ14[k]
		cbkSize = LTPVqSizes[k]

		WQ18Ptr = WQ18
		bQ14Ptr = BQ14

		rateDist = 0
		for j = 0; j < NBSubFR; j++ {
			_VQWMatEC(
				&tempIdx[j],
				&rateDistSubFR,
				bQ14Ptr,
				WQ18Ptr,
				cbkPtrQ14,
				clPtr,
				muQ8,
				cbkSize)

			rateDist = addPosSAT32(rateDist, rateDistSubFR)

			bQ14Ptr = bQ14Ptr[LTPOrder:]
			WQ18Ptr = WQ18Ptr[LTPOrder*LTPOrder:]
		}

		rateDist = min(math.MaxInt32-1, rateDist)

		if rateDist < minRateDist {
			minRateDist = rateDist
			memcpy(cbkIndex, tempIdx[:], NBSubFR)
			*periodicityIndex = k
		}

		if lowComplexity > 0 && (rateDist < LTPGainMiddleAvgRDQ14) {
			break
		}
	}

	cbkPtrQ14 = LTPVqPtrsQ14[*periodicityIndex]
	for j = 0; j < NBSubFR; j++ {
		for k = 0; k < LTPOrder; k++ {
			BQ14[j*LTPOrder+k] = cbkPtrQ14[mla(k, cbkIndex[j], LTPOrder)]
		}
	}
}

func _LTPAnalysisFilter(
	LTPres, x []int16, bufIdx int32, LTPCoefQ14 []int16, pitchL, invGainsQ16 []int32,
	subfrLength, preLength int32) {
	var (
		xPtr, xLagPtr []int16
		BtmpQ14       [LTPOrder]int16
		LTPresPtr     []int16
		k, i, j       int32
		LTPest        int32
	)

	xPtr = x[bufIdx:]
	LTPresPtr = LTPres
	for k = 0; k < NBSubFR; k++ {
		xLagPtr = x[(bufIdx+(k*subfrLength))-pitchL[k]:]
		for i = 0; i < LTPOrder; i++ {
			BtmpQ14[i] = LTPCoefQ14[k*LTPOrder+i]
		}
		for i = 0; i < subfrLength+preLength; i++ {
			LTPresPtr[i] = xPtr[i]

			LTPest = smulbb(int32(xLagPtr[LTPOrder/2]), int32(BtmpQ14[0]))
			for j = 1; j < LTPOrder; j++ {
				LTPest = smlabb(LTPest, int32(xLagPtr[LTPOrder/2-j]), int32(BtmpQ14[j]))
			}
			LTPest = rrshift(LTPest, 14)

			LTPresPtr[i] = sat16(int32(xPtr[i]) - LTPest)

			LTPresPtr[i] = int16(smulwb(invGainsQ16[k], int32(LTPresPtr[i])))

			xLagPtr = xLagPtr[1:]
		}

		LTPresPtr = LTPresPtr[subfrLength+preLength:]
		xPtr = xPtr[subfrLength:]
	}

}

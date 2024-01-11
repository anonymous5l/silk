package silk

func limitWarpedCoefs(coefsSynQ24, coefsAnaQ24 []int32, lambdaQ16, limitQ24, order int32) {
	var i, iter, ind int32
	var tmp, maxAbsQ24, chirpQ16, gainSynQ16, gainAnaQ16 int32
	var nomQ16, denQ24 int32

	lambdaQ16 = -lambdaQ16
	for i = order - 1; i > 0; i-- {
		coefsSynQ24[i-1] = smlawb(coefsSynQ24[i-1], coefsSynQ24[i], lambdaQ16)
		coefsAnaQ24[i-1] = smlawb(coefsAnaQ24[i-1], coefsAnaQ24[i], lambdaQ16)
	}

	lambdaQ16 = -lambdaQ16
	nomQ16 = smlawb(fixConst(1.0, 16), -lambdaQ16, lambdaQ16)
	denQ24 = smlawb(fixConst(1.0, 24), coefsSynQ24[0], lambdaQ16)
	gainSynQ16 = div32varQ(nomQ16, denQ24, 24)
	denQ24 = smlawb(fixConst(1.0, 24), coefsAnaQ24[0], lambdaQ16)
	gainAnaQ16 = div32varQ(nomQ16, denQ24, 24)
	for i = 0; i < order; i++ {
		coefsSynQ24[i] = smulww(gainSynQ16, coefsSynQ24[i])
		coefsAnaQ24[i] = smulww(gainAnaQ16, coefsAnaQ24[i])
	}

	for iter = 0; iter < 10; iter++ {
		maxAbsQ24 = -1
		for i = 0; i < order; i++ {
			tmp = max(abs(coefsSynQ24[i]), abs(coefsAnaQ24[i]))
			if tmp > maxAbsQ24 {
				maxAbsQ24 = tmp
				ind = i
			}
		}
		if maxAbsQ24 <= limitQ24 {
			return
		}

		for i = 1; i < order; i++ {
			coefsSynQ24[i-1] = smlawb(coefsSynQ24[i-1], coefsSynQ24[i], lambdaQ16)
			coefsAnaQ24[i-1] = smlawb(coefsAnaQ24[i-1], coefsAnaQ24[i], lambdaQ16)
		}
		gainSynQ16 = inverse32varQ(gainSynQ16, 32)
		gainAnaQ16 = inverse32varQ(gainAnaQ16, 32)

		for i = 0; i < order; i++ {
			coefsSynQ24[i] = smulww(gainSynQ16, coefsSynQ24[i])
			coefsAnaQ24[i] = smulww(gainAnaQ16, coefsAnaQ24[i])
		}

		chirpQ16 = fixConst(0.99, 16) - div32varQ(
			smulwb(maxAbsQ24-limitQ24, smlabb(fixConst(0.8, 10), fixConst(0.1, 10), iter)),
			mul(maxAbsQ24, ind+1), 22)
		bwexpander32(coefsSynQ24, order, chirpQ16)
		bwexpander32(coefsAnaQ24, order, chirpQ16)

		lambdaQ16 = -lambdaQ16
		for i = order - 1; i > 0; i-- {
			coefsSynQ24[i-1] = smlawb(coefsSynQ24[i-1], coefsSynQ24[i], lambdaQ16)
			coefsAnaQ24[i-1] = smlawb(coefsAnaQ24[i-1], coefsAnaQ24[i], lambdaQ16)
		}

		lambdaQ16 = -lambdaQ16
		nomQ16 = smlawb(fixConst(1.0, 16), -lambdaQ16, lambdaQ16)
		denQ24 = smlawb(fixConst(1.0, 24), coefsSynQ24[0], lambdaQ16)
		gainSynQ16 = div32varQ(nomQ16, denQ24, 24)
		denQ24 = smlawb(fixConst(1.0, 24), coefsAnaQ24[0], lambdaQ16)
		gainAnaQ16 = div32varQ(nomQ16, denQ24, 24)
		for i = 0; i < order; i++ {
			coefsSynQ24[i] = smulww(gainSynQ16, coefsSynQ24[i])
			coefsAnaQ24[i] = smulww(gainAnaQ16, coefsAnaQ24[i])
		}
	}
	panic(false)
}

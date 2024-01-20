package silk

func autocorr(results *slice[int32], scale *int32, inputData *slice[int16],
	inputDataSize, correlationCount int32) {
	var (
		i, lz, nRightShifts, corrCount int32
		corr64                         int64
	)

	corrCount = min(inputDataSize, correlationCount)

	corr64 = inner_prod16_aligned_64(inputData, inputData, inputDataSize)

	corr64 += 1

	lz = CLZ64(corr64)

	nRightShifts = 35 - lz
	*scale = nRightShifts

	if nRightShifts <= 0 {
		*results.ptr(0) = LSHIFT(int32(corr64), -nRightShifts)

		for i = 1; i < corrCount; i++ {
			*results.ptr(int(i)) = LSHIFT(inner_prod_aligned(inputData, inputData.off(int(i)), inputDataSize-i), -nRightShifts)
		}
	} else {
		*results.ptr(0) = int32(RSHIFT64(corr64, nRightShifts))

		for i = 1; i < corrCount; i++ {
			*results.ptr(int(i)) = int32(RSHIFT64(inner_prod16_aligned_64(inputData, inputData.off(int(i)), inputDataSize-i), nRightShifts))
		}
	}
}

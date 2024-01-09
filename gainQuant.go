package silk

const (
	Offset      = (MinQGainDB*128)/6 + 16*128
	InvScaleQ16 = (65536 * (((MaxQGainDB - MinQGainDB) * 128) / 6)) / (NLevelsQGain - 1)
)

func gainsDequant(gainQ16 []int32, ind []int32, prevInd *int32, conditional int) {
	for k := 0; k < NBSubFR; k++ {
		if k == 0 && conditional == 0 {
			*prevInd = ind[k]
		} else {
			*prevInd += ind[k] + MinDeltaGainQuant
		}

		gainQ16[k] = log2lin(i32min(smulwb(InvScaleQ16, *prevInd)+Offset, 3967))
	}
}

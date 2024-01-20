package silk

const RESAMPLER_SUPPORT_ABOVE_48KHZ = 1
const RESAMPLER_MAX_FIR_ORDER = 16
const RESAMPLER_MAX_IIR_ORDER = 6

type (
	resampleFunc1 func(*resampler_state_struct, *slice[int16], *slice[int16], int32)
	resampleFunc2 func(*slice[int32], *slice[int16], *slice[int16], int32)
)

type resampler_state_struct struct {
	sIIR               *slice[int32]
	sFIR               *slice[int32]
	sDown2             *slice[int32]
	resampler_function resampleFunc1
	up2_function       resampleFunc2
	batchSize          int32
	invRatio_Q16       int32
	FIR_Fracs          int32
	input2x            int32
	Coefs              *slice[int16]
	sDownPre           *slice[int32]
	sUpPost            *slice[int32]
	down_pre_function  resampleFunc2
	up_post_function   resampleFunc2
	batchSizePrePost   int32
	ratio_Q16          int32
	nPreDownsamplers   int32
	nPostUpsamplers    int32
	magic_number       int32
}

func (S *resampler_state_struct) init() {
	S.sDown2 = alloc[int32](2)
	S.sIIR = alloc[int32](RESAMPLER_MAX_IIR_ORDER)
	S.sFIR = alloc[int32](RESAMPLER_MAX_IIR_ORDER)
	S.sDownPre = alloc[int32](2)
	S.sUpPost = alloc[int32](2)
}

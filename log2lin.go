package silk

import "math"

func log2lin(inLog_Q7 int32) int32 {
	var out, frac_Q7 int32

	if inLog_Q7 < 0 {
		return 0
	} else if inLog_Q7 >= (31 << 7) {
		return math.MaxInt32
	}

	out = LSHIFT(1, RSHIFT(inLog_Q7, 7))
	frac_Q7 = inLog_Q7 & 0x7F
	if inLog_Q7 < 2048 {
		out = ADD_RSHIFT(out, MUL(out, SMLAWB(frac_Q7, MUL(frac_Q7, 128-frac_Q7), -174)), 7)
	} else {
		out = MLA(out, RSHIFT(out, 7), SMLAWB(frac_Q7, MUL(frac_Q7, 128-frac_Q7), -174))
	}
	return out
}

package silk

import "math"

const (
	Q_OUT      = 6
	MIN_NDELTA = 3
)

func NLSF_VQ_weights_laroia(pNLSFW_Q6, pNLSF_Q15 *slice[int32], D int32) {
	var k, tmp1_int, tmp2_int int32

	tmp1_int = max(pNLSF_Q15.idx(0), MIN_NDELTA)
	tmp1_int = DIV32_16(1<<(15+Q_OUT), int16(tmp1_int))
	tmp2_int = max(pNLSF_Q15.idx(1)-pNLSF_Q15.idx(0), MIN_NDELTA)
	tmp2_int = DIV32_16(1<<(15+Q_OUT), int16(tmp2_int))
	*pNLSFW_Q6.ptr(0) = min(tmp1_int+tmp2_int, math.MaxInt16)

	for k = 1; k < D-1; k += 2 {
		tmp1_int = max(pNLSF_Q15.idx(int(k+1))-pNLSF_Q15.idx(int(k)), MIN_NDELTA)
		tmp1_int = DIV32_16(1<<(15+Q_OUT), int16(tmp1_int))
		*pNLSFW_Q6.ptr(int(k)) = min(tmp1_int+tmp2_int, math.MaxInt16)

		tmp2_int = max(pNLSF_Q15.idx(int(k+2))-pNLSF_Q15.idx(int(k+1)), MIN_NDELTA)
		tmp2_int = DIV32_16(1<<(15+Q_OUT), int16(tmp2_int))
		*pNLSFW_Q6.ptr(int(k + 1)) = min(tmp1_int+tmp2_int, math.MaxInt16)
	}

	tmp1_int = max((1<<15)-pNLSF_Q15.idx(int(D-1)), MIN_NDELTA)
	tmp1_int = DIV32_16(1<<(15+Q_OUT), int16(tmp1_int))
	*pNLSFW_Q6.ptr(int(D - 1)) = min(tmp1_int+tmp2_int, math.MaxInt16)
}

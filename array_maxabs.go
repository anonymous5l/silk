package silk

import "math"

func int16_array_maxabs(vec *slice[int16], len int32) int16 {
	var max, i, lvl, ind int32
	if len == 0 {
		return 0
	}

	ind = len - 1
	max = SMULBB(int32(vec.idx(int(ind))), int32(vec.idx(int(ind))))
	for i = len - 2; i >= 0; i-- {
		lvl = SMULBB(int32(vec.idx(int(i))), int32(vec.idx(int(i))))
		if lvl > max {
			max = lvl
			ind = i
		}
	}

	if max >= 1073676289 {
		return math.MaxInt16
	} else {
		if vec.idx(int(ind)) < 0 {
			return -vec.idx(int(ind))
		} else {
			return vec.idx(int(ind))
		}
	}
}

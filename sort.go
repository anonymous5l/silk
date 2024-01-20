package silk

func insertion_sort_increasing(a, index *slice[int32], L, K int32) {
	var value, i, j int32

	for i = 0; i < K; i++ {
		*index.ptr(int(i)) = i
	}

	for i = 1; i < K; i++ {
		value = a.idx(int(i))
		for j = i - 1; (j >= 0) && (value < a.idx(int(j))); j-- {
			*a.ptr(int(j + 1)) = a.idx(int(j))
			*index.ptr(int(j + 1)) = index.idx(int(j))

		}
		*a.ptr(int(j + 1)) = value
		*index.ptr(int(j + 1)) = i
	}

	for i = K; i < L; i++ {
		value = a.idx(int(i))
		if value < a.idx(int(K-1)) {
			for j = K - 2; (j >= 0) && (value < a.idx(int(j))); j-- {
				*a.ptr(int(j + 1)) = a.idx(int(j))
				*index.ptr(int(j + 1)) = index.idx(int(j))
			}
			*a.ptr(int(j + 1)) = value
			*index.ptr(int(j + 1)) = i
		}
	}
}

func insertion_sort_decreasing_int16(a *slice[int16], index *slice[int32], L, K int32) {
	var i, j, value int32

	for i = 0; i < K; i++ {
		*index.ptr(int(i)) = i
	}

	for i = 1; i < K; i++ {
		value = int32(a.idx(int(i)))
		for j = i - 1; (j >= 0) && (value > int32(a.idx(int(j)))); j-- {
			*a.ptr(int(j + 1)) = a.idx(int(j))
			*index.ptr(int(j + 1)) = index.idx(int(j))
		}
		*a.ptr(int(j + 1)) = int16(value)
		*index.ptr(int(j + 1)) = i
	}

	for i = K; i < L; i++ {
		value = int32(a.idx(int(i)))
		if value > int32(a.idx(int(K-1))) {
			for j = K - 2; (j >= 0) && (value > int32(a.idx(int(j)))); j-- {
				*a.ptr(int(j + 1)) = a.idx(int(j))
				*index.ptr(int(j + 1)) = index.idx(int(j))
			}
			*a.ptr(int(j + 1)) = int16(value)
			*index.ptr(int(j + 1)) = i
		}
	}
}

func insertion_sort_increasing_all_values(a *slice[int32], L int32) {
	var value, i, j int32

	for i = 1; i < L; i++ {
		value = a.idx(int(i))
		for j = i - 1; (j >= 0) && (value < a.idx(int(j))); j-- {
			*a.ptr(int(j + 1)) = a.idx(int(j))
		}
		*a.ptr(int(j + 1)) = value
	}
}

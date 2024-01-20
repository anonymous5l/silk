package silk

func interpolate(xi, x0, x1 *slice[int32], ifact_Q2, d int32) {
	var i int32

	for i = 0; i < d; i++ {
		*xi.ptr(int(i)) = x0.idx(int(i)) + RSHIFT(MUL(x1.idx(int(i))-x0.idx(int(i)), ifact_Q2), 2)
	}
}

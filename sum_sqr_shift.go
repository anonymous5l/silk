package silk

func sum_sqr_shift(energy, shift *int32, x *slice[int16], len int32) {
	var (
		i, shft            int32
		in32, nrg_tmp, nrg int32
	)

	x32 := slice2[int32](x)
	nrg = 0
	i = 0
	shft = 0
	len--
	for i < len {
		in32 = x32.idx(int(i / 2))
		nrg = SMLABB_ovflw(nrg, in32, in32)
		nrg = SMLATT_ovflw(nrg, in32, in32)
		i += 2
		if nrg < 0 {
			nrg = int32(RSHIFT_uint(uint32(nrg), 2))
			shft = 2
			break
		}
	}
	for ; i < len; i += 2 {
		in32 = x32.idx(int(i / 2))
		nrg_tmp = SMULBB(in32, in32)
		nrg_tmp = SMLATT_ovflw(nrg_tmp, in32, in32)
		nrg = int32(ADD_RSHIFT_uint(uint32(nrg), uint32(nrg_tmp), shft))
		if nrg < 0 {
			nrg = int32(RSHIFT_uint(uint32(nrg), 2))
			shft += 2
		}
	}
	if i == len {
		nrg_tmp = SMULBB(x32.idx(int(i/2)), x32.idx(int(i/2)))
		nrg = int32(ADD_RSHIFT_uint(uint32(nrg), uint32(nrg_tmp), shft))
	}

	if uint32(nrg)&0xC0000000 != 0 {
		nrg = int32(RSHIFT_uint(uint32(nrg), 2))
		shft += 2
	}

	*shift = shft
	*energy = nrg
}

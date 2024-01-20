package silk

func combine_pulses(out, in *slice[int32], length int32) {
	var k int32
	for k = 0; k < length; k++ {
		*out.ptr(int(k)) = in.idx(int(2*k)) + in.idx(int(2*k+1))
	}
}

func encode_split(sRC *range_coder_state, p_child1, p int32, shell_table []uint16) {
	if p > 0 {
		range_encoder(sRC, p_child1, shell_table[shell_code_table_offsets[p]:])
	}
}

func decode_split(p_child1, p_child2 *int32, sRC *range_coder_state, p int32, shell_table []uint16) {

	var (
		cdf        []uint16
		cdf_middle int32
	)

	if p > 0 {
		cdf_middle = RSHIFT(p, 1)
		cdf = shell_table[shell_code_table_offsets[p]:]
		range_decoder(p_child1, sRC, cdf, cdf_middle)
		*p_child2 = p - *p_child1
	} else {
		*p_child1 = 0
		*p_child2 = 0
	}
}

func shell_encoder(sRC *range_coder_state, pulses0 *slice[int32]) {
	var (
		pulses1 = alloc[int32](8)
		pulses2 = alloc[int32](4)
		pulses3 = alloc[int32](2)
		pulses4 = alloc[int32](1)
	)

	combine_pulses(pulses1, pulses0, 8)
	combine_pulses(pulses2, pulses1, 4)
	combine_pulses(pulses3, pulses2, 2)
	combine_pulses(pulses4, pulses3, 1)

	encode_split(sRC, pulses3.idx(0), pulses4.idx(0), shell_code_table3)

	encode_split(sRC, pulses2.idx(0), pulses3.idx(0), shell_code_table2)

	encode_split(sRC, pulses1.idx(0), pulses2.idx(0), shell_code_table1)
	encode_split(sRC, pulses0.idx(0), pulses1.idx(0), shell_code_table0)
	encode_split(sRC, pulses0.idx(2), pulses1.idx(1), shell_code_table0)

	encode_split(sRC, pulses1.idx(2), pulses2.idx(1), shell_code_table1)
	encode_split(sRC, pulses0.idx(4), pulses1.idx(2), shell_code_table0)
	encode_split(sRC, pulses0.idx(6), pulses1.idx(3), shell_code_table0)

	encode_split(sRC, pulses2.idx(2), pulses3.idx(1), shell_code_table2)

	encode_split(sRC, pulses1.idx(4), pulses2.idx(2), shell_code_table1)
	encode_split(sRC, pulses0.idx(8), pulses1.idx(4), shell_code_table0)
	encode_split(sRC, pulses0.idx(10), pulses1.idx(5), shell_code_table0)

	encode_split(sRC, pulses1.idx(6), pulses2.idx(3), shell_code_table1)
	encode_split(sRC, pulses0.idx(12), pulses1.idx(6), shell_code_table0)
	encode_split(sRC, pulses0.idx(14), pulses1.idx(7), shell_code_table0)
}

func shell_decoder(pulses0 *slice[int32], sRC *range_coder_state, pulses4 int32) {
	var (
		pulses3 = alloc[int32](2)
		pulses2 = alloc[int32](4)
		pulses1 = alloc[int32](8)
	)

	decode_split(pulses3.ptr(0), pulses3.ptr(1), sRC, pulses4, shell_code_table3)

	decode_split(pulses2.ptr(0), pulses2.ptr(1), sRC, pulses3.idx(0), shell_code_table2)

	decode_split(pulses1.ptr(0), pulses1.ptr(1), sRC, pulses2.idx(0), shell_code_table1)
	decode_split(pulses0.ptr(0), pulses0.ptr(1), sRC, pulses1.idx(0), shell_code_table0)
	decode_split(pulses0.ptr(2), pulses0.ptr(3), sRC, pulses1.idx(1), shell_code_table0)

	decode_split(pulses1.ptr(2), pulses1.ptr(3), sRC, pulses2.idx(1), shell_code_table1)
	decode_split(pulses0.ptr(4), pulses0.ptr(5), sRC, pulses1.idx(2), shell_code_table0)
	decode_split(pulses0.ptr(6), pulses0.ptr(7), sRC, pulses1.idx(3), shell_code_table0)

	decode_split(pulses2.ptr(2), pulses2.ptr(3), sRC, pulses3.idx(1), shell_code_table2)

	decode_split(pulses1.ptr(4), pulses1.ptr(5), sRC, pulses2.idx(2), shell_code_table1)
	decode_split(pulses0.ptr(8), pulses0.ptr(9), sRC, pulses1.idx(4), shell_code_table0)
	decode_split(pulses0.ptr(10), pulses0.ptr(11), sRC, pulses1.idx(5), shell_code_table0)

	decode_split(pulses1.ptr(6), pulses1.ptr(7), sRC, pulses2.idx(3), shell_code_table1)
	decode_split(pulses0.ptr(12), pulses0.ptr(13), sRC, pulses1.idx(6), shell_code_table0)
	decode_split(pulses0.ptr(14), pulses0.ptr(15), sRC, pulses1.idx(7), shell_code_table0)
}

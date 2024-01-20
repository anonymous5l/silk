package silk

import (
	"math"
)

func range_encoder(psRC *range_coder_state, data int32, prob []uint16) {
	var (
		low_Q16, high_Q16   uint32
		base_tmp, range_Q32 uint32
	)
	base_Q32 := psRC.base_Q32
	range_Q16 := psRC.range_Q16
	bufferIx := psRC.bufferIx
	buffer := psRC.buffer

	if psRC.error != 0 {
		return
	}

	low_Q16 = uint32(prob[data])
	high_Q16 = uint32(prob[data+1])
	base_tmp = base_Q32
	base_Q32 += MUL_uint(range_Q16, low_Q16)
	range_Q32 = MUL_uint(range_Q16, high_Q16-low_Q16)

	if base_Q32 < base_tmp {
		bufferIxTmp := bufferIx
		for {
			bufferIxTmp--
			*buffer.ptr(int(bufferIxTmp))++
			if buffer.idx(int(bufferIxTmp)) != 0 {
				break
			}
		}
	}

	if range_Q32&0xFF000000 != 0 {
		range_Q16 = RSHIFT_uint(range_Q32, 16)
	} else {
		if range_Q32&0xFFFF0000 != 0 {
			range_Q16 = RSHIFT_uint(range_Q32, 8)
		} else {
			range_Q16 = range_Q32

			if bufferIx >= psRC.bufferLength {
				psRC.error = RANGE_CODER_WRITE_BEYOND_BUFFER
				return
			}

			*buffer.ptr(int(bufferIx)) = byte(RSHIFT_uint(base_Q32, 24))
			bufferIx++
			base_Q32 = LSHIFT_ovflw(base_Q32, 8)
		}

		if bufferIx >= psRC.bufferLength {
			psRC.error = RANGE_CODER_WRITE_BEYOND_BUFFER
			return
		}

		*buffer.ptr(int(bufferIx)) = byte(RSHIFT_uint(base_Q32, 24))

		bufferIx++
		base_Q32 = LSHIFT_ovflw(base_Q32, 8)
	}

	psRC.base_Q32 = base_Q32
	psRC.range_Q16 = range_Q16
	psRC.bufferIx = bufferIx
}

func range_encoder_multi(psRC *range_coder_state, data *slice[int32], prob [][]uint16, nSymbols int32) {
	var k int32
	for k = 0; k < nSymbols; k++ {
		range_encoder(psRC, data.idx(int(k)), prob[k])
	}
}

func range_decoder(data *int32, psRC *range_coder_state, prob []uint16, probIx int32) {
	var (
		low_Q16, high_Q16   uint32
		base_tmp, range_Q32 uint32
	)

	base_Q32 := psRC.base_Q32
	range_Q16 := psRC.range_Q16
	bufferIx := psRC.bufferIx
	buffer := psRC.buffer.off(4)

	if psRC.error != 0 {
		*data = 0
		return
	}

	high_Q16 = uint32(prob[probIx])
	base_tmp = MUL_uint(range_Q16, high_Q16)
	if base_tmp > base_Q32 {
		for {
			probIx--
			low_Q16 = uint32(prob[probIx])
			base_tmp = MUL_uint(range_Q16, low_Q16)
			if base_tmp <= base_Q32 {
				break
			}
			high_Q16 = low_Q16

			if high_Q16 == 0 {
				psRC.error = RANGE_CODER_CDF_OUT_OF_RANGE

				*data = 0
				return
			}
		}
	} else {
		for {
			low_Q16 = high_Q16
			probIx++
			high_Q16 = uint32(prob[probIx])
			base_tmp = MUL_uint(range_Q16, high_Q16)
			if base_tmp > base_Q32 {
				probIx--
				break
			}

			if high_Q16 == 0xFFFF {
				psRC.error = RANGE_CODER_CDF_OUT_OF_RANGE
				*data = 0
				return
			}
		}
	}
	*data = probIx
	base_Q32 -= MUL_uint(range_Q16, low_Q16)
	range_Q32 = MUL_uint(range_Q16, high_Q16-low_Q16)

	if range_Q32&0xFF000000 != 0 {

		range_Q16 = RSHIFT_uint(range_Q32, 16)
	} else {
		if range_Q32&0xFFFF0000 != 0 {

			range_Q16 = RSHIFT_uint(range_Q32, 8)

			if RSHIFT_uint(base_Q32, 24) != 0 {
				psRC.error = RANGE_CODER_NORMALIZATION_FAILED

				*data = 0
				return
			}
		} else {

			range_Q16 = range_Q32

			if RSHIFT(int32(base_Q32), 16) != 0 {
				psRC.error = RANGE_CODER_NORMALIZATION_FAILED

				*data = 0
				return
			}

			base_Q32 = LSHIFT_uint(base_Q32, 8)

			if bufferIx < psRC.bufferLength {

				base_Q32 |= uint32(buffer.idx(int(bufferIx)))
				bufferIx++
			}
		}

		base_Q32 = LSHIFT_uint(base_Q32, 8)

		if bufferIx < psRC.bufferLength {

			base_Q32 |= uint32(buffer.idx(int(bufferIx)))
			bufferIx++
		}
	}

	if range_Q16 == 0 {
		psRC.error = RANGE_CODER_ZERO_INTERVAL_WIDTH

		*data = 0
		return
	}

	psRC.base_Q32 = base_Q32
	psRC.range_Q16 = range_Q16
	psRC.bufferIx = bufferIx
}

func range_decoder_multi(data *slice[int32], psRC *range_coder_state, prob [][]uint16, probStartIx []int32, nSymbols int32) {
	var k int32
	for k = 0; k < nSymbols; k++ {
		range_decoder(data.ptr(int(k)), psRC, prob[k], probStartIx[k])
	}
}

func range_enc_init(psRC *range_coder_state) {
	psRC.buffer = alloc[byte](MAX_ARITHM_BYTES)
	psRC.bufferLength = MAX_ARITHM_BYTES
	psRC.range_Q16 = 0x0000FFFF
	psRC.bufferIx = 0
	psRC.base_Q32 = 0
	psRC.error = 0
}

func range_dec_init(psRC *range_coder_state, buffer *slice[byte], bufferLength int32) {
	psRC.buffer = alloc[byte](MAX_ARITHM_BYTES)
	if (bufferLength > MAX_ARITHM_BYTES) || (bufferLength < 0) {
		psRC.error = RANGE_CODER_DEC_PAYLOAD_TOO_LONG
		return
	}

	buffer.copy(psRC.buffer, int(bufferLength))
	psRC.bufferLength = bufferLength
	psRC.bufferIx = 0
	psRC.base_Q32 = LSHIFT_uint(uint32(buffer.idx(0)), 24) |
		LSHIFT_uint(uint32(buffer.idx(1)), 16) |
		LSHIFT_uint(uint32(buffer.idx(2)), 8) |
		uint32(buffer.idx(3))
	psRC.range_Q16 = 0x0000FFFF
	psRC.error = 0
}

func range_coder_get_length(psRC *range_coder_state, nBytes *int32) int32 {
	var nBits int32
	nBits = LSHIFT(psRC.bufferIx, 3) + CLZ32(int32(psRC.range_Q16-1)) - 14

	*nBytes = RSHIFT(nBits+7, 3)
	return nBits
}

func range_enc_wrap_up(psRC *range_coder_state) {
	var (
		bits_to_store, bits_in_stream, nBytes, mask int32
		base_Q24                                    uint32
	)

	base_Q24 = RSHIFT_uint(psRC.base_Q32, 8)

	bits_in_stream = range_coder_get_length(psRC, &nBytes)

	bits_to_store = bits_in_stream - LSHIFT(psRC.bufferIx, 3)

	base_Q24 += RSHIFT_uint(0x00800000, bits_to_store-1)
	base_Q24 &= LSHIFT_ovflw(math.MaxUint32, 24-bits_to_store)

	if base_Q24&0x01000000 != 0 {
		bufferIxTmp := psRC.bufferIx
		for {
			bufferIxTmp--
			*psRC.buffer.ptr(int(bufferIxTmp))++
			if psRC.buffer.idx(int(bufferIxTmp)) != 0 {
				break
			}
		}
	}

	if psRC.bufferIx < psRC.bufferLength {
		*psRC.buffer.ptr(int(psRC.bufferIx)) = byte(RSHIFT_uint(base_Q24, 16))
		psRC.bufferIx++
		if bits_to_store > 8 {
			if psRC.bufferIx < psRC.bufferLength {
				*psRC.buffer.ptr(int(psRC.bufferIx)) = byte(RSHIFT_uint(base_Q24, 8))
				psRC.bufferIx++
			}
		}
	}

	if bits_in_stream&7 != 0 {
		mask = RSHIFT(0xFF, bits_in_stream&7)
		if nBytes-1 < psRC.bufferLength {
			*psRC.buffer.ptr(int(nBytes - 1)) |= byte(mask)
		}
	}
}

func range_coder_check_after_decoding(psRC *range_coder_state) {
	var (
		bits_in_stream, nBytes, mask int32
	)

	bits_in_stream = range_coder_get_length(psRC, &nBytes)

	if nBytes-1 >= psRC.bufferLength {
		psRC.error = RANGE_CODER_DECODER_CHECK_FAILED
		return
	}

	if bits_in_stream&7 != 0 {
		mask = RSHIFT(0xFF, bits_in_stream&7)
		if (int32(psRC.buffer.idx(int(nBytes-1))) & mask) != mask {
			psRC.error = RANGE_CODER_DECODER_CHECK_FAILED
			return
		}
	}
}

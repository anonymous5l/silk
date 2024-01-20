package silk

const RESAMPLER_MAX_BATCH_SIZE_IN = 480

func gcd(a, b int32) int32 {
	var tmp int32
	for b > 0 {
		tmp = a - b*DIV32(a, b)
		a = b
		b = tmp
	}
	return a
}

func resampler_init(S *resampler_state_struct, Fs_Hz_in, Fs_Hz_out int32) int {
	var cycleLen, cyclesPerBatch, up2, down2 int32

	if Fs_Hz_in < 8000 || Fs_Hz_in > 192000 || Fs_Hz_out < 8000 || Fs_Hz_out > 192000 {
		return -1
	}

	S.init()

	if Fs_Hz_in > 96000 {
		S.nPreDownsamplers = 2
		S.down_pre_function = resampler_private_down4
	} else if Fs_Hz_in > 48000 {
		S.nPreDownsamplers = 1
		S.down_pre_function = resampler_down2
	} else {
		S.nPreDownsamplers = 0
		S.down_pre_function = nil
	}

	if Fs_Hz_out > 96000 {
		S.nPostUpsamplers = 2
		S.up_post_function = resampler_private_up4
	} else if Fs_Hz_out > 48000 {
		S.nPostUpsamplers = 1
		S.up_post_function = resampler_up2
	} else {
		S.nPostUpsamplers = 0
		S.up_post_function = nil
	}

	if S.nPreDownsamplers+S.nPostUpsamplers > 0 {
		S.ratio_Q16 = LSHIFT32(DIV32(LSHIFT32(Fs_Hz_out, 13), Fs_Hz_in), 3)
		for SMULWW(S.ratio_Q16, Fs_Hz_in) < Fs_Hz_out {
			S.ratio_Q16++
		}

		S.batchSizePrePost = DIV32_16(Fs_Hz_in, 100)

		Fs_Hz_in = RSHIFT(Fs_Hz_in, S.nPreDownsamplers)
		Fs_Hz_out = RSHIFT(Fs_Hz_out, S.nPostUpsamplers)
	}

	S.batchSize = DIV32_16(Fs_Hz_in, 100)
	if (MUL(S.batchSize, 100) != Fs_Hz_in) || (Fs_Hz_in%100 != 0) {
		cycleLen = DIV32(Fs_Hz_in, gcd(Fs_Hz_in, Fs_Hz_out))
		cyclesPerBatch = DIV32(RESAMPLER_MAX_BATCH_SIZE_IN, cycleLen)
		if cyclesPerBatch == 0 {
			S.batchSize = RESAMPLER_MAX_BATCH_SIZE_IN
		} else {
			S.batchSize = MUL(cyclesPerBatch, cycleLen)
		}
	}

	if Fs_Hz_out > Fs_Hz_in {
		if Fs_Hz_out == MUL(Fs_Hz_in, 2) {
			S.resampler_function = resampler_private_up2_HQ_wrapper
		} else {
			S.resampler_function = resampler_private_IIR_FIR
			up2 = 1
			if Fs_Hz_in > 24000 {
				S.up2_function = resampler_up2
			} else {
				S.up2_function = resampler_private_up2_HQ
			}
		}
	} else if Fs_Hz_out < Fs_Hz_in {
		if MUL(Fs_Hz_out, 4) == MUL(Fs_Hz_in, 3) {
			S.FIR_Fracs = 3
			S.Coefs = mem2Slice[int16](Resampler_3_4_COEFS)
			S.resampler_function = resampler_private_down_FIR
		} else if MUL(Fs_Hz_out, 3) == MUL(Fs_Hz_in, 2) {
			S.FIR_Fracs = 2
			S.Coefs = mem2Slice[int16](Resampler_2_3_COEFS)
			S.resampler_function = resampler_private_down_FIR
		} else if MUL(Fs_Hz_out, 2) == Fs_Hz_in {
			S.FIR_Fracs = 1
			S.Coefs = mem2Slice[int16](Resampler_1_2_COEFS)
			S.resampler_function = resampler_private_down_FIR
		} else if MUL(Fs_Hz_out, 8) == MUL(Fs_Hz_in, 3) {
			S.FIR_Fracs = 3
			S.Coefs = mem2Slice[int16](Resampler_3_8_COEFS)
			S.resampler_function = resampler_private_down_FIR
		} else if MUL(Fs_Hz_out, 3) == Fs_Hz_in {
			S.FIR_Fracs = 1
			S.Coefs = mem2Slice[int16](Resampler_1_3_COEFS)
			S.resampler_function = resampler_private_down_FIR
		} else if MUL(Fs_Hz_out, 4) == Fs_Hz_in {
			S.FIR_Fracs = 1
			down2 = 1
			S.Coefs = mem2Slice[int16](Resampler_1_2_COEFS)
			S.resampler_function = resampler_private_down_FIR
		} else if MUL(Fs_Hz_out, 6) == Fs_Hz_in {
			S.FIR_Fracs = 1
			down2 = 1
			S.Coefs = mem2Slice[int16](Resampler_1_3_COEFS)
			S.resampler_function = resampler_private_down_FIR
		} else if MUL(Fs_Hz_out, 441) == MUL(Fs_Hz_in, 80) {
			S.Coefs = mem2Slice[int16](Resampler_80_441_ARMA4_COEFS)
			S.resampler_function = resampler_private_IIR_FIR
		} else if MUL(Fs_Hz_out, 441) == MUL(Fs_Hz_in, 120) {
			S.Coefs = mem2Slice[int16](Resampler_120_441_ARMA4_COEFS)
			S.resampler_function = resampler_private_IIR_FIR
		} else if MUL(Fs_Hz_out, 441) == MUL(Fs_Hz_in, 160) {
			S.Coefs = mem2Slice[int16](Resampler_160_441_ARMA4_COEFS)
			S.resampler_function = resampler_private_IIR_FIR
		} else if MUL(Fs_Hz_out, 441) == MUL(Fs_Hz_in, 240) {
			S.Coefs = mem2Slice[int16](Resampler_240_441_ARMA4_COEFS)
			S.resampler_function = resampler_private_IIR_FIR
		} else if MUL(Fs_Hz_out, 441) == MUL(Fs_Hz_in, 320) {
			S.Coefs = mem2Slice[int16](Resampler_320_441_ARMA4_COEFS)
			S.resampler_function = resampler_private_IIR_FIR
		} else {
			S.resampler_function = resampler_private_IIR_FIR
			up2 = 1
			if Fs_Hz_in > 24000 {
				S.up2_function = resampler_up2
			} else {
				S.up2_function = resampler_private_up2_HQ
			}
		}
	} else {
		S.resampler_function = resampler_private_copy
	}

	S.input2x = up2 | down2

	S.invRatio_Q16 = LSHIFT32(DIV32(LSHIFT32(Fs_Hz_in, 14+up2-down2), Fs_Hz_out), 2)
	for SMULWW(S.invRatio_Q16, LSHIFT32(Fs_Hz_out, down2)) < LSHIFT32(Fs_Hz_in, up2) {
		S.invRatio_Q16++
	}

	S.magic_number = 123456789

	return 0
}

func resampler(S *resampler_state_struct, out, in *slice[int16], inLen int32) int {
	if S.magic_number != 123456789 {
		return -1
	}

	if S.nPreDownsamplers+S.nPostUpsamplers > 0 {
		var (
			nSamplesIn, nSamplesOut int32
			in_buf                  = alloc[int16](480)
			out_buf                 = alloc[int16](480)
		)

		for inLen > 0 {
			nSamplesIn = min(inLen, S.batchSizePrePost)
			nSamplesOut = SMULWB(S.ratio_Q16, nSamplesIn)

			if S.nPreDownsamplers > 0 {
				S.down_pre_function(S.sDownPre, in_buf, in, nSamplesIn)
				if S.nPostUpsamplers > 0 {
					S.resampler_function(S, out_buf, in_buf, RSHIFT32(nSamplesIn, S.nPreDownsamplers))
					S.up_post_function(S.sUpPost, out, out_buf, RSHIFT32(nSamplesOut, S.nPostUpsamplers))
				} else {
					S.resampler_function(S, out, in_buf, RSHIFT32(nSamplesIn, S.nPreDownsamplers))
				}
			} else {
				S.resampler_function(S, out_buf, in, RSHIFT32(nSamplesIn, S.nPreDownsamplers))
				S.up_post_function(S.sUpPost, out, out_buf, RSHIFT32(nSamplesOut, S.nPostUpsamplers))
			}

			in = in.off(int(nSamplesIn))
			out = out.off(int(nSamplesOut))
			inLen -= nSamplesIn
		}
	} else {
		S.resampler_function(S, out, in, inLen)
	}

	return 0
}

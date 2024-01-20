package silk

func resampler_private_copy(S *resampler_state_struct, out, in *slice[int16], inLen int32) {
	in.copy(out, int(inLen))
}

package silk

func LBRR_reset(psEncC *encoder_state) {
	var i int32

	for i = 0; i < MAX_LBRR_DELAY; i++ {
		psEncC.LBRR_buffer[i].usage = NO_LBRR
	}
}

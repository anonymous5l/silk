package silk

func init_decoder(psDec *decoder_state) {
	psDec.init()

	decoder_set_fs(psDec, 24)

	psDec.first_frame_after_reset = 1
	psDec.prev_inv_gain_Q16 = 65536

	CNG_Reset(psDec)

	PLC_Reset(psDec)
}

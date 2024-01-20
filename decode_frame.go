package silk

func decode_frame(psDec *decoder_state, pOut *slice[int16], pN *int16,
	pCode *slice[byte], nBytes int32, lostFlag bool, decBytes *int32) int {

	sDecCtrl := &decoder_control{}
	sDecCtrl.init()

	var (
		ret           = 0
		L, fs_Khz_old int32
		Pulses        = alloc[int32](MAX_FRAME_LENGTH)
	)

	L = psDec.frame_length
	sDecCtrl.LTP_scale_Q14 = 0

	*decBytes = 0
	if !lostFlag {
		fs_Khz_old = psDec.fs_kHz
		if psDec.nFramesDecoded == 0 {
			range_dec_init(psDec.sRC, pCode, nBytes)
		}

		decode_parameters(psDec, sDecCtrl, Pulses, 1)

		if psDec.sRC.error != 0 {
			psDec.nBytesLeft = 0

			lostFlag = true
			decoder_set_fs(psDec, fs_Khz_old)

			*decBytes = psDec.sRC.bufferLength

			if psDec.sRC.error == RANGE_CODER_DEC_PAYLOAD_TOO_LONG {
				ret = DEC_PAYLOAD_TOO_LARGE
			} else {
				ret = DEC_PAYLOAD_ERROR
			}
		} else {
			*decBytes = psDec.sRC.bufferLength - psDec.nBytesLeft
			psDec.nFramesDecoded++

			L = psDec.frame_length

			decode_core(psDec, sDecCtrl, pOut, Pulses)

			PLC(psDec, sDecCtrl, pOut, L, lostFlag)

			psDec.lossCnt = 0
			psDec.prev_sigtype = sDecCtrl.sigtype

			psDec.first_frame_after_reset = 0
		}
	}

	if lostFlag {
		PLC(psDec, sDecCtrl, pOut, L, lostFlag)
	}

	pOut.copy(psDec.outBuf, int(L))

	PLC_glue_frames(psDec, sDecCtrl, pOut, L)

	CNG(psDec, sDecCtrl, pOut, L)

	biquad(pOut, psDec.HP_B, psDec.HP_A, psDec.HPState, pOut, L)

	*pN = int16(L)

	psDec.lagPrev = sDecCtrl.pitchL.idx(NB_SUBFR - 1)

	return ret
}

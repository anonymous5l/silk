package silk

func LP_interpolate_filter_taps(B_Q28, A_Q28 []int32, ind, fac_Q16 int32) {
	var nb, na int32

	if ind < TRANSITION_INT_NUM-1 {
		if fac_Q16 > 0 {
			if fac_Q16 == SAT16(fac_Q16) {
				for nb = 0; nb < TRANSITION_NB; nb++ {
					B_Q28[nb] = SMLAWB(
						Transition_LP_B_Q28[ind][nb],
						Transition_LP_B_Q28[ind+1][nb]-
							Transition_LP_B_Q28[ind][nb],
						fac_Q16)
				}
				for na = 0; na < TRANSITION_NA; na++ {
					A_Q28[na] = SMLAWB(
						Transition_LP_A_Q28[ind][na],
						Transition_LP_A_Q28[ind+1][na]-
							Transition_LP_A_Q28[ind][na],
						fac_Q16)
				}
			} else if fac_Q16 == (1 << 15) {

				for nb = 0; nb < TRANSITION_NB; nb++ {
					B_Q28[nb] = RSHIFT(
						Transition_LP_B_Q28[ind][nb]+
							Transition_LP_B_Q28[ind+1][nb],
						1)
				}
				for na = 0; na < TRANSITION_NA; na++ {
					A_Q28[na] = RSHIFT(
						Transition_LP_A_Q28[ind][na]+
							Transition_LP_A_Q28[ind+1][na],
						1)
				}
			} else {

				for nb = 0; nb < TRANSITION_NB; nb++ {
					B_Q28[nb] = SMLAWB(
						Transition_LP_B_Q28[ind+1][nb],
						Transition_LP_B_Q28[ind][nb]-
							Transition_LP_B_Q28[ind+1][nb],
						(1<<16)-fac_Q16)
				}
				for na = 0; na < TRANSITION_NA; na++ {
					A_Q28[na] = SMLAWB(
						Transition_LP_A_Q28[ind+1][na],
						Transition_LP_A_Q28[ind][na]-
							Transition_LP_A_Q28[ind+1][na],
						(1<<16)-fac_Q16)
				}
			}
		} else {
			memcpy(B_Q28, Transition_LP_B_Q28[ind], TRANSITION_NB)
			memcpy(A_Q28, Transition_LP_A_Q28[ind], TRANSITION_NA)
		}
	} else {
		memcpy(B_Q28, Transition_LP_B_Q28[TRANSITION_INT_NUM-1], TRANSITION_NB)
		memcpy(A_Q28, Transition_LP_A_Q28[TRANSITION_INT_NUM-1], TRANSITION_NA)
	}
}

func LP_variable_cutoff(psLP *LP_state, out, in *slice[int16], frame_length int32) {
	var (
		B_Q28   [TRANSITION_NB]int32
		A_Q28   [TRANSITION_NA]int32
		fac_Q16 int32
		ind     int32
	)

	if psLP.transition_frame_no > 0 {
		if psLP.mode == 0 {
			if psLP.transition_frame_no < TRANSITION_FRAMES_DOWN {

				fac_Q16 = LSHIFT(psLP.transition_frame_no, 16-5)

				ind = RSHIFT(fac_Q16, 16)
				fac_Q16 -= LSHIFT(ind, 16)

				LP_interpolate_filter_taps(B_Q28[:], A_Q28[:], ind, fac_Q16)

				psLP.transition_frame_no++

			} else {
				LP_interpolate_filter_taps(B_Q28[:], A_Q28[:], TRANSITION_INT_NUM-1, 0)
			}
		} else {
			if psLP.transition_frame_no < TRANSITION_FRAMES_UP {
				fac_Q16 = LSHIFT(TRANSITION_FRAMES_UP-psLP.transition_frame_no, 16-6)

				ind = RSHIFT(fac_Q16, 16)
				fac_Q16 -= LSHIFT(ind, 16)

				LP_interpolate_filter_taps(B_Q28[:], A_Q28[:], ind, fac_Q16)

				psLP.transition_frame_no++

			} else {
				LP_interpolate_filter_taps(B_Q28[:], A_Q28[:], 0, 0)
			}
		}
	}

	if psLP.transition_frame_no > 0 {
		biquad_alt(in, B_Q28[:], A_Q28[:], psLP.In_LP_State[:], out, frame_length)
	} else {
		in.copy(out, int(frame_length))
	}
}

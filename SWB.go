package silk

type detectSWB struct {
	SHP8kHz               [NBSOS][2]int32
	ConsecSmplsAboveThres int32
	ActiveSpeechMS        int32
	SWBDetected           int32
	WBDetected            int32
}

func (d *detectSWB) SWBInput(samplesIn []int16) {
	var HP8kHzLen, i, shift, energy32 int32

	nSamplesIn := int32(len(samplesIn))

	inHP8kHz := make([]int16, MaxFrameLength, MaxFrameLength)
	HP8kHzLen = min(nSamplesIn, MaxFrameLength)
	HP8kHzLen = max(HP8kHzLen, 0)

	biquad(samplesIn, SWBDetectBHPQ13[0], SWBDetectAHPQ13[0],
		d.SHP8kHz[0][:], inHP8kHz, HP8kHzLen)
	for i = 1; i < NBSOS; i++ {
		biquad(inHP8kHz, SWBDetectBHPQ13[i], SWBDetectAHPQ13[i],
			d.SHP8kHz[i][:], inHP8kHz, HP8kHzLen)
	}

	sumSqrShift(&energy32, &shift, inHP8kHz, HP8kHzLen)

	if energy32 < rshift(smulbb(HP8kHzThres, HP8kHzLen), shift) {
		d.ConsecSmplsAboveThres += nSamplesIn
		if d.ConsecSmplsAboveThres > ConcecSWBSmplsThres {
			d.SWBDetected = 1
		}
	} else {
		d.ConsecSmplsAboveThres -= nSamplesIn
		d.ConsecSmplsAboveThres = max(d.ConsecSmplsAboveThres, 0)
	}

	if d.ActiveSpeechMS > WBDetectActiveSpeechMSThres && d.SWBDetected == 0 {
		d.WBDetected = 1
	}
}

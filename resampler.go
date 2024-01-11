package silk

import "errors"

const (
	ResamplerMaxIIROrder    = 6
	ResamplerMaxFIROrder    = 16
	ResamplerMaxBatchSizeIn = 480
)

type ResampleFunc func(vector []int32, out, in []int16)

type resampler struct {
	sIIR             [ResamplerMaxIIROrder]int32
	sFIR             [ResamplerMaxFIROrder]int32
	sDown2           [2]int32
	resamplerFunc    func(out, in []int16)
	up2Func          ResampleFunc
	batchSize        int32
	invRatioQ16      int32
	FIRFracs         int32
	input2x          int32
	Coefs            []int16
	sDownPre         [2]int32
	sUpPost          [2]int32
	downPreFunc      ResampleFunc
	upPostFunc       ResampleFunc
	batchSizePrePost int32
	ratioQ16         int32
	nPreDownSamplers int32
	nPostUpSamplers  int32
	magicNumber      int32
}

func newResampler(fsHzIn, fsHzOut int32) (*resampler, error) {
	if fsHzIn < KHz8000 || fsHzIn > KHz192000 || fsHzOut < KHz8000 || fsHzOut > KHz96000 {
		return nil, errors.New("in/out kHz out of range")
	}

	r := &resampler{}
	if fsHzIn > KHz96000 {
		r.nPreDownSamplers = 2
		r.nPostUpSamplers = 2
		r.downPreFunc = resamplePrivateDown4
		r.upPostFunc = resamplePrivateUp4
	} else if fsHzIn > KHz48000 {
		r.nPreDownSamplers = 1
		r.nPostUpSamplers = 1
		r.downPreFunc = resampleDown2
		r.upPostFunc = resampleUp2
	}

	if r.nPreDownSamplers+r.nPostUpSamplers > 0 {
		r.ratioQ16 = lshift(div(lshift(fsHzOut, 13), fsHzIn), 3)
		for smulww(r.ratioQ16, fsHzIn) < fsHzOut {
			r.ratioQ16++
		}
		r.batchSizePrePost = div(fsHzIn, 100)

		fsHzIn = rshift(fsHzIn, r.nPreDownSamplers)
		fsHzOut = rshift(fsHzOut, r.nPostUpSamplers)
	}

	r.batchSize = div(fsHzIn, 100)
	if mul(r.batchSize, 100) != fsHzIn || fsHzIn%100 != 0 {
		cycleLen := div(fsHzIn, gcd(fsHzIn, fsHzOut))
		cyclesPerBatch := div(ResamplerMaxBatchSizeIn, cycleLen)
		if cyclesPerBatch == 0 {
			return nil, errors.New("resampler max batch size in")
		} else {
			r.batchSize = mul(cyclesPerBatch, cycleLen)
		}
	}

	var up2, down2 int32

	if fsHzOut > fsHzIn {
		if fsHzOut == mul(fsHzIn, 2) {
			r.resamplerFunc = r.resamplePrivateUp2HQWrapper
		} else {
			r.resamplerFunc = r.resamplePrivateIIRFIR
			up2 = 1
			if fsHzIn > KHz24000 {
				r.up2Func = resampleUp2
			} else {
				r.up2Func = resamplePrivateUp2HQ
			}
		}
	} else if fsHzOut < fsHzIn {
		if mul(fsHzOut, 4) == mul(fsHzIn, 3) {
			r.FIRFracs = 3
			r.Coefs = Resampler34Coefs
			r.resamplerFunc = r.resamplePrivateDownFIR
		} else if mul(fsHzOut, 3) == mul(fsHzIn, 2) {
			r.FIRFracs = 2
			r.Coefs = Resampler23Coefs
			r.resamplerFunc = r.resamplePrivateDownFIR
		} else if mul(fsHzOut, 2) == fsHzIn {
			r.FIRFracs = 1
			r.Coefs = Resampler12Coefs
			r.resamplerFunc = r.resamplePrivateDownFIR
		} else if mul(fsHzOut, 8) == mul(fsHzIn, 3) {
			r.FIRFracs = 3
			r.Coefs = Resampler38Coefs
			r.resamplerFunc = r.resamplePrivateDownFIR
		} else if mul(fsHzOut, 3) == fsHzIn {
			r.FIRFracs = 1
			r.Coefs = Resampler13Coefs
			r.resamplerFunc = r.resamplePrivateDownFIR
		} else if mul(fsHzOut, 4) == fsHzIn {
			r.FIRFracs = 1
			down2 = 1
			r.Coefs = Resampler12Coefs
			r.resamplerFunc = r.resamplePrivateDownFIR
		} else if mul(fsHzOut, 6) == fsHzIn {
			r.FIRFracs = 1
			down2 = 1
			r.Coefs = Resampler13Coefs
			r.resamplerFunc = r.resamplePrivateDownFIR
		} else if mul(fsHzOut, 441) == mul(fsHzIn, 80) {
			r.Coefs = Resampler80441ARMA4Coefs
			r.resamplerFunc = r.resamplePrivateIIRFIR
		} else if mul(fsHzOut, 441) == mul(fsHzIn, 120) {
			r.Coefs = Resampler120441ARMA4Coefs
			r.resamplerFunc = r.resamplePrivateIIRFIR
		} else if mul(fsHzOut, 441) == mul(fsHzIn, 160) {
			r.Coefs = Resampler160441ARMA4Coefs
			r.resamplerFunc = r.resamplePrivateIIRFIR
		} else if mul(fsHzOut, 441) == mul(fsHzIn, 240) {
			r.Coefs = Resampler240441ARMA4Coefs
			r.resamplerFunc = r.resamplePrivateIIRFIR
		} else if mul(fsHzOut, 441) == mul(fsHzIn, 320) {
			r.Coefs = Resampler320441ARMA4Coefs
			r.resamplerFunc = r.resamplePrivateIIRFIR
		} else {
			r.resamplerFunc = r.resamplePrivateIIRFIR
			up2 = 1
			if fsHzIn > KHz24000 {
				r.up2Func = resampleUp2
			} else {
				r.up2Func = resamplePrivateUp2HQ
			}
		}
	} else {
		r.resamplerFunc = resamplePrivateCopy
	}

	r.input2x = up2 | down2
	r.invRatioQ16 = lshift(div(lshift(fsHzIn, 14+up2-down2), fsHzOut), 2)

	for smulww(r.invRatioQ16, lshift(fsHzOut, down2)) < lshift(fsHzIn, up2) {
		r.invRatioQ16++
	}

	r.magicNumber = 123456789

	return r, nil
}

func (r *resampler) resample(out, in []int16) error {
	if r.magicNumber != 123456789 {
		return errors.New("not init")
	}

	if r.nPreDownSamplers+r.nPostUpSamplers > 0 {
		var nSamplesIn, nSamplesOut int32
		inBuf := make([]int16, 480, 480)
		outBuf := make([]int16, 480, 480)

		for len(in) > 0 {
			nSamplesIn = min(int32(len(in)), r.batchSizePrePost)
			nSamplesOut = smulwb(r.ratioQ16, nSamplesIn)

			assert(rshift(nSamplesIn, r.nPreDownSamplers) <= 480)
			assert(rshift(nSamplesOut, r.nPostUpSamplers) <= 480)

			if r.nPreDownSamplers > 0 {
				r.downPreFunc(r.sDownPre[:], inBuf, in[:nSamplesIn])
				if r.nPostUpSamplers > 0 {
					r.resamplerFunc(outBuf, inBuf[:rshift(nSamplesIn, r.nPreDownSamplers)])
					r.upPostFunc(r.sUpPost[:], out, outBuf[:rshift(nSamplesOut, r.nPostUpSamplers)])
				} else {
					r.resamplerFunc(out, inBuf[:rshift(nSamplesIn, r.nPreDownSamplers)])
				}
			} else {
				r.resamplerFunc(outBuf, in[:rshift(nSamplesIn, r.nPreDownSamplers)])
				r.upPostFunc(r.sUpPost[:], out, outBuf[:rshift(nSamplesOut, r.nPostUpSamplers)])
			}

			in = in[nSamplesIn:]
			out = out[nSamplesOut:]
		}
	} else {
		r.resamplerFunc(out, in)
	}

	return nil
}

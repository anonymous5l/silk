package silk

import "errors"

var (
	ErrRangeCodeDecodePayloadTooLarge = errors.New("range code decode payload too large")
	ErrRangeCoderCDFOutOfRange        = errors.New("range coder CDF out of range")
	ErrRangeCoderNormalizationFailed  = errors.New("range coder normalization failed")
	ErrRangeCoderZeroIntervalWidth    = errors.New("range coder zero interval width")
	ErrRangeCoderIllegalSamplingRate  = errors.New("range coder illegal sampling rate")
	ErrRangeCoderReadBeyondBuffer     = errors.New("range coder read beyond buffer")
	ErrRangeCoderDecoderCheckFailed   = errors.New("range coder decoder check failed")
)

var (
	ErrMagicNotMatch           = errors.New("magic not match")
	ErrUnsupportedSamplingRate = errors.New("unsupported sampling rate")
)

func assert(b bool) {
	if !b {
		panic("assert")
	}
}

package silk

import "errors"

const (
	NO_ERROR = 0

	ENC_INPUT_INVALID_NO_OF_SAMPLES = -1

	ENC_FS_NOT_SUPPORTED = -2

	ENC_PACKET_SIZE_NOT_SUPPORTED = -3

	ENC_PAYLOAD_BUF_TOO_SHORT = -4

	ENC_INVALID_LOSS_RATE = -5

	ENC_INVALID_COMPLEXITY_SETTING = -6

	ENC_INVALID_INBAND_FEC_SETTING = -7

	ENC_INVALID_DTX_SETTING = -8

	ENC_INTERNAL_ERROR = -9

	DEC_INVALID_SAMPLING_FREQUENCY = -10

	DEC_PAYLOAD_TOO_LARGE = -11

	DEC_PAYLOAD_ERROR = -12
)

var (
	ErrDecodeInvalidSamplingFrequency = errors.New("invalid sampling frequency")
	ErrEncodeFSNotSupported           = errors.New("encode fs not supported")
	ErrEncodeInputInvalidNoOfSamples  = errors.New("encode input invalid no of samples")
)

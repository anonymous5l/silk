package silk

const (
	FIND_PITCH_WHITE_NOISE_FRACTION = 1e-3

	FIND_PITCH_BANDWITH_EXPANSION = 0.99

	FIND_PITCH_CORRELATION_THRESHOLD_HC_MODE = 0.7
	FIND_PITCH_CORRELATION_THRESHOLD_MC_MODE = 0.75
	FIND_PITCH_CORRELATION_THRESHOLD_LC_MODE = 0.8

	FIND_LPC_COND_FAC = 2.5e-5
	FIND_LPC_CHIRP    = 0.99995

	FIND_LTP_COND_FAC = 1e-5
	LTP_DAMPING       = 0.01
	LTP_SMOOTHING     = 0.1

	MU_LTP_QUANT_NB  = 0.03
	MU_LTP_QUANT_MB  = 0.025
	MU_LTP_QUANT_WB  = 0.02
	MU_LTP_QUANT_SWB = 0.016

	VARIABLE_HP_SMTH_COEF1 = 0.1
	VARIABLE_HP_SMTH_COEF2 = 0.015

	VARIABLE_HP_MIN_FREQ = 80.0
	VARIABLE_HP_MAX_FREQ = 150.0

	VARIABLE_HP_MAX_DELTA_FREQ = 0.4

	WB_DETECT_ACTIVE_SPEECH_LEVEL_THRES = 0.7

	SPEECH_ACTIVITY_DTX_THRES = 0.1

	LBRR_SPEECH_ACTIVITY_THRES = 0.5

	BG_SNR_DECR_dB = 4.0

	HARM_SNR_INCR_dB = 2.0

	SPARSE_SNR_INCR_dB = 2.0

	SPARSENESS_THRESHOLD_QNT_OFFSET = 0.75

	WARPING_MULTIPLIER = 0.015

	SHAPE_WHITE_NOISE_FRACTION = 1e-5

	BANDWIDTH_EXPANSION = 0.95

	LOW_RATE_BANDWIDTH_EXPANSION_DELTA = 0.01

	DE_ESSER_COEF_SWB_dB = 2.0
	DE_ESSER_COEF_WB_dB  = 1.0

	LOW_RATE_HARMONIC_BOOST = 0.1

	LOW_INPUT_QUALITY_HARMONIC_BOOST = 0.1

	HARMONIC_SHAPING = 0.3

	HIGH_RATE_OR_LOW_QUALITY_HARMONIC_SHAPING = 0.2

	HP_NOISE_COEF = 0.3

	HARM_HP_NOISE_COEF = 0.35

	INPUT_TILT = 0.05

	HIGH_RATE_INPUT_TILT = 0.1

	LOW_FREQ_SHAPING = 3.0

	LOW_QUALITY_LOW_FREQ_SHAPING_DECR = 0.5

	NOISE_FLOOR_dB = 4.0

	RELATIVE_MIN_GAIN_dB = -50.0

	GAIN_SMOOTHING_COEF = 1e-3

	SUBFR_SMTH_COEF = 0.4

	LAMBDA_OFFSET            = 1.2
	LAMBDA_SPEECH_ACT        = -0.3
	LAMBDA_DELAYED_DECISIONS = -0.05
	LAMBDA_INPUT_QUALITY     = -0.2
	LAMBDA_CODING_QUALITY    = -0.1
	LAMBDA_QUANT_OFFSET      = 1.5
)
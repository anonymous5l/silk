package silk

const (
	MaxArithmBytes = 1024

	NBSubFR = 4

	MinLPCOrder = 10
	MaxLPCOrder = 16

	LOG2INVLPCGainHighThres = 3
	LOG2INVLPCGainLowThres  = 8

	MinDeltaGainQuant = -4

	NLSFMSVQMaxCBStages = 10

	MinQGainDB   = 6
	MaxQGainDB   = 86
	NLevelsQGain = 64

	MaxFSKHZ = 24

	FrameLengthMS  = 20
	MaxFrameLength = FrameLengthMS * MaxFSKHZ

	BWEAfterLossQ16 = 63570
	BWECOEFQ16

	RandBufSize = 128
	RandBufMask = RandBufSize - 1

	PitchDriftFACQ16 = 655

	MaxPitchLagMS = 18

	NBATT = 2

	SIGTypeVoiced   = 0
	SIGTypeUnvoiced = 1

	ShellCodecFrameLength = 16
	MaxNBShellBlocks      = MaxFrameLength / ShellCodecFrameLength

	NRateLevels = 10

	OffsetVLQ10  = 32
	OffsetVHQ10  = 100
	OffsetUVLQ10 = 100
	OffsetUVHQ10 = 256

	MaxPulses = 18
	MaxLoops  = 20

	LSFCosTabSZFix            = 128
	MaxLPCStabilizeIterations = 20
	LTPOrder                  = 5

	PitchESTMinLagMS = 2
	PitchESTNBSubFR  = 4

	DecHPOrder = 2

	VPitchGainStartMinQ14 = 11469
	VPitchGainStartMaxQ14 = 15565

	NoVoiceActivity = 0
	VoiceActivity   = 1

	CNGBufMaskMax  = 255
	CNGGainSMTHQ16 = 4634
	CNGNLSFSMTHQ16 = 16348

	MaxApiFSKHZ = 48

	MoreFrames  = 1
	LastFrame   = 0
	NoLBRRRhres = 10
	LBRRVer1    = 2
	LBRRVer2    = 3
)

var (
	MagicV3 = []byte{
		0x02, 0x23, 0x21, 0x53,
		0x49, 0x4C, 0x4B, 0x5F,
		0x56, 0x33,
	}
)

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

	MaxFSkHz         = 24
	MaxInputFrames   = 5
	MaxBytesPerFrame = 1024

	FrameLengthMS  = 20
	MaxFrameLength = FrameLengthMS * MaxFSkHz

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

	DecHPOrder = 2

	VPitchGainStartMinQ14 = 11469
	VPitchGainStartMaxQ14 = 15565

	NoVoiceActivity = 0
	VoiceActivity   = 1

	CNGBufMaskMax  = 255
	CNGGainSMTHQ16 = 4634
	CNGNLSFSMTHQ16 = 16348

	MaxApiFSKHZ = 48

	MoreFrames   = 1
	LastFrame    = 0
	NoLBRRRhres  = 10
	LBRRVer1     = 2
	LBRRVer2     = 3
	MaxLBRRDelay = 2

	LAShapeMS  = 5
	LAShapeMax = LAShapeMS * MaxFSkHz

	NBSOS                       = 3
	HP8kHzThres                 = 10
	ConcecSWBSmplsThres         = 480 * 15
	WBDetectActiveSpeechMSThres = 15000

	DecisionDelay     = 32
	DecisionDelayMask = DecisionDelay - 1

	NSQLPCBufLength = DecisionDelay

	LTPBufLength = 512
	LTPMask      = LTPBufLength - 1

	MaxShapeLPCOrder = 16

	VADNBands = 4

	VADInternalSubFramesLog2 = 2
	VADInternalSubFrames     = 1 << VADInternalSubFramesLog2

	VADNoiseLevelSmoothCOEFQ16 = 1024
	VADNoiseLevelBIAS          = 50

	VADNegativeOffsetQ5 = 128
	VADSNRFactorQ16     = 45000
	VADSNRSmoothCOEFQ18 = 4096

	AccumBitsDiffThreshold = 30000000
	TargetRateTabSZ        = 8

	NoLBRR         = 0
	AddLBRRToPlus1 = 1
	AddLBRRToPlus2 = 2

	LAPitchMS  = 2
	LAPitchMax = LAPitchMS * MaxFSkHz

	FindPitchLPCWINMS  = 20 + (LAPitchMS << 1)
	FindPitchLPCWINMax = FindPitchLPCWINMS * MaxFSkHz

	MULTPQuantNB  = 0.03
	MULTPQuantMB  = 0.025
	MULTPQuantWB  = 0.02
	MULTPQuantSWB = 0.016

	PitchESTShortLAGBIASQ15    = 6554
	PitchESTPrevLAGBIASQ15     = 6554
	PitchESTFlatContourBIASQ20 = 52429

	PitchESTMinComplex = 0
	PitchESTMidComplex = 1
	PitchESTMaxComplex = 2

	PitchESTComplexityHCMode = PitchESTMaxComplex
	PitchESTComplexityMCMode = PitchESTMidComplex
	PitchESTComplexityLCMode = PitchESTMinComplex

	FindPitchWhiteNoiseFraction         = 1e-3
	FindPitchBandWithExpansion          = 0.99
	FindPitchCorrelationThresholdHCMode = 0.7
	FindPitchCorrelationThresholdMCMode = 0.75
	FindPitchCorrelationThresholdLCMode = 0.8

	NLSFMSVQFluctuationReduction = 1
	MaxNLSFMSVQSurvivors         = 16
	MaxNLSFMSVQSurvivorsLCMode   = 2
	MaxNLSFMSVQSurvivorsMCMode   = 4

	WarpingMultiplier = 0.015

	MaxDELDECStates = 4

	MaxFindPitchLPCOrder = 16

	ShapeLPCWINMax = 15 * MaxFSkHz

	InBandFECMinRateBPS = 18000
	LBRRLossThres       = 1
)

const (
	PitchESTNBSubFR = 4

	PitchESTMaxFSKHZ = 24

	PitchESTFrameLengthMS = 40

	PitchESTMaxFrameLength    = PitchESTFrameLengthMS * PitchESTMaxFSKHZ
	PitchESTMaxFrameLengthST1 = PitchESTMaxFrameLength >> 2
	PitchESTMaxFrameLengthST2 = PitchESTMaxFrameLength >> 1

	PitchESTMaxLagMS = 18
	PitchESTMinLagMS = 2
	PitchESTMaxLag   = PitchESTMaxLagMS * PitchESTMaxFSKHZ
	PitchESTMinLag   = PitchESTMinLagMS * PitchESTMaxFSKHZ

	PitchESTDSRCHLength = 24

	PitchESTMaxDecimateStateLength = 7

	PitchESTNBStage3Lags = 5

	PitchESTNBCBKSStage2    = 3
	PitchESTNBCBKSStage2Ext = 11

	PitchESTCBmn2 = 1
	PitchESTCBmx2 = 2

	PitchESTNBCBKSStage3Max = 34
	PitchESTNBCBKSStage3Mid = 24
	PitchESTNBCBKSStage3Min = 16
)

const (
	TransitionTimeUpMS     = 5120
	TransitionTimeDownMS   = 2560
	TransitionNB           = 3
	TransitionNA           = 2
	TransitionIntNum       = 5
	TransitionFramesUp     = TransitionTimeUpMS / FrameLengthMS
	TransitionFramesDown   = TransitionTimeDownMS / FrameLengthMS
	TransitionIntStepsUp   = TransitionFramesUp / (TransitionIntNum - 1)
	TransitionIntStepsDown = TransitionFramesDown / (TransitionIntNum - 1)
)

var (
	MagicV3 = []byte{
		0x02, 0x23, 0x21, 0x53,
		0x49, 0x4C, 0x4B, 0x5F,
		0x56, 0x33,
	}
)

const (
	KHz8000   = 8000
	KHz12000  = 12000
	KHz16000  = 16000
	KHz24000  = 24000
	KHz32000  = 32000
	KHz44100  = 44100
	KHz48000  = 48000
	KHz96000  = 96000
	KHz192000 = 192000
)

const (
	SWB2WBBitrateBPS = 25000
	WB2SWBBitrateBPS = 30000
	WB2MBBitrateBPS  = 14000
	MB2WBBitrateBPS  = 18000
	MB2NBBitrateBPS  = 10000
	NB2MBBitrateBPS  = 14000
)

const (
	MinTargetRateBPS = 5000
	MaxTargetRateBPS = 100000
)

const (
	VariableHPSMTHCoef1 = 0.1
	VariableHPSMTHCoef2 = 0.015

	VariableHPMinFREQ = 80.0
	VariableHPMaxFREQ = 150.0

	VariableHPMaxDeltaFREQ = 0.4
)

const (
	BGSNRDECRdB                      = 4.0
	HarmSNRINCRdB                    = 2.0
	LBRRSpeechActivityThres          = 0.5
	SparsenessThresholdQNTOffset     = 0.75
	SparseSNRINCRdB                  = 2.0
	BandWidthExpansion               = 0.95
	LowRateBandWidthExpansionDelta   = 0.01
	ShapeWhiteNoiseFraction          = 1e-5
	NoiseFloordB                     = 4.0
	RelativeMinGaindB                = -50.0
	GainSmoothingCoef                = 1e-3
	InputTilt                        = 0.05
	HighRateInputTilt                = 0.1
	DEESSERCoefSWBdB                 = 2.0
	DEESSERCoefWBdB                  = 2.0
	LowFreqShaping                   = 3.0
	LowQualityLowFreqShapingDECR     = 0.5
	HPNoiseCoef                      = 0.3
	HarmHPNoiseCoef                  = 0.35
	LowRateHarmonicBoost             = 0.1
	LowInputQualityHarmonicBoost     = 0.1
	HighRateOrLowRateHarmonicShaping = 0.2
	HarmonicShaping                  = 0.3
	SubFRSmthCoef                    = 0.4

	HarmShapeFIRTaps = 3
	LTPCorrsHeadRoom = 2
	LTPDAMping       = 0.03
	LTPSmoothing     = 0.1
	FindLTPCondFac   = 1e-5

	LTPGainMiddleAvgRDQ14 = 11010
	NBThresholds          = 11
)

package silk

type nsq_state struct {
	xq                *slice[int16]
	sLTP_shp_Q10      *slice[int32]
	sLPC_Q14          *slice[int32]
	sAR2_Q14          *slice[int32]
	sLF_AR_shp_Q12    int32
	lagPrev           int32
	sLTP_buf_idx      int32
	sLTP_shp_buf_idx  int32
	rand_seed         int32
	prev_inv_gain_Q16 int32
	rewhite_flag      int32
}

func (n *nsq_state) init() {
	n.xq = alloc[int16](2 * MAX_FRAME_LENGTH)
	n.sLTP_shp_Q10 = alloc[int32](2 * MAX_FRAME_LENGTH)
	n.sLPC_Q14 = alloc[int32](MAX_FRAME_LENGTH/NB_SUBFR + NSQ_LPC_BUF_LENGTH)
	n.sAR2_Q14 = alloc[int32](MAX_SHAPE_LPC_ORDER)
}

func (n *nsq_state) copy(to *nsq_state) {
	to.init()
	n.xq.copy(to.xq, 2*MAX_FRAME_LENGTH)
	n.sLTP_shp_Q10.copy(to.sLTP_shp_Q10, 2*MAX_FRAME_LENGTH)
	n.sLPC_Q14.copy(to.sLPC_Q14, MAX_FRAME_LENGTH/NB_SUBFR+NSQ_LPC_BUF_LENGTH)
	n.sAR2_Q14.copy(to.sAR2_Q14, MAX_SHAPE_LPC_ORDER)
	to.sLF_AR_shp_Q12 = n.sLF_AR_shp_Q12
	to.lagPrev = n.lagPrev
	to.sLTP_buf_idx = n.sLTP_buf_idx
	to.sLTP_shp_buf_idx = n.sLTP_shp_buf_idx
	to.rand_seed = n.rand_seed
	to.prev_inv_gain_Q16 = n.prev_inv_gain_Q16
	to.rewhite_flag = n.rewhite_flag
}

type LBRR_struct struct {
	payload *slice[byte]
	nBytes  int32
	usage   int32
}

func (l *LBRR_struct) init() {
	l.payload = alloc[byte](MAX_ARITHM_BYTES)
}

type VAD_state struct {
	AnaState        *slice[int32]
	AnaState1       *slice[int32]
	AnaState2       *slice[int32]
	XnrgSubfr       *slice[int32]
	NrgRatioSmth_Q8 *slice[int32]
	HPstate         int16
	NL              *slice[int32]
	inv_NL          *slice[int32]
	NoiseLevelBias  *slice[int32]
	counter         int32
}

func (v *VAD_state) init() {
	v.AnaState = alloc[int32](2)
	v.AnaState1 = alloc[int32](2)
	v.AnaState2 = alloc[int32](2)
	v.XnrgSubfr = alloc[int32](VAD_N_BANDS)
	v.NrgRatioSmth_Q8 = alloc[int32](VAD_N_BANDS)
	v.NL = alloc[int32](VAD_N_BANDS)
	v.inv_NL = alloc[int32](VAD_N_BANDS)
	v.NoiseLevelBias = alloc[int32](VAD_N_BANDS)
}

type range_coder_state struct {
	bufferLength int32
	bufferIx     int32
	base_Q32     uint32
	range_Q16    uint32
	error        int32
	buffer       *slice[byte]
}

type detect_SWB_state struct {
	S_HP_8_kHz            [NB_SOS]*slice[int32]
	ConsecSmplsAboveThres int32
	ActiveSpeech_ms       int32
	SWB_detected          int32
	WB_detected           int32
}

func (d *detect_SWB_state) init() {
	d.S_HP_8_kHz = [NB_SOS]*slice[int32]{
		alloc[int32](2),
		alloc[int32](2),
		alloc[int32](2),
	}
}

type LP_state struct {
	In_LP_State         [2]int32
	transition_frame_no int32
	mode                int32
}

func (l *LP_state) init() {
	l.In_LP_State[0] = 0
	l.In_LP_State[1] = 0
}

type NLSF_CBS struct {
	nVectors    int32
	CB_NLSF_Q15 *slice[int16]
	Rates_Q5    *slice[int16]
}

type NLSF_CB_struct struct {
	nStages       int32
	CBStages      []NLSF_CBS
	NDeltaMin_Q15 []int32
	CDF           []uint16
	StartPtr      [][]uint16
	MiddleIx      []int32
}

type encoder_state struct {
	sRC                    *range_coder_state
	sRC_LBRR               *range_coder_state
	sNSQ                   *nsq_state
	sNSQ_LBRR              *nsq_state
	In_HP_State            [2]int32
	sLP                    *LP_state
	sVAD                   *VAD_state
	LBRRprevLastGainIndex  int32
	prev_sigtype                  int32
	typeOffsetPrev                int32
	prevLag                       int32
	prev_lagIndex                 int32
	API_fs_Hz                     int32
	prev_API_fs_Hz                int32
	maxInternal_fs_kHz            int32
	fs_kHz                        int32
	fs_kHz_changed                int32
	frame_length                  int32
	subfr_length                  int32
	la_pitch                      int32
	la_shape                      int32
	shapeWinLength                int32
	TargetRate_bps                int32
	PacketSize_ms                 int32
	PacketLoss_perc               int32
	frameCounter                  int32
	Complexity                    int32
	nStatesDelayedDecision        int32
	useInterpolatedNLSFs          int32
	shapingLPCOrder               int32
	predictLPCOrder               int32
	pitchEstimationComplexity     int32
	pitchEstimationLPCOrder       int32
	pitchEstimationThreshold_Q16  int32
	LTPQuantLowComplexity         int32
	NLSF_MSVQ_Survivors           int32
	first_frame_after_reset       int32
	controlled_since_last_payload int32
	warping_Q16            int32
	inputBuf               *slice[int16]
	inputBufIx             int32
	nFramesInPayloadBuf           int32
	nBytesInPayloadBuf            int32
	frames_since_onset     int32
	psNLSF_CB              [2]NLSF_CB_struct
	LBRR_buffer            [MAX_LBRR_DELAY]LBRR_struct
	oldest_LBRR_idx        int32
	useInBandFEC                  int32
	LBRR_enabled                  int32
	LBRR_GainIncreases            int32
	bitrateDiff                   int32
	bitrate_threshold_up          int32
	bitrate_threshold_down int32
	resampler_state        resampler_state_struct
	noSpeechCounter        int32
	useDTX                        int32
	inDTX                         int32
	vadFlag                int32
	sSWBdetect             detect_SWB_state
	q                      *slice[int8]
	q_LBRR                 *slice[int8]
}

func (en *encoder_state) init() {
	en.sVAD = &VAD_state{}
	en.sVAD.init()
	en.sRC = &range_coder_state{}
	en.sRC_LBRR = &range_coder_state{}
	en.sNSQ = &nsq_state{}
	en.sNSQ_LBRR = &nsq_state{}

	en.inputBuf = alloc[int16](MAX_FRAME_LENGTH)
	en.sSWBdetect.init()
	en.resampler_state.init()
	for i := 0; i < MAX_LBRR_DELAY; i++ {
		en.LBRR_buffer[i].init()
	}
	en.q = alloc[int8](MAX_FRAME_LENGTH)
	en.q_LBRR = alloc[int8](MAX_FRAME_LENGTH)
}

type encoder_control struct {
	lagIndex          int32
	contourIndex      int32
	PERIndex          int32
	LTPIndex          *slice[int32]
	NLSFIndices       *slice[int32]
	NLSFInterpCoef_Q2 int32
	GainsIndices      *slice[int32]
	Seed              int32
	LTP_scaleIndex    int32
	RateLevelIndex    int32
	QuantOffsetType   int32
	sigtype           int32
	pitchL            *slice[int32]
	LBRR_usage        int32
}

func (ec *encoder_control) init() {
	ec.GainsIndices = alloc[int32](NB_SUBFR)
	ec.NLSFIndices = alloc[int32](NLSF_MSVQ_MAX_CB_STAGES)
	ec.LTPIndex = alloc[int32](NB_SUBFR)
	ec.pitchL = alloc[int32](NB_SUBFR)
}

type PLC_struct struct {
	pitchL_Q8         int32
	LTPCoef_Q14       *slice[int16]
	prevLPC_Q12       *slice[int16]
	last_frame_lost   int32
	rand_seed         int32
	randScale_Q14     int16
	conc_energy       int32
	conc_energy_shift int32
	prevLTP_scale_Q14 int16
	prevGain_Q16      *slice[int32]
	fs_kHz            int32
}

func (p *PLC_struct) init() {
	p.LTPCoef_Q14 = alloc[int16](LTP_ORDER)
	p.prevLPC_Q12 = alloc[int16](MAX_LPC_ORDER)
	p.prevGain_Q16 = alloc[int32](NB_SUBFR)
}

type CNG_struct struct {
	CNG_exc_buf_Q10   *slice[int32]
	CNG_smth_NLSF_Q15 *slice[int32]
	CNG_synth_state   *slice[int32]
	CNG_smth_Gain_Q16 int32
	rand_seed         int32
	fs_kHz            int32
}

func (c *CNG_struct) init() {
	c.CNG_exc_buf_Q10 = alloc[int32](MAX_FRAME_LENGTH)
	c.CNG_smth_NLSF_Q15 = alloc[int32](MAX_LPC_ORDER)
	c.CNG_synth_state = alloc[int32](MAX_LPC_ORDER)
}

type decoder_state struct {
	sRC                       *range_coder_state
	prev_inv_gain_Q16         int32
	sLTP_Q16                  *slice[int32]
	sLPC_Q14                  *slice[int32]
	exc_Q10                   *slice[int32]
	res_Q10                   *slice[int32]
	outBuf                    *slice[int16]
	lagPrev                   int32
	LastGainIndex             int32
	LastGainIndex_EnhLayer    int32
	typeOffsetPrev            int32
	HPState                   *slice[int32]
	HP_A                      []int16
	HP_B                      []int16
	fs_kHz                    int32
	prev_API_sampleRate       int32
	frame_length              int32
	subfr_length              int32
	LPC_order                 int32
	prevNLSF_Q15              *slice[int32]
	first_frame_after_reset   int32
	nBytesLeft                int32
	nFramesDecoded            int32
	nFramesInPacket           int32
	moreInternalDecoderFrames int32
	FrameTermination          int32
	resampler_state           *resampler_state_struct
	psNLSF_CB                 [2]NLSF_CB_struct
	vadFlag                   int32
	no_FEC_counter            int32
	inband_FEC_offset         int32
	sCNG                      *CNG_struct
	lossCnt                   int32
	prev_sigtype              int32
	sPLC                      *PLC_struct
}

func (d *decoder_state) init() {
	d.sRC = &range_coder_state{}

	d.resampler_state = &resampler_state_struct{}
	d.resampler_state.init()

	d.sLTP_Q16 = alloc[int32](2 * MAX_FRAME_LENGTH)
	d.sLPC_Q14 = alloc[int32](MAX_FRAME_LENGTH/NB_SUBFR + MAX_LPC_ORDER)
	d.exc_Q10 = alloc[int32](MAX_FRAME_LENGTH)
	d.res_Q10 = alloc[int32](MAX_FRAME_LENGTH)
	d.outBuf = alloc[int16](2 * MAX_FRAME_LENGTH)
	d.HPState = alloc[int32](DEC_HP_ORDER)
	d.prevNLSF_Q15 = alloc[int32](MAX_LPC_ORDER)
}

type decoder_control struct {
	pitchL            *slice[int32]
	Gains_Q16         *slice[int32]
	Seed              int32
	PredCoef_Q12      [2]*slice[int16]
	LTPCoef_Q14       *slice[int16]
	LTP_scale_Q14     int16
	PERIndex          int32
	RateLevelIndex    int32
	QuantOffsetType   int32
	sigtype           int32
	NLSFInterpCoef_Q2 int32
}

func (d *decoder_control) init() {
	d.pitchL = alloc[int32](NB_SUBFR)
	d.Gains_Q16 = alloc[int32](NB_SUBFR)
	d.LTPCoef_Q14 = alloc[int16](LTP_ORDER * NB_SUBFR)
	for i := 0; i < 2; i++ {
		d.PredCoef_Q12[i] = alloc[int16](MAX_LPC_ORDER)
	}
}

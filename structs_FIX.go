package silk

type shape_state_FIX struct {
	LastGainIndex          int32
	HarmBoost_smth_Q16     int32
	HarmShapeGain_smth_Q16 int32
	Tilt_smth_Q16          int32
}

type prefilter_state_FIX struct {
	sLTP_shp         *slice[int16]
	sAR_shp          *slice[int32]
	sLTP_shp_buf_idx int32
	sLF_AR_shp_Q12   int32
	sLF_MA_shp_Q12   int32
	sHarmHP          int32
	rand_seed        int32
	lagPrev          int32
}

func (p *prefilter_state_FIX) init() {
	p.sLTP_shp = alloc[int16](LTP_BUF_LENGTH)
	p.sAR_shp = alloc[int32](MAX_SHAPE_LPC_ORDER + 1)
}

type predict_state_FIX struct {
	pitch_LPC_win_length int32
	min_pitch_lag        int32
	max_pitch_lag        int32
	prev_NLSFq_Q15       *slice[int32]
}

func (p *predict_state_FIX) init() {
	p.prev_NLSFq_Q15 = alloc[int32](MAX_LPC_ORDER)
}

type encoder_state_FIX struct {
	sCmn                           *encoder_state
	variable_HP_smth1_Q15          int32
	variable_HP_smth2_Q15          int32
	sShape                         *shape_state_FIX
	sPrefilt                       *prefilter_state_FIX
	sPred                          *predict_state_FIX
	x_buf                          *slice[int16]
	LTPCorr_Q15                    int32
	mu_LTP_Q8                      int32
	SNR_dB_Q7                      int32
	avgGain_Q16                    int32
	avgGain_Q16_one_bit_per_sample int32
	BufferedInChannel_ms           int32
	speech_activity_Q8             int32
	prevLTPredCodGain_Q7           int32
	HPLTPredCodGain_Q7             int32
	inBandFEC_SNR_comp_Q8          int32
}

func (e *encoder_state_FIX) init() {
	e.sCmn = &encoder_state{}
	e.sCmn.init()
	e.x_buf = alloc[int16](2*MAX_FRAME_LENGTH + LA_SHAPE_MAX)
}

type encoder_control_FIX struct {
	sCmn                    encoder_control
	Gains_Q16               *slice[int32]
	PredCoef_Q12            [2]*slice[int16]
	LTPCoef_Q14             *slice[int16]
	LTP_scale_Q14           int32
	AR1_Q13                 *slice[int16]
	AR2_Q13                 *slice[int16]
	LF_shp_Q14              *slice[int32]
	GainsPre_Q14            [NB_SUBFR]int32
	HarmBoost_Q14           [NB_SUBFR]int32
	Tilt_Q14                *slice[int32]
	HarmShapeGain_Q14       *slice[int32]
	Lambda_Q10              int32
	input_quality_Q14       int32
	coding_quality_Q14      int32
	pitch_freq_low_Hz       int32
	current_SNR_dB_Q7       int32
	sparseness_Q8           int32
	predGain_Q16            int32
	LTPredCodGain_Q7        int32
	input_quality_bands_Q15 *slice[int32]
	input_tilt_Q15          int32
	ResNrg                  *slice[int32]
	ResNrgQ                 *slice[int32]
}

func (e *encoder_control_FIX) init() {
	e.sCmn.init()
	e.input_quality_bands_Q15 = alloc[int32](VAD_N_BANDS)
	e.Gains_Q16 = alloc[int32](NB_SUBFR)
	e.AR1_Q13 = alloc[int16](NB_SUBFR * MAX_SHAPE_LPC_ORDER)
	e.AR2_Q13 = alloc[int16](NB_SUBFR * MAX_SHAPE_LPC_ORDER)
	e.LTPCoef_Q14 = alloc[int16](LTP_ORDER * NB_SUBFR)
	e.HarmShapeGain_Q14 = alloc[int32](NB_SUBFR)
	e.Tilt_Q14 = alloc[int32](NB_SUBFR)
	e.LF_shp_Q14 = alloc[int32](NB_SUBFR)
	for i := 0; i < 2; i++ {
		e.PredCoef_Q12[i] = alloc[int16](MAX_LPC_ORDER)
	}
	e.ResNrg = alloc[int32](NB_SUBFR)
	e.ResNrgQ = alloc[int32](NB_SUBFR)
}

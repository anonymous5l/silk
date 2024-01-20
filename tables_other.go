package silk

const (
	LTPscale_offset                  = 2
	vadflag_offset                   = 1
	SamplingRates_offset             = 2
	NLSF_interpolation_factor_offset = 4
	FrameTermination_offset          = 2
	Seed_offset                      = 2
)

var (
	TargetRate_table_NB = []int32{
		0, 8000, 9000, 11000, 13000, 16000, 22000, MAX_TARGET_RATE_BPS,
	}
	TargetRate_table_MB = []int32{
		0, 10000, 12000, 14000, 17000, 21000, 28000, MAX_TARGET_RATE_BPS,
	}
	TargetRate_table_WB = []int32{
		0, 11000, 14000, 17000, 21000, 26000, 36000, MAX_TARGET_RATE_BPS,
	}
	TargetRate_table_SWB = []int32{
		0, 13000, 16000, 19000, 25000, 32000, 46000, MAX_TARGET_RATE_BPS,
	}
	SNR_table_Q1 = []int32{
		19, 31, 35, 39, 43, 47, 54, 64,
	}
	SNR_table_one_bit_per_sample_Q7 = []int32{
		1984, 2240, 2408, 2708,
	}
	SWB_detect_B_HP_Q13 = [][]int16{
		{575, -948, 575}, {575, -221, 575}, {575, 104, 575},
	}
	SWB_detect_A_HP_Q13 = [][]int16{
		{14613, 6868}, {12883, 7337}, {11586, 7911},
	}
	Dec_A_HP_24                   = []int16{-16220, 8030}
	Dec_B_HP_24                   = []int16{8000, -16000, 8000}
	Dec_A_HP_16                   = []int16{-16127, 7940}
	Dec_B_HP_16                   = []int16{8000, -16000, 8000}
	Dec_A_HP_12                   = []int16{-16043, 7859}
	Dec_B_HP_12                   = []int16{8000, -16000, 8000}
	Dec_A_HP_8                    = []int16{-15885, 7710}
	Dec_B_HP_8                    = []int16{8000, -16000, 8000}
	lsb_CDF                       = []uint16{0, 40000, 65535}
	LTPscale_CDF                  = []uint16{0, 32000, 48000, 65535}
	vadflag_CDF                   = []uint16{0, 22000, 65535}
	SamplingRates_table           = []int32{8, 12, 16, 24}
	SamplingRates_CDF             = []uint16{0, 16000, 32000, 48000, 65535}
	NLSF_interpolation_factor_CDF = []uint16{0, 3706, 8703, 19226, 30926, 65535}
	FrameTermination_CDF          = []uint16{0, 20000, 45000, 56000, 65535}
	Seed_CDF                      = []uint16{0, 16384, 32768, 49152, 65535}
	Quantization_Offsets_Q10      = [][]int16{
		{OFFSET_VL_Q10, OFFSET_VH_Q10}, {OFFSET_UVL_Q10, OFFSET_UVH_Q10},
	}
	LTPScales_table_Q14 = []int16{15565, 11469, 8192}
	Transition_LP_B_Q28 = [][]int32{
		{250767114, 501534038, 250767114},
		{209867381, 419732057, 209867381},
		{170987846, 341967853, 170987846},
		{131531482, 263046905, 131531482},
		{89306658, 178584282, 89306658},
	}
	Transition_LP_A_Q28 = [][]int32{
		{506393414, 239854379},
		{411067935, 169683996},
		{306733530, 116694253},
		{185807084, 77959395},
		{35497197, 57401098},
	}
)

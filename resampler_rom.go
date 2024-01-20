package silk

const (
	resampler_down2_0  = 9872
	resampler_down2_1  = 39809 - 65536
	resampler_up2_lq_0 = 8102
	resampler_up2_lq_1 = 36783 - 65536
)

var (
	resampler_up2_hq_0     = []int16{4280, 33727 - 65536}
	resampler_up2_hq_1     = []int16{16295, 54015 - 65536}
	resampler_up2_hq_notch = []int16{7864, -3604, 13107, 28508}
	Resampler_3_4_COEFS    = []int16{
		-18249, -12532,
		-97, 284, -495, 309, 10268, 20317,
		-94, 156, -48, -720, 5984, 18278,
		-45, -4, 237, -847, 2540, 14662,
	}
	Resampler_2_3_COEFS = []int16{
		-11891, -12486,
		20, 211, -657, 688, 8423, 15911,
		-44, 197, -152, -653, 3855, 13015,
	}
	Resampler_1_2_COEFS = []int16{
		2415, -13101,
		158, -295, -400, 1265, 4832, 7968,
	}
	Resampler_3_8_COEFS = []int16{
		13270, -13738,
		-294, -123, 747, 2043, 3339, 3995,
		-151, -311, 414, 1583, 2947, 3877,
		-33, -389, 143, 1141, 2503, 3653,
	}
	Resampler_1_3_COEFS = []int16{
		16643, -14000,
		-331, 19, 581, 1421, 2290, 2845,
	}
	Resampler_2_3_COEFS_LQ = []int16{
		-2797, -6507,
		4697, 10739,
		1567, 8276,
	}
	Resampler_1_3_COEFS_LQ = []int16{
		16777, -9792,
		890, 1614, 2148,
	}
	Resampler_320_441_ARMA4_COEFS = []int16{
		31454, 24746, -9706, -3386, -17911, -13243, 24797,
	}
	Resampler_240_441_ARMA4_COEFS = []int16{
		28721, 11254, 3189, -2546, -1495, -12618, 11562,
	}
	Resampler_160_441_ARMA4_COEFS = []int16{
		23492, -6457, 14358, -4856, 14654, -13008, 4456,
	}
	Resampler_120_441_ARMA4_COEFS = []int16{
		19311, -15569, 19489, -6950, 21441, -13559, 2370,
	}
	Resampler_80_441_ARMA4_COEFS = []int16{
		13248, -23849, 24126, -9486, 26806, -14286, 1065,
	}
	resampler_frac_FIR_144 = [][]int16{
		{-647, 1884, 30078},
		{-625, 1736, 30044},
		{-603, 1591, 30005},
		{-581, 1448, 29963},
		{-559, 1308, 29917},
		{-537, 1169, 29867},
		{-515, 1032, 29813},
		{-494, 898, 29755},
		{-473, 766, 29693},
		{-452, 636, 29627},
		{-431, 508, 29558},
		{-410, 383, 29484},
		{-390, 260, 29407},
		{-369, 139, 29327},
		{-349, 20, 29242},
		{-330, -97, 29154},
		{-310, -211, 29062},
		{-291, -324, 28967},
		{-271, -434, 28868},
		{-253, -542, 28765},
		{-234, -647, 28659},
		{-215, -751, 28550},
		{-197, -852, 28436},
		{-179, -951, 28320},
		{-162, -1048, 28200},
		{-144, -1143, 28077},
		{-127, -1235, 27950},
		{-110, -1326, 27820},
		{-94, -1414, 27687},
		{-77, -1500, 27550},
		{-61, -1584, 27410},
		{-45, -1665, 27268},
		{-30, -1745, 27122},
		{-15, -1822, 26972},
		{0, -1897, 26820},
		{15, -1970, 26665},
		{29, -2041, 26507},
		{44, -2110, 26346},
		{57, -2177, 26182},
		{71, -2242, 26015},
		{84, -2305, 25845},
		{97, -2365, 25673},
		{110, -2424, 25498},
		{122, -2480, 25320},
		{134, -2534, 25140},
		{146, -2587, 24956},
		{157, -2637, 24771},
		{168, -2685, 24583},
		{179, -2732, 24392},
		{190, -2776, 24199},
		{200, -2819, 24003},
		{210, -2859, 23805},
		{220, -2898, 23605},
		{229, -2934, 23403},
		{238, -2969, 23198},
		{247, -3002, 22992},
		{255, -3033, 22783},
		{263, -3062, 22572},
		{271, -3089, 22359},
		{279, -3114, 22144},
		{286, -3138, 21927},
		{293, -3160, 21709},
		{300, -3180, 21488},
		{306, -3198, 21266},
		{312, -3215, 21042},
		{318, -3229, 20816},
		{323, -3242, 20589},
		{328, -3254, 20360},
		{333, -3263, 20130},
		{338, -3272, 19898},
		{342, -3278, 19665},
		{346, -3283, 19430},
		{350, -3286, 19194},
		{353, -3288, 18957},
		{356, -3288, 18718},
		{359, -3286, 18478},
		{362, -3283, 18238},
		{364, -3279, 17996},
		{366, -3273, 17753},
		{368, -3266, 17509},
		{369, -3257, 17264},
		{371, -3247, 17018},
		{372, -3235, 16772},
		{372, -3222, 16525},
		{373, -3208, 16277},
		{373, -3192, 16028},
		{373, -3175, 15779},
		{373, -3157, 15529},
		{372, -3138, 15279},
		{371, -3117, 15028},
		{370, -3095, 14777},
		{369, -3072, 14526},
		{368, -3048, 14274},
		{366, -3022, 14022},
		{364, -2996, 13770},
		{362, -2968, 13517},
		{359, -2940, 13265},
		{357, -2910, 13012},
		{354, -2880, 12760},
		{351, -2848, 12508},
		{348, -2815, 12255},
		{344, -2782, 12003},
		{341, -2747, 11751},
		{337, -2712, 11500},
		{333, -2676, 11248},
		{328, -2639, 10997},
		{324, -2601, 10747},
		{320, -2562, 10497},
		{315, -2523, 10247},
		{310, -2482, 9998},
		{305, -2442, 9750},
		{300, -2400, 9502},
		{294, -2358, 9255},
		{289, -2315, 9009},
		{283, -2271, 8763},
		{277, -2227, 8519},
		{271, -2182, 8275},
		{265, -2137, 8032},
		{259, -2091, 7791},
		{252, -2045, 7550},
		{246, -1998, 7311},
		{239, -1951, 7072},
		{232, -1904, 6835},
		{226, -1856, 6599},
		{219, -1807, 6364},
		{212, -1758, 6131},
		{204, -1709, 5899},
		{197, -1660, 5668},
		{190, -1611, 5439},
		{183, -1561, 5212},
		{175, -1511, 4986},
		{168, -1460, 4761},
		{160, -1410, 4538},
		{152, -1359, 4317},
		{145, -1309, 4098},
		{137, -1258, 3880},
		{129, -1207, 3664},
		{121, -1156, 3450},
		{113, -1105, 3238},
		{105, -1054, 3028},
		{97, -1003, 2820},
		{89, -952, 2614},
		{81, -901, 2409},
		{73, -851, 2207},
	}
)

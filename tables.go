package silk

const (
	type_offset_CDF_offset      = 2
	gain_CDF_offset             = 32
	delta_gain_CDF_offset       = 5
	pitch_lag_NB_CDF_offset     = 43
	pitch_lag_MB_CDF_offset     = 64
	pitch_lag_WB_CDF_offset     = 86
	pitch_lag_SWB_CDF_offset    = 128
	pitch_contour_NB_CDF_offset = 5
	pitch_contour_CDF_offset    = 17
	pulses_per_block_CDF_offset = 6
	rate_levels_CDF_offset      = 4
)

var (
	type_offset_CDF = []uint16{
		0, 37522, 41030, 44212, 65535,
	}

	type_offset_joint_CDF = [][]uint16{
		{0, 57686, 61230, 62358, 65535},
		{0, 18346, 40067, 43659, 65535},
		{0, 22694, 24279, 35507, 65535},
		{0, 6067, 7215, 13010, 65535},
	}
)

var (
	gain_CDF = [][]uint16{
		{
			0, 18, 45, 94, 181, 320, 519, 777,
			1093, 1468, 1909, 2417, 2997, 3657, 4404, 5245,
			6185, 7228, 8384, 9664, 11069, 12596, 14244, 16022,
			17937, 19979, 22121, 24345, 26646, 29021, 31454, 33927,
			36438, 38982, 41538, 44068, 46532, 48904, 51160, 53265,
			55184, 56904, 58422, 59739, 60858, 61793, 62568, 63210,
			63738, 64165, 64504, 64769, 64976, 65133, 65249, 65330,
			65386, 65424, 65451, 65471, 65487, 65501, 65513, 65524,
			65535,
		},
		{
			0, 214, 581, 1261, 2376, 3920, 5742, 7632,
			9449, 11157, 12780, 14352, 15897, 17427, 18949, 20462,
			21957, 23430, 24889, 26342, 27780, 29191, 30575, 31952,
			33345, 34763, 36200, 37642, 39083, 40519, 41930, 43291,
			44602, 45885, 47154, 48402, 49619, 50805, 51959, 53069,
			54127, 55140, 56128, 57101, 58056, 58979, 59859, 60692,
			61468, 62177, 62812, 63368, 63845, 64242, 64563, 64818,
			65023, 65184, 65306, 65391, 65447, 65482, 65505, 65521,
			65535,
		},
	}

	delta_gain_CDF = []uint16{
		0, 2358, 3856, 7023, 15376, 53058, 59135, 61555,
		62784, 63498, 63949, 64265, 64478, 64647, 64783, 64894,
		64986, 65052, 65113, 65169, 65213, 65252, 65284, 65314,
		65338, 65359, 65377, 65392, 65403, 65415, 65424, 65432,
		65440, 65448, 65455, 65462, 65470, 65477, 65484, 65491,
		65499, 65506, 65513, 65521, 65528, 65535,
	}
)

var (
	pitch_lag_NB_CDF = []uint16{
		0, 194, 395, 608, 841, 1099, 1391, 1724,
		2105, 2544, 3047, 3624, 4282, 5027, 5865, 6799,
		7833, 8965, 10193, 11510, 12910, 14379, 15905, 17473,
		19065, 20664, 22252, 23814, 25335, 26802, 28206, 29541,
		30803, 31992, 33110, 34163, 35156, 36098, 36997, 37861,
		38698, 39515, 40319, 41115, 41906, 42696, 43485, 44273,
		45061, 45847, 46630, 47406, 48175, 48933, 49679, 50411,
		51126, 51824, 52502, 53161, 53799, 54416, 55011, 55584,
		56136, 56666, 57174, 57661, 58126, 58570, 58993, 59394,
		59775, 60134, 60472, 60790, 61087, 61363, 61620, 61856,
		62075, 62275, 62458, 62625, 62778, 62918, 63045, 63162,
		63269, 63368, 63459, 63544, 63623, 63698, 63769, 63836,
		63901, 63963, 64023, 64081, 64138, 64194, 64248, 64301,
		64354, 64406, 64457, 64508, 64558, 64608, 64657, 64706,
		64754, 64803, 64851, 64899, 64946, 64994, 65041, 65088,
		65135, 65181, 65227, 65272, 65317, 65361, 65405, 65449,
		65492, 65535,
	}

	pitch_lag_MB_CDF = []uint16{
		0, 132, 266, 402, 542, 686, 838, 997,
		1167, 1349, 1546, 1760, 1993, 2248, 2528, 2835,
		3173, 3544, 3951, 4397, 4882, 5411, 5984, 6604,
		7270, 7984, 8745, 9552, 10405, 11300, 12235, 13206,
		14209, 15239, 16289, 17355, 18430, 19507, 20579, 21642,
		22688, 23712, 24710, 25677, 26610, 27507, 28366, 29188,
		29971, 30717, 31427, 32104, 32751, 33370, 33964, 34537,
		35091, 35630, 36157, 36675, 37186, 37692, 38195, 38697,
		39199, 39701, 40206, 40713, 41222, 41733, 42247, 42761,
		43277, 43793, 44309, 44824, 45336, 45845, 46351, 46851,
		47347, 47836, 48319, 48795, 49264, 49724, 50177, 50621,
		51057, 51484, 51902, 52312, 52714, 53106, 53490, 53866,
		54233, 54592, 54942, 55284, 55618, 55944, 56261, 56571,
		56873, 57167, 57453, 57731, 58001, 58263, 58516, 58762,
		58998, 59226, 59446, 59656, 59857, 60050, 60233, 60408,
		60574, 60732, 60882, 61024, 61159, 61288, 61410, 61526,
		61636, 61742, 61843, 61940, 62033, 62123, 62210, 62293,
		62374, 62452, 62528, 62602, 62674, 62744, 62812, 62879,
		62945, 63009, 63072, 63135, 63196, 63256, 63316, 63375,
		63434, 63491, 63549, 63605, 63661, 63717, 63772, 63827,
		63881, 63935, 63988, 64041, 64094, 64147, 64199, 64252,
		64304, 64356, 64409, 64461, 64513, 64565, 64617, 64669,
		64721, 64773, 64824, 64875, 64925, 64975, 65024, 65072,
		65121, 65168, 65215, 65262, 65308, 65354, 65399, 65445,
		65490, 65535,
	}

	pitch_lag_WB_CDF = []uint16{
		0, 106, 213, 321, 429, 539, 651, 766,
		884, 1005, 1132, 1264, 1403, 1549, 1705, 1870,
		2047, 2236, 2439, 2658, 2893, 3147, 3420, 3714,
		4030, 4370, 4736, 5127, 5546, 5993, 6470, 6978,
		7516, 8086, 8687, 9320, 9985, 10680, 11405, 12158,
		12938, 13744, 14572, 15420, 16286, 17166, 18057, 18955,
		19857, 20759, 21657, 22547, 23427, 24293, 25141, 25969,
		26774, 27555, 28310, 29037, 29736, 30406, 31048, 31662,
		32248, 32808, 33343, 33855, 34345, 34815, 35268, 35704,
		36127, 36537, 36938, 37330, 37715, 38095, 38471, 38844,
		39216, 39588, 39959, 40332, 40707, 41084, 41463, 41844,
		42229, 42615, 43005, 43397, 43791, 44186, 44583, 44982,
		45381, 45780, 46179, 46578, 46975, 47371, 47765, 48156,
		48545, 48930, 49312, 49690, 50064, 50433, 50798, 51158,
		51513, 51862, 52206, 52544, 52877, 53204, 53526, 53842,
		54152, 54457, 54756, 55050, 55338, 55621, 55898, 56170,
		56436, 56697, 56953, 57204, 57449, 57689, 57924, 58154,
		58378, 58598, 58812, 59022, 59226, 59426, 59620, 59810,
		59994, 60173, 60348, 60517, 60681, 60840, 60993, 61141,
		61284, 61421, 61553, 61679, 61800, 61916, 62026, 62131,
		62231, 62326, 62417, 62503, 62585, 62663, 62737, 62807,
		62874, 62938, 62999, 63057, 63113, 63166, 63217, 63266,
		63314, 63359, 63404, 63446, 63488, 63528, 63567, 63605,
		63642, 63678, 63713, 63748, 63781, 63815, 63847, 63879,
		63911, 63942, 63973, 64003, 64033, 64063, 64092, 64121,
		64150, 64179, 64207, 64235, 64263, 64291, 64319, 64347,
		64374, 64401, 64428, 64455, 64481, 64508, 64534, 64560,
		64585, 64610, 64635, 64660, 64685, 64710, 64734, 64758,
		64782, 64807, 64831, 64855, 64878, 64902, 64926, 64950,
		64974, 64998, 65022, 65045, 65069, 65093, 65116, 65139,
		65163, 65186, 65209, 65231, 65254, 65276, 65299, 65321,
		65343, 65364, 65386, 65408, 65429, 65450, 65471, 65493,
		65514, 65535,
	}

	pitch_lag_SWB_CDF = []uint16{
		0, 253, 505, 757, 1008, 1258, 1507, 1755,
		2003, 2249, 2494, 2738, 2982, 3225, 3469, 3713,
		3957, 4202, 4449, 4698, 4949, 5203, 5460, 5720,
		5983, 6251, 6522, 6798, 7077, 7361, 7650, 7942,
		8238, 8539, 8843, 9150, 9461, 9775, 10092, 10411,
		10733, 11057, 11383, 11710, 12039, 12370, 12701, 13034,
		13368, 13703, 14040, 14377, 14716, 15056, 15398, 15742,
		16087, 16435, 16785, 17137, 17492, 17850, 18212, 18577,
		18946, 19318, 19695, 20075, 20460, 20849, 21243, 21640,
		22041, 22447, 22856, 23269, 23684, 24103, 24524, 24947,
		25372, 25798, 26225, 26652, 27079, 27504, 27929, 28352,
		28773, 29191, 29606, 30018, 30427, 30831, 31231, 31627,
		32018, 32404, 32786, 33163, 33535, 33902, 34264, 34621,
		34973, 35320, 35663, 36000, 36333, 36662, 36985, 37304,
		37619, 37929, 38234, 38535, 38831, 39122, 39409, 39692,
		39970, 40244, 40513, 40778, 41039, 41295, 41548, 41796,
		42041, 42282, 42520, 42754, 42985, 43213, 43438, 43660,
		43880, 44097, 44312, 44525, 44736, 44945, 45153, 45359,
		45565, 45769, 45972, 46175, 46377, 46578, 46780, 46981,
		47182, 47383, 47585, 47787, 47989, 48192, 48395, 48599,
		48804, 49009, 49215, 49422, 49630, 49839, 50049, 50259,
		50470, 50682, 50894, 51107, 51320, 51533, 51747, 51961,
		52175, 52388, 52601, 52813, 53025, 53236, 53446, 53655,
		53863, 54069, 54274, 54477, 54679, 54879, 55078, 55274,
		55469, 55662, 55853, 56042, 56230, 56415, 56598, 56779,
		56959, 57136, 57311, 57484, 57654, 57823, 57989, 58152,
		58314, 58473, 58629, 58783, 58935, 59084, 59230, 59373,
		59514, 59652, 59787, 59919, 60048, 60174, 60297, 60417,
		60533, 60647, 60757, 60865, 60969, 61070, 61167, 61262,
		61353, 61442, 61527, 61609, 61689, 61765, 61839, 61910,
		61979, 62045, 62109, 62170, 62230, 62287, 62343, 62396,
		62448, 62498, 62547, 62594, 62640, 62685, 62728, 62770,
		62811, 62852, 62891, 62929, 62967, 63004, 63040, 63075,
		63110, 63145, 63178, 63212, 63244, 63277, 63308, 63340,
		63371, 63402, 63432, 63462, 63491, 63521, 63550, 63578,
		63607, 63635, 63663, 63690, 63718, 63744, 63771, 63798,
		63824, 63850, 63875, 63900, 63925, 63950, 63975, 63999,
		64023, 64046, 64069, 64092, 64115, 64138, 64160, 64182,
		64204, 64225, 64247, 64268, 64289, 64310, 64330, 64351,
		64371, 64391, 64411, 64431, 64450, 64470, 64489, 64508,
		64527, 64545, 64564, 64582, 64600, 64617, 64635, 64652,
		64669, 64686, 64702, 64719, 64735, 64750, 64766, 64782,
		64797, 64812, 64827, 64842, 64857, 64872, 64886, 64901,
		64915, 64930, 64944, 64959, 64974, 64988, 65003, 65018,
		65033, 65048, 65063, 65078, 65094, 65109, 65125, 65141,
		65157, 65172, 65188, 65204, 65220, 65236, 65252, 65268,
		65283, 65299, 65314, 65330, 65345, 65360, 65375, 65390,
		65405, 65419, 65434, 65449, 65463, 65477, 65492, 65506,
		65521, 65535,
	}

	pitch_contour_CDF = []uint16{
		0, 372, 843, 1315, 1836, 2644, 3576, 4719,
		6088, 7621, 9396, 11509, 14245, 17618, 20777, 24294,
		27992, 33116, 40100, 44329, 47558, 50679, 53130, 55557,
		57510, 59022, 60285, 61345, 62316, 63140, 63762, 64321,
		64729, 65099, 65535,
	}

	pitch_contour_NB_CDF = []uint16{
		0, 14445, 18587, 25628, 30013, 34859, 40597, 48426,
		54460, 59033, 62990, 65535,
	}

	pulses_per_block_CDF = [][]uint16{
		{
			0, 47113, 61501, 64590, 65125, 65277, 65352, 65407,
			65450, 65474, 65488, 65501, 65508, 65514, 65516, 65520,
			65521, 65523, 65524, 65526, 65535,
		},
		{
			0, 26368, 47760, 58803, 63085, 64567, 65113, 65333,
			65424, 65474, 65498, 65511, 65517, 65520, 65523, 65525,
			65526, 65528, 65529, 65530, 65535,
		},
		{
			0, 9601, 28014, 45877, 57210, 62560, 64611, 65260,
			65447, 65500, 65511, 65519, 65521, 65525, 65526, 65529,
			65530, 65531, 65532, 65534, 65535,
		},
		{
			0, 3351, 12462, 25972, 39782, 50686, 57644, 61525,
			63521, 64506, 65009, 65255, 65375, 65441, 65471, 65488,
			65497, 65505, 65509, 65512, 65535,
		},
		{
			0, 488, 2944, 9295, 19712, 32160, 43976, 53121,
			59144, 62518, 64213, 65016, 65346, 65470, 65511, 65515,
			65525, 65529, 65531, 65534, 65535,
		},
		{
			0, 17013, 30405, 40812, 48142, 53466, 57166, 59845,
			61650, 62873, 63684, 64223, 64575, 64811, 64959, 65051,
			65111, 65143, 65165, 65183, 65535,
		},
		{
			0, 2994, 8323, 15845, 24196, 32300, 39340, 45140,
			49813, 53474, 56349, 58518, 60167, 61397, 62313, 62969,
			63410, 63715, 63906, 64056, 65535,
		},
		{
			0, 88, 721, 2795, 7542, 14888, 24420, 34593,
			43912, 51484, 56962, 60558, 62760, 64037, 64716, 65069,
			65262, 65358, 65398, 65420, 65535,
		},
		{
			0, 287, 789, 2064, 4398, 8174, 13534, 20151,
			27347, 34533, 41295, 47242, 52070, 55772, 58458, 60381,
			61679, 62533, 63109, 63519, 65535,
		},
		{
			0, 1, 3, 91, 4521, 14708, 28329, 41955,
			52116, 58375, 61729, 63534, 64459, 64924, 65092, 65164,
			65182, 65198, 65203, 65211, 65535,
		},
	}

	pulses_per_block_BITS_Q6 = [][]int16{
		{
			30, 140, 282, 444, 560, 625, 654, 677,
			731, 780, 787, 844, 859, 960, 896, 1024,
			960, 1024, 960, 821,
		},
		{
			84, 103, 164, 252, 350, 442, 526, 607,
			663, 731, 787, 859, 923, 923, 960, 1024,
			960, 1024, 1024, 875,
		},
		{
			177, 117, 120, 162, 231, 320, 426, 541,
			657, 803, 832, 960, 896, 1024, 923, 1024,
			1024, 1024, 960, 1024,
		},
		{
			275, 182, 146, 144, 166, 207, 261, 322,
			388, 450, 516, 582, 637, 710, 762, 821,
			832, 896, 923, 734,
		},
		{
			452, 303, 216, 170, 153, 158, 182, 220,
			274, 337, 406, 489, 579, 681, 896, 811,
			896, 960, 923, 1024,
		},
		{
			125, 147, 170, 202, 232, 265, 295, 332,
			368, 406, 443, 483, 520, 563, 606, 646,
			704, 739, 757, 483,
		},
		{
			285, 232, 200, 190, 193, 206, 224, 244,
			266, 289, 315, 340, 367, 394, 425, 462,
			496, 539, 561, 350,
		},
		{
			611, 428, 319, 242, 202, 178, 172, 180,
			199, 229, 268, 313, 364, 422, 482, 538,
			603, 683, 739, 586,
		},
		{
			501, 450, 364, 308, 264, 231, 212, 204,
			204, 210, 222, 241, 265, 295, 326, 362,
			401, 437, 469, 321,
		},
	}

	rate_levels_CDF = [][]uint16{
		{
			0, 2005, 12717, 20281, 31328, 36234, 45816, 57753,
			63104, 65535,
		},
		{
			0, 8553, 23489, 36031, 46295, 53519, 56519, 59151,
			64185, 65535,
		},
	}

	rate_levels_BITS_Q6 = [][]int16{
		{
			322, 167, 199, 164, 239, 178, 157, 231,
			304,
		},
		{
			188, 137, 153, 171, 204, 285, 297, 237,
			358,
		},
	}

	max_pulses_table = []int32{
		6, 8, 12, 18,
	}

	shell_code_table0 = []uint16{
		0, 32748, 65535, 0, 9505, 56230, 65535, 0,
		4093, 32204, 61720, 65535, 0, 2285, 16207, 48750,
		63424, 65535, 0, 1709, 9446, 32026, 55752, 63876,
		65535, 0, 1623, 6986, 21845, 45381, 59147, 64186,
		65535,
	}

	shell_code_table1 = []uint16{
		0, 32691, 65535, 0, 12782, 52752, 65535, 0,
		4847, 32665, 60899, 65535, 0, 2500, 17305, 47989,
		63369, 65535, 0, 1843, 10329, 32419, 55433, 64277,
		65535, 0, 1485, 7062, 21465, 43414, 59079, 64623,
		65535, 0, 0, 4841, 14797, 31799, 49667, 61309,
		65535, 65535, 0, 0, 0, 8032, 21695, 41078,
		56317, 65535, 65535, 65535,
	}

	shell_code_table2 = []uint16{
		0, 32615, 65535, 0, 14447, 50912, 65535, 0,
		6301, 32587, 59361, 65535, 0, 3038, 18640, 46809,
		62852, 65535, 0, 1746, 10524, 32509, 55273, 64278,
		65535, 0, 1234, 6360, 21259, 43712, 59651, 64805,
		65535, 0, 1020, 4461, 14030, 32286, 51249, 61904,
		65100, 65535, 0, 851, 3435, 10006, 23241, 40797,
		55444, 63009, 65252, 65535, 0, 0, 2075, 7137,
		17119, 31499, 46982, 58723, 63976, 65535, 65535, 0,
		0, 0, 3820, 11572, 23038, 37789, 51969, 61243,
		65535, 65535, 65535, 0, 0, 0, 0, 6882,
		16828, 30444, 44844, 57365, 65535, 65535, 65535, 65535,
		0, 0, 0, 0, 0, 10093, 22963, 38779,
		54426, 65535, 65535, 65535, 65535, 65535,
	}

	shell_code_table3 = []uint16{
		0, 32324, 65535, 0, 15328, 49505, 65535, 0,
		7474, 32344, 57955, 65535, 0, 3944, 19450, 45364,
		61873, 65535, 0, 2338, 11698, 32435, 53915, 63734,
		65535, 0, 1506, 7074, 21778, 42972, 58861, 64590,
		65535, 0, 1027, 4490, 14383, 32264, 50980, 61712,
		65043, 65535, 0, 760, 3022, 9696, 23264, 41465,
		56181, 63253, 65251, 65535, 0, 579, 2256, 6873,
		16661, 31951, 48250, 59403, 64198, 65360, 65535, 0,
		464, 1783, 5181, 12269, 24247, 39877, 53490, 61502,
		64591, 65410, 65535, 0, 366, 1332, 3880, 9273,
		18585, 32014, 45928, 56659, 62616, 64899, 65483, 65535,
		0, 286, 1065, 3089, 6969, 14148, 24859, 38274,
		50715, 59078, 63448, 65091, 65481, 65535, 0, 0,
		482, 2010, 5302, 10408, 18988, 30698, 43634, 54233,
		60828, 64119, 65288, 65535, 65535, 0, 0, 0,
		1006, 3531, 7857, 14832, 24543, 36272, 47547, 56883,
		62327, 64746, 65535, 65535, 65535, 0, 0, 0,
		0, 1863, 4950, 10730, 19284, 29397, 41382, 52335,
		59755, 63834, 65535, 65535, 65535, 65535, 0, 0,
		0, 0, 0, 2513, 7290, 14487, 24275, 35312,
		46240, 55841, 62007, 65535, 65535, 65535, 65535, 65535,
		0, 0, 0, 0, 0, 0, 3606, 9573,
		18764, 28667, 40220, 51290, 59924, 65535, 65535, 65535,
		65535, 65535, 65535, 0, 0, 0, 0, 0,
		0, 0, 4879, 13091, 23376, 36061, 49395, 59315,
		65535, 65535, 65535, 65535, 65535, 65535, 65535,
	}

	shell_code_table_offsets = []uint16{
		0, 0, 3, 7, 12, 18, 25, 33,
		42, 52, 63, 75, 88, 102, 117, 133,
		150, 168, 187,
	}
)

var (
	sign_CDF = []uint16{
		37840, 36944, 36251, 35304,
		34715, 35503, 34529, 34296,
		34016, 47659, 44945, 42503,
		40235, 38569, 40254, 37851,
		37243, 36595, 43410, 44121,
		43127, 40978, 38845, 40433,
		38252, 37795, 36637, 59159,
		55630, 51806, 48073, 45036,
		48416, 43857, 42678, 41146,
	}
)

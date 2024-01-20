package silk

const (
	PITCH_EST_MAX_FS_KHZ = 24

	PITCH_EST_FRAME_LENGTH_MS = 40

	PITCH_EST_MAX_FRAME_LENGTH      = (PITCH_EST_FRAME_LENGTH_MS * PITCH_EST_MAX_FS_KHZ)
	PITCH_EST_MAX_FRAME_LENGTH_ST_1 = (PITCH_EST_MAX_FRAME_LENGTH >> 2)
	PITCH_EST_MAX_FRAME_LENGTH_ST_2 = (PITCH_EST_MAX_FRAME_LENGTH >> 1)

	PITCH_EST_MAX_LAG_MS = 18
	PITCH_EST_MIN_LAG_MS = 2
	PITCH_EST_MAX_LAG    = (PITCH_EST_MAX_LAG_MS * PITCH_EST_MAX_FS_KHZ)
	PITCH_EST_MIN_LAG    = (PITCH_EST_MIN_LAG_MS * PITCH_EST_MAX_FS_KHZ)

	PITCH_EST_NB_SUBFR = 4

	PITCH_EST_D_SRCH_LENGTH = 24

	PITCH_EST_MAX_DECIMATE_STATE_LENGTH = 7

	PITCH_EST_NB_STAGE3_LAGS = 5

	PITCH_EST_NB_CBKS_STAGE2     = 3
	PITCH_EST_NB_CBKS_STAGE2_EXT = 11

	PITCH_EST_NB_CBKS_STAGE3_MAX = 34
	PITCH_EST_NB_CBKS_STAGE3_MID = 24
	PITCH_EST_NB_CBKS_STAGE3_MIN = 16
)

var (
	CB_lags_stage2 = [][]int16{
		{0, 2, -1, -1, -1, 0, 0, 1, 1, 0, 1},
		{0, 1, 0, 0, 0, 0, 0, 1, 0, 0, 0},
		{0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 0},
		{0, -1, 2, 1, 0, 1, 1, 0, 0, -1, -1},
	}
	CB_lags_stage3 = [][]int16{
		{-9, -7, -6, -5, -5, -4, -4, -3, -3, -2, -2, -2, -1, -1, -1, 0, 0, 0, 1, 1, 0, 1, 2, 2, 2, 3, 3, 4, 4, 5, 6, 5, 6, 8},
		{-3, -2, -2, -2, -1, -1, -1, -1, -1, 0, 0, -1, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 1, 0, 1, 1, 2, 1, 2, 2, 2, 2, 3},
		{3, 3, 2, 2, 2, 2, 1, 2, 1, 1, 0, 1, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, -1, 0, 0, -1, -1, -1, -1, -1, -2, -2, -2},
		{9, 8, 6, 5, 6, 5, 4, 4, 3, 3, 2, 2, 2, 1, 0, 1, 1, 0, 0, 0, -1, -1, -1, -2, -2, -2, -3, -3, -4, -4, -5, -5, -6, -7},
	}
	Lag_range_stage3 = [][][]int16{
		{
			{-2, 6},
			{-1, 5},
			{-1, 5},
			{-2, 7},
		},
		{
			{-4, 8},
			{-1, 6},
			{-1, 6},
			{-4, 9},
		},
		{
			{-9, 12},
			{-3, 7},
			{-2, 7},
			{-7, 13},
		},
	}
	cbk_sizes_stage3 = []int16{
		PITCH_EST_NB_CBKS_STAGE3_MIN,
		PITCH_EST_NB_CBKS_STAGE3_MID,
		PITCH_EST_NB_CBKS_STAGE3_MAX,
	}
	cbk_offsets_stage3 = []int16{
		((PITCH_EST_NB_CBKS_STAGE3_MAX - PITCH_EST_NB_CBKS_STAGE3_MIN) >> 1),
		((PITCH_EST_NB_CBKS_STAGE3_MAX - PITCH_EST_NB_CBKS_STAGE3_MID) >> 1),
		0,
	}
)

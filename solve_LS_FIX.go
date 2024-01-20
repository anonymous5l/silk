package silk

type inv_D_t struct {
	Q36_part int32
	Q48_part int32
}

func solve_LDL_FIX(A *slice[int32], M int32, b, x_Q16 *slice[int32]) {
	var (
		L_Q16 = alloc[int32](MAX_MATRIX_SIZE * MAX_MATRIX_SIZE)
		Y     = alloc[int32](MAX_MATRIX_SIZE)
		inv_D = [MAX_MATRIX_SIZE]inv_D_t{}
	)

	LDL_factorize_FIX(A, M, L_Q16, inv_D[:])

	LS_SolveFirst_FIX(L_Q16, M, b, Y)

	LS_divide_Q16_FIX(Y, inv_D[:], M)

	LS_SolveLast_FIX(L_Q16, M, Y, x_Q16)
}

func LDL_factorize_FIX(A *slice[int32], M int32, L_Q16 *slice[int32], inv_D []inv_D_t) {
	var (
		i, j, k, status, loop_count                          int32
		ptr1, ptr2                                           *slice[int32]
		diag_min_value, tmp_32, err                          int32
		v_Q0                                                 = alloc[int32](MAX_MATRIX_SIZE)
		D_Q0                                                 = alloc[int32](MAX_MATRIX_SIZE)
		one_div_diag_Q36, one_div_diag_Q40, one_div_diag_Q48 int32
	)

	status = 1
	diag_min_value = max(SMMUL(ADD_SAT32(A.idx(0), A.idx(int(SMULBB(M, M)-1))), FIX_CONST(FIND_LTP_COND_FAC, 31)), 1<<9)
	for loop_count = 0; loop_count < M && status == 1; loop_count++ {
		status = 0
		for j = 0; j < M; j++ {
			ptr1 = matrix_adr(L_Q16, j, 0, M)
			tmp_32 = 0
			for i = 0; i < j; i++ {
				*v_Q0.ptr(int(i)) = SMULWW(D_Q0.idx(int(i)), ptr1.idx(int(i)))
				tmp_32 = SMLAWW(tmp_32, v_Q0.idx(int(i)), ptr1.idx(int(i)))
			}
			tmp_32 = SUB32(*matrix_ptr(A, j, j, M), tmp_32)

			if tmp_32 < diag_min_value {
				tmp_32 = SUB32(SMULBB(loop_count+1, diag_min_value), tmp_32)
				for i = 0; i < M; i++ {
					*matrix_ptr(A, i, i, M) = ADD32(*matrix_ptr(A, i, i, M), tmp_32)
				}
				status = 1
				break
			}
			*D_Q0.ptr(int(j)) = tmp_32

			one_div_diag_Q36 = INVERSE32_varQ(tmp_32, 36)
			one_div_diag_Q40 = LSHIFT(one_div_diag_Q36, 4)
			err = SUB32(1<<24, SMULWW(tmp_32, one_div_diag_Q40))
			one_div_diag_Q48 = SMULWW(err, one_div_diag_Q40)

			inv_D[j].Q36_part = one_div_diag_Q36
			inv_D[j].Q48_part = one_div_diag_Q48

			*matrix_ptr(L_Q16, j, j, M) = 65536
			ptr1 = matrix_adr(A, j, 0, M)
			ptr2 = matrix_adr(L_Q16, j+1, 0, M)
			for i = j + 1; i < M; i++ {
				tmp_32 = 0
				for k = 0; k < j; k++ {
					tmp_32 = SMLAWW(tmp_32, v_Q0.idx(int(k)), ptr2.idx(int(k)))
				}
				tmp_32 = SUB32(ptr1.idx(int(i)), tmp_32)

				*matrix_ptr(L_Q16, i, j, M) = ADD32(SMMUL(tmp_32, one_div_diag_Q48),
					RSHIFT(SMULWW(tmp_32, one_div_diag_Q36), 4))

				ptr2 = ptr2.off(int(M))
			}
		}
	}

}

func LS_divide_Q16_FIX(T *slice[int32], inv_D []inv_D_t, M int32) {
	var i, tmp_32, one_div_diag_Q36, one_div_diag_Q48 int32

	for i = 0; i < M; i++ {
		one_div_diag_Q36 = inv_D[i].Q36_part
		one_div_diag_Q48 = inv_D[i].Q48_part

		tmp_32 = T.idx(int(i))
		*T.ptr(int(i)) = ADD32(SMMUL(tmp_32, one_div_diag_Q48), RSHIFT(SMULWW(tmp_32, one_div_diag_Q36), 4))
	}
}

func LS_SolveFirst_FIX(L_Q16 *slice[int32], M int32, b *slice[int32], x_Q16 *slice[int32]) {
	var (
		i, j   int32
		ptr32  *slice[int32]
		tmp_32 int32
	)

	for i = 0; i < M; i++ {
		ptr32 = matrix_adr(L_Q16, i, 0, M)
		tmp_32 = 0
		for j = 0; j < i; j++ {
			tmp_32 = SMLAWW(tmp_32, ptr32.idx(int(j)), x_Q16.idx(int(j)))
		}
		*x_Q16.ptr(int(i)) = SUB32(b.idx(int(i)), tmp_32)
	}
}

func LS_SolveLast_FIX(L_Q16 *slice[int32], M int32, b *slice[int32], x_Q16 *slice[int32]) {
	var (
		i, j   int32
		ptr32  *slice[int32]
		tmp_32 int32
	)

	for i = M - 1; i >= 0; i-- {
		ptr32 = matrix_adr(L_Q16, 0, i, M)
		tmp_32 = 0
		for j = M - 1; j > i; j-- {
			tmp_32 = SMLAWW(tmp_32, ptr32.idx(int(SMULBB(j, M))), x_Q16.idx(int(j)))
		}
		*x_Q16.ptr(int(i)) = SUB32(b.idx(int(i)), tmp_32)
	}
}

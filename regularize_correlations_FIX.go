package silk

func regularize_correlations_FIX(XX, xx *slice[int32], noise, D int32) {
	var i int32
	for i = 0; i < D; i++ {
		*matrix_ptr(XX, i, i, D) = ADD32(*matrix_ptr(XX, i, i, D), noise)
	}
	*xx.ptr(0) += noise
}

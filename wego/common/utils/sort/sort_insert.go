package sort

func SortInsert(s SortData, desc bool) {
	N := s.Len()
	for i := 1; i < N; i++ {
		insert(s, i, desc)
	}
}

func insert(s SortData, N int, desc bool) {
	var k int
	if desc {
		for k = 0; k <= N; k++ {
			if s.Compare(N, k) {
				break
			}
		}
		if k == N {
			return
		}
		for i := k + 1; i <= N; i++ {
			s.Swap(k, i)
		}
	} else {
		for k = 0; k <= N-1; k++ {
			if !s.Compare(N, k) {
				break
			}
		}
		if k == N {
			return
		}
		for i := k + 1; i <= N; i++ {
			s.Swap(k, i)
		}
	}
}

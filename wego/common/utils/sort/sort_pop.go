package sort

func SortPop(s SortData, desc bool) {
	N := s.Len()
	if desc {
		for i := 0; i < N; i++ {
			for j := 0; j < N-i-1; j++ {
				if s.Compare(j+1, j) {
					s.Swap(j, j+1)
				}
			}
		}
	} else {
		for i := 0; i < N; i++ {
			for j := 0; j < N-i-1; j++ {
				if s.Compare(j, j+1) {
					s.Swap(j, j+1)
				}
			}
		}
	}
}

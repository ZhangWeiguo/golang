package sort

func SortEasySelect(s SortData, desc bool) {
	N := s.Len()
	var k int
	if desc {
		for i := 0; i < N; i++ {
			k = i
			for j := i + 1; j < N; j++ {
				if !s.Compare(i, j) {
					k = j
				}
			}
			if k != i {
				s.Swap(i, k)
			}
		}
	} else {
		for i := 0; i < N; i++ {
			k = i
			for j := i + 1; j < N; j++ {
				if s.Compare(i, j) {
					k = j
				}
			}
			if k != i {
				s.Swap(i, k)
			}
		}
	}
}

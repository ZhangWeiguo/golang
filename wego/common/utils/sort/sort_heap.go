package sort

func sift(s SortData, k int, m int, desc bool) {
	i := k
	j := 2*i + 1
	if desc {
		for {
			if j > m {
				return
			}
			if j < m && s.Compare(j, j+1) {
				j++
			}
			if s.Compare(j, i) {
				return
			} else {
				s.Swap(i, j)
				i = j
				j = 2*i + 1
			}
		}
	} else {
		for {
			if j > m {
				return
			}
			if j < m && s.Compare(j+1, j) {
				j++
			}
			if s.Compare(i, j) {
				return
			} else {
				s.Swap(i, j)
				i = j
				j = 2*i + 1
			}
		}
	}

}

func SortHeap(s SortData, desc bool) {
	N := s.Len()
	for i := N / 2; i >= 1; i-- {
		sift(s, i-1, N-1, desc)
	}
	for i := 1; i < N; i++ {
		s.Swap(0, N-i)
		sift(s, 0, N-i-1, desc)
	}
	return
}

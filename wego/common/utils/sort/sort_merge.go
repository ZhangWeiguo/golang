package sort

func index(s SortData, L []int, start, N int) {
	for i := start; i < start+N; i++ {
		s.Swap(i, L[i-start])
		change(L, i, L[i-start])
	}
}

func change(L []int, x int, n int) {
	for k, v := range L {
		if v == x {
			L[k] = n
		}
	}
}

func merge(s SortData, start, k, end int, desc bool) {
	n1 := k - start + 1
	n2 := end - k
	N := n1 + n2
	i, j := start, k+1
	var L []int
	L = make([]int, N)
	k0 := 0
	if desc {
		for {
			if i <= k && j <= end {
				if s.Compare(i, j) {
					L[k0] = i
					i++
				} else {
					L[k0] = j
					j++
				}
			} else {
				if i <= k && j > end {
					L[k0] = i
					i++
				}
				if i > k && j <= end {
					L[k0] = j
					j++
				}
				if i > k && j > end {
					break
				}
			}
			k0++
		}
	} else {
		for {
			if i <= k && j <= end {
				if !s.Compare(i, j) {
					L[k0] = i
					i++
				} else {
					L[k0] = j
					j++
				}
			} else {
				if i <= k && j > end {
					L[k0] = i
					i++
				}
				if i > k && j <= end {
					L[k0] = j
					j++
				}
				if i > k && j > end {
					break
				}
			}
			k0++
		}
	}
	index(s, L, start, N)
}

func sortMerge(s SortData, start, end int, desc bool) {
	N := end - start + 1
	k := start + N/2 - 1
	if N > 1 {
		sortMerge(s, start, k, desc)
		sortMerge(s, k+1, end, desc)
		merge(s, start, k, end, desc)
		return
	}
	return
}

func SortMerge(s SortData, desc bool) {
	N := s.Len()
	sortMerge(s, 0, N-1, desc)
}

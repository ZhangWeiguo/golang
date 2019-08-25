package sort

func partition(s SortData, start, end int, desc bool) int {
	m := start
	n := end
	if desc {
		for {
			if m >= n {
				break
			}
			for {
				// fmt.Println(n, m)
				if m < n && s.Compare(m, n) {
					n--
				} else {
					break
				}
			}
			if m < n {
				s.Swap(m, n)
				m++
			}
			for {
				if m < n && s.Compare(m, n) {
					m++
				} else {
					break
				}
			}
			if m < n {
				s.Swap(m, n)
				n--
			}
		}
	} else {
		for {
			if m >= n {
				break
			}
			for {
				// fmt.Println(n, m)
				if m < n && s.Compare(n, m) {
					n--
				} else {
					break
				}
			}
			if m < n {
				s.Swap(m, n)
				m++
			}
			for {
				if m < n && s.Compare(n, m) {
					m++
				} else {
					break
				}
			}
			if m < n {
				s.Swap(m, n)
				n--
			}
		}
	}
	return n
}

func sortQuick(s SortData, start, end int, desc bool) {
	N := end - start + 1
	if N <= 1 {
		return
	}
	i := partition(s, start, end, desc)
	sortQuick(s, start, i-1, desc)
	sortQuick(s, i+1, end, desc)
}

func SortQuick(s SortData, desc bool) {
	N := s.Len()
	sortQuick(s, 0, N-1, desc)
}

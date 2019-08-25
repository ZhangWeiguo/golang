package sort

type SortData interface {
	Len() int              // len(s)
	Swap(i, j int)         // s[i],s[j] = s[j],s[i]
	Compare(i, j int) bool // if s[i]>s[j] true else false
}

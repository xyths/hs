package hs

type FloatTuple struct {
	Key   interface{}
	Value float64
}
type KVSlice []FloatTuple

func (s KVSlice) Len() int {
	return len(s)
}
func (s KVSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s KVSlice) Less(i, j int) bool {
	return s[i].Value < s[j].Value
}

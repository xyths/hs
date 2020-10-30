package hs

import (
	"sort"
	"testing"
)

func TestKVSlice(t *testing.T) {
	t.Run("string tuple", func(t *testing.T) {
		tests := []struct {
			K string
			V float64
		}{
			{"a", -1},
			{"b", 9},
			{"c", 4},
		}
		var s KVSlice
		for _, tt := range tests {
			s = append(s, FloatTuple{
				Key:   tt.K,
				Value: tt.V,
			})
		}
		t.Logf("before: %v", s)
		sort.Sort(sort.Reverse(s))
		t.Logf("after: %v", s)
	})
	t.Run("object tuple", func(t *testing.T) {
		type Symbol struct {
			Name string
		}
		tests := []struct {
			K Symbol
			V float64
		}{
			{Symbol{Name: "a"}, -1},
			{Symbol{Name: "b"}, 9},
			{Symbol{Name: "c"}, 4},
		}
		var s KVSlice
		for _, tt := range tests {
			s = append(s, FloatTuple{
				Key:   tt.K,
				Value: tt.V,
			})
		}
		t.Logf("before: %v", s)
		sort.Sort(sort.Reverse(s))
		t.Logf("after: %v", s)
	})
}

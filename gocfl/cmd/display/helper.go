package display

import (
	"cmp"
	"slices"
)

func InsertSortedUniqueFunc[S interface{ ~[]E }, E any](s S, e E, cmp func(E, E) int) S {
	i, found := slices.BinarySearchFunc(s, e, cmp)
	if found {
		return s
	}
	s = append(s, e)
	copy(s[i+1:], s[i:])
	s[i] = e
	return s
}

func InsertSortedUnique[S interface{ ~[]E }, E cmp.Ordered](s S, e E) S {
	i, found := slices.BinarySearch(s, e)
	if found {
		return s
	}
	s = append(s, e)
	copy(s[i+1:], s[i:])
	s[i] = e
	return s
}

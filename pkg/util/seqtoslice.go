package util

import "iter"

func SeqToSlice[K any](i iter.Seq[K]) []K {
	if i == nil {
		return nil
	}
	result := []K{}
	for k := range i {
		result = append(result, k)
	}
	return result
}

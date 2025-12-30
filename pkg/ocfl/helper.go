package ocfl

import "iter"

func SeqToSlice[K any](i iter.Seq[K]) []K {
	result := []K{}
	for k := range i {
		result = append(result, k)
	}
	return result
}

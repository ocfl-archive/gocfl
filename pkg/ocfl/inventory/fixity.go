package inventory

import (
	"github.com/je4/utils/v2/pkg/checksum"
)

type Fixity map[checksum.DigestAlgorithm]map[string][]string

func (f Fixity) Checksums(filename string) map[checksum.DigestAlgorithm]string {
	result := map[checksum.DigestAlgorithm]string{}
	for da, dfs := range f {
		for d, fs := range dfs {
			found := false
			for _, fname := range fs {
				if fname == filename {
					result[da] = d
					found = true
					break
				}
				if found {
					break
				}
			}
			if found {
				break
			}
		}
	}
	return result
}

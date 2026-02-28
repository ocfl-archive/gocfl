package inventoryimpl

import (
	"iter"
	"maps"
	"slices"
	"strings"

	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/util"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfllogger"
)

func NewFixityBase(allowed []checksum.DigestAlgorithm, logger ocfllogger.OCFLLogger) *FixityBase {
	f := &FixityBase{
		fixity:  map[checksum.DigestAlgorithm]map[string][]string{},
		err:     nil,
		logger:  logger,
		allowed: allowed,
	}
	return f
}

type FixityBase struct {
	fixity                 map[checksum.DigestAlgorithm]map[string][]string
	err                    error
	fixityDigestAlgorithms []checksum.DigestAlgorithm
	logger                 ocfllogger.OCFLLogger
	allowed                []checksum.DigestAlgorithm
}

func (f *FixityBase) Checksums(s string) map[checksum.DigestAlgorithm]string {
	var result = map[checksum.DigestAlgorithm]string{}
	for alg, csFiles := range f.fixity {
		var found bool
		for cs, files := range csFiles {
			for _, file := range files {
				if file == s {
					result[alg] = cs
				}
				found = true
				break
			}
			if found {
				break
			}
		}
	}
	return result
}

func (f *FixityBase) WithAlgorithms(algorithms ...checksum.DigestAlgorithm) inventory.Fixity {
	for _, alg := range algorithms {
		if !slices.Contains(f.allowed, alg) {
			f.logger.Warn().Msgf("Fixity based algorithm %s is not allowed", alg)
			continue
		}
		if _, ok := f.fixity[alg]; !ok {
			f.fixity[alg] = map[string][]string{}
		}
	}
	return f
}

func (f *FixityBase) AddFile(manifestFilename string, digests map[checksum.DigestAlgorithm]string) (bool, error) {
	var modified bool
	for alg, digest := range digests {
		if _, ok := f.fixity[alg]; !ok {
			f.fixity[alg] = map[string][]string{}
		}
		if _, ok := f.fixity[alg][digest]; !ok {
			f.fixity[alg][digest] = []string{}
		}
		if !slices.Contains(f.fixity[alg][digest], manifestFilename) {
			f.fixity[alg][digest] = append(f.fixity[alg][digest], manifestFilename)
			modified = true
		}
	}
	return modified, nil
}

func (f *FixityBase) GetDigestAlgorithms() iter.Seq[checksum.DigestAlgorithm] {
	return maps.Keys(f.fixity)
}

func (f *FixityBase) Iterate(alg checksum.DigestAlgorithm) func(yield func(digest string, internal []string) bool) {
	return func(yield func(digest string, internal []string) bool) {
		dfiles, ok := f.fixity[alg]
		if !ok {
			return
		}
		for digest, files := range dfiles {
			if !yield(digest, files) {
				return
			}
		}
		return
	}
}

func (f *FixityBase) GetFiles(alg checksum.DigestAlgorithm, digest string) ([]string, error) {
	dfiles, ok := f.fixity[alg]
	if !ok {
		return nil, errors.Wrapf(inventory.DigestAlgNotFound, "digest algorithm '%s'", alg)
	}
	files, ok := dfiles[digest]
	if !ok {
		return nil, errors.Wrapf(inventory.DigestNotFound, "digest  '%s-%s'", alg, digest)
	}
	return files, nil
}

func (f *FixityBase) Err() error {
	if f == nil {
		return nil
	}
	return f.err
}

func (f *FixityBase) Equals(fixity inventory.Fixity) bool {
	fixity2, ok := fixity.(*FixityBase)
	if !ok {
		return false
	}
	if len(f.fixity) != len(fixity2.fixity) {
		return false
	}
	for digestAlg, state := range f.fixity {
		state2, ok := fixity2.fixity[digestAlg]
		if !ok {
			return false
		}
		if len(state) != len(state2) {
			return false
		}
		for digest, files := range state {
			files2, ok := state2[digest]
			if !ok {
				return false
			}
			if len(files) != len(files2) {
				return false
			}
			for i := range files {
				if files[i] != files2[i] {
					return false
				}
			}
		}
	}
	return true
}

func (f *FixityBase) CopyFrom(fixity inventory.Fixity) error {
	f.fixity = map[checksum.DigestAlgorithm]map[string][]string{}
	for digestAlg := range fixity.GetDigestAlgorithms() {
		f.fixity[digestAlg] = map[string][]string{}
		for digest, files := range fixity.Iterate(digestAlg) {
			f.fixity[digestAlg][digest] = make([]string, len(files))
			copy(f.fixity[digestAlg][digest], files)
		}
	}
	return nil
}

func (f *FixityBase) Check(fileManifest map[checksum.DigestAlgorithm]map[string][]string) error {
	for digestAlg, fixity := range f.fixity {
		// check calculated digests
		if fileManifest != nil {
			csFiles, ok := fileManifest[digestAlg]
			if !ok {
				return errors.Errorf("checksum for '%s' not created", digestAlg)
			}
			for digest, files := range fixity {
				csFilenames, ok := csFiles[digest]
				if !ok {
					csFilenames, ok = csFiles[strings.ToLower(digest)]
					if !ok {
						f.logger.ValidationError(validation.E093, "fixity digest '%s' for file(s) %v not found in content", digest, files)
						continue
					}
				}
				for _, path := range files {
					if !slices.Contains(csFilenames, path) {
						f.logger.ValidationError(validation.E093, "invalid fixity digest for file '%s'", path)
					}
				}
			}
		}
		// check consistency and format
		for digest, files := range fixity {
			digests := []string{}
			lowerDigest := strings.ToLower(digest)
			if _, found := slices.BinarySearch(digests, lowerDigest); found {
				f.logger.ValidationError(validation.E097, "fixity '%s' digest '%s' is duplicate", digestAlg, digest)
			} else {
				digests = util.SliceInsertSorted(digests, lowerDigest)
				//digests = append(digests, lowerDigest)
			}

			for _, path := range files {
				if path[0] == '/' || path[len(path)-1] == '/' {
					f.logger.ValidationError(validation.E100, "invalid path '%s' in fixity", path)
				}
				if path == "" {
					f.logger.ValidationError(validation.E099, "empty path in fixity")
				}
				path2 := path
				if path[0] == '/' {
					path2 = path[1:]
				}
				elements := strings.Split(path2, "/")
				for _, element := range elements {
					if slices.Contains([]string{"", ".", ".."}, element) {
						f.logger.ValidationError(validation.E099, "invalid path '%s' in fixity", path)
					}
				}
			}
		}

	}
	return nil
}

func (f *FixityBase) String() string {
	result := "fixity for "
	for digestAlg := range f.GetDigestAlgorithms() {
		result += digestAlg.String() + "/"
	}
	return strings.TrimSuffix(result, "/")
}

func (f *FixityBase) Finalize(inCreation bool) error {
	for alg := range f.GetDigestAlgorithms() {
		if !slices.Contains(f.fixityDigestAlgorithms, alg) {
			f.fixityDigestAlgorithms = append(f.fixityDigestAlgorithms, alg)
		}
	}
	return nil
}

var _ inventory.Fixity = (*FixityBase)(nil)

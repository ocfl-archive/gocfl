package inventory

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"emperror.dev/errors"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
)

func NewManifestBase() *ManifestBase {
	return &ManifestBase{
		manifest: map[string][]string{},
	}
}

type ManifestBase struct {
	manifest map[string][]string
	err      error
}

func (s *ManifestBase) Finalize(val validation.Validation, factory Factory, creation bool) error {
	if s.manifest == nil {
		s.manifest = map[string][]string{}
	}
	return nil
}

func (s *ManifestBase) Check(val validation.Validation, version string, manifestDigests []string, manifestDigestsLower []string) error {
	if s.Err() != nil {
		val.AddValidationError(validation.E050, "invalid state format in version '%s': %v", version, s.Err().Error())
	}
	for digest, paths := range s.IterateFiles() {
		// massive performance boost by using sorted manifest
		if _, found := slices.BinarySearch(manifestDigests, digest); !found {
			if _, found := slices.BinarySearch(manifestDigestsLower, strings.ToLower(digest)); found {
				val.AddValidationError(validation.E096, "wrong digest case in version '%s' - '%s'", version, digest)
			} else {
				val.AddValidationError(validation.E050, "digest not in manifest of versions '%s' - '%s'", version, digest)
			}
		}
		for _, path := range paths {
			if path[0] == '/' || path[len(path)-1] == '/' {
				val.AddValidationError(validation.E053, "invalid path '%s' in state for version '%s'", path, version)
			}
			if path == "" {
				val.AddValidationError(validation.E051, "empty path in state for version '%s'", version)
			}
			path2 := path
			if path[0] == '/' {
				path2 = path[1:]
			}
			elements := strings.Split(path2, "/")
			for _, element := range elements {
				if slices.Contains([]string{"", ".", ".."}, element) {
					val.AddValidationError(validation.E052, "invalid path '%s' in state for version '%s'", path, version)
				}
			}
		}
	}
	return nil
}

func (s *ManifestBase) CopyFrom(manifest Manifest) Manifest {
	s.err = manifest.Err()

	s.manifest = make(map[string][]string)
	for k, vs := range manifest.IterateFiles() {
		newVs := make([]string, len(vs))
		copy(newVs, vs)
		s.manifest[k] = newVs
	}
	return s
}

func (s *ManifestBase) Err() error {
	return s.err
}

func (s *ManifestBase) Equals(manifest Manifest) bool {
	if s == nil || manifest == nil {
		return false
	}
	manifestB, ok := manifest.(*ManifestBase)
	if !ok {
		return false
	}
	if s.err.Error() != manifest.Err().Error() {
		return false
	}
	if len(s.manifest) != len(manifestB.manifest) {
		return false
	}
	for k, v := range s.manifest {
		v2, ok := manifestB.manifest[k]
		if !ok {
			return false
		}
		if slices.Compare(v, v2) != 0 {
			return false
		}
	}
	return true
}

func (s *ManifestBase) String() string {
	var num int64
	var unique int64
	for _, v := range s.manifest {
		unique++
		num += int64(len(v))
	}
	return fmt.Sprintf("%d files (%d unique)", num, unique)
}

func (s *ManifestBase) IterateFiles() func(yield func(digest string, external []string) bool) {
	return func(yield func(digest string, external []string) bool) {
		for digest, files := range s.manifest {
			if !yield(digest, files) {
				return
			}
		}
	}
}

func (s *ManifestBase) GetFiles(digest string) ([]string, error) {
	files, ok := s.manifest[digest]
	if !ok {
		return nil, errors.Wrapf(DigestNotFound, "digest %s", digest)
	}
	return files, nil
}

func (s *ManifestBase) UnmarshalJSON(data []byte) error {
	s.manifest = map[string][]string{}
	if err := json.Unmarshal(data, &s.manifest); err != nil {
		s.err = errors.Wrapf(err, "cannot unmarshal state %s", string(data))
		return nil
	}
	return nil
}

func (s *ManifestBase) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.manifest)
}

var _ Manifest = (*ManifestBase)(nil)

package inventory

import (
	"encoding/json"
	"fmt"
	"iter"
	"slices"
	"strings"

	"emperror.dev/errors"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
)

func NewManifestBase(factory Factory) *ManifestBase {
	return &ManifestBase{
		manifest: map[string][]string{},
	}
}

type ManifestBase struct {
	manifest map[string][]string
	err      error
}

func (s *ManifestBase) GetFilesFlat() iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, internal := range s.IterateFiles() {
			for _, file := range internal {
				if !yield(file) {
					return
				}
			}
		}
	}
}

func (s *ManifestBase) AddFile(filename string, digest string) (bool, error) {
	if _, ok := s.manifest[digest]; !ok {
		s.manifest[digest] = []string{}
	}
	if slices.Contains(s.manifest[digest], filename) {
		return false, nil
	}
	s.manifest[digest] = append(s.manifest[digest], filename)
	return true, nil
}

func (s *ManifestBase) GetDuplicates(digest string) []string {
	// not necessary but fast...
	if digest == "" {
		return nil
	}
	for cs, files := range s.IterateFiles() {
		if cs == digest {
			return files
		}
	}
	return nil
}

func (s *ManifestBase) Finalize(val validation.Validation, factory Factory, creation bool) error {
	if s.manifest == nil {
		s.manifest = map[string][]string{}
	}
	return nil
}

func (s *ManifestBase) Check(val validation.Validation, csFiles map[string][]string) error {
	if s.Err() != nil {
		return errors.Wrap(s.Err(), "manifest has errors")
	}
	for digest, files := range s.IterateFiles() {
		csFilenames, ok := csFiles[strings.ToLower(digest)]
		if !ok {
			val.AddValidationError(validation.E092, "digest '%s' for file(s) %v not found in content", digest, files)
			continue
		}
		for _, file := range files {
			if !slices.Contains(csFilenames, file) {
				val.AddValidationError(validation.E092, "invalid digest for file '%s'", file)
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
	if s == nil {
		return nil
	}
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

func (s *ManifestBase) IterateFiles() func(yield func(digest string, internal []string) bool) {
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

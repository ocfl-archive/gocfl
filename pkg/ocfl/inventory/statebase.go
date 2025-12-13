package inventory

import (
	"encoding/json"
	"fmt"

	"emperror.dev/errors"
)

var DigestNotFound = errors.New("Digest not found")

type StateBase struct {
	State map[string][]string
	err   error
}

func (s *StateBase) String() string {
	var num int64
	var unique int64
	for _, v := range s.State {
		unique++
		num += int64(len(v))
	}
	return fmt.Sprintf("%d files (%d unique)", num, unique)
}

func (s *StateBase) IterateFiles() func(yield func(external []string, digest string) bool) {
	return func(yield func(external []string, digest string) bool) {
		for digest, files := range s.State {
			if !yield(files, digest) {
				return
			}
		}
	}
}

func (s *StateBase) GetFiles(digest string) ([]string, error) {
	files, ok := s.State[digest]
	if !ok {
		return nil, errors.Wrapf(DigestNotFound, "digest %s", digest)
	}
	return files, nil
}

func (s *StateBase) UnmarshalJSON(data []byte) error {
	s.State = map[string][]string{}
	if err := json.Unmarshal(data, &s.State); err != nil {
		s.err = errors.Wrapf(err, "cannot unmarshal state %s", string(data))
		return nil
	}
	return nil
}

func (s *StateBase) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.State)
}

var _ State = (*StateBase)(nil)

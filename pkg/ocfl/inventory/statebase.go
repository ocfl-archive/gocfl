package inventory

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"emperror.dev/errors"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
)

type StateBase struct {
	State map[string][]string
	err   error
}

func (s *StateBase) CopyFile(stateFilename, digest string) (bool, error) {
	var modified bool
	if _, ok := s.State[digest]; !ok {
		return false, errors.Errorf("digest %s not found", digest)
	}
	if !slices.Contains(s.State[digest], stateFilename) {
		s.State[digest] = append(s.State[digest], stateFilename)
		modified = true
	}
	return modified, nil
}

func (s *StateBase) RenameFile(oldStateFilename, newStateFilename string) (bool, error) {
	var newState = map[string][]string{}
	modified := false
	for cs, paths := range s.IterateFiles() {
		var newPaths = make([]string, 0, len(paths))
		for _, path := range paths {
			if path == oldStateFilename {
				newPaths = append(newPaths, newStateFilename)
				modified = true
			} else {
				newPaths = append(newPaths, path)
			}
		}
		if len(newPaths) > 0 {
			newState[cs] = newPaths
		}
	}
	if modified {
		s.State = newState
	}
	return modified, nil

}

func (s *StateBase) FileChecksum(path string) string {
	for d, ps := range s.IterateFiles() {
		if slices.Contains(ps, path) {
			return d
		}
	}
	return ""
}

func (s *StateBase) Check(val validation.Validation, version string, manifestDigests []string, manifestDigestsLower []string) error {
	logPaths := []string{}
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
			logPaths = append(logPaths, paths...)
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
	// check logical paths for prefixes
	slices.Sort(logPaths)
	for j := 0; j < len(logPaths)-1; j++ {
		prefix := strings.TrimSuffix(logPaths[j], "/") + "/"
		if strings.HasPrefix(logPaths[j+1], prefix) {
			val.AddValidationError(validation.E095, "logical path '%s' is prefix of '%s'", logPaths[j], logPaths[j+1])
		}
	}
	return nil
}

func (s *StateBase) CopyFrom(state State) error {
	state2, ok := state.(*StateBase)
	if !ok {
		return errors.WithStack(StateTypeDifferent)
	}
	s.err = state2.Err()

	s.State = make(map[string][]string)
	for k, vs := range state2.IterateFiles() {
		newVs := make([]string, len(vs))
		copy(newVs, vs)
		s.State[k] = newVs
	}
	return nil
}

func (s *StateBase) Err() error {
	return s.err
}

func (s *StateBase) Equals(state State) bool {
	if s == nil || state == nil {
		return false
	}
	stateB, ok := state.(*StateBase)
	if !ok {
		return false
	}
	if s.err.Error() != state.Err().Error() {
		return false
	}
	if len(s.State) != len(stateB.State) {
		return false
	}
	for k, v := range s.State {
		v2, ok := stateB.State[k]
		if !ok {
			return false
		}
		if slices.Compare(v, v2) != 0 {
			return false
		}
	}
	return true
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

func (s *StateBase) IterateFiles() func(yield func(digest string, external []string) bool) {
	return func(yield func(digest string, external []string) bool) {
		for digest, files := range s.State {
			if !yield(digest, files) {
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

func (s *StateBase) DeleteFile(stateFilename string) (bool, error) {
	var newState = map[string][]string{}
	modified := false
	for cs, paths := range s.IterateFiles() {
		var newPaths = make([]string, 0, len(paths))
		for _, path := range paths {
			if path == stateFilename {
				modified = true
				continue
			}
			newPaths = append(newPaths, path)
		}
		if len(newPaths) > 0 {
			newState[cs] = newPaths
		}
	}
	if modified {
		s.State = newState
	}
	return modified, nil
}

var _ State = (*StateBase)(nil)

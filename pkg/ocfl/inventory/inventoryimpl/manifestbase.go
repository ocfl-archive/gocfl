package inventoryimpl

import (
	"encoding/json"
	"fmt"
	"iter"
	"slices"
	"strings"

	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/util"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
)

func NewManifestBase(logger zLogger.ZLogger) *ManifestBase {
	return &ManifestBase{
		manifest: map[string][]string{},
		logger:   logger,
	}
}

type ManifestBase struct {
	manifest map[string][]string
	err      error
	logger   zLogger.ZLogger
}

func (manifest *ManifestBase) GetFilesFlat() iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, internal := range manifest.Iterate() {
			for _, file := range internal {
				if !yield(file) {
					return
				}
			}
		}
	}
}

func (manifest *ManifestBase) AddFile(filename string, digest string) (bool, error) {
	if _, ok := manifest.manifest[digest]; !ok {
		manifest.manifest[digest] = []string{}
	}
	if slices.Contains(manifest.manifest[digest], filename) {
		return false, nil
	}
	manifest.manifest[digest] = append(manifest.manifest[digest], filename)
	return true, nil
}

func (manifest *ManifestBase) GetDuplicates(digest string) []string {
	// not necessary but fast...
	if digest == "" {
		return nil
	}
	for cs, files := range manifest.Iterate() {
		if cs == digest {
			return files
		}
	}
	return nil
}

func (manifest *ManifestBase) Finalize(val validation.Validation, factory inventory.Factory, creation bool) error {
	if manifest.manifest == nil {
		manifest.manifest = map[string][]string{}
	}
	return nil
}

func (manifest *ManifestBase) Check(val validation.Validation, csFiles map[string][]string, versionDigests []string) error {
	if manifest.Err() != nil {
		return errors.Wrap(manifest.Err(), "manifest has errors")
	}
	slices.Sort(versionDigests)
	versionDigests = slices.Compact(versionDigests)
	if csFiles != nil {
		for digest, files := range manifest.Iterate() {
			csFilenames, ok := csFiles[strings.ToLower(digest)]
			if !ok {
				val.AddValidationError(validation.E092, "digest '%manifest' for file(manifest) %v not found in content", digest, files)
				continue
			}
			for _, file := range files {
				if !slices.Contains(csFilenames, file) {
					val.AddValidationError(validation.E092, "invalid digest for file '%manifest'", file)
				}
			}
		}
	}
	digests := []string{}
	allPaths := []string{}
	for digest, paths := range manifest.Iterate() {
		//		digest = strings.ToLower(digest)
		if slices.Contains(digests, digest) {
			val.AddValidationError(validation.E096, "manifest digest '%manifest' is duplicate", digest)
		} else {
			digests = util.SliceInsertSorted(digests, digest)
			//digests = append(digests, digest)
			if _, found := slices.BinarySearch(versionDigests, digest); !found {
				//if !slices.Contains(versionDigests, digest) {
				val.AddValidationError(validation.E107, "digest '%manifest' does not appear in any version", digest)
			}
		}
		for _, path := range paths {
			//allPaths = sliceInsertSorted(allPaths, path)
			allPaths = append(allPaths, path)
			if path[0] == '/' || path[len(path)-1] == '/' {
				val.AddValidationError(validation.E100, "invalid path '%manifest' in manifest", path)
			}
			if path == "" {
				val.AddValidationError(validation.E099, "empty path in manifest")
			}
			path2 := path
			if path[0] == '/' {
				path2 = path[1:]
			}
			elements := strings.Split(path2, "/")
			for _, element := range elements {
				if slices.Contains([]string{"", ".", ".."}, element) {
					val.AddValidationError(validation.E099, "invalid path '%manifest' in manifest", path)
				}
			}

		}

	}
	slices.Sort(allPaths)
	for j := 0; j < len(allPaths)-1; j++ {
		prefix := strings.TrimRight(allPaths[j+1], "/") + "/"
		if strings.HasPrefix(allPaths[j], prefix) {
			val.AddValidationError(validation.E101, "content path '%s' is prefix or equal to '%s' in manifest", allPaths[j], prefix)
		}
	}

	return nil
}

func (manifest *ManifestBase) CopyFrom(manifest2 inventory.Manifest) inventory.Manifest {
	manifest.err = manifest2.Err()

	manifest.manifest = make(map[string][]string)
	for k, vs := range manifest2.Iterate() {
		newVs := make([]string, len(vs))
		copy(newVs, vs)
		manifest.manifest[k] = newVs
	}
	return manifest
}

func (manifest *ManifestBase) Err() error {
	if manifest == nil {
		return nil
	}
	return manifest.err
}

func (manifest *ManifestBase) Equals(manifest2 inventory.Manifest) bool {
	if manifest == nil || manifest2 == nil {
		return false
	}
	manifest2B, ok := manifest2.(*ManifestBase)
	if !ok {
		return false
	}
	if (manifest.Err() == nil && manifest2.Err() != nil) || (manifest.Err() != nil && manifest2.Err() == nil) {
		return false
	}
	if (manifest.Err() != nil && manifest2.Err() != nil) && manifest.Err().Error() != manifest2.Err().Error() {
		return false
	}
	if len(manifest.manifest) != len(manifest2B.manifest) {
		return false
	}
	for k, v := range manifest.manifest {
		v2, ok := manifest2B.manifest[k]
		if !ok {
			return false
		}
		if slices.Compare(v, v2) != 0 {
			return false
		}
	}
	return true
}

func (manifest *ManifestBase) String() string {
	var num int64
	var unique int64
	for _, v := range manifest.manifest {
		unique++
		num += int64(len(v))
	}
	return fmt.Sprintf("%d files (%d unique)", num, unique)
}

func (manifest *ManifestBase) Iterate() func(yield func(digest string, internal []string) bool) {
	return func(yield func(digest string, external []string) bool) {
		for digest, files := range manifest.manifest {
			if !yield(digest, files) {
				return
			}
		}
	}
}

func (manifest *ManifestBase) GetFiles(digest string) ([]string, error) {
	files, ok := manifest.manifest[digest]
	if !ok {
		return nil, errors.Wrapf(inventory.DigestNotFound, "digest %manifest", digest)
	}
	return files, nil
}

func (manifest *ManifestBase) UnmarshalJSON(data []byte) error {
	manifest.manifest = map[string][]string{}
	if err := json.Unmarshal(data, &manifest.manifest); err != nil {
		manifest.err = errors.Wrapf(err, "cannot unmarshal state %manifest", string(data))
		return nil
	}
	return nil
}

func (manifest *ManifestBase) MarshalJSON() ([]byte, error) {
	return json.Marshal(manifest.manifest)
}

var _ inventory.Manifest = (*ManifestBase)(nil)

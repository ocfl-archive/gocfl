package ocfl

import (
	"strings"
	"time"

	"emperror.dev/errors"
	"github.com/google/uuid"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/indexer/v3/pkg/indexer"
	"golang.org/x/exp/maps"
)

type FileMetadata struct {
	Checksums    map[checksum.DigestAlgorithm]string
	InternalName []string
	VersionName  map[string][]string
	Extension    map[string]any
}

type FilesMetadata map[string]*FileMetadata

func (fms FilesMetadata) Obfuscate() error {
	var stringMap = map[string]string{}
	var errs = []error{}
	for _, fm := range fms {
		names := []string{}
		for _, str := range fm.InternalName {
			parts := strings.SplitN(str, "/", 3)
			if len(parts) != 3 {
				return errors.Errorf("invalid version filename '%s'", str)
			}
			replace, ok := stringMap[parts[2]]
			if !ok {
				replace = uuid.New().String()
				stringMap[parts[2]] = replace
			}
			names = append(names, strings.Join([]string{parts[0], parts[1], replace}, "/"))
		}
		fm.InternalName = names
	}
	//clear(stringMap)
	for _, fm := range fms {
		for _, key := range maps.Keys(fm.VersionName) {
			strs := fm.VersionName[key]
			names := []string{}
			for _, str := range strs {
				replace, ok := stringMap[str]
				if !ok {
					replace = uuid.New().String()
					stringMap[str] = replace
				}
				names = append(names, replace)
			}
			fm.VersionName[key] = names
		}
	}
	for _, fm := range fms {
		idx, hasIndexer := fm.Extension["NNNN-indexer"]
		clear(fm.Extension)
		if hasIndexer {
			idx2, ok := idx.(*indexer.ResultV2)
			if ok {
				idx2.Metadata = map[string]any{}
				idx2.Errors = map[string]string{}
				fm.Extension["NNNN-indexer"] = idx2
			}
		}
	}
	for _, fm := range fms {
		fm.Checksums = map[checksum.DigestAlgorithm]string{}
	}
	newMeta := FilesMetadata{}
	for _, fm := range fms {
		newMeta[uuid.New().String()] = fm
	}
	clear(fms)
	for key, fm := range newMeta {
		fms[key] = fm
	}
	return errors.Combine(errs...)
}

type VersionMetadata struct {
	Created time.Time
	Message string
	Name    string
	Address string
}

type ObjectMetadata struct {
	ID              string
	DigestAlgorithm checksum.DigestAlgorithm
	Head            string
	Versions        map[string]*VersionMetadata
	Files           FilesMetadata
	Extension       any
}

func (om *ObjectMetadata) Obfuscate() error {
	var metafile any
	var hasMetafile = false
	extension := map[string]any{}
	if extMap, ok := om.Extension.(map[string]any); ok {
		if metafile, hasMetafile = extMap["NNNN-metafile"]; hasMetafile {
			extension["NNNN-metafile"] = metafile
		}
	}
	om.Extension = extension
	return errors.WithStack(om.Files.Obfuscate())
}

type StorageRootMetadata struct {
	Objects map[string]*ObjectMetadata
}

func (srm *StorageRootMetadata) Obfuscate() error {
	var errs = []error{}
	for _, om := range srm.Objects {
		if err := om.Obfuscate(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Combine(errs...)
}

package inventory

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"emperror.dev/errors"
	"github.com/google/uuid"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/indexer/v3/pkg/indexer"
	"golang.org/x/exp/maps"
)

// FileMetadata contains metadata for a specific file across its versions.
type FileMetadata struct {
	Checksums    map[checksum.DigestAlgorithm]string // Additional fixity checksums
	InternalName []string                            // Physical file paths
	VersionName  map[string][]string                 // Logical file names per version
	Extension    map[string]any                      // Extension metadata
}

// FilesMetadata is a map of file IDs to their metadata.
type FilesMetadata map[string]*FileMetadata

// Obfuscate replaces sensitive file names with random UUIDs and removes sensitive parts of metadata.
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

// VersionMetadata contains metadata for a specific OCFL version.
type VersionMetadata struct {
	Created time.Time // Creation timestamp
	Message string    // Version message
	Name    string    // User name
	Address string    // User address
}

// Metadata represents the flattened metadata of an OCFL object.
type Metadata struct {
	ID              string                      // Object ID
	DigestAlgorithm checksum.DigestAlgorithm    // Primary digest algorithm
	Head            *VersionNumber              // Current head version
	Versions        map[string]*VersionMetadata // Version metadata by version number string
	Files           FilesMetadata               // File metadata by file ID
	Extension       any                         // Object level extension metadata
}

// Obfuscate sensitive information in the object metadata.
func (om *Metadata) Obfuscate() error {
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

func (om *Metadata) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Object ID: %s\n", om.ID))
	sb.WriteString(fmt.Sprintf("Digest Algorithm: %s\n", om.DigestAlgorithm))
	if om.Head != nil {
		sb.WriteString(fmt.Sprintf("Head: %s\n", om.Head.String()))
	}

	sb.WriteString("\nVersions:\n")
	vKeys := make([]string, 0, len(om.Versions))
	for k := range om.Versions {
		vKeys = append(vKeys, k)
	}
	sort.Strings(vKeys)
	for _, k := range vKeys {
		v := om.Versions[k]
		sb.WriteString(fmt.Sprintf("  %s:\n", k))
		sb.WriteString(fmt.Sprintf("    Created: %s\n", v.Created.Format(time.RFC3339)))
		if v.Message != "" {
			sb.WriteString(fmt.Sprintf("    Message: %s\n", v.Message))
		}
		if v.Name != "" || v.Address != "" {
			sb.WriteString(fmt.Sprintf("    User: %s <%s>\n", v.Name, v.Address))
		}
	}

	sb.WriteString("\nFiles:\n")
	fKeys := make([]string, 0, len(om.Files))
	for k := range om.Files {
		fKeys = append(fKeys, k)
	}
	sort.Strings(fKeys)
	for _, k := range fKeys {
		fm := om.Files[k]
		sb.WriteString(fmt.Sprintf("  [%s]:\n", k))
		if len(fm.InternalName) > 0 {
			sb.WriteString(fmt.Sprintf("    Physical: %s\n", strings.Join(fm.InternalName, ", ")))
		}
		vkNames := make([]string, 0, len(fm.VersionName))
		for vk := range fm.VersionName {
			vkNames = append(vkNames, vk)
		}
		sort.Strings(vkNames)
		for _, vk := range vkNames {
			sb.WriteString(fmt.Sprintf("    Version %s: %s\n", vk, strings.Join(fm.VersionName[vk], ", ")))
		}
	}

	return sb.String()
}

// StorageRootMetadata contains metadata for all objects in a storage root.
type StorageRootMetadata struct {
	Objects map[string]*Metadata // Metadata map by object ID
}

// Obfuscate sensitive information in the storage root metadata.
func (srm *StorageRootMetadata) Obfuscate() error {
	var errs = []error{}
	for _, om := range srm.Objects {
		if err := om.Obfuscate(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Combine(errs...)
}

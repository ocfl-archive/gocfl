package ocfl

import "github.com/je4/gocfl/v2/pkg/checksum"

type FileMetadata struct {
	Checksums    map[checksum.DigestAlgorithm]string
	InternalName []string
	VersionName  map[string][]string
	Extension    map[string]any
}

type ObjectMetadata struct {
	ID    string
	Files map[string]*FileMetadata
}

type StorageRootMetadata struct {
	Objects map[string]*ObjectMetadata
}

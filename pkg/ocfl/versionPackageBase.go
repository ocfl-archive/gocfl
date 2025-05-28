package ocfl

import (
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/zLogger"
)

type VersionPackagesSpec string

const (
	VersionPackageSpec2_0 VersionPackagesSpec = "https://ocfl.io/2.0/spec/#version-package"
)

func newVersionPackageBase(
	logger zLogger.ZLogger,
) *VersionPackagesBase {
	return &VersionPackagesBase{
		logger:   logger,
		Versions: make(map[string]*PackageVersionBase),
		Manifest: make(map[string][]string),
	}
}

type MetadataBase struct {
	Format        string `json:"format"`
	FormatVersion string `json:"formatVersion"`
	Extension     string `json:"extension,omitempty"`
}

type PackageVersionBase struct {
	Metadata *MetadataBase `json:"metadata"`
	Packages []string      `json:"packages"`
}

type VersionPackagesBase struct {
	DigestAlgorithm checksum.DigestAlgorithm       `json:"digestAlgorithm"`
	Type            VersionPackagesSpec            `json:"type"`
	Manifest        map[string][]string            `json:"manifest"`
	Versions        map[string]*PackageVersionBase `json:"versions"`
	logger          zLogger.ZLogger                `json:"-"`
}

func (v *VersionPackagesBase) GetDigestAlgorithm() checksum.DigestAlgorithm {
	return v.DigestAlgorithm
}

func (v *VersionPackagesBase) GetSpec() VersionPackagesSpec {
	return v.Type
}

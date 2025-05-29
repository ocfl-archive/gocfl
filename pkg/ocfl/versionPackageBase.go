package ocfl

import (
	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/zLogger"
)

type VersionPackagesSpec string

const (
	VersionPackageSpec2_0 VersionPackagesSpec = "https://ocfl.io/2.0/spec/#version-package"
)

func newVersionPackageBase(
	digestAlgorithm checksum.DigestAlgorithm,
	logger zLogger.ZLogger,
) *VersionPackagesBase {
	return &VersionPackagesBase{
		DigestAlgorithm: digestAlgorithm,
		Type:            VersionPackageSpec2_0,
		logger:          logger,
		Versions:        make(map[string]*PackageVersionBase),
		Manifest:        make(map[string][]string),
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

// VersionPackagesBase is the base structure for version packages in OCFL.
// It may be nil if OCFL version is below 2.0.
type VersionPackagesBase struct {
	DigestAlgorithm checksum.DigestAlgorithm       `json:"digestAlgorithm"`
	Type            VersionPackagesSpec            `json:"type"`
	Manifest        map[string][]string            `json:"manifest"`
	Versions        map[string]*PackageVersionBase `json:"versions"`
	logger          zLogger.ZLogger                `json:"-"`
}

func (v *VersionPackagesBase) AddVersion(version string, versionType VersionPackagesType, versionTypeVersion string, fileDigest map[string]string) error {
	if v == nil || versionType == VersionPlain {
		return nil // No version package for plain type or nil version packages
	}
	versionTypeString, ok := VersionPackageTypeString[versionType]
	if !ok {
		return errors.Errorf("unknown version package type '%s'", versionType)
	}
	pv := &PackageVersionBase{
		Metadata: &MetadataBase{
			Format:        versionTypeString,
			FormatVersion: versionTypeVersion,
		},
		Packages: []string{},
	}
	for file, digest := range fileDigest {
		if _, ok := v.Manifest[digest]; !ok {
			v.Manifest[digest] = []string{}
		}
		v.Manifest[digest] = append(v.Manifest[digest], file)
		pv.Packages = append(pv.Packages, file)
	}
	v.Versions[version] = pv
	return nil
}

func (v *VersionPackagesBase) Init(digestAlgorithm checksum.DigestAlgorithm) error {
	if v == nil {
		return nil
	}
	v.DigestAlgorithm = digestAlgorithm
	return nil
}

func (v *VersionPackagesBase) IsEmpty() bool {
	return v == nil || len(v.Versions) == 0
}

func (v *VersionPackagesBase) GetDigestAlgorithm() checksum.DigestAlgorithm {
	if v == nil {
		return ""
	}
	return v.DigestAlgorithm
}

func (v *VersionPackagesBase) GetSpec() VersionPackagesSpec {
	if v == nil {
		return ""
	}
	return v.Type
}

var _ VersionPackages = (*VersionPackagesBase)(nil)

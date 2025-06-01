package ocfl

import (
	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/zLogger"
	"io"
	"io/fs"
	"slices"
)

type VersionPackagesSpec string

const (
	VersionPackageSpec2_0 VersionPackagesSpec = "https://ocfl.io/2.0/spec/#version-package"
)

func newVersionPackagesBase(
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

func (v *VersionPackagesBase) GetVersions() []string {
	result := make([]string, 0, len(v.Versions))
	for version := range v.Versions {
		result = append(result, version)
	}
	return result
}

func (v *VersionPackagesBase) HasPart(name string) bool {
	for _, pv := range v.Versions {
		if slices.Contains(pv.Packages, name) {
			return true // Found the part in one of the version packages
		}
	}
	return false // Part not found in any version package
}

func (v *VersionPackagesBase) GetVersion(version string) (*PackageVersionBase, bool) {
	if v == nil || v.IsEmpty() {
		return nil, false // No version packages or nil version packages
	}
	pv, ok := v.Versions[version]
	return pv, ok
}

func (v *VersionPackagesBase) GetFS(version string, object Object) (fs.FS, io.Closer, error) {
	if v == nil || v.IsEmpty() {
		return object.GetFS(), io.NopCloser(nil), nil // No version packages or nil version packages
	}
	pv, ok := v.Versions[version]
	if !ok {
		return object.GetFS(), io.NopCloser(nil), nil // No version package for this version
	}
	switch pv.Metadata.Format {
	case "zip":
		return NewMultiZIPFS(object.GetFS(), pv.Packages, v.logger)
	default:
		return nil, nil, errors.Errorf("unknown version package format '%s'", pv.Metadata.Format)
	}
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

package inventory

import "github.com/je4/utils/v2/pkg/checksum"

type Factory interface {
	NewVersions() Versions
	NewVersion() Version
	NewState() State
	NewUser() User
	NewManifest() Manifest
	NewFixity(algorithms []checksum.DigestAlgorithm) Fixity
}

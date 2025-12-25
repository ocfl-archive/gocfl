package inventory

type Factory interface {
	NewVersions() Versions
	NewVersion() Version
	NewState() State
	NewUser() User
	NewManifest() Manifest
}

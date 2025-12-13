package inventory

type Factory interface {
	NewVersions() Versions
	NewVersion() Version
	NewState() State
}

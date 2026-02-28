package inventory

import ()

type User interface {
	String() string
	Equals(other User) bool
	GetAddress() string
	GetName() string
	Err() error
	Finalize()
	WithAddress(address string) User
	WithName(name string) User
	Check(version *VersionNumber) error
}

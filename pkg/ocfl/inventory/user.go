package inventory

import "github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"

type User interface {
	String() string
	Equals(other User) bool
	GetAddress() string
	GetName() string
	Err() error
	Finalize()
	WithAddress(address string) User
	WithName(name string) User
	Check(val validation.Validation, version *VersionNumber) error
}

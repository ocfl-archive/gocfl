package inventory

import "github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"

type User interface {
	String() string
	Equals(other User) bool
	GetAddress() (string, error)
	GetName() (string, error)
	Err() error
	Finalize()
	SetAddress(address string) User
	SetName(name string) User
	Check(val validation.Validation, version string) error
}

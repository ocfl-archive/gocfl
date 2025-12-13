package inventory

import "github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"

type Version interface {
	String() string
	Equals(other Version) bool
	Finalize(inCreation bool, val validation.Validation) error
}

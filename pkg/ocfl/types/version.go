package types

import (
	"time"

	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
)

type Version interface {
	Operations

	String() string
	Equals(other Version) bool
	Finalize(val validation.Validation, factory Factory, inCreation bool) error
	WithCreated(t time.Time) Version
	GetCreated() time.Time
	WithMessage(msg string) Version
	GetMessage() string
	WithState(state State) Version
	GetState() State
	WithUser(user User) Version
	GetUser() User
	WithVersion(number *VersionNumber) Version
	GetVersionNumber() *VersionNumber
	Check(val validation.Validation, manifestDigests, manifestDigestsLower []string) error
	Err() error
	FileChecksum(path string) string
}

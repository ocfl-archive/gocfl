package interfaces

import (
	"time"

	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
)

type Version interface {
	Operations

	String() string
	Equals(other Version) bool
	Finalize(val validation.Validation, factory Factory, inCreation bool) error
	GetState() State
	GetUser() User
	GetCreated() time.Time
	GetMessage() string
	WithCreated(t time.Time) Version
	WithMessage(msg string) Version
	WithState(state State) Version
	WithUser(user User) Version
	Check(val validation.Validation, manifestDigests, manifestDigestsLower []string) error
	Err() error
	FileChecksum(path string) string
	SetVersion(number *VersionNumber)
}

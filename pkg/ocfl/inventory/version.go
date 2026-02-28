package inventory

import (
	"time"
)

type Version interface {
	Operations

	String() string
	Equals(other Version) bool
	Finalize(inCreation bool) error
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
	Check(manifestDigests, manifestDigestsLower []string) error
	Err() error
	FileChecksum(path string) string
	InCreation() bool
}

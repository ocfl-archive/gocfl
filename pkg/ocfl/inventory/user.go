package inventory

// User represents the author of an OCFL version.
type User interface {
	// String returns a string representation of the user.
	String() string
	// Equals checks if two users are equal.
	Equals(other User) bool
	// GetAddress returns the user's address.
	GetAddress() string
	// GetName returns the user's name.
	GetName() string
	// Err returns any error encountered during user operations.
	Err() error
	// Finalize completes the user creation or update.
	Finalize()
	// WithAddress sets the user's address.
	WithAddress(address string) User
	// WithName sets the user's name.
	WithName(name string) User
	// Check verifies the user integrity.
	Check(version *VersionNumber) error
}

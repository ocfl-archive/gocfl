package inventory

// Operations defines the common file operations for inventory components (Versions, State).
type Operations interface {
	// DeleteFile removes a file from the state.
	DeleteFile(stateFilename string) (bool, error)
	// RenameFile changes the logical path of a file.
	RenameFile(oldStateFilename, newStateFilename string) (bool, error)
	// CopyFile copies a file within the state.
	CopyFile(stateFilename, digest string) (bool, error)
	// EchoDelete performs a dry-run or logging of file deletions.
	EchoDelete(existing []string, pathPrefix string) (bool, error)
	// AddFile adds a file to the state.
	AddFile(stateFilename string, digest string) (bool, error)
}

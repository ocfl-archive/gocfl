package inventory

type Operations interface {
	DeleteFile(stateFilename string) (bool, error)
	RenameFile(oldStateFilename, newStateFilename string) (bool, error)
	CopyFile(stateFilename, digest string) (bool, error)
}

package types

type Operations interface {
	DeleteFile(stateFilename string) (bool, error)
	RenameFile(oldStateFilename, newStateFilename string) (bool, error)
	CopyFile(stateFilename, digest string) (bool, error)
	EchoDelete(existing []string, pathPrefix string) (bool, error)
	AddFile(stateFilename string, digest string) (bool, error)
}

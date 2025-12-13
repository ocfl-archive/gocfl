package inventory

type State interface {
	String() string
	IterateFiles() func(yield func(external []string, digest string) bool)
	GetFiles(digest string) ([]string, error)
}

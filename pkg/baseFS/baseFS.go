package baseFS

type BaseFS interface {
	Valid(path string) bool
}

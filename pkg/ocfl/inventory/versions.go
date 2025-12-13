package inventory

type Versions interface {
	String() string
	Iterate() func(yield func(versionString string, version Version) bool)
	Equals(other Versions) bool
	Get(versionString string) (Version, bool)
}

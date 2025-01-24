package version

var (
	Version = "dev-0.0.0"
	Commit  = "000000000000000000000000000000000badf00d"
	Date    = "1970-01-01T00:00:01Z"
	BuiltBy = "dev"
)

// ShortCommit returns a short commit hash.
func ShortCommit() string {
	return Commit[:6]
}

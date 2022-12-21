package osfs

import (
	"go.ub.unibas.ch/gocfl/v2/pkg/baseFS"
	"net/url"
)

type BaseFS struct {
}

func NewBaseFS() (baseFS.BaseFS, error) {
	return &BaseFS{}, nil
}

func (*BaseFS) Valid(path string) bool {
	if u, err := url.Parse(path); err == nil {
		return u.Scheme == "file"
	}
	return true
}

var (
	_ baseFS.BaseFS = &BaseFS{}
)

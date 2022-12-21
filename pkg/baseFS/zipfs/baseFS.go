package zipfs

import (
	"go.ub.unibas.ch/gocfl/v2/pkg/baseFS"
	"go.ub.unibas.ch/gocfl/v2/pkg/ocfl"
	"strings"
)

type BaseFS struct {
}

func NewBaseFS() (baseFS.BaseFS, error) {
	return &BaseFS{}, nil
}

func (b *BaseFS) Valid(path string) bool {
	return strings.HasSuffix(strings.ToLower(path), ".zip")
}

func (b *BaseFS) GetFS(path string) (ocfl.OCFLFS, error) {
	//tempFilename := path + ".tmp"
	return nil, nil
}

var (
	_ baseFS.BaseFS = &BaseFS{}
)

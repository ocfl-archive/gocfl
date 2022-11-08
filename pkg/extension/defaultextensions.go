package extension

import (
	"emperror.dev/errors"
	"go.ub.unibas.ch/gocfl/v2/pkg/ocfl"
)

func NewDefaultStorageRootExtension() (ocfl.Extension, error) {
	var err error
	var cfg = &DirectCleanConfig{
		ExtensionConfig:             &ocfl.ExtensionConfig{ExtensionName: DirectCleanName},
		MaxPathnameLen:              32000,
		MaxFilenameLen:              127,
		WhitespaceReplacementString: " ",
		ReplacementString:           "_",
		UTFEncode:                   true,
	}
	layout, err := NewDirectClean(cfg)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot initialize %s", cfg.ExtensionName)
	}
	return layout, nil
}

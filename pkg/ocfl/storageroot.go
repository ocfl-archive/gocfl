package ocfl

import (
	"github.com/goph/emperror"
	"github.com/op/go-logging"
	"io/fs"
)

type OCFLStorageRoot struct {
	fs                OCFLFS
	changed           bool
	logger            *logging.Logger
	hierarchyTopLevel bool
}

// NewOCFL creates an empty OCFL structure
func NewOCFLStorageRoot(fs OCFLFS, hierarchyTopLevel bool, logger *logging.Logger) (*OCFLStorageRoot, error) {
	ocfl := &OCFLStorageRoot{fs: fs, hierarchyTopLevel: hierarchyTopLevel, logger: logger}

	if err := ocfl.Init(); err != nil {
		return nil, emperror.Wrap(err, "cannot initialize ocfl")
	}
	return ocfl, nil
}

func (osr *OCFLStorageRoot) Init() error {
	osr.logger.Debug()

	// first check whether ocfl is not empty
	fp, err := osr.fs.Open(rootConformanceDeclaration)
	if err != nil {
		if err != fs.ErrNotExist {
			return emperror.Wrap(err, "cannot initialize OCFL layout")
		}
		_, err := osr.fs.Create(rootConformanceDeclaration)
		if err != nil {
			return emperror.Wrapf(err, "cannot create %s", rootConformanceDeclaration)
		}
	} else {
		if err := fp.Close(); err != nil {
			return emperror.Wrapf(err, "cannot close %s", rootConformanceDeclaration)
		}
	}

	return nil
}

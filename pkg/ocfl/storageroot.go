package ocfl

import (
	"bytes"
	"fmt"
	"github.com/goph/emperror"
	"github.com/op/go-logging"
	"gitlab.switch.ch/ub-unibas/gocfl/v2/pkg/storagelayout"
	"io"
	"io/fs"
)

type OCFLStorageRoot struct {
	fs      OCFLFS
	changed bool
	logger  *logging.Logger
	layout  storagelayout.StorageLayout
}

// NewOCFL creates an empty OCFL structure
func NewOCFLStorageRoot(fs OCFLFS, layout storagelayout.StorageLayout, logger *logging.Logger) (*OCFLStorageRoot, error) {
	ocfl := &OCFLStorageRoot{fs: fs, layout: layout, logger: logger}

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
		// write default storage layout into extension folder
		if err != fs.ErrNotExist {
			return emperror.Wrap(err, "cannot initialize OCFL layout")
		}
		_, err := osr.fs.Create(rootConformanceDeclaration)
		if err != nil {
			return emperror.Wrapf(err, "cannot create %s", rootConformanceDeclaration)
		}
		extConfig, err := osr.fs.Create(fmt.Sprintf("extensions/%s/config.json", osr.layout.Name()))
		if err != nil {
			return emperror.Wrapf(err, "cannot create %s", rootConformanceDeclaration)
		}
		defer extConfig.Close()
		if err := osr.layout.WriteConfig(extConfig); err != nil {
			return emperror.Wrap(err, "cannot write config")
		}
	} else {
		// read storage layout from extension folder...
		if err := fp.Close(); err != nil {
			return emperror.Wrapf(err, "cannot close %s", rootConformanceDeclaration)
		}
		exts, err := osr.fs.ReadDir("extensions")
		if err != nil {
			// if directory does not exists - no problem
			if err != fs.ErrNotExist {
				return emperror.Wrap(err, "cannot read extensions folder")
			}
			exts = []fs.DirEntry{}
		}
		var layout storagelayout.StorageLayout
		for _, extFolder := range exts {
			extConfig := fmt.Sprintf("extensions/%s/config.json", extFolder.Name())
			configReader, err := osr.fs.Open(extConfig)
			if err != nil {
				return emperror.Wrapf(err, "cannot open %s for reading", extConfig)
			}
			buf := bytes.NewBuffer(nil)
			if _, err := io.Copy(buf, configReader); err != nil {
				return emperror.Wrapf(err, "cannot read %s", extConfig)
			}
			if layout, err = storagelayout.NewStorageLayout(buf.Bytes()); err != nil {
				osr.logger.Warningf("%s not a storage layout: %v", extConfig, err)
				continue
			}
		}
		if layout == nil {
			// ...or set to default
			if layout, err = storagelayout.NewDefaultStorageLayout(); err != nil {
				return emperror.Wrap(err, "cannot initiate default storage layout")
			}
		}
	}

	return nil
}

func (osr *OCFLStorageRoot) GetObjectFolders() ([]string, error) {
	dirs, err := osr.fs.ReadDir("")
	if err != nil {
		return nil, emperror.Wrap(err, "cannot read folders of storage root")
	}
	var result = []string{}
	for _, dir := range dirs {
		if !dir.IsDir() || dir.Name() == "extensions" {
			continue
		}
		result = append(result, dir.Name())
	}
	return result, nil
}

func (osr *OCFLStorageRoot) OpenObjectFolder(folder string) (*OCFLObject, error) {
	return NewOCFLObject(osr.fs, folder, "", osr.logger)
}

func (osr *OCFLStorageRoot) OpenObject(id string) (*OCFLObject, error) {
	folder, err := osr.layout.ID2Path(id)
	if err != nil {
		return nil, emperror.Wrapf(err, "cannot build path for %s", id)
	}
	return NewOCFLObject(osr.fs, folder, id, osr.logger)
}

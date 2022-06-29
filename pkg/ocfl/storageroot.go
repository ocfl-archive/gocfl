package ocfl

import (
	"bytes"
	"fmt"
	"github.com/goph/emperror"
	"github.com/op/go-logging"
	"gitlab.switch.ch/ub-unibas/gocfl/v2/pkg/storagelayout"
	"io"
	"io/fs"
	"regexp"
)

type OCFLStorageRoot struct {
	fs      OCFLFS
	changed bool
	logger  *logging.Logger
	layout  storagelayout.StorageLayout
	version string
}

// NewOCFL creates an empty OCFL structure
func NewOCFLStorageRoot(fs OCFLFS, layout storagelayout.StorageLayout, logger *logging.Logger) (*OCFLStorageRoot, error) {
	ocfl := &OCFLStorageRoot{fs: fs, layout: layout, logger: logger}

	if err := ocfl.Init(); err != nil {
		return nil, emperror.Wrap(err, "cannot initialize ocfl")
	}
	return ocfl, nil
}

var NAMASTEVersionRegexp = regexp.MustCompile("[0-9]+=ocfl_([0-9]+\\.[0-9]+)")

func (osr *OCFLStorageRoot) Init() error {
	osr.logger.Debug()

	// init failed
	files, err := osr.GetFiles()
	if err != nil {
		return emperror.Wrap(err, "cannot get root files")
	}
	for _, file := range files {
		matches := NAMASTEVersionRegexp.FindStringSubmatch(file)
		if matches != nil {
			osr.version = matches[1]
			break
		}
	}
	folders, err := osr.GetFolders()
	if err != nil {
		return emperror.Wrap(err, "cannot get root folders")
	}
	if osr.version == "" && (len(files) > 0 || len(folders) > 0) {
		return ErrorE003 // ‘[The version declaration] must be a file in the base directory of the OCFL Object Root giving the OCFL version in the filename.’
	}
	if osr.version == "" {
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
		osr.version = VERSION
	} else {
		// read storage layout from extension folder...
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

func (osr *OCFLStorageRoot) GetFiles() ([]string, error) {
	dirs, err := osr.fs.ReadDir("")
	if err != nil {
		return nil, emperror.Wrap(err, "cannot read folders of storage root")
	}
	var result = []string{}
	for _, dir := range dirs {
		if dir.IsDir() {
			continue
		}
		result = append(result, dir.Name())
	}
	return result, nil
}

func (osr *OCFLStorageRoot) GetFolders() ([]string, error) {
	dirs, err := osr.fs.ReadDir("")
	if err != nil {
		return nil, emperror.Wrap(err, "cannot read folders of storage root")
	}
	var result = []string{}
	for _, dir := range dirs {
		if !dir.IsDir() || dir.Name() == "." || dir.Name() == ".." {
			continue
		}
		result = append(result, dir.Name())
	}
	return result, nil
}

func (osr *OCFLStorageRoot) GetObjectFolders() ([]string, error) {
	dirs, err := osr.GetFolders()
	if err != nil {
		return nil, err
	}
	var result = []string{}
	for _, dir := range dirs {
		if dir == "extensions" {
			continue
		}
		result = append(result, dir)
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

func (osr *OCFLStorageRoot) Check() error {
	// https://ocfl.io/1.0/spec/validation-codes.html
	objectFolders, err := osr.GetObjectFolders()
	if err != nil {
		return emperror.Wrapf(err, "cannot get object folders")
	}
	multiError := emperror.NewMultiErrorBuilder()
	for _, objectFolder := range objectFolders {
		osr.logger.Infof("checking folder %s", objectFolder)
		obj, err := NewOCFLObject(osr.fs, objectFolder, "", osr.logger)
		if err != nil {
			return emperror.Wrapf(err, "cannot instantiate OCFL Object at %s", objectFolder)
		}
		if err := obj.Check(); err != nil {
			multiError.Add(err)
		}
	}
	return multiError.ErrOrNil()
}

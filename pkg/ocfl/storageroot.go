package ocfl

import (
	"bytes"
	"emperror.dev/emperror"
	"emperror.dev/errors"
	"fmt"
	"github.com/op/go-logging"
	"go.ub.unibas.ch/gocfl/v2/pkg/extension"
	"io"
	"io/fs"
	"regexp"
)

type OCFLStorageRoot struct {
	fs      OCFLFS
	changed bool
	logger  *logging.Logger
	layout  extension.StorageLayout
	version string
}

var rootConformanceDeclaration = fmt.Sprintf("0=ocfl_%s", VERSION)

// NewOCFL creates an empty OCFL structure
func NewOCFLStorageRoot(fs OCFLFS, layout extension.StorageLayout, logger *logging.Logger) (*OCFLStorageRoot, error) {
	ocfl := &OCFLStorageRoot{fs: fs, layout: layout, logger: logger}

	if err := ocfl.Init(); err != nil {
		return nil, errors.Wrap(err, "cannot initialize ocfl")
	}
	return ocfl, nil
}

var NAMASTEVersionRegexp = regexp.MustCompile("[0-9]+=ocfl_([0-9]+\\.[0-9]+)")

func (osr *OCFLStorageRoot) Init() error {
	osr.logger.Debug()

	// init failed
	files, err := osr.GetFiles()
	if err != nil {
		return errors.Wrap(err, "cannot get root files")
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
		return errors.Wrap(err, "cannot get root folders")
	}
	if osr.version == "" && (len(files) > 0 || len(folders) > 0) {
		return errors.WithStack(ErrorE069) // ‘An OCFL Storage Root MUST contain a Root Conformance Declaration identifying it as such. (https://ocfl.io/1.0/spec/#E069)’
	}
	if osr.version == "" {
		_, err := osr.fs.Create(rootConformanceDeclaration)
		if err != nil {
			return errors.Wrapf(err, "cannot create %s", rootConformanceDeclaration)
		}
		extConfig, err := osr.fs.Create(fmt.Sprintf("extensions/%s/config.json", osr.layout.Name()))
		if err != nil {
			return errors.Wrapf(err, "cannot create %s", rootConformanceDeclaration)
		}
		defer extConfig.Close()
		if err := osr.layout.WriteConfig(extConfig); err != nil {
			return errors.Wrap(err, "cannot write config")
		}
		osr.version = VERSION
	} else {
		// read storage layout from extension folder...
		exts, err := osr.fs.ReadDir("extensions")
		if err != nil {
			// if directory does not exists - no problem
			if err != fs.ErrNotExist {
				return errors.Wrap(err, "cannot read extensions folder")
			}
			exts = []fs.DirEntry{}
		}
		var layout extension.StorageLayout
		for _, extFolder := range exts {
			extConfig := fmt.Sprintf("extensions/%s/config.json", extFolder.Name())
			configReader, err := osr.fs.Open(extConfig)
			if err != nil {
				return errors.Wrapf(err, "cannot open %s for reading", extConfig)
			}
			buf := bytes.NewBuffer(nil)
			if _, err := io.Copy(buf, configReader); err != nil {
				return errors.Wrapf(err, "cannot read %s", extConfig)
			}
			if layout, err = extension.NewStorageLayout(buf.Bytes()); err != nil {
				osr.logger.Warningf("%s not a storage layout: %v", extConfig, err)
				continue
			}
		}
		if layout == nil {
			// ...or set to default
			if layout, err = extension.NewDefaultStorageLayout(); err != nil {
				return errors.Wrap(err, "cannot initiate default storage layout")
			}
		}
	}
	return nil
}

func (osr *OCFLStorageRoot) GetFiles() ([]string, error) {
	dirs, err := osr.fs.ReadDir("")
	if err != nil {
		return nil, errors.Wrap(err, "cannot read folders of storage root")
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
		return nil, errors.Wrap(err, "cannot read folders of storage root")
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
		return nil, errors.Wrapf(err, "cannot build path for %s", id)
	}
	return NewOCFLObject(osr.fs, folder, id, osr.logger)
}

func (osr *OCFLStorageRoot) Check() error {
	// https://ocfl.io/1.0/spec/validation-codes.html
	objectFolders, err := osr.GetObjectFolders()
	if err != nil {
		return errors.Wrapf(err, "cannot get object folders")
	}
	multiError := emperror.NewMultiErrorBuilder()
	for _, objectFolder := range objectFolders {
		osr.logger.Infof("checking folder %s", objectFolder)
		obj, err := NewOCFLObject(osr.fs, objectFolder, "", osr.logger)
		if err != nil {
			return errors.Wrapf(err, "cannot instantiate OCFL Object at %s", objectFolder)
		}
		if err := obj.Check(); err != nil {
			multiError.Add(err)
		}
	}
	return multiError.ErrOrNil()
}

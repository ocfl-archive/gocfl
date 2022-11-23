package cmd

import (
	"context"
	"emperror.dev/errors"
	"fmt"
	"github.com/op/go-logging"
	"go.ub.unibas.ch/gocfl/v2/data/defaultextensions"
	"go.ub.unibas.ch/gocfl/v2/pkg/extension"
	"go.ub.unibas.ch/gocfl/v2/pkg/genericfs"
	"go.ub.unibas.ch/gocfl/v2/pkg/ocfl"
	"go.ub.unibas.ch/gocfl/v2/pkg/osfs"
	"go.ub.unibas.ch/gocfl/v2/pkg/zipfs"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func initExtensionFactory(extensionFactory *ocfl.ExtensionFactory) error {
	extensionFactory.AddCreator(extension.DirectCleanName, func(fs ocfl.OCFLFS) (ocfl.Extension, error) {
		return extension.NewDirectCleanFS(fs)
	})

	extensionFactory.AddCreator(extension.PathDirectName, func(fs ocfl.OCFLFS) (ocfl.Extension, error) {
		return extension.NewPathDirectFS(fs)
	})

	extensionFactory.AddCreator(extension.StorageLayoutFlatDirectName, func(fs ocfl.OCFLFS) (ocfl.Extension, error) {
		return extension.NewStorageLayoutFlatDirectFS(fs)
	})

	extensionFactory.AddCreator(extension.StorageLayoutHashAndIdNTupleName, func(fs ocfl.OCFLFS) (ocfl.Extension, error) {
		return extension.NewStorageLayoutHashAndIdNTupleFS(fs)
	})

	extensionFactory.AddCreator(extension.StorageLayoutHashedNTupleName, func(fs ocfl.OCFLFS) (ocfl.Extension, error) {
		return extension.NewStorageLayoutHashedNTupleFS(fs)
	})

	extensionFactory.AddCreator(extension.StorageLayoutPairTreeName, func(fs ocfl.OCFLFS) (ocfl.Extension, error) {
		return extension.NewStorageLayoutPairTreeFS(fs)
	})

	return nil
}

func initDefaultExtensions(extensionFactory *ocfl.ExtensionFactory, defaultExtensionFolder string, logger *logging.Logger) (storageRootExtensions, objectExtensions []ocfl.Extension, err error) {
	var dExtDirFS fs.FS
	if defaultExtensionFolder == "" {
		dExtDirFS = defaultextensions.DefaultExtensionFS
	} else {
		dExtDirFS = os.DirFS(defaultExtensionFolder)
	}
	ofs, err := genericfs.NewGenericFS(dExtDirFS, ".", logger)
	if err != nil {
		err = errors.Wrapf(err, "cannot create generic fs for %v", dExtDirFS)
		return
	}
	subFS, err := ofs.SubFS("storageroot")
	if err != nil {
		err = errors.Wrapf(err, "cannot create subfs'%s'for %v", "storageroot", dExtDirFS)
		return
	}
	storageRootExtensions, err = extensionFactory.LoadExtensions(subFS)
	if err != nil {
		err = errors.Wrapf(err, "cannot load extension folder %v", ofs)
		return
	}
	subFS, err = ofs.SubFS("object")
	if err != nil {
		err = errors.Wrapf(err, "cannot create subfs'%s'for %v", "object", dExtDirFS)
		return
	}
	objectExtensions, err = extensionFactory.LoadExtensions(subFS)
	if err != nil {
		err = errors.Wrapf(err, "cannot load extension folder %v", ofs)
		return
	}
	return
}

func OpenRO(ocflPath string, logger *logging.Logger) (ocfl.OCFLFS, error) {
	var ocfs ocfl.OCFLFS
	var err error

	var zipSize int64
	var zipReader *os.File
	var zipWriter *os.File

	var zipFile string
	//var objectPath string
	if strings.HasSuffix(strings.ToLower(ocflPath), ".zip") {
		zipFile = ocflPath
	} else {
		if pos := strings.Index(ocflPath, ".zip/"); pos != -1 {
			zipFile = (ocflPath)[0 : pos+4]
			//objectPath = (*target)[pos+4:]
		}
	}
	if zipFile != "" {
		stat, err := os.Stat(zipFile)
		if err != nil {
			log.Print(errors.Wrapf(err, "%s does not exist. creating new file", zipFile))
		} else {
			zipSize = stat.Size()
			if zipReader, err = os.Open(zipFile); err != nil {
				return nil, errors.Wrapf(err, "cannot open zipfile %s", zipFile)
			}
		}
		ocfs, err = zipfs.NewFSIO(zipReader, zipSize, zipWriter, ".", logger)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot create zipfs")
		}
	} else {
		ocfs, err = osfs.NewFSIO(ocflPath, logger)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot create osfs")
		}
	}
	return ocfs, nil
}

func OpenRW(ocflPath string, logger *logging.Logger) (ocfl.OCFLFS, error) {
	var ocfs ocfl.OCFLFS
	var err error

	var zipSize int64
	var zipReader *os.File
	var zipWriter *os.File

	ocflPath = filepath.ToSlash(filepath.Clean(ocflPath))

	if strings.HasSuffix(strings.ToLower(ocflPath), ".zip") {
		stat, err := os.Stat(ocflPath)
		if err != nil {
			if !os.IsNotExist(err) {
				log.Print(errors.Wrapf(err, "%s does not exist. creating new file", ocflPath))
			}
		} else {
			zipSize = stat.Size()
			if zipReader, err = os.Open(ocflPath); err != nil {
				return nil, errors.Wrapf(err, "cannot open zipfile %s", ocflPath)
			}
		}
		tempFile := fmt.Sprintf("%s.tmp", ocflPath)
		if zipWriter, err = os.Create(tempFile); err != nil {
			logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
			panic(err)
		}

		ocfs, err = zipfs.NewFSIO(zipReader, zipSize, zipWriter, ".", logger)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot create zipfs")
		}
	} else {
		ocfs, err = osfs.NewFSIO(ocflPath, logger)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot create osfs")
		}
	}
	return ocfs, nil
}

func showStatus(ctx context.Context) error {
	status, err := ocfl.GetValidationStatus(ctx)
	if err != nil {
		return errors.Wrap(err, "cannot get status of validation")
	}
	status.Compact()
	context := ""
	errs := 0
	for _, err := range status.Errors {
		if err.Code[0] == 'E' {
			errs++
		}
		if err.Context != context {
			fmt.Printf("\n[%s]\n", err.Context)
			context = err.Context
		}
		fmt.Printf("   #%s - %s [%s]\n", err.Code, err.Description, err.Description2)
		//logger.Infof("ERROR: %v", err)
	}
	if errs > 0 {
		fmt.Printf("\n%d errors found\n", errs)
	} else {
		fmt.Printf("\nno errors found\n")
	}
	/*
		for _, err := range status.Warnings {
			if err.Context != context {
				fmt.Printf("\n[%s]\n", err.Context)
				context = err.Context
			}
			fmt.Printf("   Validation Warning #%s - %s [%s]\n", err.Code, err.Description, err.Description2)
			//logger.Infof("WARN:  %v", err)
		}
		fmt.Println("\n")

	*/
	return nil
}

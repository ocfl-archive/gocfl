package cmd

import (
	"context"
	"emperror.dev/errors"
	"fmt"
	"github.com/op/go-logging"
	"github.com/spf13/viper"
	defaultextensions_object "go.ub.unibas.ch/gocfl/v2/data/defaultextensions/object"
	defaultextensions_storageroot "go.ub.unibas.ch/gocfl/v2/data/defaultextensions/storageroot"
	"go.ub.unibas.ch/gocfl/v2/pkg/baseFS"
	"go.ub.unibas.ch/gocfl/v2/pkg/baseFS/genericfs"
	"go.ub.unibas.ch/gocfl/v2/pkg/baseFS/osfs"
	"go.ub.unibas.ch/gocfl/v2/pkg/baseFS/s3fs"
	"go.ub.unibas.ch/gocfl/v2/pkg/baseFS/zipfs"
	"go.ub.unibas.ch/gocfl/v2/pkg/checksum"
	"go.ub.unibas.ch/gocfl/v2/pkg/extension"
	"go.ub.unibas.ch/gocfl/v2/pkg/ocfl"
	"io/fs"
	"os"
)

func initExtensionFactory(logger *logging.Logger, params map[string]map[string]string) (*ocfl.ExtensionFactory, error) {
	extensionFactory, err := ocfl.NewExtensionFactory(logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot instantiate extension factory")
	}

	extensionFactory.AddCreator(extension.DirectCleanName, func(fsys ocfl.OCFLFSRead) (ocfl.Extension, error) {
		return extension.NewDirectCleanFS(fsys)
	})

	extensionFactory.AddCreator(extension.PathDirectName, func(fsys ocfl.OCFLFSRead) (ocfl.Extension, error) {
		return extension.NewPathDirectFS(fsys)
	})

	extensionFactory.AddCreator(extension.StorageLayoutFlatDirectName, func(fsys ocfl.OCFLFSRead) (ocfl.Extension, error) {
		return extension.NewStorageLayoutFlatDirectFS(fsys)
	})

	extensionFactory.AddCreator(extension.StorageLayoutHashAndIdNTupleName, func(fsys ocfl.OCFLFSRead) (ocfl.Extension, error) {
		return extension.NewStorageLayoutHashAndIdNTupleFS(fsys)
	})

	extensionFactory.AddCreator(extension.StorageLayoutHashedNTupleName, func(fsys ocfl.OCFLFSRead) (ocfl.Extension, error) {
		return extension.NewStorageLayoutHashedNTupleFS(fsys)
	})

	extensionFactory.AddCreator(extension.StorageLayoutPairTreeName, func(fsys ocfl.OCFLFSRead) (ocfl.Extension, error) {
		return extension.NewStorageLayoutPairTreeFS(fsys)
	})

	extensionFactory.AddCreator(ocfl.ExtensionManagerName, func(fsys ocfl.OCFLFSRead) (ocfl.Extension, error) {
		return ocfl.NewInitialDummyFS(fsys)
	})

	extensionFactory.AddCreator(extension.ContentSubPathName, func(fsys ocfl.OCFLFSRead) (ocfl.Extension, error) {
		return extension.NewContentSubPathFS(fsys)
	})

	extensionFactory.AddCreator(extension.MetaFileName, func(fsys ocfl.OCFLFSRead) (ocfl.Extension, error) {
		ps, ok := params[extension.MetaFileName]
		if !ok {
			return nil, errors.Errorf("no flags or config entries for extension '%s'", extension.MetaFileName)
		}
		return extension.NewMetaFileFS(fsys, ps)
	})

	return extensionFactory, nil
}

func GetExtensionParams() map[string][]ocfl.ExtensionExternalParam {
	var result = map[string][]ocfl.ExtensionExternalParam{}

	result[extension.IndexerName] = extension.GetIndexerParams()
	result[extension.MetaFileName] = extension.GetMetaFileParams()

	return result
}

func initDefaultExtensions(extensionFactory *ocfl.ExtensionFactory, storageRootExtensionsFolder, objectExtensionsFolder string, logger *logging.Logger) (storageRootExtensions, objectExtensions []ocfl.Extension, err error) {
	var dStorageRootExtDirFS, dObjectExtDirFS fs.FS
	if storageRootExtensionsFolder == "" {
		dStorageRootExtDirFS = defaultextensions_storageroot.DefaultStorageRootExtensionFS
	} else {
		dStorageRootExtDirFS = os.DirFS(storageRootExtensionsFolder)
	}
	osrfs, err := genericfs.NewFS(dStorageRootExtDirFS, ".", logger)
	if err != nil {
		err = errors.Wrapf(err, "cannot create generic fs for %v", dStorageRootExtDirFS)
		return
	}
	if objectExtensionsFolder == "" {
		dObjectExtDirFS = defaultextensions_object.DefaultObjectExtensionFS
	} else {
		dObjectExtDirFS = os.DirFS(objectExtensionsFolder)
	}
	oofs, err := genericfs.NewFS(dObjectExtDirFS, ".", logger)
	if err != nil {
		err = errors.Wrapf(err, "cannot create generic fs for %v", dObjectExtDirFS)
		return
	}
	storageRootExtensions, err = extensionFactory.LoadExtensions(osrfs)
	if err != nil {
		err = errors.Wrapf(err, "cannot load extension folder %v", osrfs)
		return
	}
	objectExtensions, err = extensionFactory.LoadExtensions(oofs)
	if err != nil {
		err = errors.Wrapf(err, "cannot load extension folder %v", oofs)
		return
	}
	return
}

func initializeFSFactory(logger *logging.Logger) (*baseFS.Factory, error) {
	fsFactory, err := baseFS.NewFactory()
	if err != nil {
		return nil, errors.Wrap(err, "cannot create filesystem factory")
	}

	zipFS, err := zipfs.NewBaseFS(logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create zip base filesystem factory")
	}
	fsFactory.Add(zipFS)

	// do S3 FS base instance
	endpoint := viper.GetString("S3Endpoint")
	accessKeyID := viper.GetString("S3AccessKeyID")
	secretAccessKey := viper.GetString("S3SecretAccessKey")
	s3FS, err := s3fs.NewBaseFS(endpoint, accessKeyID, secretAccessKey, "", true, logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create s3 base filesystem factory")
	}
	fsFactory.Add(s3FS)

	osFS, err := osfs.NewBaseFS(logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create os base filesystem factory")
	}
	fsFactory.Add(osFS)

	return fsFactory, nil
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

func addObjectByPath(
	storageRoot ocfl.StorageRoot,
	fixity []checksum.DigestAlgorithm,
	defaultExtensions []ocfl.Extension,
	checkDuplicates bool,
	id, userName, userAddress, message string,
	sourceFS ocfl.OCFLFSRead, area string,
	areaPaths map[string]ocfl.OCFLFSRead,
	echo bool) (bool, error) {
	var o ocfl.Object
	exists, err := storageRoot.ObjectExists(flagObjectID)
	if err != nil {
		return false, errors.Wrapf(err, "cannot check for existence of %s", id)
	}
	if exists {
		o, err = storageRoot.LoadObjectByID(id)
		if err != nil {
			return false, errors.Wrapf(err, "cannot load object %s", id)
		}
	} else {
		o, err = storageRoot.CreateObject(id, storageRoot.GetVersion(), storageRoot.GetDigest(), fixity, defaultExtensions)
		if err != nil {
			return false, errors.Wrapf(err, "cannot create object %s", id)
		}
	}
	if err := o.StartUpdate(message, userName, userAddress, echo); err != nil {
		return false, errors.Wrapf(err, "cannot start update for object %s", id)
	}

	if err := o.AddFolder(sourceFS, checkDuplicates, area); err != nil {
		return false, errors.Wrapf(err, "cannot add folder '%s' to '%s'", sourceFS, id)
	}
	if areaPaths != nil {
		for a, aPath := range areaPaths {
			if err := o.AddFolder(aPath, checkDuplicates, a); err != nil {
				return false, errors.Wrapf(err, "cannot add area '%s' folder '%s' to '%s'", a, aPath, id)
			}
		}
	}

	if err := o.Close(); err != nil {
		return false, errors.Wrapf(err, "cannot close object '%s'", id)
	}

	return o.IsModified(), nil
}

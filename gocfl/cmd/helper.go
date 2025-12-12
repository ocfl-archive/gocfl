package cmd

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"time"

	"emperror.dev/errors"
	"github.com/je4/filesystem/v3/pkg/osfsrw"
	"github.com/je4/filesystem/v3/pkg/s3fsrw"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/je4/filesystem/v3/pkg/zipfs"
	"github.com/je4/filesystem/v3/pkg/zipfsrw"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/keepass2kms"
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v2/config"
	defaultextensions_object "github.com/ocfl-archive/gocfl/v2/data/defaultextensions/object"
	defaultextensions_storageroot "github.com/ocfl-archive/gocfl/v2/data/defaultextensions/storageroot"
	ocflextension "github.com/ocfl-archive/gocfl/v2/pkg/extension"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/storageroot"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
	"github.com/ocfl-archive/gocfl/v2/pkg/subsystem/migration"
	"github.com/ocfl-archive/gocfl/v2/pkg/subsystem/thumbnail"
	ironmaiden "github.com/ocfl-archive/indexer/v3/pkg/indexer"
	"github.com/spf13/cobra"
	"github.com/tink-crypto/tink-go/v2/core/registry"
)

func startTimer() *timer {
	t := &timer{}
	t.Start()
	return t
}

type timer struct {
	start time.Time
}

func (t *timer) Start() {
	t.start = time.Now()
}

func (t *timer) String() string {
	delta := time.Now().Sub(t.start)
	return delta.String()
}

func InitExtensionFactory(extensionParams map[string]string, indexerAddr string, indexerLocalCache bool, indexerActions *ironmaiden.ActionDispatcher, migration *migration.Migration, thumbnail *thumbnail.Thumbnail, sourceFS fs.FS, logger zLogger.ZLogger) (*extension.ExtensionFactory, error) {
	logger.Debug().Msgf("initializing ExtensionFactory")
	extensionFactory, err := extension.NewExtensionFactory(extensionParams, logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot instantiate extension factory")
	}

	logger.Debug().Msgf("adding creator for extension %s", ocflextension.InitialName)
	extensionFactory.AddCreator(ocflextension.InitialName, func(fsys fs.FS) (extension.Extension, error) {
		return ocflextension.NewInitialFS(fsys)
	})

	logger.Debug().Msgf("adding creator for extension %s", ocflextension.GOCFLExtensionManagerName)
	extensionFactory.AddCreator(ocflextension.GOCFLExtensionManagerName, func(fsys fs.FS) (extension.Extension, error) {
		// return ocfl.NewInitialDummyFS(fsys)
		return ocflextension.NewGOCFLExtensionManagerFS(fsys)
	})

	logger.Debug().Msgf("adding creator for extension %s", ocflextension.DigestAlgorithmsName)
	extensionFactory.AddCreator(ocflextension.DigestAlgorithmsName, func(fsys fs.FS) (extension.Extension, error) {
		return ocflextension.NewDigestAlgorithmsFS(fsys)
	})

	logger.Debug().Msgf("adding creator for extension %s", ocflextension.StorageLayoutFlatDirectName)
	extensionFactory.AddCreator(ocflextension.StorageLayoutFlatDirectName, func(fsys fs.FS) (extension.Extension, error) {
		return ocflextension.NewStorageLayoutFlatDirectFS(fsys)
	})

	logger.Debug().Msgf("adding creator for extension %s", ocflextension.StorageLayoutHashAndIdNTupleName)
	extensionFactory.AddCreator(ocflextension.StorageLayoutHashAndIdNTupleName, func(fsys fs.FS) (extension.Extension, error) {
		return ocflextension.NewStorageLayoutHashAndIdNTupleFS(fsys)
	})

	logger.Debug().Msgf("adding creator for extension %s", ocflextension.StorageLayoutHashedNTupleName)
	extensionFactory.AddCreator(ocflextension.StorageLayoutHashedNTupleName, func(fsys fs.FS) (extension.Extension, error) {
		return ocflextension.NewStorageLayoutHashedNTupleFS(fsys)
	})

	logger.Debug().Msgf("adding creator for extension %s", ocflextension.FlatOmitPrefixStorageLayoutName)
	extensionFactory.AddCreator(ocflextension.FlatOmitPrefixStorageLayoutName, func(fsys fs.FS) (extension.Extension, error) {
		return ocflextension.NewFlatOmitPrefixStorageLayoutFS(fsys)
	})

	logger.Debug().Msgf("adding creator for extension %s", ocflextension.NTupleOmitPrefixStorageLayoutName)
	extensionFactory.AddCreator(ocflextension.NTupleOmitPrefixStorageLayoutName, func(fsys fs.FS) (extension.Extension, error) {
		return ocflextension.NewNTupleOmitPrefixStorageLayoutFS(fsys)
	})

	logger.Debug().Msgf("adding creator for extension %s", ocflextension.DirectCleanName)
	extensionFactory.AddCreator(ocflextension.DirectCleanName, func(fsys fs.FS) (extension.Extension, error) {
		return ocflextension.NewDirectCleanFS(fsys)
	})

	logger.Debug().Msgf("adding creator for extension %s", ocflextension.LegacyDirectCleanName)
	extensionFactory.AddCreator(ocflextension.LegacyDirectCleanName, func(fsys fs.FS) (extension.Extension, error) {
		return ocflextension.NewLegacyDirectCleanFS(fsys)
	})

	logger.Debug().Msgf("adding creator for extension %s", ocflextension.PathDirectName)
	extensionFactory.AddCreator(ocflextension.PathDirectName, func(fsys fs.FS) (extension.Extension, error) {
		return ocflextension.NewPathDirectFS(fsys)
	})

	logger.Debug().Msgf("adding creator for extension %s", ocflextension.StorageLayoutPairTreeName)
	extensionFactory.AddCreator(ocflextension.StorageLayoutPairTreeName, func(fsys fs.FS) (extension.Extension, error) {
		return ocflextension.NewStorageLayoutPairTreeFS(fsys)
	})

	logger.Debug().Msgf("adding creator for extension %s", ocflextension.ContentSubPathName)
	extensionFactory.AddCreator(ocflextension.ContentSubPathName, func(fsys fs.FS) (extension.Extension, error) {
		return ocflextension.NewContentSubPathFS(fsys)
	})

	logger.Debug().Msgf("adding creator for extension %s", ocflextension.MetaFileName)
	extensionFactory.AddCreator(ocflextension.MetaFileName, func(fsys fs.FS) (extension.Extension, error) {
		return ocflextension.NewMetaFileFS(fsys)
	})

	logger.Debug().Msgf("adding creator for extension %s", ocflextension.TimestampName)
	extensionFactory.AddCreator(ocflextension.TimestampName, func(fsys fs.FS) (extension.Extension, error) {
		return ocflextension.NewTimestampFS(fsys, logger)
	})

	logger.Debug().Msgf("adding creator for extension %s", ocflextension.IndexerName)
	extensionFactory.AddCreator(ocflextension.IndexerName, func(fsys fs.FS) (extension.Extension, error) {
		ext, err := ocflextension.NewIndexerFS(fsys, indexerAddr, indexerActions, indexerLocalCache, logger)
		if err != nil {
			return nil, errors.Wrap(err, "cannot create new indexer from filesystem")
		}
		return ext, nil
	})

	logger.Debug().Msgf("adding creator for extension %s", ocflextension.MigrationName)
	extensionFactory.AddCreator(ocflextension.MigrationName, func(fsys fs.FS) (extension.Extension, error) {
		return ocflextension.NewMigrationFS(fsys, migration, logger)
	})

	logger.Debug().Msgf("adding creator for extension %s", ocflextension.ThumbnailName)
	extensionFactory.AddCreator(ocflextension.ThumbnailName, func(fsys fs.FS) (extension.Extension, error) {
		return ocflextension.NewThumbnailFS(fsys, thumbnail, logger)
	})

	logger.Debug().Msgf("adding creator for extension %s", ocflextension.FilesystemName)
	extensionFactory.AddCreator(ocflextension.FilesystemName, func(fsys fs.FS) (extension.Extension, error) {
		return ocflextension.NewFilesystemFS(fsys, logger)
	})

	logger.Debug().Msgf("adding creator for extension %s", ocflextension.METSName)
	extensionFactory.AddCreator(ocflextension.METSName, func(fsys fs.FS) (extension.Extension, error) {
		return ocflextension.NewMetsFS(fsys, logger)
	})

	return extensionFactory, nil
}

func GetExtensionParams() []*extension.ExtensionExternalParam {
	var result = []*extension.ExtensionExternalParam{}

	result = append(result, ocflextension.GetIndexerParams()...)
	result = append(result, ocflextension.GetMetaFileParams()...)
	result = append(result, ocflextension.GetMetsParams()...)
	result = append(result, ocflextension.GetContentSubPathParams()...)
	result = append(result, ocflextension.GetTimestampParams()...)

	return result
}

func GetExtensionParamValues(cmd *cobra.Command, conf *config.GOCFLConfig) map[string]string {
	var result = map[string]string{}
	extParams := GetExtensionParams()
	for _, param := range extParams {
		name, value := param.GetParam(cmd, conf)
		if name != "" {
			result[name] = value
		}
	}
	return result
}

func initDefaultExtensions(extensionFactory *extension.ExtensionFactory, storageRootExtensionsFolder, objectExtensionsFolder string, logger zLogger.ZLogger) (storageRootExtensions storageroot.ExtensionManager, objectExtensions object.ExtensionManager, err error) {
	var dStorageRootExtDirFS, dObjectExtDirFS fs.FS
	if storageRootExtensionsFolder == "" {
		dStorageRootExtDirFS = defaultextensions_storageroot.DefaultStorageRootExtensionFS
	} else {
		dStorageRootExtDirFS, err = osfsrw.NewFS(storageRootExtensionsFolder, true, logger)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "cannot create filesystem for storage root extensions folder %v", storageRootExtensionsFolder)
		}
	}
	if objectExtensionsFolder == "" {
		dObjectExtDirFS = defaultextensions_object.DefaultObjectExtensionFS
	} else {
		dObjectExtDirFS, err = osfsrw.NewFS(objectExtensionsFolder, true, logger)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "cannot create filesystem for object extensions folder %v", objectExtensionsFolder)
		}
	}
	_storageRootExtensions, err := extensionFactory.LoadExtensions(dStorageRootExtDirFS, nil)
	if err != nil {
		err = errors.Wrapf(err, "cannot load extension folder %v", dStorageRootExtDirFS)
		return
	}
	_objectExtensions, err := extensionFactory.LoadExtensions(dObjectExtDirFS, nil)
	if err != nil {
		err = errors.Wrapf(err, "cannot load extension folder %v", dObjectExtDirFS)
		return
	}
	return _storageRootExtensions.(storageroot.ExtensionManager), _objectExtensions.(object.ExtensionManager), nil
}

func initializeFSFactory(zipDigests []checksum.DigestAlgorithm, aesConfig *config.AESConfig, s3Config *config.S3Config, noCompression, readOnly bool, logger zLogger.ZLogger) (*writefs.Factory, error) {
	if zipDigests == nil {
		zipDigests = []checksum.DigestAlgorithm{checksum.DigestSHA512}
	}
	if aesConfig == nil {
		aesConfig = &config.AESConfig{}
	}
	if s3Config == nil {
		s3Config = &config.S3Config{}
	}

	fsFactory, err := writefs.NewFactory()
	if err != nil {
		return nil, errors.Wrap(err, "cannot create filesystem factory")
	}

	if readOnly {
		if err := fsFactory.Register(zipfs.NewCreateFSFunc(logger), "\\.zip$", writefs.HighFS); err != nil {
			return nil, errors.Wrap(err, "cannot register zipfs")
		}
	} else {
		// todo: allow different KMS clients
		if aesConfig.Enable {
			db, err := keepass2kms.LoadKeePassDBFromFile(string(aesConfig.KeepassFile), string(aesConfig.KeepassKey))
			if err != nil {
				return nil, errors.Wrapf(err, "cannot load keepass file '%s'", aesConfig.KeepassFile)
			}
			client, err := keepass2kms.NewClient(db, filepath.Base(string(aesConfig.KeepassFile)))
			if err != nil {
				return nil, errors.Wrap(err, "cannot create keepass2kms client")
			}
			registry.RegisterKMSClient(client)

			if err := fsFactory.Register(zipfsrw.NewCreateFSEncryptedChecksumFunc(noCompression, zipDigests, string(aesConfig.KeepassEntry), logger), "\\.zip$", writefs.HighFS); err != nil {
				return nil, errors.Wrap(err, "cannot register FSEncryptedChecksum")
			}
		} else {
			if err := fsFactory.Register(zipfsrw.NewCreateFSChecksumFunc(noCompression, zipDigests, logger), "\\.zip$", writefs.HighFS); err != nil {
				return nil, errors.Wrap(err, "cannot register FSChecksum")
			}
		}
	}
	if err := fsFactory.Register(osfsrw.NewCreateFSFunc(logger), "", writefs.LowFS); err != nil {
		return nil, errors.Wrap(err, "cannot register osfs")
	}
	if s3Config.Endpoint != "" {
		if err := fsFactory.Register(
			s3fsrw.NewCreateFSFunc(
				map[string]*s3fsrw.S3Access{
					"switch": {
						string(s3Config.AccessKeyID),
						string(s3Config.AccessKey),
						string(s3Config.Endpoint),
						true,
					},
				},
				s3fsrw.ARNRegexStr,
				false,
				nil,
				"",
				"",
				logger,
			),
			s3fsrw.ARNRegexStr,
			writefs.MediumFS,
		); err != nil {
			return nil, errors.Wrap(err, "cannot register s3fs")
		}
	}
	return fsFactory, nil
}

func showStatus(ctx context.Context, logger zLogger.ZLogger) error {
	status, err := validation.GetValidationStatus(ctx)
	if err != nil {
		return errors.Wrap(err, "cannot get status of validation")
	}
	status.Compact()
	contextString := ""
	errs := 0
	for _, err := range status.Errors {
		if err.Code[0] == 'E' {
			errs++
		}
		if err.Context != contextString {
			fmt.Printf("\n[%s]\n", err.Context)
			contextString = err.Context
		}
		fmt.Printf("   #%s - %s [%s]\n", err.Code, err.Description, err.Description2)
		//logger.Info().Msgf("ERROR: %v", err)
	}
	if errs > 0 {
		fmt.Printf("\n%d errors found\n", errs)
	} else {
		fmt.Printf("\nno errors found\n")
	}
	return nil
}

func LoadObjectByID(sr storageroot.StorageRoot, extensionFactory *extension.ExtensionFactory, id string, logger zLogger.ZLogger) (object.Object, error) {
	folder, err := sr.IdToFolder(id)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot load object %s", id)
	}
	fsys, err := writefs.Sub(sr.GetFS(), folder)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create subfs for %v / %s", sr.GetFS(), folder)
	}
	obj, err := object.LoadObject(context.Background(), fsys, extensionFactory, logger)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot load object %s", id)
	}
	return obj, nil
}

func addObjectByPath(
	sr storageroot.StorageRoot,
	fixity []checksum.DigestAlgorithm,
	extensionFactory *extension.ExtensionFactory,
	extensionManager object.ExtensionManager,
	checkDuplicates bool,
	id, userName, userAddress, message string,
	sourceFS fs.FS, area string,
	areaPaths map[string]fs.FS,
	echo bool,
	logger zLogger.ZLogger,
) (bool, error) {
	if fixity == nil {
		fixity = []checksum.DigestAlgorithm{}
	}
	var o object.Object
	exists, err := sr.ObjectExists(flagObjectID)
	if err != nil {
		return false, errors.Wrapf(err, "cannot check for existence of %s", id)
	}
	if exists {
		o, err = LoadObjectByID(sr, extensionFactory, id, logger)
		if err != nil {
			return false, errors.Wrapf(err, "cannot load object %s", id)
		}
		// if we update, fixity is taken from last object version
		f := o.GetInventory().GetFixity()
		for alg, _ := range f {
			fixity = append(fixity, alg)
		}
	} else {
		o, err = object.CreateObject(context.Background(), id, sr.GetVersion(), sr.GetDigest(), fixity, extensionFactory, extensionManager, sr.GetFS(), logger)
		if err != nil {
			return false, errors.Wrapf(err, "cannot create object %s", id)
		}
	}
	versionFS, err := o.StartUpdate(sourceFS, message, userName, userAddress, echo)
	if err != nil {
		return false, errors.Wrapf(err, "cannot start update for object %s", id)
	}

	if err := o.AddFolder(sourceFS, versionFS, checkDuplicates, area); err != nil {
		return false, errors.Wrapf(err, "cannot add folder '%s' to '%s'", sourceFS, id)
	}
	if areaPaths != nil {
		for a, aPath := range areaPaths {
			if err := o.AddFolder(aPath, versionFS, checkDuplicates, a); err != nil {
				return false, errors.Wrapf(err, "cannot add area '%s' folder '%s' to '%s'", a, aPath, id)
			}
		}
	}
	if err := o.EndUpdate(); err != nil {
		return false, errors.Wrapf(err, "cannot end update for object '%s'", id)
	}

	if err := o.Close(); err != nil {
		return false, errors.Wrapf(err, "cannot close object '%s'", id)
	}

	return o.IsModified(), nil
}

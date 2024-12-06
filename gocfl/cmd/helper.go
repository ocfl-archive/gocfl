package cmd

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"time"

	"emperror.dev/errors"
	"github.com/google/tink/go/core/registry"
	"github.com/je4/filesystem/v3/pkg/osfsrw"
	"github.com/je4/filesystem/v3/pkg/s3fsrw"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/je4/filesystem/v3/pkg/zipfs"
	"github.com/je4/filesystem/v3/pkg/zipfsrw"
	ironmaiden "github.com/je4/indexer/v3/pkg/indexer"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/keepass2kms"
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v2/config"
	defaultextensions_object "github.com/ocfl-archive/gocfl/v2/data/defaultextensions/object"
	defaultextensions_storageroot "github.com/ocfl-archive/gocfl/v2/data/defaultextensions/storageroot"
	"github.com/ocfl-archive/gocfl/v2/pkg/extension"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl"
	"github.com/ocfl-archive/gocfl/v2/pkg/subsystem/migration"
	"github.com/ocfl-archive/gocfl/v2/pkg/subsystem/thumbnail"
	"github.com/spf13/cobra"
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

// InitExtensionFactory initializes the extension factory so that they
// can be called upon within the primary GOCL runner.
func InitExtensionFactory(extensionParams map[string]string, indexerAddr string, indexerLocalCache bool, indexerActions *ironmaiden.ActionDispatcher, migration *migration.Migration, thumbnail *thumbnail.Thumbnail, sourceFS fs.FS, logger zLogger.ZLogger) (*ocfl.ExtensionFactory, error) {
	logger.Debug().Msgf("initializing ExtensionFactory")
	extensionFactory, err := ocfl.NewExtensionFactory(extensionParams, logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot instantiate extension factory")
	}

	logger.Debug().Msgf("adding creator for extension %s", extension.InitialName)
	extensionFactory.AddCreator(extension.InitialName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewInitialFS(fsys)
	})

	logger.Debug().Msgf("adding creator for extension %s", extension.GOCFLExtensionManagerName)
	extensionFactory.AddCreator(extension.GOCFLExtensionManagerName, func(fsys fs.FS) (ocfl.Extension, error) {
		// return ocfl.NewInitialDummyFS(fsys)
		return extension.NewGOCFLExtensionManagerFS(fsys)
	})

	logger.Debug().Msgf("adding creator for extension %s", extension.DigestAlgorithmsName)
	extensionFactory.AddCreator(extension.DigestAlgorithmsName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewDigestAlgorithmsFS(fsys)
	})

	logger.Debug().Msgf("adding creator for extension %s", extension.StorageLayoutFlatDirectName)
	extensionFactory.AddCreator(extension.StorageLayoutFlatDirectName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewStorageLayoutFlatDirectFS(fsys)
	})

	logger.Debug().Msgf("adding creator for extension %s", extension.StorageLayoutHashAndIdNTupleName)
	extensionFactory.AddCreator(extension.StorageLayoutHashAndIdNTupleName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewStorageLayoutHashAndIdNTupleFS(fsys)
	})

	logger.Debug().Msgf("adding creator for extension %s", extension.StorageLayoutHashedNTupleName)
	extensionFactory.AddCreator(extension.StorageLayoutHashedNTupleName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewStorageLayoutHashedNTupleFS(fsys)
	})

	logger.Debug().Msgf("adding creator for extension %s", extension.FlatOmitPrefixStorageLayoutName)
	extensionFactory.AddCreator(extension.FlatOmitPrefixStorageLayoutName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewFlatOmitPrefixStorageLayoutFS(fsys)
	})

	logger.Debug().Msgf("adding creator for extension %s", extension.NTupleOmitPrefixStorageLayoutName)
	extensionFactory.AddCreator(extension.NTupleOmitPrefixStorageLayoutName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewNTupleOmitPrefixStorageLayoutFS(fsys)
	})

	logger.Debug().Msgf("adding creator for extension %s", extension.DirectCleanName)
	extensionFactory.AddCreator(extension.DirectCleanName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewDirectCleanFS(fsys)
	})

	logger.Debug().Msgf("adding creator for extension %s", extension.LegacyDirectCleanName)
	extensionFactory.AddCreator(extension.LegacyDirectCleanName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewLegacyDirectCleanFS(fsys)
	})

	logger.Debug().Msgf("adding creator for extension %s", extension.PathDirectName)
	extensionFactory.AddCreator(extension.PathDirectName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewPathDirectFS(fsys)
	})

	logger.Debug().Msgf("adding creator for extension %s", extension.StorageLayoutPairTreeName)
	extensionFactory.AddCreator(extension.StorageLayoutPairTreeName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewStorageLayoutPairTreeFS(fsys)
	})

	logger.Debug().Msgf("adding creator for extension %s", extension.ContentSubPathName)
	extensionFactory.AddCreator(extension.ContentSubPathName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewContentSubPathFS(fsys)
	})

	logger.Debug().Msgf("adding creator for extension %s", extension.MetaFileName)
	extensionFactory.AddCreator(extension.MetaFileName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewMetaFileFS(fsys)
	})

	logger.Debug().Msgf("adding creator for extension %s", extension.IndexerName)
	extensionFactory.AddCreator(extension.IndexerName, func(fsys fs.FS) (ocfl.Extension, error) {
		ext, err := extension.NewIndexerFS(
			fsys, indexerAddr, indexerActions, indexerLocalCache, logger, ErrorFactory,
		)
		if err != nil {
			return nil, errors.Wrap(err, "cannot create new indexer from filesystem")
		}
		return ext, nil
	})

	logger.Debug().Msgf("adding creator for extension %s", extension.MigrationName)
	extensionFactory.AddCreator(extension.MigrationName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewMigrationFS(fsys, migration, logger)
	})

	logger.Debug().Msgf("adding creator for extension %s", extension.ThumbnailName)
	extensionFactory.AddCreator(extension.ThumbnailName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewThumbnailFS(fsys, thumbnail, logger)
	})

	logger.Debug().Msgf("adding creator for extension %s", extension.FilesystemName)
	extensionFactory.AddCreator(extension.FilesystemName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewFilesystemFS(fsys, logger)
	})

	logger.Debug().Msgf("adding creator for extension %s", extension.METSName)
	extensionFactory.AddCreator(extension.METSName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewMetsFS(fsys, logger)
	})

	return extensionFactory, nil
}

func GetExtensionParams() []*ocfl.ExtensionExternalParam {
	var result = []*ocfl.ExtensionExternalParam{}

	result = append(result, extension.GetIndexerParams()...)
	result = append(result, extension.GetMetaFileParams()...)
	result = append(result, extension.GetMetsParams()...)
	result = append(result, extension.GetContentSubPathParams()...)

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

func initDefaultExtensions(extensionFactory *ocfl.ExtensionFactory, storageRootExtensionsFolder, objectExtensionsFolder string, logger zLogger.ZLogger) (storageRootExtensions, objectExtensions ocfl.ExtensionManager, err error) {
	var dStorageRootExtDirFS, dObjectExtDirFS fs.FS
	if storageRootExtensionsFolder == "" {
		dStorageRootExtDirFS = defaultextensions_storageroot.DefaultStorageRootExtensionFS
	} else {
		dStorageRootExtDirFS, err = osfsrw.NewFS(storageRootExtensionsFolder, logger)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "cannot create filesystem for storage root extensions folder %v", storageRootExtensionsFolder)
		}
	}
	if objectExtensionsFolder == "" {
		dObjectExtDirFS = defaultextensions_object.DefaultObjectExtensionFS
	} else {
		dObjectExtDirFS, err = osfsrw.NewFS(objectExtensionsFolder, logger)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "cannot create filesystem for object extensions folder %v", objectExtensionsFolder)
		}
	}
	storageRootExtensions, err = extensionFactory.LoadExtensions(dStorageRootExtDirFS, nil)
	if err != nil {
		err = errors.Wrapf(err, "cannot load extension folder %v", dStorageRootExtDirFS)
		return
	}
	objectExtensions, err = extensionFactory.LoadExtensions(dObjectExtDirFS, nil)
	if err != nil {
		err = errors.Wrapf(err, "cannot load extension folder %v", dObjectExtDirFS)
		return
	}
	return
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
	status, err := ocfl.GetValidationStatus(ctx)
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

func addObjectByPath(
	storageRoot ocfl.StorageRoot,
	fixity []checksum.DigestAlgorithm,
	extensionManager ocfl.ExtensionManager,
	checkDuplicates bool,
	id, userName, userAddress, message string,
	sourceFS fs.FS, area string,
	areaPaths map[string]fs.FS,
	echo bool) (bool, error) {
	if fixity == nil {
		fixity = []checksum.DigestAlgorithm{}
	}
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
		// if we update, fixity is taken from last object version
		f := o.GetInventory().GetFixity()
		for alg, _ := range f {
			fixity = append(fixity, alg)
		}
	} else {
		o, err = storageRoot.CreateObject(id, storageRoot.GetVersion(), storageRoot.GetDigest(), fixity, extensionManager)
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

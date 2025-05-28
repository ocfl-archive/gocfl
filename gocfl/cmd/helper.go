package cmd

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"time"

	"github.com/je4/filesystem/v3/pkg/osfsrw"
	"github.com/je4/filesystem/v3/pkg/s3fsrw"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/je4/filesystem/v3/pkg/zipfs"
	"github.com/je4/filesystem/v3/pkg/zipfsrw"
	statickms "github.com/je4/utils/v2/pkg/StaticKMS"
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

func logExtInit(logger zLogger.ZLogger, ext string) {
	logger.Debug().Any(
		ErrorFactory.LogError(
			ErrorExtensionInit,
			fmt.Sprintf("adding creator for extension: %s", ext), nil),
	).Msg("")
}

// InitExtensionFactory initializes the extension factory so that they
// can be called upon within the primary GOCL runner.
func InitExtensionFactory(
	extensionParams map[string]string,
	indexerAddr string,
	indexerLocalCache bool,
	indexerActions *ironmaiden.ActionDispatcher,
	migration *migration.Migration,
	thumbnail *thumbnail.Thumbnail,
	sourceFS fs.FS,
	logger zLogger.ZLogger,
	tempDir string,
) (*ocfl.ExtensionFactory, error) {

	logger.Debug().Any(
		ErrorFactory.LogError(ErrorExtensionInit, "initializing ExtensionFactory", nil),
	).Msg("")

	extensionFactory, err := ocfl.NewExtensionFactory(extensionParams, logger)
	if err != nil {
		err = ErrorFactory.NewError(ErrorExtensionInitErr, "cannot instantiate extension factory", err)
		return nil, err
	}

	logExtInit(logger, extension.InitialName)
	extensionFactory.AddCreator(extension.InitialName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewInitialFS(fsys)
	})

	logExtInit(logger, extension.GOCFLExtensionManagerName)
	extensionFactory.AddCreator(extension.GOCFLExtensionManagerName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewGOCFLExtensionManagerFS(fsys)
	})

	logExtInit(logger, extension.DigestAlgorithmsName)
	extensionFactory.AddCreator(extension.DigestAlgorithmsName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewDigestAlgorithmsFS(fsys)
	})

	logExtInit(logger, extension.StorageLayoutFlatDirectName)
	extensionFactory.AddCreator(extension.StorageLayoutFlatDirectName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewStorageLayoutFlatDirectFS(fsys)
	})

	logExtInit(logger, extension.StorageLayoutHashAndIdNTupleName)
	extensionFactory.AddCreator(extension.StorageLayoutHashAndIdNTupleName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewStorageLayoutHashAndIdNTupleFS(fsys)
	})

	logExtInit(logger, extension.StorageLayoutHashedNTupleName)
	extensionFactory.AddCreator(extension.StorageLayoutHashedNTupleName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewStorageLayoutHashedNTupleFS(fsys)
	})

	logExtInit(logger, extension.FlatOmitPrefixStorageLayoutName)
	extensionFactory.AddCreator(extension.FlatOmitPrefixStorageLayoutName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewFlatOmitPrefixStorageLayoutFS(fsys)
	})

	logExtInit(logger, extension.NTupleOmitPrefixStorageLayoutName)
	extensionFactory.AddCreator(extension.NTupleOmitPrefixStorageLayoutName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewNTupleOmitPrefixStorageLayoutFS(fsys)
	})

	logExtInit(logger, extension.DirectCleanName)
	extensionFactory.AddCreator(extension.DirectCleanName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewDirectCleanFS(fsys)
	})

	logExtInit(logger, extension.LegacyDirectCleanName)
	extensionFactory.AddCreator(extension.LegacyDirectCleanName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewLegacyDirectCleanFS(fsys)
	})

	logExtInit(logger, extension.PathDirectName)
	extensionFactory.AddCreator(extension.PathDirectName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewPathDirectFS(fsys)
	})

	logExtInit(logger, extension.StorageLayoutPairTreeName)
	extensionFactory.AddCreator(extension.StorageLayoutPairTreeName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewStorageLayoutPairTreeFS(fsys)
	})

	logExtInit(logger, extension.ContentSubPathName)
	extensionFactory.AddCreator(extension.ContentSubPathName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewContentSubPathFS(fsys)
	})

	logExtInit(logger, extension.MetaFileName)
	extensionFactory.AddCreator(extension.MetaFileName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewMetaFileFS(fsys)
	})

	logExtInit(logger, extension.ROCrateFileName)
	extensionFactory.AddCreator(extension.ROCrateFileName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewROCrateFileFS(fsys)
	})

	logExtInit(logger, extension.IndexerName)
	extensionFactory.AddCreator(extension.IndexerName, func(fsys fs.FS) (ocfl.Extension, error) {
		ext, err := extension.NewIndexerFS(
			fsys, indexerAddr, indexerActions, indexerLocalCache, logger, ErrorFactory, tempDir,
		)
		if err != nil {
			err = ErrorFactory.NewError(ErrorExtensionInitErr, "cannot create new indexer from filesystem", err)
			return nil, err
		}
		return ext, nil
	})

	logExtInit(logger, extension.MigrationName)
	extensionFactory.AddCreator(extension.MigrationName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewMigrationFS(fsys, migration, logger, tempDir)
	})

	logExtInit(logger, extension.ThumbnailName)
	extensionFactory.AddCreator(extension.ThumbnailName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewThumbnailFS(fsys, thumbnail, logger, ErrorFactory, tempDir)
	})

	logExtInit(logger, extension.FilesystemName)
	extensionFactory.AddCreator(extension.FilesystemName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewFilesystemFS(fsys, logger)
	})

	logExtInit(logger, extension.METSName)
	extensionFactory.AddCreator(extension.METSName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewMetsFS(fsys, logger)
	})
	logExtInit(logger, extension.TimestampName)
	extensionFactory.AddCreator(extension.TimestampName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewTimestampFS(fsys, logger)
	})

	return extensionFactory, nil
}

func GetExtensionParams() []*ocfl.ExtensionExternalParam {
	var result = []*ocfl.ExtensionExternalParam{}
	result = append(result, extension.GetIndexerParams()...)
	result = append(result, extension.GetMetaFileParams()...)
	result = append(result, extension.GetMetsParams()...)
	result = append(result, extension.GetROCrateFileParams()...)
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

func initDefaultExtensions(
	extensionFactory *ocfl.ExtensionFactory,
	storageRootExtensionsFolder,
	objectExtensionsFolder string,
	logger zLogger.ZLogger,
) (storageRootExtensions, objectExtensions ocfl.ExtensionManager, err error) {
	var dStorageRootExtDirFS, dObjectExtDirFS fs.FS
	if storageRootExtensionsFolder == "" {
		dStorageRootExtDirFS = defaultextensions_storageroot.DefaultStorageRootExtensionFS
	} else {
		dStorageRootExtDirFS, err = osfsrw.NewFS(storageRootExtensionsFolder, true, logger)
		if err != nil {
			err = ErrorFactory.NewError(
				ErrorExtensionInitErr,
				fmt.Sprintf("cannot create filesystem for storage root extensions folder %v", storageRootExtensionsFolder),
				err,
			)
			return nil, nil, err
		}
	}
	if objectExtensionsFolder == "" {
		dObjectExtDirFS = defaultextensions_object.DefaultObjectExtensionFS
	} else {
		dObjectExtDirFS, err = osfsrw.NewFS(objectExtensionsFolder, true, logger)
		if err != nil {
			err = ErrorFactory.NewError(
				ErrorExtensionInitErr,
				fmt.Sprintf("cannot create filesystem for object extensions folder %v", objectExtensionsFolder),
				err,
			)
			return nil, nil, err
		}
	}
	storageRootExtensions, err = extensionFactory.LoadExtensions(dStorageRootExtDirFS, nil)
	if err != nil {
		err = ErrorFactory.NewError(
			ErrorExtensionInitErr,
			fmt.Sprintf("cannot load extension folder %v", dStorageRootExtDirFS),
			err,
		)
		return nil, nil, err
	}
	objectExtensions, err = extensionFactory.LoadExtensions(dObjectExtDirFS, nil)
	if err != nil {
		err = ErrorFactory.NewError(
			ErrorExtensionInitErr,
			fmt.Sprintf("cannot load extension folder %v", dObjectExtDirFS),
			err,
		)
		return nil, nil, err
	}
	return
}

func initializeFSFactory(zipDigests []checksum.DigestAlgorithm, aesConfig *config.AESConfig, s3Config *config.S3Config, noCompression, readOnly bool, logger zLogger.ZLogger) (*writefs.Factory, error) {
	//return nil, errors.Wrapf(errors.New("testing 123"), "cannot create filesystem factory")
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
		err = ErrorFactory.NewError(
			ErrorFS, "cannot create filesystem factory", err,
		)
		return nil, err
	}

	if readOnly {
		if err := fsFactory.Register(zipfs.NewCreateFSFunc(logger), "\\.zip$", writefs.HighFS); err != nil {
			err = ErrorFactory.NewError(
				ErrorFS, "cannot register zipfs", err,
			)
			return nil, err
		}
	} else {
		if aesConfig.Enable {
			var client registry.KMSClient
			if aesConfig.Key != "" {
				logger.Info().Msgf("using static KMS client")
				client, err = statickms.NewClient(string(aesConfig.Key))
				if err != nil {
					err = ErrorFactory.NewError(
						ErrorFS,
						"cannot create static kms client",
						err,
					)
					return nil, err
				}
			} else {
				logger.Info().Msgf("using keepass2kms client with file '%s'", aesConfig.KeepassFile)
				db, err := keepass2kms.LoadKeePassDBFromFile(string(aesConfig.KeepassFile), string(aesConfig.KeepassKey))
				if err != nil {
					err = ErrorFactory.NewError(
						ErrorFS,
						fmt.Sprintf("cannot load keepass file '%s'", aesConfig.KeepassFile),
						err,
					)
					return nil, err
				}
				client, err = keepass2kms.NewClient(db, filepath.Base(string(aesConfig.KeepassFile)))
				if err != nil {
					err = ErrorFactory.NewError(
						ErrorFS,
						"cannot create keepass2kms client",
						err,
					)
					return nil, err
				}
			}
			registry.RegisterKMSClient(client)

			if err := fsFactory.Register(zipfsrw.NewCreateFSEncryptedChecksumFunc(noCompression, zipDigests, string(aesConfig.KeepassEntry), logger), "\\.zip$", writefs.HighFS); err != nil {
				err = ErrorFactory.NewError(
					ErrorFS,
					"cannot register FSEncryptedChecksum",
					err,
				)
				return nil, err
			}
		} else {
			if err := fsFactory.Register(zipfsrw.NewCreateFSChecksumFunc(noCompression, zipDigests, logger), "\\.zip$", writefs.HighFS); err != nil {
				err = ErrorFactory.NewError(
					ErrorFS,
					"cannot register FSChecksum",
					err,
				)
				return nil, err
			}
		}
	}
	if err := fsFactory.Register(osfsrw.NewCreateFSFunc(logger), "", writefs.LowFS); err != nil {
		err = ErrorFactory.NewError(
			ErrorFS, "cannot register osfs", err,
		)
		return nil, err
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
			err = ErrorFactory.NewError(
				ErrorFS,
				"cannot register s3fs",
				err,
			)
			return nil, err
		}
	}
	return fsFactory, nil
}

// showStatus providees a way to list status errors on close of the
// GOCFL object.
func showStatus(ctx context.Context, logger zLogger.ZLogger) error {
	status, err := ocfl.GetValidationStatus(ctx)
	if err != nil {
		err = ErrorFactory.NewError(ErrorValidationStatus, "cannot get status of validation", err)
		return err
	}
	status.Compact()
	contextString := ""
	errs := 0
	for _, err := range status.Errors {
		if err.Code[0] == 'E' {
			errs++
		}
		if err.Context != contextString {
			logger.Info().Any(
				ErrorFactory.LogError(
					ErrorValidationStatus,
					fmt.Sprintf("[%s]", err.Context),
					nil,
				),
			).Msg("")
			contextString = err.Context
		}

		logger.Info().Any(
			ErrorFactory.LogError(
				ErrorValidationStatus,
				fmt.Sprintf("#%s - %s [%s]", err.Code, err.Description, err.Description2),
				nil,
			),
		).Msg("")
	}
	if errs > 0 {
		logger.Error().Any(
			ErrorFactory.LogError(
				ErrorValidationStatus,
				fmt.Sprintf("'%d'errors found", errs),
				nil,
			),
		).Msg("")
	} else {
		logger.Info().Any(
			ErrorFactory.LogError(
				ErrorValidationStatus, "no errors found", nil,
			),
		).Msg("")
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
	echo bool,
	logger zLogger.ZLogger) (bool, error) {
	if fixity == nil {
		fixity = []checksum.DigestAlgorithm{}
	}
	var o ocfl.Object
	exists, err := storageRoot.ObjectExists(flagObjectID)
	if err != nil {
		logger.Error().Any(
			errorTopic,
			ErrorFactory.NewError(
				ErrorOCFLCreation,
				fmt.Sprintf("cannot check for existence of %s", id),
				err,
			)).Msg("")
		err = ErrorFactory.NewError(
			ErrorOCFLCreation,
			fmt.Sprintf("cannot check for existence of %s", id),
			err,
		)
		return false, err
	}
	if exists {
		o, err = storageRoot.LoadObjectByID(id)
		if err != nil {
			logger.Error().Any(
				errorTopic,
				ErrorFactory.NewError(
					ErrorOCFLCreation,
					fmt.Sprintf("cannot load object %s", id),
					err,
				)).Msg("")
			err = ErrorFactory.NewError(
				ErrorOCFLCreation,
				fmt.Sprintf("cannot load object %s", id),
				err,
			)
			return false, err
		}
		// if we update, fixity is taken from last object version
		f := o.GetInventory().GetFixity()
		for alg, _ := range f {
			fixity = append(fixity, alg)
		}
	} else {
		o, err = storageRoot.CreateObject(id, storageRoot.GetVersion(), storageRoot.GetDigest(), fixity, extensionManager)
		if err != nil {
			logger.Error().Any(
				errorTopic,
				ErrorFactory.NewError(
					ErrorOCFLCreation,
					fmt.Sprintf("cannot create object %s", id),
					err,
				)).Msg("")
			err = ErrorFactory.NewError(
				ErrorOCFLCreation,
				fmt.Sprintf("cannot create object %s", id),
				err,
			)
			return false, err
		}
	}
	err = o.StartUpdate(sourceFS, message, userName, userAddress, echo)
	if err != nil {
		logger.Error().Any(
			errorTopic,
			ErrorFactory.NewError(
				ErrorOCFLCreation,
				fmt.Sprintf("cannot start update for object %s", id),
				err,
			)).Msg("")
		err = ErrorFactory.NewError(
			ErrorOCFLCreation,
			fmt.Sprintf("cannot start update for object %s", id),
			err,
		)
		return false, err
	}
	if err := o.AddFolder(sourceFS, checkDuplicates, area); err != nil {
		logger.Error().Any(
			errorTopic,
			ErrorFactory.NewError(
				ErrorOCFLCreation,
				fmt.Sprintf("cannot add folder %s to %s", sourceFS, id),
				err,
			)).Msg("")
		err = ErrorFactory.NewError(
			ErrorOCFLCreation,
			fmt.Sprintf("cannot add folder '%s' to '%s'", sourceFS, id),
			err,
		)
		return false, err
	}
	if areaPaths != nil {
		for a, aPath := range areaPaths {
			if err := o.AddFolder(aPath, checkDuplicates, a); err != nil {
				logger.Error().Any(
					errorTopic,
					ErrorFactory.NewError(
						ErrorOCFLCreation,
						fmt.Sprintf("cannot add area '%s' folder '%s' to '%s'", a, aPath, id),
						err,
					)).Msg("")
				err = ErrorFactory.NewError(
					ErrorOCFLCreation,
					fmt.Sprintf("cannot add area '%s' folder '%s' to '%s'", a, aPath, id),
					err,
				)
				return false, err
			}
		}
	}
	if err := o.EndUpdate(); err != nil {
		logger.Error().Any(
			errorTopic,
			ErrorFactory.NewError(
				ErrorOCFLCreation,
				fmt.Sprintf("cannot end update for object %s", id),
				err,
			)).Msg("")
		err = ErrorFactory.NewError(
			ErrorOCFLEnd,
			fmt.Sprintf("cannot end update for object '%s'", id),
			err,
		)
		return false, err
	}
	if err := o.Close(); err != nil {
		logger.Error().Any(
			errorTopic,
			ErrorFactory.NewError(
				ErrorOCFLCreation,
				fmt.Sprintf("cannot close object %s", id),
				err,
			)).Msg("")
		err = ErrorFactory.NewError(
			ErrorOCFLEnd,
			fmt.Sprintf("cannot close object '%s'", id),
			err,
		)
		return false, err
	}
	return o.IsModified(), nil
}

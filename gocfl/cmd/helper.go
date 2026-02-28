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
	"github.com/ocfl-archive/gocfl/v2/config"
	defaultextensions_object "github.com/ocfl-archive/gocfl/v2/data/defaultextensions/object"
	defaultextensions_storageroot "github.com/ocfl-archive/gocfl/v2/data/defaultextensions/storageroot"
	ocflextension "github.com/ocfl-archive/gocfl/v2/pkg/extension"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension/extensionimpl"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/factory/factoryimpl"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/functions"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/ocflerrors"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/storageroot"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/util"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfllogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/streamfs"
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

func InitExtensionFactory(extensionParams map[string]string, indexerAddr string, indexerLocalCache bool, indexerActions *ironmaiden.ActionDispatcher, migration *migration.Migration, thumbnail *thumbnail.Thumbnail, sourceFS fs.FS, logger ocfllogger.OCFLLogger) (*extensionimpl.Factory, error) {
	logger.Debug().Msgf("initializing ExtensionFactory")
	extensionFactory, err := extensionimpl.NewFactory(extensionParams, logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot instantiate extension factory")
	}

	logger.Debug().Msgf("adding creator for extension %s", ocflextension.InitialName)
	extensionFactory.AddCreator(ocflextension.InitialName, func(fsys fs.FS) (extension.Extension, error) {
		initial := ocflextension.NewInitial(logger)
		if err := initial.Load(fsys); err != nil {
			return nil, errors.Wrap(err, "cannot load initial extension")
		}
		return initial, nil
	})

	logger.Debug().Msgf("adding creator for extension %s", ocflextension.GOCFLExtensionManagerName)
	extensionFactory.AddCreator(ocflextension.GOCFLExtensionManagerName, func(fsys fs.FS) (extension.Extension, error) {
		ext := ocflextension.NewGOCFLExtensionManager(logger)
		if err := ext.Load(fsys); err != nil {
			return nil, errors.Wrap(err, "cannot load gocfl extension manager")
		}
		return ext, nil
	})

	logger.Debug().Msgf("adding creator for extension %s", ocflextension.DigestAlgorithmsName)
	extensionFactory.AddCreator(ocflextension.DigestAlgorithmsName, func(fsys fs.FS) (extension.Extension, error) {
		ext := ocflextension.NewDigestAlgorithms(logger)
		if err := ext.Load(fsys); err != nil {
			return nil, errors.Wrap(err, "cannot load digest algorithms extension")
		}
		return ext, nil
	})

	logger.Debug().Msgf("adding creator for extension %s", ocflextension.StorageLayoutFlatDirectName)
	extensionFactory.AddCreator(ocflextension.StorageLayoutFlatDirectName, func(fsys fs.FS) (extension.Extension, error) {
		ext := ocflextension.NewStorageLayoutFlatDirect(logger)
		if err := ext.Load(fsys); err != nil {
			return nil, errors.Wrap(err, "cannot load storage layout flat direct extension")
		}
		return ext, nil
	})

	logger.Debug().Msgf("adding creator for extension %s", ocflextension.StorageLayoutHashAndIdNTupleName)
	extensionFactory.AddCreator(ocflextension.StorageLayoutHashAndIdNTupleName, func(fsys fs.FS) (extension.Extension, error) {
		ext := ocflextension.NewStorageLayoutHashAndIdNTuple(logger)
		if err := ext.Load(fsys); err != nil {
			return nil, errors.Wrap(err, "cannot load storage layout hash and id n-tuple extension")
		}
		return ext, nil
	})

	logger.Debug().Msgf("adding creator for extension %s", ocflextension.StorageLayoutHashedNTupleName)
	extensionFactory.AddCreator(ocflextension.StorageLayoutHashedNTupleName, func(fsys fs.FS) (extension.Extension, error) {
		ext := ocflextension.NewStorageLayoutHashedNTuple(logger)
		if err := ext.Load(fsys); err != nil {
			return nil, errors.Wrap(err, "cannot load storage layout hashed n-tuple extension")
		}
		return ext, nil
	})

	logger.Debug().Msgf("adding creator for extension %s", ocflextension.FlatOmitPrefixStorageLayoutName)
	extensionFactory.AddCreator(ocflextension.FlatOmitPrefixStorageLayoutName, func(fsys fs.FS) (extension.Extension, error) {
		ext := ocflextension.NewFlatOmitPrefixStorageLayout(logger)
		if err := ext.Load(fsys); err != nil {
			return nil, errors.Wrap(err, "cannot load flat omit prefix storage layout extension")
		}
		return ext, nil
	})

	logger.Debug().Msgf("adding creator for extension %s", ocflextension.NTupleOmitPrefixStorageLayoutName)
	extensionFactory.AddCreator(ocflextension.NTupleOmitPrefixStorageLayoutName, func(fsys fs.FS) (extension.Extension, error) {
		ext := ocflextension.NewNTupleOmitPrefixStorageLayout(logger)
		if err := ext.Load(fsys); err != nil {
			return nil, errors.Wrap(err, "cannot load n-tuple omit prefix storage layout extension")
		}
		return ext, nil
	})

	logger.Debug().Msgf("adding creator for extension %s", ocflextension.DirectCleanName)
	extensionFactory.AddCreator(ocflextension.DirectCleanName, func(fsys fs.FS) (extension.Extension, error) {
		ext := ocflextension.NewDirectClean(logger)
		if err := ext.Load(fsys); err != nil {
			return nil, errors.Wrap(err, "cannot load direct clean extension")
		}
		return ext, nil
	})

	logger.Debug().Msgf("adding creator for extension %s", ocflextension.LegacyDirectCleanName)
	extensionFactory.AddCreator(ocflextension.LegacyDirectCleanName, func(fsys fs.FS) (extension.Extension, error) {
		ext := ocflextension.NewLegacyDirectClean(logger)
		if err := ext.Load(fsys); err != nil {
			return nil, errors.Wrap(err, "cannot load legacy direct clean extension")
		}
		return ext, nil
	})

	logger.Debug().Msgf("adding creator for extension %s", ocflextension.StorageLayoutPairTreeName)
	extensionFactory.AddCreator(ocflextension.StorageLayoutPairTreeName, func(fsys fs.FS) (extension.Extension, error) {
		ext := ocflextension.NewStorageLayoutPairTree(logger)
		if err := ext.Load(fsys); err != nil {
			return nil, errors.Wrap(err, "cannot load storage layout pair tree extension")
		}
		return ext, nil
	})

	logger.Debug().Msgf("adding creator for extension %s", ocflextension.ContentSubPathName)
	extensionFactory.AddCreator(ocflextension.ContentSubPathName, func(fsys fs.FS) (extension.Extension, error) {
		ext := ocflextension.NewContentSubPath(logger)
		if err := ext.Load(fsys); err != nil {
			return nil, errors.Wrap(err, "cannot load content sub path extension")
		}
		return ext, nil
	})

	logger.Debug().Msgf("adding creator for extension %s", ocflextension.MetaFileName)
	extensionFactory.AddCreator(ocflextension.MetaFileName, func(fsys fs.FS) (extension.Extension, error) {
		ext := ocflextension.NewMetaFile(logger, nil)
		if err := ext.Load(fsys); err != nil {
			return nil, errors.Wrap(err, "cannot load meta file extension")
		}
		return ext, nil
	})

	logger.Debug().Msgf("adding creator for extension %s", ocflextension.TimestampName)
	extensionFactory.AddCreator(ocflextension.TimestampName, func(fsys fs.FS) (extension.Extension, error) {
		ext := ocflextension.NewTimestamp(logger)
		if err := ext.Load(fsys); err != nil {
			return nil, errors.Wrap(err, "cannot load timestamp extension")
		}
		return ext, nil
	})

	logger.Debug().Msgf("adding creator for extension %s", ocflextension.IndexerName)
	extensionFactory.AddCreator(ocflextension.IndexerName, func(fsys fs.FS) (extension.Extension, error) {
		ext := ocflextension.NewIndexer(logger, indexerAddr, indexerActions, indexerLocalCache)
		if err := ext.Load(fsys); err != nil {
			return nil, errors.Wrap(err, "cannot create new indexer from filesystem")
		}
		return ext, nil
	})

	logger.Debug().Msgf("adding creator for extension %s", ocflextension.MigrationName)
	extensionFactory.AddCreator(ocflextension.MigrationName, func(fsys fs.FS) (extension.Extension, error) {
		ext := ocflextension.NewMigration(logger, migration)
		if err := ext.Load(fsys); err != nil {
			return nil, errors.Wrap(err, "cannot load migration extension")
		}
		return ext, nil
	})

	logger.Debug().Msgf("adding creator for extension %s", ocflextension.ThumbnailName)
	extensionFactory.AddCreator(ocflextension.ThumbnailName, func(fsys fs.FS) (extension.Extension, error) {
		ext := ocflextension.NewThumbnail(logger, thumbnail)
		if err := ext.Load(fsys); err != nil {
			return nil, errors.Wrap(err, "cannot load thumbnail extension")
		}
		return ext, nil
	})

	logger.Debug().Msgf("adding creator for extension %s", ocflextension.FilesystemName)
	extensionFactory.AddCreator(ocflextension.FilesystemName, func(fsys fs.FS) (extension.Extension, error) {
		ext := ocflextension.NewFilesystem(logger)
		if err := ext.Load(fsys); err != nil {
			return nil, errors.Wrap(err, "cannot load filesystem extension")
		}
		return ext, nil
	})

	logger.Debug().Msgf("adding creator for extension %s", ocflextension.METSName)
	extensionFactory.AddCreator(ocflextension.METSName, func(fsys fs.FS) (extension.Extension, error) {
		ext := ocflextension.NewMets(logger)
		if err := ext.Load(fsys); err != nil {
			return nil, errors.Wrap(err, "cannot load mets extension")
		}
		return ext, nil
	})

	return extensionFactory, nil
}

func GetExtensionParams() []*extensionimpl.ExtensionExternalParam {
	var result = []*extensionimpl.ExtensionExternalParam{}

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

func initDefaultExtensions(ver version.OCFLVersion, extensionFactory *extensionimpl.Factory, storageRootExtensionsFolder, objectExtensionsFolder string, logger ocfllogger.OCFLLogger) (storageRootExtensions storageroot.ExtensionManager, objectExtensions object.ExtensionManager, err error) {
	var dStorageRootExtDirFS, dObjectExtDirFS fs.FS
	if storageRootExtensionsFolder == "" {
		dStorageRootExtDirFS = defaultextensions_storageroot.DefaultStorageRootExtensionFS
	} else {
		dStorageRootExtDirFS, err = osfsrw.NewFS(storageRootExtensionsFolder, true, logger.Logger())
		if err != nil {
			return nil, nil, errors.Wrapf(err, "cannot create filesystem for storage root extensions folder %v", storageRootExtensionsFolder)
		}
	}
	if objectExtensionsFolder == "" {
		dObjectExtDirFS = defaultextensions_object.DefaultObjectExtensionFS
	} else {
		dObjectExtDirFS, err = osfsrw.NewFS(objectExtensionsFolder, true, logger.Logger())
		if err != nil {
			return nil, nil, errors.Wrapf(err, "cannot create filesystem for object extensions folder %v", objectExtensionsFolder)
		}
	}
	_storageRootExtensions, err := extensionFactory.LoadExtensions(dStorageRootExtDirFS, ver)
	if err != nil {
		err = errors.Wrapf(err, "cannot load extension folder %v", dStorageRootExtDirFS)
		return
	}
	_objectExtensions, err := extensionFactory.LoadExtensions(dObjectExtDirFS, ver)
	if err != nil {
		err = errors.Wrapf(err, "cannot load extension folder %v", dObjectExtDirFS)
		return
	}
	return _storageRootExtensions.(storageroot.ExtensionManager), _objectExtensions.(object.ExtensionManager), nil
}

func initializeFSFactory(zipDigests []checksum.DigestAlgorithm, aesConfig *config.AESConfig, s3Config *config.S3Config, noCompression, readOnly bool, logger ocfllogger.OCFLLogger) (*writefs.Factory, error) {
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
		if err := fsFactory.Register(zipfs.NewCreateFSFunc(logger.Logger()), "\\.zip$", writefs.HighFS); err != nil {
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

			if err := fsFactory.Register(zipfsrw.NewCreateFSEncryptedChecksumFunc(noCompression, zipDigests, string(aesConfig.KeepassEntry), logger.Logger()), "\\.zip$", writefs.HighFS); err != nil {
				return nil, errors.Wrap(err, "cannot register FSEncryptedChecksum")
			}
		} else {
			if err := fsFactory.Register(zipfsrw.NewCreateFSChecksumFunc(noCompression, zipDigests, logger.Logger()), "\\.zip$", writefs.HighFS); err != nil {
				return nil, errors.Wrap(err, "cannot register FSChecksum")
			}
		}
	}
	if err := fsFactory.Register(osfsrw.NewCreateFSFunc(logger.Logger()), "", writefs.LowFS); err != nil {
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
				logger.Logger(),
			),
			s3fsrw.ARNRegexStr,
			writefs.MediumFS,
		); err != nil {
			return nil, errors.Wrap(err, "cannot register s3fs")
		}
	}
	return fsFactory, nil
}

func showStatus(ctx context.Context, logger ocfllogger.OCFLLogger) error {
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

func LoadObjectByID(sr storageroot.StorageRoot, extensionFactory *extensionimpl.Factory, id string, logger ocfllogger.OCFLLogger) (object.Object, error) {
	folder, err := sr.IdToFolder(id)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot load object %s", id)
	}
	fsys, err := writefs.Sub(sr.GetReadFS(), folder)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create subfs for %v / %s", sr.GetReadFS(), folder)
	}
	obj, err := functions.LoadObject(context.Background(), fsys, extensionFactory, logger)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot load object %s", id)
	}
	return obj, nil
}

func addObjectByPath(
	ctx context.Context,
	sr storageroot.StorageRoot,
	fixity []checksum.DigestAlgorithm,
	extensionFactory *extensionimpl.Factory,
	extensionManager object.ExtensionManager,
	checkDuplicates bool,
	id, userName, userAddress, message string,
	sourceFS fs.FS, area string,
	areaPaths map[string]fs.FS,
	echo bool,
	logger ocfllogger.OCFLLogger,
) (bool, error) {
	if fixity == nil {
		fixity = []checksum.DigestAlgorithm{}
	}
	var o object.Object
	objPath, err := sr.IdToFolder(id)
	if err != nil {
		return false, errors.Wrapf(err, "cannot create folder for id %s", id)
	}
	objectFS, err := streamfs.Sub(sr.GetWriteFS(), objPath)
	if err != nil {
		return false, errors.Wrapf(err, "cannot create subfs %v / %s for id %s", sr.GetWriteFS(), objPath, id)
	}
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
		for alg := range f.GetDigestAlgorithms() {
			fixity = append(fixity, alg)
		}
	} else {
		o, err = functions.CreateObject(ctx, id, sr.GetVersion(), sr.GetDigest(), fixity, extensionFactory, extensionManager, objectFS, logger)
		if err != nil {
			return false, errors.Wrapf(err, "cannot create object %s", id)
		}
	}
	versionWriter, err := o.StartUpdate(objectFS, message, userName, userAddress, echo)
	if err != nil {
		return false, errors.Wrapf(err, "cannot start update for object %s", id)
	}
	defer func() {
		if versionWriter != nil {
			versionWriter.Close()
		}
	}()
	if err := versionWriter.AddFolder(sourceFS, checkDuplicates, area); err != nil {
		return false, errors.Wrapf(err, "cannot add folder '%s' to '%s'", sourceFS, id)
	}
	if areaPaths != nil {
		for a, aPath := range areaPaths {
			if err := versionWriter.AddFolder(aPath, checkDuplicates, a); err != nil {
				return false, errors.Wrapf(err, "cannot add area '%s' folder '%s' to '%s'", a, aPath, id)
			}
		}
	}
	if err := versionWriter.Close(); err != nil {
		return false, errors.Wrapf(err, "cannot close version writer for object %s", id)
	}
	versionWriter = nil
	return o.GetInventory().IsModified(), nil
}

func CreateStorageRoot(ctx context.Context, objectWriteFS streamfs.FS, ver version.OCFLVersion, extensionFactory *extensionimpl.Factory, extensionManager storageroot.ExtensionManager, digest checksum.DigestAlgorithm, logger ocfllogger.OCFLLogger) (storageroot.StorageRoot, error) {
	fact := factoryimpl.NewFactory(ver, extensionFactory, logger)
	storageRoot := fact.NewStorageRoot(ctx).WithReadFS(objectWriteFS).WithWriteFS(objectWriteFS).WithExtensionManager(extensionManager).WithDigestAlgorithm(digest)

	init := storageRoot.GetInitializer()
	defer init.Close()
	if err := init.Init(); err != nil {
		return nil, errors.Wrap(err, "cannot initialize storage root")
	}

	return storageRoot, nil
}

func LoadStorageRoot(ctx context.Context, storageRootFS streamfs.FS, extensionFactory *extensionimpl.Factory, logger ocfllogger.OCFLLogger) (storageroot.StorageRoot, error) {
	ver, err := util.GetVersion(storageRootFS, ".", "ocfl_")
	if err != nil && !errors.Is(err, ocflerrors.ErrVersionNone) {
		return nil, errors.WithStack(err)
	}
	if ver == "" {
		dirs, err := fs.ReadDir(storageRootFS, ".")
		if err != nil {
			return nil, errors.WithStack(err)
		}
		if len(dirs) > 0 {
			logger.ValidationError(validation.E069, "storage root %s not empty without version information", storageRootFS)
		}
		ver = version.Version1_1
	}
	logger.WithVersion(ver)
	fact := factoryimpl.NewFactory(ver, extensionFactory, logger)
	storageRoot := fact.NewStorageRoot(ctx).WithReadFS(storageRootFS).WithWriteFS(storageRootFS)
	loader := storageRoot.GetLoader(extensionFactory)
	defer loader.Close()
	if err := loader.Load(); err != nil {
		return nil, errors.Wrap(err, "cannot load storage root")
	}
	return storageRoot, nil
}

func LoadStorageRootRO(ctx context.Context, storageRootFS fs.FS, extensionFactory *extensionimpl.Factory, logger ocfllogger.OCFLLogger) (storageroot.StorageRoot, error) {
	ver, err := util.GetVersion(storageRootFS, ".", "ocfl_")
	if err != nil && !errors.Is(err, ocflerrors.ErrVersionNone) {
		return nil, errors.WithStack(err)
	}
	if ver == "" {
		dirs, err := fs.ReadDir(storageRootFS, ".")
		if err != nil {
			return nil, errors.WithStack(err)
		}
		if len(dirs) > 0 {
			logger.ValidationError(validation.E069, "storage root %s not empty without version information", storageRootFS)
		}
		ver = version.Version1_1
	}
	logger.WithVersion(ver)
	fact := factoryimpl.NewFactory(ver, extensionFactory, logger)
	storageRoot := fact.NewStorageRoot(ctx).WithReadFS(storageRootFS)
	loader := storageRoot.GetLoader(extensionFactory)
	defer loader.Close()
	if err := loader.Load(); err != nil {
		return nil, errors.Wrap(err, "cannot load storage root")
	}
	return storageRoot, nil
}

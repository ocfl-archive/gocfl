package cmd

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"emperror.dev/errors"
	"github.com/je4/filesystem/v3/pkg/osfsrw"
	"github.com/je4/filesystem/v3/pkg/s3fsrw"
	"github.com/je4/filesystem/v3/pkg/vfsrw"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/je4/filesystem/v3/pkg/zipfs"
	"github.com/je4/filesystem/v3/pkg/zipfsrw"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/keepass2kms"
	"github.com/ocfl-archive/gocfl/v2/config"
	defaultextensions_object "github.com/ocfl-archive/gocfl/v2/data/defaultextensions/object"
	defaultextensions_storageroot "github.com/ocfl-archive/gocfl/v2/data/defaultextensions/storageroot"
	"github.com/ocfl-archive/gocfl/v2/pkg/appendfs"
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
	"github.com/ocfl-archive/gocfl/v2/pkg/subsystem/migration"
	"github.com/ocfl-archive/gocfl/v2/pkg/subsystem/thumbnail"
	ironmaiden "github.com/ocfl-archive/indexer/v3/pkg/indexer"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/spf13/cobra"
	"github.com/tink-crypto/tink-go/v2/core/registry"
)

func firstOrSecond[T any](first bool, a T, b T) T {
	if first {
		return a
	}
	return b
}

type ExtensionManager interface {
	extension.ManagerCore
}

func LoadExtensionManager[T ExtensionManager](fact extension.Factory, fsys fs.FS) (T, error) {
	m, err := fact.LoadExtensionManager(fsys)
	if err != nil {
		var result T
		return result, errors.Wrapf(err, "loading extension manager")
	}
	tVal, ok := m.(T)
	if !ok {
		var result T
		return result, errors.Errorf("failed to cast extension manager to expected type %T", result)
	}
	return tVal, nil
}

type AddFS interface {
	AddFS(name string, fsys fs.FS)
}

func getLocalFSConfig() map[string]*vfsrw.VFS {
	var result = map[string]*vfsrw.VFS{}
	if runtime.GOOS == "windows" {
		partitions, _ := disk.Partitions(false)
		for _, partition := range partitions {
			if len(partition.Mountpoint) < 2 {
				continue
			}
			if partition.Mountpoint[1] != ':' {
				continue
			}
			result[strings.ToLower(partition.Mountpoint[:1])] = &vfsrw.VFS{
				Name:     strings.ToLower(partition.Mountpoint[:1]),
				Type:     "os",
				ReadOnly: false,
				OS: &vfsrw.OS{
					BaseDir:          partition.Mountpoint + "/",
					ZipAsFolderCache: 1,
				},
			}
		}
	} else {
		result["root"] = &vfsrw.VFS{
			Name:     "root",
			Type:     "os",
			ReadOnly: false,
			OS: &vfsrw.OS{
				BaseDir:          "/",
				ZipAsFolderCache: 1,
			},
		}

	}
	return result
}

func resolveExtensionParam(cmd *cobra.Command, name, extensionName, param, defaultValue string) string {
	configValue := conf.Extension[extensionName][param]
	flagValue, _ := cmd.Flags().GetString(name)

	if configValue == "" {
		return flagValue
	}
	if flagValue != "" && flagValue != defaultValue {
		return flagValue
	}
	return configValue
}

func getExtensionParams(cmd *cobra.Command) (map[string]string, error) {
	var extensionParams = map[string]string{}
	if err := extension.GetExtensionParamValues(cmd.Name(), func(name, extensionName, param, defaultValue string) {
		if name == "" {
			return
		}
		extensionParams[name] = resolveExtensionParam(cmd, name, extensionName, param, defaultValue)
	}); err != nil {
		return nil, errors.Wrap(err, "cannot get extension params")
	}
	return extensionParams, nil
}

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

func path2vfs(pathStr string) (string, error) {
	pathStr = filepath.ToSlash(pathStr)
	if runtime.GOOS == "windows" {
		if len(pathStr) > 2 && pathStr[1] == ':' {
			pathStr = "vfs://" + path.Join(strings.ToLower(pathStr[:1]), pathStr[2:])
		} else {
			wd, err := os.Getwd()
			if err != nil {
				return "", errors.Wrap(err, "getting working directory")
			}
			pathStr = path.Join(filepath.ToSlash(wd), pathStr)
			pathStr = "vfs://" + path.Join(strings.ToLower(pathStr[:1]), pathStr[2:])
		}
	} else {
		if pathStr[0] != '/' {
			wd, err := os.Getwd()
			if err != nil {
				return "", errors.Wrap(err, "getting working directory")
			}
			pathStr = path.Join(wd, pathStr)
		}
		pathStr = "vfs://root" + pathStr
	}
	return pathStr, nil
}

func RegisterComplexExtensions(
	fss map[string]fs.FS,
	sourceFS fs.FS,
	indexerAddr string,
	indexerLocalCache bool,
	indexerConf ironmaiden.IndexerConfig,
	migrationConf *config.Migration,
	thumbnailConf *config.Thumbnail,
	logger ocfllogger.OCFLLogger,
) error {

	mig, err := migration.GetMigrations(migrationConf)
	if err != nil {
		return errors.Wrap(err, "cannot get migrations")
	}
	mig.SetSourceFS(sourceFS)

	thumb, err := thumbnail.GetThumbnails(thumbnailConf)
	if err != nil {
		logger.Error().Err(err).Msg("cannot get thumbnails")
		return errors.Wrap(err, "cannot get thumbnails")
	}
	thumb.SetSourceFS(sourceFS)
	extension.RegisterExtension(
		ocflextension.IndexerName,
		func() (extension.Extension, error) {
			ext, err := ocflextension.NewIndexer(indexerAddr, fss, indexerConf, indexerLocalCache, logger)
			if err != nil {
				return nil, err
			}
			return ext, nil
		},
		ocflextension.GetIndexerParams)
	extension.RegisterExtension(
		ocflextension.MigrationName,
		func() (extension.Extension, error) {
			return ocflextension.NewMigration(mig), nil
		},
		nil)
	extension.RegisterExtension(
		ocflextension.ThumbnailName,
		func() (extension.Extension, error) {
			return ocflextension.NewThumbnail(thumb), nil
		},
		nil)

	return nil
}

func InitDefaultExtensions(
	ver version.OCFLVersion,
	extensionFactory *extensionimpl.Factory,
	storageRootExtensionsFolder,
	objectExtensionsFolder string,
	logger ocfllogger.OCFLLogger,
) (storageRootExtensions storageroot.ExtensionManager, objectExtensions object.ExtensionManager, err error) {
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
	_storageRootExtensions, err := extensionFactory.LoadExtensionManager(dStorageRootExtDirFS)
	if err != nil {
		err = errors.Wrapf(err, "cannot load extension folder %v", dStorageRootExtDirFS)
		return
	}
	_objectExtensions, err := extensionFactory.LoadExtensionManager(dObjectExtDirFS)
	if err != nil {
		err = errors.Wrapf(err, "cannot load extension folder %v", dObjectExtDirFS)
		return
	}
	return _storageRootExtensions.(storageroot.ExtensionManager), _objectExtensions.(object.ExtensionManager), nil
}

// todo: use filesystem VFS
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

func showStatus(logger ocfllogger.OCFLLogger) error {
	contextString := ""
	errs := 0
	for _, err := range logger.ValidationErrors() {
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
	var ofs fs.FS = sr.GetWriteFS()
	if ofs == nil {
		ofs = sr.GetReadFS()
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
	objectFS, err := appendfs.Sub(sr.GetWriteFS(), objPath)
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
		if extensionManager == nil {
			return false, errors.New("extension manager is nil")
		}
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
			if err := versionWriter.Close(); err != nil {
				logger.Error().Err(err).Msg("cannot close version writer")
			}
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

func CreateStorageRoot(ctx context.Context, objectWriteFS appendfs.FS, ver version.OCFLVersion, extensionFactory *extensionimpl.Factory, extensionManager storageroot.ExtensionManager, digest checksum.DigestAlgorithm, logger ocfllogger.OCFLLogger) (storageroot.StorageRoot, error) {
	fact := factoryimpl.NewFactory(ver, extensionFactory, logger)
	storageRoot := fact.NewStorageRoot(ctx).WithReadFS(objectWriteFS).WithWriteFS(objectWriteFS).WithExtensionManager(extensionManager).WithDigestAlgorithm(digest)

	init := storageRoot.GetInitializer()
	defer init.Close()
	if err := init.Init(); err != nil {
		return nil, errors.Wrap(err, "cannot initialize storage root")
	}

	return storageRoot, nil
}

func loadStorageRootInternal(
	ctx context.Context,
	readFS fs.FS,
	writeFS appendfs.FS,
	extensionFactory *extensionimpl.Factory,
	logger ocfllogger.OCFLLogger,
) (storageroot.StorageRoot, error) {
	ver, err := util.GetVersion(readFS, ".", "ocfl_")
	if err != nil && !errors.Is(err, ocflerrors.ErrVersionNone) {
		return nil, errors.WithStack(err)
	}
	if ver == "" {
		dirs, err := fs.ReadDir(readFS, ".")
		if err != nil {
			return nil, errors.WithStack(err)
		}
		if len(dirs) > 0 {
			logger.ValidationError(validation.E069, "storage root %s not empty without version information", readFS)
		}
		ver = version.Version1_1
	}

	logger.WithVersion(ver)
	fact := factoryimpl.NewFactory(ver, extensionFactory, logger)

	storageRoot := fact.NewStorageRoot(ctx).WithReadFS(readFS)
	if writeFS != nil {
		storageRoot = storageRoot.WithWriteFS(writeFS)
	}

	loader := storageRoot.GetLoader(extensionFactory)
	defer loader.Close()

	if err := loader.Load(); err != nil {
		return nil, errors.Wrap(err, "cannot load storage root")
	}
	return storageRoot, nil
}

func LoadStorageRoot(
	ctx context.Context,
	storageRootFS appendfs.FS,
	extensionFactory *extensionimpl.Factory,
	logger ocfllogger.OCFLLogger,
) (storageroot.StorageRoot, error) {
	return loadStorageRootInternal(ctx, storageRootFS, storageRootFS, extensionFactory, logger)
}

func LoadStorageRootRO(
	ctx context.Context,
	storageRootFS fs.FS,
	extensionFactory *extensionimpl.Factory,
	logger ocfllogger.OCFLLogger,
) (storageroot.StorageRoot, error) {
	return loadStorageRootInternal(ctx, storageRootFS, nil, extensionFactory, logger)
}

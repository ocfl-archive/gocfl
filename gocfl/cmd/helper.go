package cmd

import (
	"bufio"
	"context"
	"emperror.dev/errors"
	"fmt"
	"github.com/google/tink/go/core/registry"
	"github.com/google/tink/go/tink"
	"github.com/je4/filesystem/v2/pkg/osfsrw"
	"github.com/je4/filesystem/v2/pkg/writefs"
	"github.com/je4/filesystem/v2/pkg/zipfs"
	"github.com/je4/filesystem/v2/pkg/zipfsrw"
	defaultextensions_object "github.com/je4/gocfl/v2/data/defaultextensions/object"
	defaultextensions_storageroot "github.com/je4/gocfl/v2/data/defaultextensions/storageroot"
	"github.com/je4/gocfl/v2/pkg/baseFS"
	"github.com/je4/gocfl/v2/pkg/extension"
	"github.com/je4/gocfl/v2/pkg/migration"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	ironmaiden "github.com/je4/indexer/v2/pkg/indexer"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/keepass2kms"
	"github.com/op/go-logging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
	"time"
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

func initExtensionFactory(extensionParams map[string]string, indexerAddr string, indexerActions *ironmaiden.ActionDispatcher, migration *migration.Migration, sourceFS fs.FS, logger *logging.Logger) (*ocfl.ExtensionFactory, error) {
	logger.Debugf("initializing ExtensionFactory")
	extensionFactory, err := ocfl.NewExtensionFactory(extensionParams, logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot instantiate extension factory")
	}

	logger.Debugf("adding creator for extension %s", extension.DigestAlgorithmsName)
	extensionFactory.AddCreator(extension.DigestAlgorithmsName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewDigestAlgorithmsFS(fsys)
	})

	logger.Debugf("adding creator for extension %s", extension.StorageLayoutFlatDirectName)
	extensionFactory.AddCreator(extension.StorageLayoutFlatDirectName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewStorageLayoutFlatDirectFS(fsys)
	})

	logger.Debugf("adding creator for extension %s", extension.StorageLayoutHashAndIdNTupleName)
	extensionFactory.AddCreator(extension.StorageLayoutHashAndIdNTupleName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewStorageLayoutHashAndIdNTupleFS(fsys)
	})

	logger.Debugf("adding creator for extension %s", extension.StorageLayoutHashedNTupleName)
	extensionFactory.AddCreator(extension.StorageLayoutHashedNTupleName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewStorageLayoutHashedNTupleFS(fsys)
	})

	logger.Debugf("adding creator for extension %s", extension.FlatOmitPrefixStorageLayoutName)
	extensionFactory.AddCreator(extension.FlatOmitPrefixStorageLayoutName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewFlatOmitPrefixStorageLayoutFS(fsys)
	})

	logger.Debugf("adding creator for extension %s", extension.NTupleOmitPrefixStorageLayoutName)
	extensionFactory.AddCreator(extension.NTupleOmitPrefixStorageLayoutName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewNTupleOmitPrefixStorageLayoutFS(fsys)
	})

	logger.Debugf("adding creator for extension %s", extension.DirectCleanName)
	extensionFactory.AddCreator(extension.DirectCleanName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewDirectCleanFS(fsys)
	})

	logger.Debugf("adding creator for extension %s", extension.PathDirectName)
	extensionFactory.AddCreator(extension.PathDirectName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewPathDirectFS(fsys)
	})

	logger.Debugf("adding creator for extension %s", extension.StorageLayoutPairTreeName)
	extensionFactory.AddCreator(extension.StorageLayoutPairTreeName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewStorageLayoutPairTreeFS(fsys)
	})

	logger.Debugf("adding creator for extension %s", ocfl.ExtensionManagerName)
	extensionFactory.AddCreator(ocfl.ExtensionManagerName, func(fsys fs.FS) (ocfl.Extension, error) {
		return ocfl.NewInitialDummyFS(fsys)
	})

	logger.Debugf("adding creator for extension %s", extension.ContentSubPathName)
	extensionFactory.AddCreator(extension.ContentSubPathName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewContentSubPathFS(fsys)
	})

	logger.Debugf("adding creator for extension %s", extension.MetaFileName)
	extensionFactory.AddCreator(extension.MetaFileName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewMetaFileFS(fsys)
	})

	logger.Debugf("adding creator for extension %s", extension.IndexerName)
	extensionFactory.AddCreator(extension.IndexerName, func(fsys fs.FS) (ocfl.Extension, error) {
		ext, err := extension.NewIndexerFS(fsys, indexerAddr, indexerActions)
		if err != nil {
			return nil, errors.Wrap(err, "cannot create new indexer from filesystem")
		}
		return ext, nil
	})

	logger.Debugf("adding creator for extension %s", extension.MigrationName)
	extensionFactory.AddCreator(extension.MigrationName, func(fsys fs.FS) (ocfl.Extension, error) {
		return extension.NewMigrationFS(fsys, migration)
	})

	return extensionFactory, nil
}

func GetExtensionParams() []*ocfl.ExtensionExternalParam {
	var result = []*ocfl.ExtensionExternalParam{}

	result = append(result, extension.GetIndexerParams()...)
	result = append(result, extension.GetMetaFileParams()...)
	result = append(result, extension.GetContentSubPathParams()...)

	return result
}

func GetExtensionParamValues(cmd *cobra.Command) map[string]string {
	var result = map[string]string{}
	extParams := GetExtensionParams()
	for _, param := range extParams {
		name, value := param.GetParam(cmd)
		if name != "" {
			result[name] = value
		}
	}
	return result
}

func initDefaultExtensions(extensionFactory *ocfl.ExtensionFactory, storageRootExtensionsFolder, objectExtensionsFolder string, logger *logging.Logger) (storageRootExtensions, objectExtensions []ocfl.Extension, err error) {
	var dStorageRootExtDirFS, dObjectExtDirFS fs.FS
	if storageRootExtensionsFolder == "" {
		dStorageRootExtDirFS = defaultextensions_storageroot.DefaultStorageRootExtensionFS
	} else {
		dStorageRootExtDirFS = osfsrw.NewFS(storageRootExtensionsFolder)
	}
	if objectExtensionsFolder == "" {
		dObjectExtDirFS = defaultextensions_object.DefaultObjectExtensionFS
	} else {
		dObjectExtDirFS = osfsrw.NewFS(objectExtensionsFolder)
	}
	storageRootExtensions, err = extensionFactory.LoadExtensions(dStorageRootExtDirFS)
	if err != nil {
		err = errors.Wrapf(err, "cannot load extension folder %v", dStorageRootExtDirFS)
		return
	}
	objectExtensions, err = extensionFactory.LoadExtensions(dObjectExtDirFS)
	if err != nil {
		err = errors.Wrapf(err, "cannot load extension folder %v", dObjectExtDirFS)
		return
	}
	return
}

func NewZipFSCreator(factory *baseFS.Factory, zipDigests []checksum.DigestAlgorithm, RW bool, noCompression bool, aes bool, aead tink.AEAD, aad []byte, clear bool) baseFS.CreateFS {
	return func(path string) (fs.FS, error) {
		parts := strings.Split(path, "/")
		if len(parts) < 2 {
			return nil, errors.Errorf("invalid path %s", path)
		}
		folder := strings.Join(parts[0:len(parts)-2], "/")
		zipFile := parts[len(parts)-1]
		baseFSys, err := factory.Get(folder)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot get filesystem for folder %s", folder)
		}
		if RW {

		} else {
			fp, err := baseFSys.Open(zipFile)
			if err != nil {
				return nil, errors.Wrapf(err, "cannot open file %s", zipFile)
			}
			zfs, err := zipfs.NewFS()
		}
	}
}

func initializeFSFactory(prefix string, cmd *cobra.Command, zipDigests []checksum.DigestAlgorithm, logger *logging.Logger) (*baseFS.Factory, error) {
	if zipDigests == nil {
		zipDigests = []checksum.DigestAlgorithm{}
	}
	prefix = strings.TrimRight(prefix, ".") + "."

	fsFactory, err := baseFS.NewFactory()
	if err != nil {
		return nil, errors.Wrap(err, "cannot create filesystem factory")
	}

	flagAES := viper.GetBool(prefix + "AES")

	keePassFile := viper.GetString(prefix + "KeePassFile")
	keePassEntry := viper.GetString(prefix + "KeePassEntry")
	keePassKey := viper.GetString(prefix + "KeePassKey")
	// todo: allow different KMS clients
	if flagAES {
		db, err := keepass2kms.LoadKeePassDBFromFile(keePassFile, keePassKey)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot load keepass file '%s'", keePassFile)
		}
		client, err := keepass2kms.NewClient(db, filepath.Base(keePassFile))
		if err != nil {
			return nil, errors.Wrap(err, "cannot create keepass2kms client")
		}
		registry.RegisterKMSClient(client)
	}

	flagNoCompression := viper.GetBool(prefix + "NoCompression")

	// ----------------------------------------------------
	var fpat io.ReaderAt
	var size int64
	var fp fs.File
	var ok bool
	newpath := path
	// if target file exists, open it and create a zipfs
	stat, err := fs.Stat(baseFS, path)
	if err == nil {
		size = stat.Size()
		fp, err = baseFS.Open(path)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot open zip file '%s'", path)
		}
		fpat, ok = fp.(io.ReaderAt)
		if !ok {
			return nil, errors.Errorf("cannot cast file '%s' to io.WriterAt", path)
		}
		newpath = newpath + ".tmp"
	}
	// create new file
	newfp, err := writefs.Create(baseFS, newpath)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create zip file '%s'", newpath)
	}
	// add a buffer to the file
	newFPBuffer := bufio.NewWriterSize(newfp, 1024*1024)

	// ----------------------------------------------------
	zipFS, err := zipfsrw.NewZipFSRW().NewFS(zipDigests, flagNoCompression, flagAES, keePassEntry, logger)
	//zipFS, err := zipfs.NewBaseFS(zipDigests, flagNoCompression, flagAES, keePassEntry, logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create zip base filesystem factory")
	}
	fsFactory.Add(zipFS)

	// do S3 FS base instance
	endpoint := viper.GetString("S3Endpoint")
	accessKeyID := viper.GetString("S3AccessKeyID")
	secretAccessKey := viper.GetString("S3SecretAccessKey")
	region := viper.GetString("S3Region")
	s3FS, err := s3fs.NewBaseFS(endpoint, accessKeyID, secretAccessKey, region, true, logger)
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
	sourceFS fs.FS, area string,
	areaPaths map[string]fs.FS,
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
	if err := o.EndUpdate(); err != nil {
		return false, errors.Wrapf(err, "cannot end update for object '%s'", id)
	}

	if err := o.Close(); err != nil {
		return false, errors.Wrapf(err, "cannot close object '%s'", id)
	}

	return o.IsModified(), nil
}

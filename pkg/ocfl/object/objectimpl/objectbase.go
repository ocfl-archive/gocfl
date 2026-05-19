package objectimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"slices"
	"strings"

	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/filesystem/pkg/appendfs"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

type ObjectBaseConfig struct{}

// NewObjectBase creates an empty ObjectBase structure
func NewObjectBase(ctx context.Context, fact factory.FactoryObject, defaultVersion version.OCFLVersion, extensionFactory extension.Factory[object.ExtensionManager], config any, logger ocfllogger.OCFLLogger) object.Object {
	objectbaseConfig, ok := config.(*ObjectBaseConfig)
	if config != nil && !ok {
		logger.Error().Msg("invalid config type for objectbase")
	}
	objectBase := &ObjectBase{
		extensionFactory: extensionFactory,
		//extensionManager: extensionManager.(object.ExtensionManager),
		ctx: ctx,
		//		fsys: nil,
		i: fact.NewInventory(ctx).WithWriteable(),
		//versionFolders:     []string{},
		versionInventories: map[string]inventory.Inventory{},
		changed:            false,
		logger:             logger,
		digest:             "",
		echo:               false,
		updateFiles:        []string{},
		area:               "",
		factory:            fact,
		config:             objectbaseConfig,
	}
	return objectBase
}

type ObjectBase struct {
	//	storageRoot        storageroot.StorageRoot
	extensionFactory extension.Factory[object.ExtensionManager]
	extensionManager object.ExtensionManager
	ctx              context.Context
	readFS           fs.FS
	writeFS          appendfs.FS
	i                inventory.Inventory
	//versionFolders     []string
	versionInventories map[string]inventory.Inventory
	changed            bool
	logger             ocfllogger.OCFLLogger
	//version            version.OCFLVersion
	digest      checksum.DigestAlgorithm
	echo        bool
	updateFiles []string
	area        string
	factory     factory.FactoryObject
	config      *ObjectBaseConfig
}

func (objectBase *ObjectBase) GetWriteFS() appendfs.FS {
	return objectBase.writeFS
}

func (objectBase *ObjectBase) GetFactory() factory.FactoryObject {
	return objectBase.factory
}

func (objectBase *ObjectBase) WithInventory(inv inventory.Inventory) object.Object {
	objectBase.i = inv
	return objectBase
}

func (objectBase *ObjectBase) WithExtensionManager(manager object.ExtensionManager) object.Object {
	objectBase.extensionManager = manager
	return objectBase
}

func (objectBase *ObjectBase) WithReadFS(fsys fs.FS) object.Object {
	objectBase.readFS = fsys
	return objectBase
}

func (objectBase *ObjectBase) WithWriteFS(fsys appendfs.FS) object.Object {
	objectBase.writeFS = fsys
	return objectBase
}

func (objectBase *ObjectBase) GetExtensionManager() object.ExtensionManager {
	return objectBase.extensionManager
}

func (objectBase *ObjectBase) IsModified() bool { return objectBase.i.IsModified() }

func (objectBase *ObjectBase) Stat(w io.Writer, statInfo []object.StatInfo) error {
	fmt.Fprintf(w, "[%s] Path: %s\n", objectBase.i.GetID(), objectBase.i.GetDigestAlgorithm())
	i := objectBase.i
	fmt.Fprintf(w, "[%s] Head: %s\n", objectBase.i.GetID(), i.GetHead())
	fixity := i.GetFixity()
	algs := []string{}
	for alg := range fixity.GetDigestAlgorithms() {
		algs = append(algs, string(alg))
	}
	_, _ = fmt.Fprintf(w, "[%s] Fixity: %s\n", objectBase.i.GetID(), strings.Join(algs, ", "))
	manifest := i.GetManifest()
	cnt := 0
	for _, files := range manifest.Iterate() {
		cnt += len(files)
	}

	var uniqueFileCount int
	for _, _ = range manifest.Iterate() {
		uniqueFileCount++
	}
	_, _ = fmt.Fprintf(w, "[%s] Manifest: %v files (%v unique files)\n", objectBase.i.GetID(), cnt, uniqueFileCount)
	if slices.Contains(statInfo, object.StatObjectVersions) || len(statInfo) == 0 {
		for vString, ver := range i.GetVersions().Iterate() {
			fmt.Fprintf(w, "[%s] Version %s\n", objectBase.i.GetID(), vString)
			fmt.Fprintf(w, "[%s]     User: %s (%s)\n", objectBase.i.GetID(), ver.GetUser().GetName(), ver.GetUser().GetAddress())
			fmt.Fprintf(w, "[%s]     Created: %s\n", objectBase.i.GetID(), ver.GetCreated().String())
			fmt.Fprintf(w, "[%s]     Message: %s\n", objectBase.i.GetID(), ver.GetMessage())
			if slices.Contains(statInfo, object.StatObjectVersionState) || len(statInfo) == 0 {
				for cs, sList := range ver.GetState().Iterate() {
					for _, s := range sList {
						_, _ = fmt.Fprintf(w, "[%s]        %s\n", objectBase.i.GetID(), s)
						if slices.Contains(statInfo, object.StatObjectManifest) || len(statInfo) == 0 {
							ms, err := manifest.GetFiles(cs)
							if err != nil {
								if errors.Is(err, inventory.DigestNotFound) {
									continue
								}
								return errors.Wrapf(err, "cannot get files for manifest '%s'", objectBase.i.GetID())
							}
							for _, m := range ms {
								fmt.Fprintf(w, "[%s]           %s\n", objectBase.i.GetID(), m)
							}
						}
					}
				}
			}
		}
	}
	if slices.Contains(statInfo, object.StatObjectExtensionConfigs) || len(statInfo) == 0 {
		data, err := json.MarshalIndent(objectBase.extensionManager.GetConfig(), "", "  ")
		if err != nil {
			return errors.Wrap(err, "cannot marshal ExtensionManagerConfig")
		}
		fmt.Fprintf(w, "[%s] Initial Extension:\n---\n%s\n---\n", objectBase.i.GetID(), string(data))
		fmt.Fprintf(w, "[%s] Extension Configurations:\n", objectBase.i.GetID())
		for _, ext := range objectBase.extensionManager.GetExtensions() {
			cfg := ext.GetConfig()
			str, _ := json.MarshalIndent(cfg, "", "  ")

			fmt.Fprintf(w, "---\n%s\n", str)
		}
	}
	return nil
}

func (objectBase *ObjectBase) CreateInventory(id string, digestAlg checksum.DigestAlgorithm, fixityAlgs []checksum.DigestAlgorithm) (inventory.Inventory, error) {
	fixity := objectBase.factory.NewFixity(objectBase.ctx).WithAlgorithms(fixityAlgs...)
	inv := objectBase.factory.NewInventory(objectBase.ctx).
		WithID(id).
		WithDigestAlgorithm(digestAlg).
		WithFixity(fixity)

	/*
		inventory, err := inventory.NewInventory(objectBase.ctx, "new", objectBase.GetOCFLVersion(), objectBase.logger)
		if err != nil {
			return nil, errors.Wrap(err, "cannot create empty inventory")
		}
	*/
	/*
		if err := inventory.Init(id, digest, fixity); err != nil {
			return nil, errors.Wrap(err, "cannot initialize empty inventory")
		}
	*/

	return inv, inv.Finalize(true)
}

func (objectBase *ObjectBase) GetInventory() inventory.Inventory {
	if objectBase.i == nil {
		objectBase.i = objectBase.factory.NewInventory(objectBase.ctx)
	}
	return objectBase.i
}

func (objectBase *ObjectBase) GetDigestAlgorithm() checksum.DigestAlgorithm {
	return objectBase.i.GetDigestAlgorithm()
}

func (objectBase *ObjectBase) StartUpdate(msg string, UserName string, UserAddress string, echo bool) (object.VersionWriter, error) {
	objectBase.logger.Debug().Msgf("'%s' / '%s' / '%s'", msg, UserName, UserAddress)
	vw := objectBase.factory.NewVersionWriter(objectBase.ctx).
		WithObject(objectBase).
		WithEcho(echo)
	if err := vw.Init(msg, UserName, UserAddress); err != nil {
		return nil, errors.Wrap(err, "cannot initialize version writer")
	}
	return vw, nil
}

func (objectBase *ObjectBase) GetID() string {
	if objectBase.i == nil {
		return ""
	}
	return objectBase.i.GetID()
}

func (objectBase *ObjectBase) GetReadFS() fs.FS {
	return objectBase.readFS
}

func (objectBase *ObjectBase) GetChecker() object.Checker {
	return objectBase.factory.NewChecker(objectBase.ctx).WithObject(objectBase)
}

func (objectBase *ObjectBase) GetOCFLVersion() version.OCFLVersion {
	return objectBase.factory.GetVersion()
}

func (objectBase *ObjectBase) GetLoader() object.Loader {
	return objectBase.factory.NewLoader(objectBase.ctx).WithObject(objectBase).SetExtensionFactory(objectBase.extensionFactory)
}

func (objectBase *ObjectBase) GetInitializer() object.Initializer {
	return objectBase.factory.NewInitializer(objectBase.ctx).WithObject(objectBase)
}

func (objectBase *ObjectBase) GetExtractor() object.Extractor {
	return objectBase.factory.NewExtractor(objectBase.ctx).WithObject(objectBase)
}

/*
func (objectBase *ObjectBase) GetAreaPath(area string) (string, error) {
	path, err := objectBase.extensionManager.GetAreaPath(area)
	return path, errors.WithStack(err)
}
*/

var _ object.Object = (*ObjectBase)(nil)

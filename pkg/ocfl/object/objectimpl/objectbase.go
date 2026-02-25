package objectimpl

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"slices"
	"strconv"
	"strings"

	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension/extensionimpl"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfllogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/streamfs"
)

// NewObjectBase creates an empty ObjectBase structure
func NewObjectBase(ctx context.Context, fact factory.Factory, defaultVersion version.OCFLVersion, extensionFactory *extensionimpl.Factory, logger ocfllogger.OCFLLogger) *ObjectBase {
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
	}
	return objectBase
}

type ObjectBase struct {
	//	storageRoot        storageroot.StorageRoot
	extensionFactory *extensionimpl.Factory
	extensionManager object.ExtensionManager
	ctx              context.Context
	//fsys             fs.FS
	i inventory.Inventory
	//versionFolders     []string
	versionInventories map[string]inventory.Inventory
	changed            bool
	logger             ocfllogger.OCFLLogger
	//version            version.OCFLVersion
	digest      checksum.DigestAlgorithm
	echo        bool
	updateFiles []string
	area        string
	factory     factory.Factory
}

func (objectBase *ObjectBase) GetFactory() factory.Factory {
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

func (objectBase *ObjectBase) GetExtensionManager() object.ExtensionManager {
	return objectBase.extensionManager
}

func (objectBase *ObjectBase) IsModified() bool { return objectBase.i.IsModified() }

func (objectBase *ObjectBase) GetMetadata() (*inventory.Metadata, error) {
	inv := objectBase.i
	if inv == nil {
		return nil, errors.Errorf("inventory is nil")
	}

	result := &inventory.Metadata{
		ID:              objectBase.i.GetID(),
		Head:            inv.GetHead(),
		Files:           map[string]*inventory.FileMetadata{},
		DigestAlgorithm: objectBase.i.GetDigestAlgorithm(),
		Versions:        map[string]*inventory.VersionMetadata{},
	}
	versions := inv.GetVersions()
	versionStrings := []string{}
	for v, ver := range versions.Iterate() {
		result.Versions[v.String()] = &inventory.VersionMetadata{
			Created: ver.GetCreated(),
			Message: ver.GetMessage(),
			Name:    ver.GetUser().GetName(),
			Address: ver.GetUser().GetAddress(),
		}
		versionStrings = append(versionStrings, v.String())
	}
	// sort version strings in ascending order
	slices.SortFunc(versionStrings, func(a, b string) int {
		a = strings.TrimPrefix(a, "v0")
		b = strings.TrimPrefix(b, "v0")
		ia, _ := strconv.Atoi(a)
		ib, _ := strconv.Atoi(b)
		return cmp.Compare(ia, ib)
	})
	extensionMetadata, err := objectBase.extensionManager.GetMetadata(objectBase)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot get extension metadata for object '%s'", objectBase.i.GetID())
	}
	if objectMeta, ok := extensionMetadata[""]; ok {
		/*
			for key, val := range objectMeta {
				result.Extension[key] = val
			}

		*/
		result.Extension = objectMeta
	}
	manifest := inv.GetManifest()
	fixity := inv.GetFixity()
	for digest, fnames := range manifest.Iterate() {
		if len(fnames) == 0 {
			continue
		}
		fm := &inventory.FileMetadata{
			Checksums:    map[checksum.DigestAlgorithm]string{},
			InternalName: fnames,
			VersionName:  map[string][]string{},
			Extension:    map[string]any{},
		}
		fm.Checksums = fixity.Checksums(fnames[0])
		for v, ver := range versions.Iterate() {
			for d, externalNames := range ver.GetState().Iterate() {
				if digest == d {
					if _, ok := fm.VersionName[v.String()]; !ok {
						fm.VersionName[v.String()] = []string{}
					}
					fm.VersionName[v.String()] = append(fm.VersionName[v.String()], externalNames...)
					break
				}
			}
		}
		if emAny, ok := extensionMetadata[digest]; ok {
			if em, ok := emAny.(map[string]any); ok {
				fm.Extension = em
			}
		}
		result.Files[digest] = fm
	}
	return result, nil
}

func (objectBase *ObjectBase) Stat(w io.Writer, statInfo []object.StatInfo) error {
	fmt.Fprintf(w, "[%s] Path: %s\n", objectBase.i.GetID(), objectBase.i.GetDigestAlgorithm())
	i := objectBase.i
	fmt.Fprintf(w, "[%s] Head: %s\n", objectBase.i.GetID(), i.GetHead())
	fixity := i.GetFixity()
	algs := []string{}
	for alg := range fixity.GetDigestAlgorithms() {
		algs = append(algs, string(alg))
	}
	fmt.Fprintf(w, "[%s] Fixity: %s\n", objectBase.i.GetID(), strings.Join(algs, ", "))
	manifest := i.GetManifest()
	cnt := 0
	for _, fs := range manifest.Iterate() {
		cnt += len(fs)
	}

	var uniqueFileCount int
	for _, _ = range manifest.Iterate() {
		uniqueFileCount++
	}
	fmt.Fprintf(w, "[%s] Manifest: %v files (%v unique files)\n", objectBase.i.GetID(), cnt, uniqueFileCount)
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
	inventory := objectBase.factory.NewInventory(objectBase.ctx).
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

	return inventory, inventory.Finalize(true)
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

func (objectBase *ObjectBase) StartUpdate(objectFS streamfs.FS, msg string, UserName string, UserAddress string, echo bool) (object.VersionWriter, error) {
	objectBase.logger.Debug().Msgf("'%s' / '%s' / '%s'", msg, UserName, UserAddress)

	vw, err := NewVersionWriter(objectBase, objectFS, echo, objectBase.logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create version writer")
	}
	return vw, nil
}

func (objectBase *ObjectBase) GetID() string {
	if objectBase.i == nil {
		return ""
	}
	return objectBase.i.GetID()
}

func (objectBase *ObjectBase) GetOCFLVersion() version.OCFLVersion {
	return objectBase.factory.GetVersion()
}

func (objectBase *ObjectBase) GetLoader(sourceFS fs.FS, extensionFactory extension.Factory) object.Loader {
	return objectBase.factory.NewLoader(objectBase.ctx).WithObject(objectBase).WithFS(sourceFS).WithExtensionFactory(extensionFactory)
}

func (objectBase *ObjectBase) GetInitializer(objectFS streamfs.FS) object.Initializer {
	return objectBase.factory.NewInitializer(objectBase.ctx).WithObject(objectBase).WithFS(objectFS)
}

func (objectBase *ObjectBase) GetChecker(objectFS fs.FS) object.Checker {
	return objectBase.factory.NewChecker(objectBase.ctx).WithObject(objectBase).WithFS(objectFS)
}

func (objectBase *ObjectBase) GetExtractor(fsys fs.FS) object.Extractor {
	return objectBase.factory.NewExtractor(objectBase.ctx).WithObject(objectBase).WithFS(fsys, nil)
}

/*
func (objectBase *ObjectBase) GetAreaPath(area string) (string, error) {
	path, err := objectBase.extensionManager.GetAreaPath(area)
	return path, errors.WithStack(err)
}
*/

var _ object.Object = (*ObjectBase)(nil)

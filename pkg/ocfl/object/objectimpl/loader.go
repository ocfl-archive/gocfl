package objectimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/filesystem/pkg/writefs"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/validation"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

func NewLoader(ctx context.Context, factory factory.FactoryObject, logger ocfllogger.OCFLLogger) *Loader {
	return &Loader{
		ctx:     ctx,
		factory: factory,
		logger:  logger.With("task", "loader"),
	}
}

type Loader struct {
	object.Object
	ctx              context.Context
	extensionFactory extension.Factory[object.ExtensionManager]
	logger           ocfllogger.OCFLLogger
	factory          factory.FactoryObject
}

func (loader *Loader) GetFS() fs.FS {
	return loader.GetReadFS()
}

func (loader *Loader) Load() error {
	if loader.GetReadFS() == nil {
		return errors.New("object read FS is nil")
	}
	if err := loader.loadExtensionManager(); err != nil {
		return errors.Wrap(err, "loading extension manager")
	}
	if err := loader.loadInventory(); err != nil {
		return errors.Wrap(err, "cannot load inventory.json")
	}
	return nil
}

func (loader *Loader) SetExtensionFactory(factory extension.Factory[object.ExtensionManager]) object.Loader {
	loader.extensionFactory = factory
	return loader
}

func (loader *Loader) WithObject(o object.Object) object.Loader {
	loader.Object = o
	return loader
}

func (loader *Loader) findInventoryFile() (string, error) {
	// for version 1.0 and 1.1 there MUST be an inventory.json in the object root
	if slices.Contains([]version.OCFLVersion{version.Version1_0, version.Version1_1}, loader.GetOCFLVersion()) {
		return "inventory.json", nil
	}
	dirs, err := fs.ReadDir(loader.GetReadFS(), ".")
	if err != nil {
		return "", errors.Wrapf(err, "failed to read directory %v", loader.GetReadFS())
	}
	for _, d := range dirs {
		if d.IsDir() {
			continue
		}
		if d.Name() == "inventory.json" {
			return "inventory.json", nil
		}
	}
	var headNumber int64
	var p string
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		folderName := d.Name()
		if len(folderName) == 0 || folderName[0] != 'v' {
			continue
		}
		num, err := strconv.ParseInt(strings.TrimLeft(folderName[1:], "0"), 10, 64)
		if err != nil {
			return "", errors.Wrapf(err, "failed to parse version number from folder name '%s'", folderName)
		}
		if num > headNumber {
			p = path.Join(folderName, "inventory.json")
			headNumber = num
		}
	}
	if headNumber == 0 {
		return "", errors.Errorf("failed to find inventory.json in %v", loader.GetReadFS())
	}
	return p, nil
}

func unmarshalInventoryData(ctx context.Context, data []byte, ver version.OCFLVersion, fact factory.FactoryObject, logger ocfllogger.OCFLLogger) (inventory.Inventory, error) {
	anyMap := map[string]any{}
	if err := json.Unmarshal(data, &anyMap); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal json '%s'", string(data))
	}
	var iVer version.OCFLVersion
	t, ok := anyMap["type"]
	if !ok {
		return nil, errors.New("no type in inventory")
	}
	sStr, ok := t.(string)
	if !ok {
		return nil, errors.Errorf("type not a string in inventory - '%v'", t)
	}
	switch sStr {
	case "https://ocfl.io/1.1/spec/#inventory":
		iVer = version.Version1_1
	case "https://ocfl.io/1.0/spec/#inventory":
		iVer = version.Version1_0
	case "https://ocfl.io/2.0/spec/#inventory":
		iVer = version.Version2_0
	default:
		// if we don't know anything use the old stuff
		return nil, errors.Errorf("unsupported inventory type '%s'", sStr)
	}
	if _, ok := anyMap["manifest"]; !ok {
		logger.ValidationError(validation.E041, "manifest not found in inventory")
	}
	if _, ok := anyMap["versions"]; !ok {
		logger.ValidationError(validation.E041, "versions not found in inventory")
	}
	// if necessary, use factory with an older version
	oldFactVersion := fact.GetVersion()
	if oldFactVersion != iVer {
		oldFact := fact
		fact = fact.WithNewVersion(iVer)
		defer func() {
			fact = oldFact
		}()
	}
	inv := fact.NewInventory(ctx)
	if err := json.Unmarshal(data, inv); err != nil {
		// now lets try it again
		jsonMap := map[string]any{}
		// check for json format error
		if err2 := json.Unmarshal(data, &jsonMap); err2 != nil {
			logger.ValidationError(validation.E033, "json unmarshal error: %v", err2)
			logger.ValidationError(validation.E034, "json unmarshal error: %v", err2)
		} else {
			if _, ok := jsonMap["head"].(string); !ok {
				logger.ValidationError(validation.E040, "json head not a string: %v", jsonMap["head"])
			}
		}
		//return nil, errors.Wrapf(err, "cannot marshal data - '%s'", string(data))
	}

	return inv, inv.Finalize(false)
}

var inventorySideCarFormat = regexp.MustCompile(`^([a-fA-F0-9]+)\s+inventory.json$`)

func getInventorySidecarChecksum(objectFS fs.FS, sidecarPath string, logger ocfllogger.OCFLLogger) (string, error) {
	sidecarBytes, err := fs.ReadFile(objectFS, sidecarPath)
	if err != nil {
		if errors.Is(errors.Cause(err), fs.ErrNotExist) {
			logger.ValidationError(validation.E058, "sidecar '%v/%s' does not exist", objectFS, sidecarPath)
		} else {
			logger.ValidationError(validation.E060, "cannot read sidecar '%v/%s'", objectFS, sidecarPath)
		}
		return "", errors.Wrapf(err, "cannot read sidecar '%v/%s'", objectFS, sidecarPath)
		//		objectBase.logger.ValidationError(E058, "cannot read '%s': %v", sidecarPath, err)
	}

	digestString := strings.TrimSpace(string(sidecarBytes))
	matches := inventorySideCarFormat.FindStringSubmatch(digestString)
	if len(matches) == 0 {
		logger.ValidationError(validation.E061, "no suffix \" inventory.json\" in '%v/%s'", objectFS, sidecarPath)
		return "", errors.New(fmt.Sprintf("invalid digest file for inventory - '%s'", sidecarPath))
	}

	return matches[1], nil
}

func (loader *Loader) loadInventoryFile(filename string) (inventory.Inventory, error) {
	inv, _, err := loadInventoryFile(loader.ctx, loader.GetReadFS(), filename, loader.GetOCFLVersion(), loader.factory, loader.logger)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot load inventory file '%s'", filename)
	}
	return inv, nil
}
func loadInventoryFile(ctx context.Context, objectFS fs.FS, filename string, ver version.OCFLVersion, fact factory.FactoryObject, logger ocfllogger.OCFLLogger) (inventory.Inventory, string, error) {
	// load inventory file
	inventoryBytes, err := fs.ReadFile(objectFS, filename)
	if err != nil {
		if errors.Is(errors.Cause(err), fs.ErrNotExist) {
			logger.ValidationError(validation.E063, "file '%v/%s' does not exist", objectFS, filename)
			return nil, "", errors.Wrapf(err, "inventory file '%v/%s' does not exist", objectFS, filename)
		}
		return nil, "", errors.Wrapf(err, "failed to read inventory file '%v/%s'", objectFS, filename)
	}
	inv, err := unmarshalInventoryData(ctx, inventoryBytes, ver, fact, logger)
	if err != nil {
		return nil, "", errors.Wrap(err, "cannot unmarshal inventory object")
	}
	if _, writeable := objectFS.(writefs.AppendFS); writeable {
		inv.WithWriteable()
	}

	digest := inv.GetDigestAlgorithm()

	// check digest for inventory
	sidecarPath := fmt.Sprintf("%s.%s", filename, digest)
	digestString, err := getInventorySidecarChecksum(objectFS, sidecarPath, logger)
	if err == nil {
		h, err := checksum.GetHash(digest)
		if err != nil {
			logger.ValidationError(validation.E060, "invalid hash %s in '%v/%s'", digest, objectFS, sidecarPath)
			return nil, "", errors.New(fmt.Sprintf("invalid digest file for inventory - '%s'", string(digest)))
		}
		h.Reset()
		h.Write(inventoryBytes)
		sumBytes := h.Sum(nil)
		inventoryDigestString := fmt.Sprintf("%x", sumBytes)
		if digestString != inventoryDigestString {
			logger.ValidationError(validation.E060, "'%s' != '%s'", digestString, inventoryDigestString)
		}
	}
	return inv, digestString, nil
}

func (loader *Loader) loadInventory() error {
	filename, err := loader.findInventoryFile()

	inv, err := loader.loadInventoryFile(filename)
	if err != nil {
		return errors.Wrapf(err, "failed to load inventory file '%s'", filename)
	}
	loader.WithInventory(inv)
	return nil
}

func (loader *Loader) loadExtensionManager() error {
	extensionFS, err := writefs.Sub(loader.GetReadFS(), "extensions")
	if err != nil {
		return errors.Wrapf(err, "cannot create subfs of %v for folder '%s'", loader.GetReadFS(), "extensions")
	}
	manager, err := loader.extensionFactory.LoadExtensionManager(extensionFS)
	if err != nil {
		loader.logger.ValidationError(validation.W000, "cannot initialize all extensions in folder '%s': %v", extensionFS, err)
		if manager == nil {
			return errors.Wrap(err, "cannot create extension manager")
		}
	}
	loader.Object.WithExtensionManager(manager.(object.ExtensionManager))
	return nil
}

var _ object.Loader = (*Loader)(nil)

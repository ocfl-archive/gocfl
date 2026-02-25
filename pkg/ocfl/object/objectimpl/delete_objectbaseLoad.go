//go:build not

package objectimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"strings"

	"emperror.dev/errors"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfllogger"
)

func (objectBase *ObjectBase) Load(fsys fs.FS) error {
	extFolder, err := writefs.Sub(fsys, "extensions")
	if err != nil {
		return errors.Wrapf(err, "cannot create subfs of %v for folder '%s'", fsys, "extensions")
	}
	manager, err := objectBase.extensionFactory.LoadExtensionManager(extFolder, objectBase.i.GetOCFLVersion())
	if err != nil {
		objectBase.logger.ValidationError(validation.W000, "cannot initialize all extensions in folder '%s': %v", extFolder, err)
		if manager == nil {
			return errors.Wrap(err, "cannot create extension manager")
		}
	}
	objectBase.extensionManager = manager.(object.ExtensionManager)

	// load the inventory
	inv, err := loadInventory(objectBase.ctx, fsys, inventory.VersionNumber{}, objectBase.factory, objectBase.logger)
	if err != nil {
		return errors.Wrap(err, "cannot load inventory.json of root")
	}
	objectBase.i = inv
	objectBase.logger = objectBase.logger.WithVersion(inv.GetOCFLVersion())
	return nil
}

// helper functions
func loadInventory(ctx context.Context, fsys fs.FS, ver inventory.VersionNumber, factory factory.Factory, logger ocfllogger.OCFLLogger) (inventory.Inventory, error) {
	var filename string
	var err error
	if ver.Int() == 0 {
		filename, err = findInventoryFile(fsys)
	} else {
		filename = path.Join("/", ver.String(), "inventory.json")
	}
	inv, err := loadInventoryFile(ctx, fsys, filename, factory, logger)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load inventory file '%s'", filename)
	}

	return inv, nil
}

func loadInventoryFile(ctx context.Context, fsys fs.FS, filename string, factory factory.Factory, logger ocfllogger.OCFLLogger) (inventory.Inventory, error) {
	// load inventory file
	inventoryBytes, err := fs.ReadFile(fsys, filename)
	if err != nil {
		if errors.Is(errors.Cause(err), fs.ErrNotExist) {
			return nil, errors.Wrapf(err, "inventory file '%s' does not exist", filename)
		}
		return nil, errors.Wrapf(err, "failed to read inventory file '%s'", filename)
	}
	inv, err := unmarshalInventoryData(ctx, inventoryBytes, factory, logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot initiate inventory object")
	}
	digest := inv.GetDigestAlgorithm()

	// check digest for inventory
	sidecarPath := fmt.Sprintf("%s.%s", filename, digest)
	sidecarBytes, err := fs.ReadFile(fsys, sidecarPath)
	if err != nil {
		if errors.Is(errors.Cause(err), fs.ErrNotExist) {
			logger.ValidationError(validation.E058, "sidecar '%v/%s' does not exist", fsys, sidecarPath)
		} else {
			logger.ValidationError(validation.E060, "cannot read sidecar '%v/%s'", fsys, sidecarPath)
		}
		//		objectBase.logger.ValidationError(E058, "cannot read '%s': %v", sidecarPath, err)
	} else {
		digestString := strings.TrimSpace(string(sidecarBytes))
		//if !strings.HasSuffix(digestString, " inventory.json") {
		matches := inventorySideCarFormat.FindStringSubmatch(digestString)
		if /* matches == nil || */ len(matches) == 0 {
			logger.ValidationError(validation.E061, "no suffix \" inventory.json\" in '%v/%s'", fsys, sidecarPath)
		} else {
			//digestString = strings.TrimSpace(strings.TrimSuffix(digestString, " inventory.json"))
			digestString = matches[1]
			h, err := checksum.GetHash(digest)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("invalid digest file for inventory - '%s'", string(digest)))
			}
			h.Reset()
			h.Write(inventoryBytes)
			sumBytes := h.Sum(nil)
			inventoryDigestString := fmt.Sprintf("%x", sumBytes)
			if digestString != inventoryDigestString {
				logger.ValidationError(validation.E060, "'%s' != '%s'", digestString, inventoryDigestString)
			}
		}
	}
	return inv, inv.Finalize(false)
}

func unmarshalInventoryData(ctx context.Context, data []byte, factory factory.Factory, logger ocfllogger.OCFLLogger) (inventory.Inventory, error) {
	anyMap := map[string]any{}
	if err := json.Unmarshal(data, &anyMap); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal json '%s'", string(data))
	}
	var ver version.OCFLVersion
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
		ver = version.Version1_1
	case "https://ocfl.io/1.0/spec/#inventory":
		ver = version.Version1_0
	case "https://ocfl.io/2.0/spec/#inventory":
		ver = version.Version2_0
	default:
		// if we don't know anything use the old stuff
		return nil, errors.Errorf("unsupported inventory type '%s'", sStr)
	}
	if ver != factory.GetVersion() {
		return nil, errors.Errorf("inventory version '%s' does not match expected version '%s'", ver, factory.GetVersion())
	}
	inv := factory.NewInventory(ctx).WithWriteable()
	/*
		inventory, err := inventory.NewInventory(objectBase.ctx, folder, ver, objectBase.logger)
		if err != nil {
			return nil, errors.Wrap(err, "cannot create empty inventory")
		}
	*/
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

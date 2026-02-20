package objectimpl

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"strconv"
	"strings"

	"emperror.dev/errors"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
)

func (objectBase *ObjectBase) findInventoryFile(fsys fs.FS) (string, error) {
	dirs, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return "", errors.Wrapf(err, "failed to read directory %v", fsys)
	}
	for _, d := range dirs {
		if d.IsDir() {
			continue
		}
		if d.Name() == "inventory.json" {
			return "/inventory.json", nil
		}
	}
	var headNumber int64
	var p string
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		folderName := d.Name()
		if folderName[0] != 'v' {
			continue
		}
		num, err := strconv.ParseInt(strings.TrimLeft(folderName[1:], "0"), 10, 64)
		if err != nil {
			return "", errors.Wrapf(err, "failed to parse version number from folder name '%s'", folderName)
		}
		if num > headNumber {
			p = path.Join("/", folderName, "inventory.json")
			headNumber = num
		}
	}
	if headNumber == 0 {
		return "", errors.Errorf("failed to find inventory.json in %v", fsys)
	}
	return p, nil
}

func (objectBase *ObjectBase) unmarshalInventoryData(data []byte) (inventory.Inventory, error) {
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
	if ver != objectBase.version {
		return nil, errors.Errorf("inventory version '%s' does not match expected version '%s'", ver, objectBase.version)
	}
	inv := objectBase.factory.NewInventory(objectBase.ctx).WithWriteable()
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
			objectBase.logger.ValidationError(ver, validation.E033, "json unmarshal error: %v", err2)
			objectBase.logger.ValidationError(ver, validation.E034, "json unmarshal error: %v", err2)
		} else {
			if _, ok := jsonMap["head"].(string); !ok {
				objectBase.logger.ValidationError(ver, validation.E040, "json head not a string: %v", jsonMap["head"])
			}
		}
		//return nil, errors.Wrapf(err, "cannot marshal data - '%s'", string(data))
	}

	return inv, objectBase.i.Finalize(false)
}

func (objectBase *ObjectBase) loadInventory(fsys fs.FS) (err error) {
	filename, err := objectBase.findInventoryFile(fsys)
	// load inventory file
	inventoryBytes, err := fs.ReadFile(fsys, filename)
	if err != nil {
		if errors.Is(errors.Cause(err), fs.ErrNotExist) {
			return err
		}
		objectBase.i = objectBase.factory.NewInventory(objectBase.ctx).WithWriteable()
		return nil
	}
	objectBase.i, err = objectBase.unmarshalInventoryData(inventoryBytes)
	if err != nil {
		return errors.Wrap(err, "cannot initiate inventory object")
	}
	digest := objectBase.i.GetDigestAlgorithm()

	// check digest for inventory
	sidecarPath := fmt.Sprintf("%s.%s", filename, digest)
	sidecarBytes, err := fs.ReadFile(fsys, sidecarPath)
	if err != nil {
		if errors.Is(errors.Cause(err), fs.ErrNotExist) {
			objectBase.logger.ValidationError(objectBase.version, validation.E058, "sidecar '%v/%s' does not exist", fsys, sidecarPath)
		} else {
			objectBase.logger.ValidationError(objectBase.version, validation.E060, "cannot read sidecar '%v/%s'", fsys, sidecarPath)
		}
		//		objectBase.AddValidationError(E058, "cannot read '%s': %v", sidecarPath, err)
	} else {
		digestString := strings.TrimSpace(string(sidecarBytes))
		//if !strings.HasSuffix(digestString, " inventory.json") {
		matches := inventorySideCarFormat.FindStringSubmatch(digestString)
		if /* matches == nil || */ len(matches) == 0 {
			objectBase.logger.ValidationError(objectBase.i.GetOCFLVersion(), validation.E061, "no suffix \" inventory.json\" in '%v/%s'", fsys, sidecarPath)
		} else {
			//digestString = strings.TrimSpace(strings.TrimSuffix(digestString, " inventory.json"))
			digestString = matches[1]
			h, err := checksum.GetHash(digest)
			if err != nil {
				return errors.New(fmt.Sprintf("invalid digest file for inventory - '%s'", string(digest)))
			}
			h.Reset()
			h.Write(inventoryBytes)
			sumBytes := h.Sum(nil)
			inventoryDigestString := fmt.Sprintf("%x", sumBytes)
			if digestString != inventoryDigestString {
				objectBase.logger.ValidationError(objectBase.i.GetOCFLVersion(), validation.E060, "'%s' != '%s'", digestString, inventoryDigestString)
			}
		}
	}
	return objectBase.i.Finalize(false)
}

func (objectBase *ObjectBase) Load(fsys fs.FS) error {
	extFolder, err := writefs.Sub(fsys, "extensions")
	if err != nil {
		return errors.Wrapf(err, "cannot create subfs of %v for folder '%s'", fsys, "extensions")
	}
	manager, err := objectBase.extensionFactory.LoadExtensionManager(extFolder, objectBase.version)
	if err != nil {
		objectBase.logger.ValidationError(objectBase.version, validation.W000, "cannot initialize all extensions in folder '%s': %v", extFolder, err)
		if manager == nil {
			return errors.Wrap(err, "cannot create extension manager")
		}
	}
	objectBase.extensionManager = manager.(object.ExtensionManager)

	// load the inventory
	if err = objectBase.loadInventory(fsys); err != nil {
		return errors.Wrap(err, "cannot load inventory.json of root")
	}
	return nil
}

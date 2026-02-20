package objectimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"regexp"
	"strings"

	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/util"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfllogger"
)

func NewObjectLoader(ctx context.Context, fsys fs.FS, fact factory.Factory, logger ocfllogger.OCFLLogger) *objectLoader {
	return &objectLoader{
		ctx:     ctx,
		fsys:    fsys,
		factory: fact,
		logger:  logger,
	}
}

type objectLoader struct {
	ctx     context.Context
	fsys    fs.FS
	factory factory.Factory
	logger  ocfllogger.OCFLLogger
}

func (loader *objectLoader) Load() (object.Object, error) {
	ver, err := util.GetVersion(loader.fsys, ".", "ocfl_object_")
	if err != nil {
		return nil, errors.Wrap(err, "getting object version")
	}
	if err := loader.factory.SetVersion(ver); err != nil {
		return nil, errors.Wrap(err, "setting version")
	}
	object := loader.factory.NewObject(loader.ctx)
	if err := object.Load(); err != nil {
		return nil, errors.Wrap(err, "loading object")
	}
	return nil, errors.New("not implemented")
}

func (loader *objectLoader) loadInventory(data []byte) (inventory.Inventory, error) {
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
	loader.factory.SetVersion(ver)
	inv := loader.factory.NewInventory(loader.ctx).WithWriteable()
	/*
		inventory, err := inventory.NewInventory(loader.ctx, folder, ver, loader.logger)
		if err != nil {
			return nil, errors.Wrap(err, "cannot create empty inventory")
		}
	*/
	if err := json.Unmarshal(data, inv); err != nil {
		// now lets try it again
		jsonMap := map[string]any{}
		// check for json format error
		if err2 := json.Unmarshal(data, &jsonMap); err2 != nil {
			loader.logger.ValidationError(ver, validation.E033, "json unmarshal error: %v", err2)
			loader.logger.ValidationError(ver, validation.E034, "json unmarshal error: %v", err2)
		} else {
			if _, ok := jsonMap["head"].(string); !ok {
				loader.logger.ValidationError(ver, validation.E040, "json head not a string: %v", jsonMap["head"])
			}
		}
		//return nil, errors.Wrapf(err, "cannot marshal data - '%s'", string(data))
	}

	return inv, inv.Finalize(false)
}

var inventorySideCarFormat = regexp.MustCompile(`^([a-fA-F0-9]+)\s+inventory.json$`)

// loadInventory loads inventory from existing Object
func (loader *objectLoader) LoadInventory() (inventory.Inventory, error) {
	filename, err := util.FindInventoryFile(loader.fsys)
	// load inventory file
	inventoryBytes, err := fs.ReadFile(loader.fsys, filename)
	if err != nil {
		if errors.Is(errors.Cause(err), fs.ErrNotExist) {
			return nil, err
		}
		inv := loader.factory.NewInventory(loader.ctx).WithWriteable()
		return inv, nil
	}
	inv, err := loader.loadInventory(inventoryBytes)
	if err != nil {
		return nil, errors.Wrap(err, "cannot initiate inventory object")
	}
	digest := inv.GetDigestAlgorithm()

	// check digest for inventory
	sidecarPath := fmt.Sprintf("%s.%s", filename, digest)
	sidecarBytes, err := fs.ReadFile(loader.fsys, sidecarPath)
	if err != nil {
		if errors.Is(errors.Cause(err), fs.ErrNotExist) {
			validation.AddValidationErrors(loader.ctx, validation.GetValidationError(inv.GetOCFLVersion(), validation.E058).AppendDescription("sidecar '%v/%s' does not exist", loader.fsys, sidecarPath))
		} else {
			validation.AddValidationErrors(loader.ctx, validation.GetValidationError(inv.GetOCFLVersion(), validation.E060).AppendDescription("cannot read sidecar '%v/%s': %v", loader.fsys, sidecarPath, err.Error()))
		}
		//		loader.AddValidationError(E058, "cannot read '%s': %v", sidecarPath, err)
	} else {
		digestString := strings.TrimSpace(string(sidecarBytes))
		//if !strings.HasSuffix(digestString, " inventory.json") {
		matches := inventorySideCarFormat.FindStringSubmatch(digestString)
		if /* matches == nil || */ len(matches) == 0 {
			validation.AddValidationErrors(loader.ctx, validation.GetValidationError(inv.GetOCFLVersion(), validation.E061).AppendDescription("no suffix \" inventory.json\" in '%v/%s'", loader.fsys, sidecarPath))
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
				validation.AddValidationErrors(loader.ctx, validation.GetValidationError(inv.GetOCFLVersion(), validation.E060).AppendContext("'%s' != '%s'", digestString, inventoryDigestString))
			}
		}
	}
	return inv, inv.Finalize(false)
}

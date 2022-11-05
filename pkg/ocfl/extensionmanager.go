package ocfl

import (
	"emperror.dev/errors"
	"fmt"
)

type ExtensionManager struct {
	extensions        []Extension
	storagerootPath   []StoragerootPath
	objectContentPath []ObjectContentPath
}

func NewExtensionManager() (*ExtensionManager, error) {
	m := &ExtensionManager{
		extensions:        []Extension{},
		storagerootPath:   []StoragerootPath{},
		objectContentPath: []ObjectContentPath{},
	}
	return m, nil
}

func (manager *ExtensionManager) Add(ext Extension) error {
	manager.extensions = append(manager.extensions, ext)
	if srp, ok := ext.(StoragerootPath); ok {
		manager.storagerootPath = append(manager.storagerootPath, srp)
	}
	if ocp, ok := ext.(ObjectContentPath); ok {
		manager.objectContentPath = append(manager.objectContentPath, ocp)
	}
	return nil
}

func (manager *ExtensionManager) BuildStoragerootPath(storageroot StorageRoot, id string) (string, error) {
	var errs = []error{}
	for _, srp := range manager.storagerootPath {
		p, err := srp.BuildStorageRootPath(storageroot, id)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if p != "" {
			return p, errors.Combine(errs...)
		}
	}
	return id, errors.Combine(errs...)
}

func (manager *ExtensionManager) BuildObjectContentPath(object Object, path string) (string, error) {
	var errs = []error{}
	for _, ocp := range manager.objectContentPath {
		p, err := ocp.BuildObjectContentPath(object, path)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if p != "" {
			return p, errors.Combine(errs...)
		}
	}
	return path, errors.Combine(errs...)
}

func (manager *ExtensionManager) StoreConfigs(fs OCFLFS) error {
	for _, ext := range manager.extensions {
		configName := fmt.Sprintf("%s/config.json", ext.GetName())
		fp, err := fs.Create(configName)
		if err != nil {
			return errors.Wrapf(err, "cannot create '%s'", configName)
		}
		if err := ext.WriteConfig(fp); err != nil {
			fp.Close()
			return errors.Wrapf(err, "cannot store '%s'", configName)
		}
		fp.Close()
	}
	return nil
}

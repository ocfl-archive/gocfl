package ocfl

import (
	"emperror.dev/errors"
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
		subfs, err := fs.SubFS(ext.GetName())
		if err != nil {
			return errors.Wrapf(err, "cannot create subfs of %v for folder %s", fs, ext.GetName())
		}
		if err := ext.WriteConfig(subfs); err != nil {
			return errors.Wrapf(err, "cannot store '%s'", ext.GetName())
		}
	}
	return nil
}

func (manager *ExtensionManager) StoreRootLayout(fs OCFLFS) error {
	for _, ext := range manager.storagerootPath {
		if err := ext.WriteLayout(fs); err != nil {
			return errors.Wrapf(err, "cannot store '%v'", ext)
		}
	}
	return nil
}

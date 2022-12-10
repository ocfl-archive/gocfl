package ocfl

import (
	"emperror.dev/errors"
)

type ExtensionManager struct {
	extensions        []Extension
	storagerootPath   []ExtensionStoragerootPath
	objectContentPath []ExtensionObjectContentPath
	contentChange     []ExtensionContentChange
	objectChange      []ExtensionObjectChange
}

func NewExtensionManager() (*ExtensionManager, error) {
	m := &ExtensionManager{
		extensions:        []Extension{},
		storagerootPath:   []ExtensionStoragerootPath{},
		objectContentPath: []ExtensionObjectContentPath{},
		objectChange:      []ExtensionObjectChange{},
	}
	return m, nil
}

func (manager *ExtensionManager) Add(ext Extension) error {
	manager.extensions = append(manager.extensions, ext)

	if srp, ok := ext.(ExtensionStoragerootPath); ok {
		manager.storagerootPath = append(manager.storagerootPath, srp)
	}
	if ocp, ok := ext.(ExtensionObjectContentPath); ok {
		manager.objectContentPath = append(manager.objectContentPath, ocp)
	}
	if occ, ok := ext.(ExtensionContentChange); ok {
		manager.contentChange = append(manager.contentChange, occ)
	}
	if occ, ok := ext.(ExtensionObjectChange); ok {
		manager.objectChange = append(manager.objectChange, occ)
	}
	return nil
}

func (manager *ExtensionManager) StoreConfigs() error {
	for _, ext := range manager.extensions {
		if err := ext.WriteConfig(); err != nil {
			return errors.Wrapf(err, "cannot store '%s'", ext.GetName())
		}
	}
	return nil
}

// StoragerootPath
func (manager *ExtensionManager) StoreRootLayout(fs OCFLFS) error {
	for _, ext := range manager.storagerootPath {
		if err := ext.WriteLayout(fs); err != nil {
			return errors.Wrapf(err, "cannot store '%v'", ext)
		}
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

// ObjectContentPath
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

// ContentChange
func (manager *ExtensionManager) AddFileBefore(object Object, source, dest string) error {
	var errs = []error{}
	for _, ocp := range manager.contentChange {
		if err := ocp.AddFileBefore(object, source, dest); err != nil {
			errs = append(errs, err)
			continue
		}
	}
	return errors.Combine(errs...)
}
func (manager *ExtensionManager) UpdateFileBefore(object Object, source, dest string) error {
	var errs = []error{}
	for _, ocp := range manager.contentChange {
		if err := ocp.UpdateFileBefore(object, source, dest); err != nil {
			errs = append(errs, err)
			continue
		}
	}
	return errors.Combine(errs...)
}
func (manager *ExtensionManager) DeleteFileBefore(object Object, dest string) error {
	var errs = []error{}
	for _, ocp := range manager.contentChange {
		if err := ocp.DeleteFileBefore(object, dest); err != nil {
			errs = append(errs, err)
			continue
		}
	}
	return errors.Combine(errs...)
}
func (manager *ExtensionManager) AddFileAfter(object Object, source, dest string) error {
	var errs = []error{}
	for _, ocp := range manager.contentChange {
		if err := ocp.AddFileAfter(object, source, dest); err != nil {
			errs = append(errs, err)
			continue
		}
	}
	return errors.Combine(errs...)
}
func (manager *ExtensionManager) UpdateFileAfter(object Object, source, dest string) error {
	var errs = []error{}
	for _, ocp := range manager.contentChange {
		if err := ocp.UpdateFileAfter(object, source, dest); err != nil {
			errs = append(errs, err)
			continue
		}
	}
	return errors.Combine(errs...)
}
func (manager *ExtensionManager) DeleteFileAfter(object Object, dest string) error {
	var errs = []error{}
	for _, ocp := range manager.contentChange {
		if err := ocp.DeleteFileAfter(object, dest); err != nil {
			errs = append(errs, err)
			continue
		}
	}
	return errors.Combine(errs...)
}

// ObjectChange
func (manager *ExtensionManager) UpdateObjectBefore(object Object) error {
	var errs = []error{}
	for _, ocp := range manager.objectChange {
		if err := ocp.UpdateObjectBefore(object); err != nil {
			errs = append(errs, err)
			continue
		}
	}
	return errors.Combine(errs...)
}
func (manager *ExtensionManager) UpdateObjectAfter(object Object) error {
	var errs = []error{}
	for _, ocp := range manager.objectChange {
		if err := ocp.UpdateObjectAfter(object); err != nil {
			errs = append(errs, err)
			continue
		}
	}
	return errors.Combine(errs...)
}

func (manager *ExtensionManager) SetFS(subfs OCFLFS) {
	for _, ext := range manager.extensions {
		extFS, err := subfs.SubFS(ext.GetName())
		if err != nil {
			panic(err)
		}
		ext.SetFS(extFS)
	}
}

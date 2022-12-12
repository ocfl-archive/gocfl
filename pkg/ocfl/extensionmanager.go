package ocfl

import (
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"go.ub.unibas.ch/gocfl/v2/pkg/checksum"
	"golang.org/x/exp/slices"
)

const ExtensionManagerName = "NNNN-gocfl-extension-manager"
const ExtensionManagerDescription = "initial extension for sorted exclusion and sorted execution"

type ExtensionManager struct {
	*ExtensionManagerConfig
	extensions         []Extension
	storageRootPath    []ExtensionStorageRootPath
	objectContentPath  []ExtensionObjectContentPath
	objectExternalPath []ExtensionObjectExternalPath
	contentChange      []ExtensionContentChange
	objectChange       []ExtensionObjectChange
	fixityDigest       []ExtensionFixityDigest
	fs                 OCFLFS
}

type ExtensionManagerConfig struct {
	*ExtensionConfig
	Sort      map[string][]string   `json:"sort"`
	Exclusion map[string][][]string `json:"exclusion"`
}

func NewExtensionManager() (*ExtensionManager, error) {
	m := &ExtensionManager{
		ExtensionManagerConfig: &ExtensionManagerConfig{
			ExtensionConfig: &ExtensionConfig{ExtensionName: ExtensionManagerName},
			Sort:            map[string][]string{},
			Exclusion:       map[string][][]string{},
		},
		extensions:        []Extension{},
		storageRootPath:   []ExtensionStorageRootPath{},
		objectContentPath: []ExtensionObjectContentPath{},
		objectChange:      []ExtensionObjectChange{},
		fixityDigest:      []ExtensionFixityDigest{},
	}
	return m, nil
}

func (manager *ExtensionManager) Add(ext Extension) error {
	// set extensionmanager config...
	if ext.GetName() == ExtensionManagerName {
		dummy, ok := ext.(*InitialDummy)
		if !ok {
			return errors.Errorf("extension '%s' not of type *InitialDummy", ext)
		}
		manager.ExtensionManagerConfig = dummy.ExtensionManagerConfig
		return nil
	}
	manager.extensions = append(manager.extensions, ext)

	if srp, ok := ext.(ExtensionStorageRootPath); ok {
		manager.storageRootPath = append(manager.storageRootPath, srp)
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
	if occ, ok := ext.(ExtensionFixityDigest); ok {
		manager.fixityDigest = append(manager.fixityDigest, occ)
	}
	if occ, ok := ext.(ExtensionObjectExternalPath); ok {
		manager.objectExternalPath = append(manager.objectExternalPath, occ)
	}
	return nil
}

func (manager *ExtensionManager) SetFS(subfs OCFLFS) {
	for _, ext := range manager.extensions {
		extFS, err := subfs.SubFS(ext.GetName())
		if err != nil {
			panic(err)
		}
		ext.SetFS(extFS)
	}
	var err error
	manager.fs, err = subfs.SubFS("initial")
	if err != nil {
		panic(err)
	}
}

func sortExtensions(list []Extension, sortName []string) {
	sortFunc := func(aExt, bExt Extension) bool {
		aName := aExt.GetName()
		bName := bExt.GetName()
		var aNum, bNum int
		for num, name := range sortName {
			if name == aName {
				aNum = num
			}
			if name == bName {
				bNum = num
			}
		}
		return aNum < bNum
	}
	slices.SortFunc(list, sortFunc)
}

func excludeExtensions(list []Extension, exclusionSort []string) []Extension {
	sortFunc := func(aExt, bExt Extension) bool {
		aName := aExt.GetName()
		bName := bExt.GetName()
		var aNum, bNum int
		for num, name := range exclusionSort {
			if name == aName {
				aNum = num
			}
			if name == bName {
				bNum = num
			}
		}
		return aNum < bNum
	}
	// first sort list
	slices.SortFunc(list, sortFunc)

	// find out, which have to be deleted
	var first = true
	var remove = []int{}
	for num, entry := range list {
		ext, ok := entry.(Extension)
		if !ok {
			panic(fmt.Sprintf("%v not an Extension", entry))
		}
		if slices.Contains(exclusionSort, ext.GetName()) {
			if first {
				first = false
				continue
			}
			remove = append(remove, num)
		}
	}

	// delete the remove-list from back
	// todo: check garbage collector behaviour
	for i := len(remove) - 1; i >= 0; i-- {
		list = slices.Delete(list, remove[i], remove[i]+1)
	}
	return list
}

func (manager *ExtensionManager) organize(list []Extension, name string) []Extension {
	if len(list) == 0 {
		return list
	}
	if excludes, ok := manager.Exclusion[name]; ok {
		for _, exclude := range excludes {
			list = excludeExtensions(list, exclude)
		}
	}
	if sort, ok := manager.Sort[name]; ok {
		sortExtensions(list, sort)
	}
	return list
}

func (manager *ExtensionManager) Finalize() {
	var extList = []Extension{}
	for _, e := range manager.storageRootPath {
		extList = append(extList, e.(Extension))
	}
	extList = manager.organize(extList, "StorageRootPath")
	manager.storageRootPath = []ExtensionStorageRootPath{}
	for _, e := range extList {
		manager.storageRootPath = append(manager.storageRootPath, e.(ExtensionStorageRootPath))
	}

	extList = []Extension{}
	for _, e := range manager.objectContentPath {
		extList = append(extList, e.(Extension))
	}
	extList = manager.organize(extList, "ObjectContentPath")
	manager.objectContentPath = []ExtensionObjectContentPath{}
	for _, e := range extList {
		manager.objectContentPath = append(manager.objectContentPath, e.(ExtensionObjectContentPath))
	}
}

// Extension
func (manager *ExtensionManager) GetName() string {
	return ExtensionManagerName
}
func (manager *ExtensionManager) WriteConfig() error {
	for _, ext := range manager.extensions {
		if err := ext.WriteConfig(); err != nil {
			return errors.Wrapf(err, "cannot store '%s'", ext.GetName())
		}
	}

	configWriter, err := manager.fs.Create("config.json")
	if err != nil {
		return errors.Wrap(err, "cannot open config.json")
	}
	defer configWriter.Close()
	jenc := json.NewEncoder(configWriter)
	jenc.SetIndent("", "   ")
	if err := jenc.Encode(manager.ExtensionManagerConfig); err != nil {
		return errors.Wrapf(err, "cannot encode config to file")
	}
	return nil
}

// StorageRootPath
func (manager *ExtensionManager) StoreRootLayout(fs OCFLFS) error {
	for _, ext := range manager.storageRootPath {
		if err := ext.WriteLayout(fs); err != nil {
			return errors.Wrapf(err, "cannot store '%v'", ext)
		}
	}
	return nil
}
func (manager *ExtensionManager) BuildStorageRootPath(storageRoot StorageRoot, id string) (string, error) {
	var errs = []error{}
	for _, srp := range manager.storageRootPath {
		p, err := srp.BuildStorageRootPath(storageRoot, id)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if p != "" {
			id = p
		}
	}
	return id, errors.Combine(errs...)
}
func (manager *ExtensionManager) WriteLayout(fs OCFLFS) error {
	if len(manager.storageRootPath) == 0 {
		return nil
	}
	if len(manager.storageRootPath) == 1 {
		return manager.storageRootPath[0].WriteLayout(fs)
	}
	configWriter, err := fs.Create("ocfl_layout.json")
	if err != nil {
		return errors.Wrap(err, "cannot open ocfl_layout.json")
	}
	defer configWriter.Close()
	jenc := json.NewEncoder(configWriter)
	jenc.SetIndent("", "   ")
	if err := jenc.Encode(struct {
		Extension   string `json:"extension"`
		Description string `json:"description"`
	}{
		Extension:   ExtensionManagerName,
		Description: ExtensionManagerDescription,
	}); err != nil {
		return errors.Wrapf(err, "cannot encode config to file")
	}
	return nil
}

// ObjectContentPath
func (manager *ExtensionManager) BuildObjectContentPath(object Object, originalPath string, area string) (string, error) {
	var errs = []error{}
	for _, ocp := range manager.objectContentPath {
		p, err := ocp.BuildObjectContentPath(object, originalPath, area)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if p != "" {
			originalPath = p
			//return p, errors.Combine(errs...)
		}
	}
	return originalPath, errors.Combine(errs...)
}

// ObjectExternalPath
func (manager *ExtensionManager) BuildObjectExternalPath(object Object, originalPath string, area string) (string, error) {
	var errs = []error{}
	for _, ocp := range manager.objectExternalPath {
		p, err := ocp.BuildObjectExternalPath(object, originalPath, area)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if p != "" {
			originalPath = p
			//return p, errors.Combine(errs...)
		}
	}
	return originalPath, errors.Combine(errs...)
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

// FixityDigests
func (manager *ExtensionManager) GetFixityDigests() []checksum.DigestAlgorithm {
	var digests = []checksum.DigestAlgorithm{}
	for _, ext := range manager.fixityDigest {
		digests = append(digests, ext.GetFixityDigests()...)
	}
	slices.Sort(digests)
	digests = slices.Compact(digests)
	return digests
}

// check interface satisfaction
var (
	_ Extension                   = &ExtensionManager{}
	_ ExtensionStorageRootPath    = &ExtensionManager{}
	_ ExtensionObjectContentPath  = &ExtensionManager{}
	_ ExtensionObjectExternalPath = &ExtensionManager{}
	_ ExtensionContentChange      = &ExtensionManager{}
	_ ExtensionObjectChange       = &ExtensionManager{}
	_ ExtensionFixityDigest       = &ExtensionManager{}
)

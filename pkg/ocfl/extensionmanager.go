package ocfl

import (
	"emperror.dev/errors"
	"encoding/json"
	"github.com/je4/gocfl/v2/pkg/checksum"
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
	objectExtractPath  []ExtensionObjectExtractPath
	fs                 OCFLFS
}

func (manager *ExtensionManager) GetConfigString() string {
	str, _ := json.MarshalIndent(manager.ExtensionManagerConfig, "", "  ")
	return string(str)
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
	if occ, ok := ext.(ExtensionObjectExtractPath); ok {
		manager.objectExtractPath = append(manager.objectExtractPath, occ)
	}
	return nil
}

func (manager *ExtensionManager) SetFS(subfs OCFLFS) {
	for _, ext := range manager.extensions {
		extFS, err := subfs.SubFSRW(ext.GetName())
		if err != nil {
			panic(err)
		}
		ext.SetFS(extFS)
	}
	var err error
	manager.fs, err = subfs.SubFSRW("initial")
	if err != nil {
		panic(err)
	}
}

func sortExtensions[E Extension](list []E, sortName []string) {
	sortFunc := func(aExt, bExt E) bool {
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

func excludeExtensions[E Extension](list []E, exclusionSort []string) []E {
	sortFunc := func(aExt, bExt E) bool {
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
		if slices.Contains(exclusionSort, entry.GetName()) {
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

func organize[E Extension](manager *ExtensionManager, list []E, name string) []E {
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
	manager.storageRootPath = organize(manager, manager.storageRootPath, ExtensionStorageRootPathName)
	manager.objectContentPath = organize(manager, manager.objectContentPath, ExtensionObjectContentPathName)
	manager.objectExtractPath = organize(manager, manager.objectExtractPath, ExtensionObjectExtractPathName)
	manager.objectExternalPath = organize(manager, manager.objectExternalPath, ExtensionObjectExternalPathName)
	manager.contentChange = organize(manager, manager.contentChange, ExtensionContentChangeName)
	manager.objectChange = organize(manager, manager.objectChange, ExtensionObjectChangeName)
	manager.fixityDigest = organize(manager, manager.fixityDigest, ExtensionFixityDigestName)

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

func (manager *ExtensionManager) SetParams(params map[string]string) error {
	for _, ext := range manager.extensions {
		if err := ext.SetParams(params); err != nil {
			return errors.Wrapf(err, "cannot set params of '%s'", ext.GetName())
		}
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
func (manager *ExtensionManager) BuildObjectExternalPath(object Object, originalPath string) (string, error) {
	var errs = []error{}
	for _, ocp := range manager.objectExternalPath {
		p, err := ocp.BuildObjectExternalPath(object, originalPath)
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

func (manager *ExtensionManager) BuildObjectExtractPath(object Object, originalPath string) (string, error) {
	var err error
	for _, ext := range manager.objectExtractPath {
		originalPath, err = ext.BuildObjectExtractPath(object, originalPath)
		if err != nil {
			return "", errors.Wrapf(err, "cannot call BuildObjectExraPath")
		}
	}
	return originalPath, nil
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
	_ ExtensionObjectExtractPath  = &ExtensionManager{}
)

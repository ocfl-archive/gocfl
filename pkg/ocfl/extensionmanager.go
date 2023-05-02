package ocfl

import (
	"bufio"
	"emperror.dev/errors"
	"encoding/json"
	"github.com/je4/filesystem/v2/pkg/writefs"
	"github.com/je4/utils/v2/pkg/checksum"
	iou "github.com/je4/utils/v2/pkg/io"
	"golang.org/x/exp/slices"
	"io"
	"io/fs"
	"sync"
)

const ExtensionManagerName = "NNNN-gocfl-extension-manager"
const ExtensionManagerDescription = "initial extension for sorted exclusion and sorted execution"

type ExtensionManager struct {
	*ExtensionManagerConfig
	extensions         []Extension
	storageRootPath    []ExtensionStorageRootPath
	objectContentPath  []ExtensionObjectContentPath
	objectExternalPath []ExtensionObjectStatePath
	contentChange      []ExtensionContentChange
	objectChange       []ExtensionObjectChange
	fixityDigest       []ExtensionFixityDigest
	objectExtractPath  []ExtensionObjectExtractPath
	metadata           []ExtensionMetadata
	area               []ExtensionArea
	stream             []ExtensionStream
	newVersion         []ExtensionNewVersion
	fsys               fs.FS
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
		metadata:          []ExtensionMetadata{},
		area:              []ExtensionArea{},
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
	if occ, ok := ext.(ExtensionObjectStatePath); ok {
		manager.objectExternalPath = append(manager.objectExternalPath, occ)
	}
	if occ, ok := ext.(ExtensionObjectExtractPath); ok {
		manager.objectExtractPath = append(manager.objectExtractPath, occ)
	}
	if meta, ok := ext.(ExtensionMetadata); ok {
		manager.metadata = append(manager.metadata, meta)
	}
	if area, ok := ext.(ExtensionArea); ok {
		manager.area = append(manager.area, area)
	}
	if stream, ok := ext.(ExtensionStream); ok {
		manager.stream = append(manager.stream, stream)
	}
	if newversion, ok := ext.(ExtensionNewVersion); ok {
		manager.newVersion = append(manager.newVersion, newversion)
	}
	return nil
}

func (manager *ExtensionManager) SetFS(fsys fs.FS) {
	for _, ext := range manager.extensions {
		extFS, err := fs.Sub(fsys, ext.GetName())
		if err != nil {
			panic(err)
		}
		ext.SetFS(extFS)
	}
	var err error
	manager.fsys, err = fs.Sub(fsys, "initial")
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
	manager.metadata = organize(manager, manager.metadata, ExtensionMetadataName)
	manager.area = organize(manager, manager.area, ExtensionAreaName)
	manager.stream = organize(manager, manager.stream, ExtensionStreamName)
	manager.newVersion = organize(manager, manager.newVersion, ExtensionNewVersionName)
}

// Extension
func (manager *ExtensionManager) IsRegistered() bool {
	return false
}
func (manager *ExtensionManager) GetName() string {
	return ExtensionManagerName
}
func (manager *ExtensionManager) WriteConfig() error {
	for _, ext := range manager.extensions {
		if err := ext.WriteConfig(); err != nil {
			return errors.Wrapf(err, "cannot store '%s'", ext.GetName())
		}
	}

	if len(manager.Exclusion) != 0 || len(manager.Sort) != 0 {
		configWriter, err := writefs.Create(manager.fsys, "config.json")
		if err != nil {
			return errors.Wrap(err, "cannot open config.json")
		}
		defer configWriter.Close()
		jenc := json.NewEncoder(configWriter)
		jenc.SetIndent("", "   ")
		if err := jenc.Encode(manager.ExtensionManagerConfig); err != nil {
			return errors.Wrapf(err, "cannot encode config to file")
		}
	}
	return nil
}

// StorageRootPath
func (manager *ExtensionManager) StoreRootLayout(fsys fs.FS) error {
	for _, ext := range manager.storageRootPath {
		if err := ext.WriteLayout(fsys); err != nil {
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
func (manager *ExtensionManager) WriteLayout(fsys fs.FS) error {
	if len(manager.storageRootPath) == 0 {
		return nil
	}
	if len(manager.storageRootPath) == 1 {
		return manager.storageRootPath[0].WriteLayout(fsys)
	}
	configWriter, err := writefs.Create(fsys, "ocfl_layout.json")
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
func (manager *ExtensionManager) BuildObjectManifestPath(object Object, originalPath string, area string) (string, error) {
	var errs = []error{}
	for _, ocp := range manager.objectContentPath {
		p, err := ocp.BuildObjectManifestPath(object, originalPath, area)
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
func (manager *ExtensionManager) BuildObjectStatePath(object Object, originalPath string, area string) (string, error) {
	var errs = []error{}
	for _, ocp := range manager.objectExternalPath {
		p, err := ocp.BuildObjectStatePath(object, originalPath, area)
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
func (manager *ExtensionManager) AddFileBefore(object Object, sourceFS fs.FS, source string, dest string, area string, isDir bool) error {
	var errs = []error{}
	for _, ocp := range manager.contentChange {
		if err := ocp.AddFileBefore(object, sourceFS, source, dest, area, isDir); err != nil {
			errs = append(errs, err)
			continue
		}
	}
	return errors.Combine(errs...)
}
func (manager *ExtensionManager) UpdateFileBefore(object Object, sourceFS fs.FS, source, dest, area string, isDir bool) error {
	var errs = []error{}
	for _, ocp := range manager.contentChange {
		if err := ocp.UpdateFileBefore(object, sourceFS, source, dest, area, isDir); err != nil {
			errs = append(errs, err)
			continue
		}
	}
	return errors.Combine(errs...)
}
func (manager *ExtensionManager) DeleteFileBefore(object Object, dest string, area string) error {
	var errs = []error{}
	for _, ocp := range manager.contentChange {
		if err := ocp.DeleteFileBefore(object, dest, area); err != nil {
			errs = append(errs, err)
			continue
		}
	}
	return errors.Combine(errs...)
}
func (manager *ExtensionManager) AddFileAfter(object Object, sourceFS fs.FS, source []string, internalPath, digest, area string, isDir bool) error {
	var errs = []error{}
	for _, ocp := range manager.contentChange {
		if err := ocp.AddFileAfter(object, sourceFS, source, internalPath, digest, area, isDir); err != nil {
			errs = append(errs, err)
			continue
		}
	}
	return errors.Combine(errs...)
}
func (manager *ExtensionManager) UpdateFileAfter(object Object, sourceFS fs.FS, source, dest, area string, isDir bool) error {
	var errs = []error{}
	for _, ocp := range manager.contentChange {
		if err := ocp.UpdateFileAfter(object, sourceFS, source, dest, area, isDir); err != nil {
			errs = append(errs, err)
			continue
		}
	}
	return errors.Combine(errs...)
}
func (manager *ExtensionManager) DeleteFileAfter(object Object, dest string, area string) error {
	var errs = []error{}
	for _, ocp := range manager.contentChange {
		if err := ocp.DeleteFileAfter(object, dest, area); err != nil {
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

func (manager *ExtensionManager) BuildObjectExtractPath(object Object, originalPath string, area string) (string, error) {
	var err error
	for _, ext := range manager.objectExtractPath {
		originalPath, err = ext.BuildObjectExtractPath(object, originalPath, area)
		if err != nil {
			return "", errors.Wrapf(err, "cannot call BuildObjectExtractPath")
		}
	}
	return originalPath, nil
}

func (manager *ExtensionManager) GetMetadata(object Object) (map[string]any, error) {
	var metaResult = map[string]map[string]any{}
	for _, ext := range manager.metadata {
		meta, err := ext.GetMetadata(object)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot call GetMetadata() from extension '%s'", ext.GetName())
		}
		for h, val := range meta {
			if _, ok := metaResult[h]; !ok {
				metaResult[h] = map[string]any{}
			}
			name := ext.GetName()
			metaResult[h][name] = val
		}
	}
	var result = map[string]any{}
	for h, val := range metaResult {
		result[h] = val
	}
	return result, nil
}

func (manager *ExtensionManager) GetAreaPath(object Object, area string) (string, error) {
	var errs = []error{}
	for _, ext := range manager.area {
		path, err := ext.GetAreaPath(object, area)
		if err != nil {
			errs = append(errs, errors.Wrapf(err, "cannot call GetArea(%s) from extension '%s'", area, ext.GetName()))
		}
		if path != "" {
			return path, nil
		}
	}

	return "", errors.Combine(errs...)
}

// NewVersion
func (manager *ExtensionManager) NeedNewVersion(object Object) (bool, error) {
	for _, ext := range manager.newVersion {
		need, err := ext.NeedNewVersion(object)
		if err != nil {
			return false, errors.Wrapf(err, "cannot call NeedNewVersion() from extension '%s'", ext.GetName())
		}
		if need {
			return true, nil
		}
	}
	return false, nil
}

func (manager *ExtensionManager) DoNewVersion(object Object) error {
	for _, ext := range manager.newVersion {
		if err := ext.DoNewVersion(object); err != nil {
			return errors.Wrapf(err, "cannot call NeedNewVersion() from extension '%s'", ext.GetName())
		}
	}
	return nil
}

// Stream
func (manager *ExtensionManager) StreamObject(object Object, reader io.Reader, stateFiles []string, dest string) error {
	if len(manager.stream) == 0 {
		_, _ = io.Copy(io.Discard, reader)
		return nil
	}
	var wg = sync.WaitGroup{}
	writer := []*iou.WriteIgnoreCloser{}
	extErrors := make(chan error, len(manager.stream))
	for _, ext := range manager.stream {
		wg.Add(1)
		pr, pw := io.Pipe()
		writer = append(writer, iou.NewWriteIgnoreCloser(pw))
		go func(r io.Reader, extension ExtensionStream) {
			defer wg.Done()
			if err := extension.StreamObject(object, r, stateFiles, dest); err != nil {
				extErrors <- errors.Wrapf(err, "cannot call StreamObject() from extension '%s' for object '%s'", extension.GetName(), object.GetID())
			}
			// discard remaining data
			_, _ = io.Copy(io.Discard, r)
		}(pr, ext)
	}
	var ws = []io.Writer{}
	for _, w := range writer {
		ws = append(ws, bufio.NewWriterSize(w, 1024*1024))
	}
	multiWriter := io.MultiWriter(ws...)
	_, err := io.Copy(multiWriter, reader)
	for _, w := range ws {
		// it's sure that w is a bufio.Writer
		if err := w.(*bufio.Writer).Flush(); err != nil {
			return errors.Wrap(err, "cannot flush buffer")
		}
	}
	for _, w := range writer {
		w.ForceClose()
	}
	if err != nil {
		return errors.Wrap(err, "cannot copy stream to actions")
	}
	// wait for all actions to finish
	wg.Wait()
	close(extErrors)
	var errs = []error{}
	for err := range extErrors {
		errs = append(errs, err)
	}
	return errors.Combine(errs...)
}

// check interface satisfaction
var (
	_ Extension                  = (*ExtensionManager)(nil)
	_ ExtensionStorageRootPath   = (*ExtensionManager)(nil)
	_ ExtensionObjectContentPath = (*ExtensionManager)(nil)
	_ ExtensionObjectStatePath   = (*ExtensionManager)(nil)
	_ ExtensionContentChange     = (*ExtensionManager)(nil)
	_ ExtensionObjectChange      = (*ExtensionManager)(nil)
	_ ExtensionFixityDigest      = (*ExtensionManager)(nil)
	_ ExtensionObjectExtractPath = (*ExtensionManager)(nil)
	_ ExtensionMetadata          = (*ExtensionManager)(nil)
	_ ExtensionArea              = (*ExtensionManager)(nil)
	_ ExtensionStream            = (*ExtensionManager)(nil)
	_ ExtensionNewVersion        = (*ExtensionManager)(nil)
)

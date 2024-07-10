package extension

import (
	"bufio"
	"cmp"
	"emperror.dev/errors"
	"encoding/json"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/je4/utils/v2/pkg/checksum"
	iou "github.com/je4/utils/v2/pkg/io"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl"
	"golang.org/x/exp/slices"
	"io"
	"io/fs"
	"sync"
)

const GOCFLExtensionManagerName = "NNNN-gocfl-extension-manager"
const GOCFLExtensionManagerDescription = "initial extension for sorted exclusion and sorted execution"

func NewGOCFLExtensionManagerFS(fsys fs.FS) (*GOCFLExtensionManager, error) {
	var config = &ocfl.ExtensionManagerConfig{
		ExtensionConfig: &ocfl.ExtensionConfig{
			ExtensionName: GOCFLExtensionManagerName,
		},
		Sort:      map[string][]string{},
		Exclusion: map[string][][]string{},
	}
	if fsys != nil {
		fp, err := fsys.Open("config.json")
		if err != nil {
			return nil, errors.Wrap(err, "cannot open config.json")
		}
		defer fp.Close()
		data, err := io.ReadAll(fp)
		if err != nil {
			return nil, errors.Wrap(err, "cannot read config.json")
		}
		if err := json.Unmarshal(data, config); err != nil {
			return nil, errors.Wrapf(err, "cannot unmarshal ContentSubPathConfig '%s'", string(data))
		}
	}

	return NewGOCFLExtensionManager(config)
}

func NewGOCFLExtensionManager(config *ocfl.ExtensionManagerConfig) (*GOCFLExtensionManager, error) {
	m := &GOCFLExtensionManager{
		ExtensionManagerConfig: config,
		extensions:             []ocfl.Extension{},
		storageRootPath:        []ocfl.ExtensionStorageRootPath{},
		objectContentPath:      []ocfl.ExtensionObjectContentPath{},
		objectChange:           []ocfl.ExtensionObjectChange{},
		fixityDigest:           []ocfl.ExtensionFixityDigest{},
		metadata:               []ocfl.ExtensionMetadata{},
		area:                   []ocfl.ExtensionArea{},
	}
	return m, nil
}

type GOCFLExtensionManager struct {
	*ocfl.ExtensionManagerConfig
	extensions         []ocfl.Extension
	storageRootPath    []ocfl.ExtensionStorageRootPath
	objectContentPath  []ocfl.ExtensionObjectContentPath
	objectExternalPath []ocfl.ExtensionObjectStatePath
	contentChange      []ocfl.ExtensionContentChange
	objectChange       []ocfl.ExtensionObjectChange
	fixityDigest       []ocfl.ExtensionFixityDigest
	objectExtractPath  []ocfl.ExtensionObjectExtractPath
	metadata           []ocfl.ExtensionMetadata
	area               []ocfl.ExtensionArea
	stream             []ocfl.ExtensionStream
	newVersion         []ocfl.ExtensionNewVersion
	fsys               fs.FS
	initial            ocfl.ExtensionInitial
}

func (manager *GOCFLExtensionManager) Terminate() error {
	var errs = []error{}
	for _, ext := range manager.extensions {
		if err := ext.Terminate(); err != nil {
			errs = append(errs, errors.Wrapf(err, "cannot terminate '%s'", ext.GetName()))
		}
	}
	return errors.Combine(errs...)
}

func (manager *GOCFLExtensionManager) SetInitial(initial ocfl.ExtensionInitial) {
	manager.initial = initial
}

func (manager *GOCFLExtensionManager) GetExtensions() []ocfl.Extension {
	return manager.extensions
}

func (manager *GOCFLExtensionManager) GetFS() fs.FS {
	return manager.fsys
}

func (manager *GOCFLExtensionManager) GetConfig() any {
	return manager.ExtensionManagerConfig
}

func (manager *GOCFLExtensionManager) GetConfigName(extName string) (any, error) {
	for _, ext := range manager.extensions {
		if ext.GetName() == extName {
			return ext.GetConfig(), nil
		}
	}
	return nil, errors.Errorf("extension '%s' not active", extName)
}

func (manager *GOCFLExtensionManager) Add(ext ocfl.Extension) error {
	// set extensionmanager config...
	if ext.GetName() == GOCFLExtensionManagerName {
		return errors.Errorf("cannot add extension '%s' to itself", GOCFLExtensionManagerName)
	}
	manager.extensions = append(manager.extensions, ext)

	if srp, ok := ext.(ocfl.ExtensionStorageRootPath); ok {
		manager.storageRootPath = append(manager.storageRootPath, srp)
	}
	if ocp, ok := ext.(ocfl.ExtensionObjectContentPath); ok {
		manager.objectContentPath = append(manager.objectContentPath, ocp)
	}
	if occ, ok := ext.(ocfl.ExtensionContentChange); ok {
		manager.contentChange = append(manager.contentChange, occ)
	}
	if occ, ok := ext.(ocfl.ExtensionObjectChange); ok {
		manager.objectChange = append(manager.objectChange, occ)
	}
	if occ, ok := ext.(ocfl.ExtensionFixityDigest); ok {
		manager.fixityDigest = append(manager.fixityDigest, occ)
	}
	if occ, ok := ext.(ocfl.ExtensionObjectStatePath); ok {
		manager.objectExternalPath = append(manager.objectExternalPath, occ)
	}
	if occ, ok := ext.(ocfl.ExtensionObjectExtractPath); ok {
		manager.objectExtractPath = append(manager.objectExtractPath, occ)
	}
	if meta, ok := ext.(ocfl.ExtensionMetadata); ok {
		manager.metadata = append(manager.metadata, meta)
	}
	if area, ok := ext.(ocfl.ExtensionArea); ok {
		manager.area = append(manager.area, area)
	}
	if stream, ok := ext.(ocfl.ExtensionStream); ok {
		manager.stream = append(manager.stream, stream)
	}
	if newversion, ok := ext.(ocfl.ExtensionNewVersion); ok {
		manager.newVersion = append(manager.newVersion, newversion)
	}
	return nil
}

func (manager *GOCFLExtensionManager) SetFS(fsys fs.FS, create bool) {
	if fsys == nil {
		return
	}
	for _, ext := range append(manager.extensions, manager.initial) {
		if ext == nil {
			continue
		}
		extFS, err := fs.Sub(fsys, ext.GetName())
		if err != nil {
			if create && errors.Is(err, fs.ErrNotExist) {
				extFS, err = writefs.SubFSCreate(fsys, ext.GetName())
			}
			if err != nil {
				panic(err)
			}
		}
		ext.SetFS(extFS, create)
	}
	var err error
	manager.fsys, err = fs.Sub(fsys, manager.GetName())
	if err != nil {
		if create && errors.Is(err, fs.ErrNotExist) {
			manager.fsys, err = writefs.SubFSCreate(fsys, manager.GetName())
		}
		if err != nil {
			panic(err)
		}
	}
}

func (manager *GOCFLExtensionManager) GetFSName(extName string) (fs.FS, error) {
	for _, ext := range manager.extensions {
		if ext.GetName() == extName {
			return ext.GetFS(), nil
		}
	}
	return nil, errors.Errorf("extension '%s' not active", extName)
}

func sortExtensions[E ocfl.Extension](list []E, sortName []string) {
	sortFunc := func(aExt, bExt E) int {
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
		return cmp.Compare(aNum, bNum)
	}
	slices.SortFunc(list, sortFunc)
}

func excludeExtensions[E ocfl.Extension](list []E, exclusionSort []string) []E {
	sortFunc := func(aExt, bExt E) int {
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
		return cmp.Compare(aNum, bNum)
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

func organize[E ocfl.Extension](manager *GOCFLExtensionManager, list []E, name string) []E {
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

func (manager *GOCFLExtensionManager) Finalize() {
	manager.storageRootPath = organize(manager, manager.storageRootPath, ocfl.ExtensionStorageRootPathName)
	manager.objectContentPath = organize(manager, manager.objectContentPath, ocfl.ExtensionObjectContentPathName)
	manager.objectExtractPath = organize(manager, manager.objectExtractPath, ocfl.ExtensionObjectExtractPathName)
	manager.objectExternalPath = organize(manager, manager.objectExternalPath, ocfl.ExtensionObjectExternalPathName)
	manager.contentChange = organize(manager, manager.contentChange, ocfl.ExtensionContentChangeName)
	manager.objectChange = organize(manager, manager.objectChange, ocfl.ExtensionObjectChangeName)
	manager.fixityDigest = organize(manager, manager.fixityDigest, ocfl.ExtensionFixityDigestName)
	manager.metadata = organize(manager, manager.metadata, ocfl.ExtensionMetadataName)
	manager.area = organize(manager, manager.area, ocfl.ExtensionAreaName)
	manager.stream = organize(manager, manager.stream, ocfl.ExtensionStreamName)
	manager.newVersion = organize(manager, manager.newVersion, ocfl.ExtensionNewVersionName)
}

// Extension
func (manager *GOCFLExtensionManager) IsRegistered() bool {
	return false
}
func (manager *GOCFLExtensionManager) GetName() string {
	return GOCFLExtensionManagerName
}
func (manager *GOCFLExtensionManager) WriteConfig() error {
	for _, ext := range append(manager.extensions, manager.initial) {
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
func (manager *GOCFLExtensionManager) StoreRootLayout(fsys fs.FS) error {
	for _, ext := range manager.storageRootPath {
		if err := ext.WriteLayout(fsys); err != nil {
			return errors.Wrapf(err, "cannot store '%v'", ext)
		}
	}
	return nil
}
func (manager *GOCFLExtensionManager) BuildStorageRootPath(storageRoot ocfl.StorageRoot, id string) (string, error) {
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
func (manager *GOCFLExtensionManager) WriteLayout(fsys fs.FS) error {
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
		Extension:   GOCFLExtensionManagerName,
		Description: GOCFLExtensionManagerDescription,
	}); err != nil {
		return errors.Wrapf(err, "cannot encode config to file")
	}
	return nil
}

func (manager *GOCFLExtensionManager) SetParams(params map[string]string) error {
	for _, ext := range manager.extensions {
		if err := ext.SetParams(params); err != nil {
			return errors.Wrapf(err, "cannot set params of '%s'", ext.GetName())
		}
	}
	return nil
}

// ObjectContentPath
func (manager *GOCFLExtensionManager) BuildObjectManifestPath(object ocfl.Object, originalPath string, area string) (string, error) {
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
func (manager *GOCFLExtensionManager) BuildObjectStatePath(object ocfl.Object, originalPath string, area string) (string, error) {
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
func (manager *GOCFLExtensionManager) AddFileBefore(object ocfl.Object, sourceFS fs.FS, source string, dest string, area string, isDir bool) error {
	var errs = []error{}
	for _, ocp := range manager.contentChange {
		if err := ocp.AddFileBefore(object, sourceFS, source, dest, area, isDir); err != nil {
			errs = append(errs, err)
			continue
		}
	}
	return errors.Combine(errs...)
}
func (manager *GOCFLExtensionManager) UpdateFileBefore(object ocfl.Object, sourceFS fs.FS, source, dest, area string, isDir bool) error {
	var errs = []error{}
	for _, ocp := range manager.contentChange {
		if err := ocp.UpdateFileBefore(object, sourceFS, source, dest, area, isDir); err != nil {
			errs = append(errs, err)
			continue
		}
	}
	return errors.Combine(errs...)
}
func (manager *GOCFLExtensionManager) DeleteFileBefore(object ocfl.Object, dest string, area string) error {
	var errs = []error{}
	for _, ocp := range manager.contentChange {
		if err := ocp.DeleteFileBefore(object, dest, area); err != nil {
			errs = append(errs, err)
			continue
		}
	}
	return errors.Combine(errs...)
}
func (manager *GOCFLExtensionManager) AddFileAfter(object ocfl.Object, sourceFS fs.FS, source []string, internalPath, digest, area string, isDir bool) error {
	var errs = []error{}
	for _, ocp := range manager.contentChange {
		if err := ocp.AddFileAfter(object, sourceFS, source, internalPath, digest, area, isDir); err != nil {
			errs = append(errs, err)
			continue
		}
	}
	return errors.Combine(errs...)
}
func (manager *GOCFLExtensionManager) UpdateFileAfter(object ocfl.Object, sourceFS fs.FS, source, dest, area string, isDir bool) error {
	var errs = []error{}
	for _, ocp := range manager.contentChange {
		if err := ocp.UpdateFileAfter(object, sourceFS, source, dest, area, isDir); err != nil {
			errs = append(errs, err)
			continue
		}
	}
	return errors.Combine(errs...)
}
func (manager *GOCFLExtensionManager) DeleteFileAfter(object ocfl.Object, dest string, area string) error {
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
func (manager *GOCFLExtensionManager) UpdateObjectBefore(object ocfl.Object) error {
	var errs = []error{}
	for _, ocp := range manager.objectChange {
		if err := ocp.UpdateObjectBefore(object); err != nil {
			errs = append(errs, err)
			continue
		}
	}
	return errors.Combine(errs...)
}
func (manager *GOCFLExtensionManager) UpdateObjectAfter(object ocfl.Object) error {
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
func (manager *GOCFLExtensionManager) GetFixityDigests() []checksum.DigestAlgorithm {
	var digests = []checksum.DigestAlgorithm{}
	for _, ext := range manager.fixityDigest {
		digests = append(digests, ext.GetFixityDigests()...)
	}
	slices.Sort(digests)
	digests = slices.Compact(digests)
	return digests
}

func (manager *GOCFLExtensionManager) BuildObjectExtractPath(object ocfl.Object, originalPath string, area string) (string, error) {
	var err error
	for _, ext := range manager.objectExtractPath {
		originalPath, err = ext.BuildObjectExtractPath(object, originalPath, area)
		if err != nil {
			return "", errors.Wrapf(err, "cannot call BuildObjectExtractPath")
		}
	}
	return originalPath, nil
}

func (manager *GOCFLExtensionManager) GetMetadata(object ocfl.Object) (map[string]any, error) {
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

func (manager *GOCFLExtensionManager) GetAreaPath(object ocfl.Object, area string) (string, error) {
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
func (manager *GOCFLExtensionManager) NeedNewVersion(object ocfl.Object) (bool, error) {
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

func (manager *GOCFLExtensionManager) DoNewVersion(object ocfl.Object) error {
	for _, ext := range manager.newVersion {
		if err := ext.DoNewVersion(object); err != nil {
			return errors.Wrapf(err, "cannot call NeedNewVersion() from extension '%s'", ext.GetName())
		}
	}
	return nil
}

// Stream
func (manager *GOCFLExtensionManager) StreamObject(object ocfl.Object, reader io.Reader, stateFiles []string, dest string) error {
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
		go func(r io.Reader, extension ocfl.ExtensionStream) {
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
	_ ocfl.ExtensionManager = (*GOCFLExtensionManager)(nil)
)

package extension

import (
	"bufio"
	"cmp"
	"encoding/json"
	"io"
	"io/fs"
	"sync"

	"emperror.dev/errors"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/je4/utils/v2/pkg/checksum"
	iou "github.com/je4/utils/v2/pkg/io"
	"github.com/ocfl-archive/gocfl/v3/pkg/appendfs"
	extension2 "github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/storageroot"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
	"golang.org/x/exp/slices"
)

const GOCFLExtensionManagerName = "NNNN-gocfl-extension-manager"
const GOCFLExtensionManagerDescription = "initial extension for sorted exclusion and sorted execution"

func init() {
	extension2.RegisterExtension(GOCFLExtensionManagerName, NewGOCFLExtensionManager, nil)
}

func NewGOCFLExtensionManager() (extension2.Extension, error) {
	var config = &extension2.ManagerConfig{
		ExtensionConfig: &extension2.ExtensionConfig{
			ExtensionName: GOCFLExtensionManagerName,
		},
		Sort:      map[string][]string{},
		Exclusion: map[string][][]string{},
	}
	m := &GOCFLExtensionManager{
		ManagerConfig:     config,
		extensions:        []extension2.Extension{},
		storageRootPath:   []storageroot.ExtensionStorageRootPath{},
		objectContentPath: []object.ExtensionObjectContentPath{},
		objectChange:      []object.ExtensionObjectChange{},
		fixityDigest:      []object.ExtensionFixityDigest{},
		metadata:          []object.ExtensionMetadata{},
		area:              []object.ExtensionArea{},
	}
	return m, nil
}

type GOCFLExtensionManager struct {
	*extension2.ManagerConfig
	extensions         []extension2.Extension
	storageRootPath    []storageroot.ExtensionStorageRootPath
	objectContentPath  []object.ExtensionObjectContentPath
	objectExternalPath []object.ExtensionObjectStatePath
	contentChange      []object.ExtensionContentChange
	objectChange       []object.ExtensionObjectChange
	fixityDigest       []object.ExtensionFixityDigest
	objectExtractPath  []object.ExtensionObjectExtractPath
	metadata           []object.ExtensionMetadata
	area               []object.ExtensionArea
	stream             []object.ExtensionStream
	newVersion         []object.ExtensionNewVersion
	initial            extension2.Initial
	logger             ocfllogger.OCFLLogger
}

func (manager *GOCFLExtensionManager) WithLogger(logger ocfllogger.OCFLLogger) extension2.Extension {
	manager.logger = logger.With("extension", GOCFLExtensionManagerName)
	return manager
}

func (manager *GOCFLExtensionManager) Load(fsys fs.FS) error {
	data, err := fs.ReadFile(fsys, "config.json")
	if err != nil {
		return errors.Wrap(err, "cannot read config.json")
	}

	if err := json.Unmarshal(data, manager.ManagerConfig); err != nil {
		return errors.Wrapf(err, "cannot unmarshal ExtensionManagerConfig '%s'", string(data))
	}
	return nil
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

func (manager *GOCFLExtensionManager) SetInitial(initial extension2.Initial) {
	manager.initial = initial
}

func (manager *GOCFLExtensionManager) GetExtensions() []extension2.Extension {
	return manager.extensions
}

func (manager *GOCFLExtensionManager) GetConfig() any {
	return manager.ManagerConfig
}

func (manager *GOCFLExtensionManager) GetConfigName(extName string) (any, error) {
	for _, ext := range manager.extensions {
		if ext.GetName() == extName {
			return ext.GetConfig(), nil
		}
	}
	return nil, errors.Errorf("extension '%s' not active", extName)
}

func (manager *GOCFLExtensionManager) Add(ext extension2.Extension) error {
	// set extensionmanager config...
	if ext.GetName() == GOCFLExtensionManagerName {
		return errors.Errorf("cannot add extension '%s' to itself", GOCFLExtensionManagerName)
	}
	manager.extensions = append(manager.extensions, ext)

	if srp, ok := ext.(storageroot.ExtensionStorageRootPath); ok {
		manager.storageRootPath = append(manager.storageRootPath, srp)
	}
	if ocp, ok := ext.(object.ExtensionObjectContentPath); ok {
		manager.objectContentPath = append(manager.objectContentPath, ocp)
	}
	if occ, ok := ext.(object.ExtensionContentChange); ok {
		manager.contentChange = append(manager.contentChange, occ)
	}
	if occ, ok := ext.(object.ExtensionObjectChange); ok {
		manager.objectChange = append(manager.objectChange, occ)
	}
	if occ, ok := ext.(object.ExtensionFixityDigest); ok {
		manager.fixityDigest = append(manager.fixityDigest, occ)
	}
	if occ, ok := ext.(object.ExtensionObjectStatePath); ok {
		manager.objectExternalPath = append(manager.objectExternalPath, occ)
	}
	if occ, ok := ext.(object.ExtensionObjectExtractPath); ok {
		manager.objectExtractPath = append(manager.objectExtractPath, occ)
	}
	if meta, ok := ext.(object.ExtensionMetadata); ok {
		manager.metadata = append(manager.metadata, meta)
	}
	if area, ok := ext.(object.ExtensionArea); ok {
		manager.area = append(manager.area, area)
	}
	if stream, ok := ext.(object.ExtensionStream); ok {
		manager.stream = append(manager.stream, stream)
	}
	if newversion, ok := ext.(object.ExtensionNewVersion); ok {
		manager.newVersion = append(manager.newVersion, newversion)
	}
	return nil
}

func (manager *GOCFLExtensionManager) GetFSName(extName string) (fs.FS, error) {
	return nil, errors.Errorf("extension '%s' not active", extName)
}

func sortExtensions[E extension2.Extension](list []E, sortName []string) {
	sortFunc := func(aExt, bExt E) int {
		/*
			if aExt == nil {
				return -1
			}
			if bExt == nil {
				return 1
			}
		*/
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

func excludeExtensions[E extension2.Extension](list []E, exclusionSort []string) []E {
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
		/*
			if entry == nil {
				continue
			}
		*/
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

func organize[E extension2.Extension](manager *GOCFLExtensionManager, list []E, name string) []E {
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
	manager.storageRootPath = organize(manager, manager.storageRootPath, object.ExtensionStorageRootPathName)
	manager.objectContentPath = organize(manager, manager.objectContentPath, object.ExtensionObjectContentPathName)
	manager.objectExtractPath = organize(manager, manager.objectExtractPath, object.ExtensionObjectExtractPathName)
	manager.objectExternalPath = organize(manager, manager.objectExternalPath, object.ExtensionObjectExternalPathName)
	manager.contentChange = organize(manager, manager.contentChange, object.ExtensionContentChangeName)
	manager.objectChange = organize(manager, manager.objectChange, object.ExtensionObjectChangeName)
	manager.fixityDigest = organize(manager, manager.fixityDigest, object.ExtensionFixityDigestName)
	manager.metadata = organize(manager, manager.metadata, object.ExtensionMetadataName)
	manager.area = organize(manager, manager.area, object.ExtensionAreaName)
	manager.stream = organize(manager, manager.stream, object.ExtensionStreamName)
	manager.newVersion = organize(manager, manager.newVersion, object.ExtensionNewVersionName)
}

// Extension
func (manager *GOCFLExtensionManager) IsRegistered() bool {
	return false
}
func (manager *GOCFLExtensionManager) GetName() string {
	return GOCFLExtensionManagerName
}
func (manager *GOCFLExtensionManager) WriteConfig(fsys appendfs.FS) error {
	if manager.initial == nil {
		return errors.New("no extension manager initial")
	}
	for _, ext := range append(manager.extensions, manager.initial) {
		subFS, err := appendfs.Sub(fsys, ext.GetName())
		if err != nil {
			return errors.Wrapf(err, "cannot create sub filesystem for %v/%s", fsys, ext.GetName())
		}
		if err := ext.WriteConfig(subFS); err != nil {
			return errors.Wrapf(err, "cannot store '%s'", ext.GetName())
		}
	}

	if len(manager.Exclusion) != 0 || len(manager.Sort) != 0 {
		subFS, err := appendfs.Sub(fsys, manager.GetName())
		if err != nil {
			return errors.Wrapf(err, "cannot create sub filesystem for %v/%s", fsys, manager.GetName())
		}
		configWriter, err := writefs.Create(subFS, "config.json")
		if err != nil {
			return errors.Wrap(err, "cannot open config.json")
		}
		defer configWriter.Close()
		jenc := json.NewEncoder(configWriter)
		jenc.SetIndent("", "   ")
		if err := jenc.Encode(manager.ManagerConfig); err != nil {
			return errors.Wrapf(err, "cannot encode config to file")
		}
	}
	return nil
}

// StorageRootPath
func (manager *GOCFLExtensionManager) StoreRootLayout(fsys appendfs.FS) error {
	for _, ext := range manager.storageRootPath {
		if err := ext.WriteLayout(fsys); err != nil {
			return errors.Wrapf(err, "cannot store '%v'", ext)
		}
	}
	return nil
}
func (manager *GOCFLExtensionManager) BuildStorageRootPath(storageRoot storageroot.StorageRoot, id string) (string, error) {
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
func (manager *GOCFLExtensionManager) WriteLayout(fsys appendfs.FS) error {
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
func (manager *GOCFLExtensionManager) BuildObjectManifestPath(originalPath string, area string) (string, error) {
	var errs = []error{}
	for _, ocp := range manager.objectContentPath {
		p, err := ocp.BuildObjectManifestPath(originalPath, area)
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
func (manager *GOCFLExtensionManager) BuildObjectStatePath(originalPath string, area string) (string, error) {
	var errs = []error{}
	for _, ocp := range manager.objectExternalPath {
		p, err := ocp.BuildObjectStatePath(originalPath, area)
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
func (manager *GOCFLExtensionManager) AddFileBefore(object object.VersionWriter, sourceFS fs.FS, source string, dest string, area string, isDir bool) error {
	var errs = []error{}
	for _, ocp := range manager.contentChange {
		if err := ocp.AddFileBefore(object, sourceFS, source, dest, area, isDir); err != nil {
			errs = append(errs, err)
			continue
		}
	}
	return errors.Combine(errs...)
}
func (manager *GOCFLExtensionManager) UpdateFileBefore(object object.VersionWriter, sourceFS fs.FS, source, dest, area string, isDir bool) error {
	var errs = []error{}
	for _, ocp := range manager.contentChange {
		if err := ocp.UpdateFileBefore(object, sourceFS, source, dest, area, isDir); err != nil {
			errs = append(errs, err)
			continue
		}
	}
	return errors.Combine(errs...)
}
func (manager *GOCFLExtensionManager) DeleteFileBefore(versionWriter object.VersionWriter, dest string, area string) error {
	var errs = []error{}
	for _, ocp := range manager.contentChange {
		if err := ocp.DeleteFileBefore(versionWriter, dest, area); err != nil {
			errs = append(errs, err)
			continue
		}
	}
	return errors.Combine(errs...)
}
func (manager *GOCFLExtensionManager) AddFileAfter(versionWriter object.VersionWriter, sourceFS fs.FS, source []string, internalPath, digest, area string, isDir bool) error {
	var errs = []error{}
	for _, ocp := range manager.contentChange {
		if err := ocp.AddFileAfter(versionWriter, sourceFS, source, internalPath, digest, area, isDir); err != nil {
			errs = append(errs, err)
			continue
		}
	}
	return errors.Combine(errs...)
}
func (manager *GOCFLExtensionManager) UpdateFileAfter(object object.VersionWriter, sourceFS fs.FS, source, area string, isDir bool) error {
	var errs = []error{}
	for _, ocp := range manager.contentChange {
		if err := ocp.UpdateFileAfter(object, sourceFS, source, area, isDir); err != nil {
			errs = append(errs, err)
			continue
		}
	}
	return errors.Combine(errs...)
}
func (manager *GOCFLExtensionManager) DeleteFileAfter(object object.VersionWriter, dest string, area string) error {
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
func (manager *GOCFLExtensionManager) UpdateObjectBefore(object object.VersionWriter) error {
	var errs = []error{}
	for _, ocp := range manager.objectChange {
		if err := ocp.UpdateObjectBefore(object); err != nil {
			errs = append(errs, err)
			continue
		}
	}
	return errors.Combine(errs...)
}
func (manager *GOCFLExtensionManager) UpdateObjectAfter(object object.VersionWriter) error {
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

func (manager *GOCFLExtensionManager) BuildObjectExtractPath(originalPath string, area string) (string, error) {
	var err error
	for _, ext := range manager.objectExtractPath {
		originalPath, err = ext.BuildObjectExtractPath(originalPath, area)
		if err != nil {
			return "", errors.Wrapf(err, "cannot call BuildObjectExtractPath")
		}
	}
	return originalPath, nil
}

func (manager *GOCFLExtensionManager) GetMetadata(sourceFS fs.FS, obj object.Object) (map[string]any, error) {
	var metaResult = map[string]map[string]any{}
	for _, ext := range manager.metadata {
		meta, err := ext.GetMetadata(sourceFS, obj)
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

func (manager *GOCFLExtensionManager) GetAreaPath(area string) (string, error) {
	var errs = []error{}
	for _, ext := range manager.area {
		path, err := ext.GetAreaPath(area)
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
func (manager *GOCFLExtensionManager) NeedNewVersion(object object.VersionWriter) (bool, error) {
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

func (manager *GOCFLExtensionManager) DoNewVersion(versionWriter object.VersionWriter) error {
	for _, ext := range manager.newVersion {
		if err := ext.DoNewVersion(versionWriter); err != nil {
			return errors.Wrapf(err, "cannot call NeedNewVersion() from extension '%s'", ext.GetName())
		}
	}
	return nil
}

// Stream
func (manager *GOCFLExtensionManager) StreamObject(versionWriter object.VersionWriter, reader io.Reader, stateFiles []string, dest string) error {
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
		go func(r io.Reader, ext object.ExtensionStream) {
			defer wg.Done()
			if err := ext.StreamObject(versionWriter, r, stateFiles, dest); err != nil {
				extErrors <- errors.Wrapf(err, "cannot call StreamObject() from extension '%s' for object '%s'", ext.GetName(), versionWriter.GetID())
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
	_ extension2.ManagerCore = (*GOCFLExtensionManager)(nil)
)

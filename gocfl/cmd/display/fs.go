package display

import (
	"io/fs"
	"math"
	"path"
	"strings"
	"time"

	"emperror.dev/errors"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/inventory"
	"golang.org/x/exp/slices"
)

type FSType uint8

const (
	FSTypeStage FSType = iota
	FSTypeManifest
)

const (
	StateDir    = "state"
	ManifestDir = "manifest"
	DataDir     = "data"
)

func NewObjectFS(obj ocfl.Object) (*ObjectFS, error) {
	/*
		metadata, err := obj.GetMetadata()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get metadata")
		}
	*/
	return &ObjectFS{
		object: obj,
		//		metadata:  metadata,
		inventory: obj.GetInventory(),
		manifest:  obj.GetInventory().GetManifest(),
	}, nil
}

type ObjectFS struct {
	object ocfl.Object
	//	metadata  *ocfl.ObjectMetadata
	inventory inventory.Inventory
	manifest  map[string][]string
}

func (o *ObjectFS) readDir(name string, num int) (files []fs.DirEntry, err error) {
	name = strings.TrimSuffix(name, "/")
	if name == "." {
		if num == -1 {
			num = math.MaxInt
		}
		result := []fs.DirEntry{
			NewObjectFileInfoDir(StateDir, time.Time{}),
			NewObjectFileInfoDir(ManifestDir, time.Time{}),
			NewObjectFileInfoDir(DataDir, time.Time{}),
		}
		if num > len(result) {
			return result, nil
		}
		return result[:num], nil
	}
	name = path.Clean(name)
	parts := strings.SplitN(name, "/", 2)
	if len(parts) == 0 {
		return nil, errors.Wrapf(fs.ErrNotExist, "invalid path: %s", name)
	}
	part2 := ""
	if len(parts) == 2 {
		part2 = parts[1]
	}
	switch parts[0] {
	case DataDir:
		return o.readDirData(part2, num)
	case ManifestDir:
		return o.readDirManifest(part2, num)
	case StateDir:
		return o.readDirState(part2, num)
	}
	return nil, errors.Wrapf(fs.ErrNotExist, "invalid path: %s", name)
}
func (o *ObjectFS) openState(name string) (fs.File, error) {
	parts := strings.SplitN(name, "/", 2)
	if len(parts) != 2 {
		return nil, errors.Wrapf(fs.ErrNotExist, "invalid state path: %s", name)
	}
	var versionStr string
	if parts[0] == "latest" {
		versionStr = o.inventory.GetHead()
	} else {
		versionStr = parts[0]
	}
	path := parts[1]
	versions := o.inventory.GetVersions()
	version, ok := versions[versionStr]
	if !ok {
		return nil, errors.Wrapf(fs.ErrNotExist, "invalid version: %s", version)
	}
	for cs, paths := range version.State.State {
		if !slices.Contains(paths, path) {
			continue
		}
		realpaths, ok := o.manifest[cs]
		if !ok {
			return nil, errors.Wrapf(fs.ErrNotExist, "unknown checksum %s for version %s and path %s", cs, versionStr, path)
		}
		if len(realpaths) == 0 {
			return nil, errors.Wrapf(fs.ErrNotExist, "no files found for checksum %s for version %s and path %s", cs, versionStr, path)
		}
		fsys := o.object.GetFS()
		return fsys.Open(realpaths[0])
	}
	return nil, errors.Wrapf(fs.ErrNotExist, "invalid state path: %s", name)
}
func (o *ObjectFS) statState(name string) (fs.FileInfo, error) {
	parts := strings.SplitN(name, "/", 2)
	if len(parts) != 2 {
		return nil, errors.Wrapf(fs.ErrNotExist, "invalid state path: %s", name)
	}
	var versionStr string
	if parts[0] == "latest" {
		versionStr = o.inventory.GetHead()
	} else {
		versionStr = parts[0]
	}
	path := parts[1]
	versions := o.inventory.GetVersions()
	version, ok := versions[versionStr]
	if !ok {
		return nil, errors.Wrapf(fs.ErrNotExist, "invalid version: %s", version)
	}
	for cs, paths := range version.State.State {
		if !slices.Contains(paths, path) {
			continue
		}
		realpaths, ok := o.manifest[cs]
		if !ok {
			return nil, errors.Wrapf(fs.ErrNotExist, "unknown checksum %s for version %s and path %s", cs, versionStr, path)
		}
		if len(realpaths) == 0 {
			return nil, errors.Wrapf(fs.ErrNotExist, "no files found for checksum %s for version %s and path %s", cs, versionStr, path)
		}
		fsys := o.object.GetFS()
		return fs.Stat(fsys, realpaths[0])
	}
	return nil, errors.Wrapf(fs.ErrNotExist, "invalid state path: %s", name)
}
func (o *ObjectFS) readDirState(name string, num int) (files []fs.DirEntry, err error) {
	if num == -1 {
		num = math.MaxInt
	}
	files = []fs.DirEntry{}
	name = strings.TrimSuffix(name, "/")
	if name == "" {
		versions := o.inventory.GetVersions()
		for versionStr, version := range versions {
			files = append(files, NewObjectFileInfoDir(versionStr, version.Created.Time))
		}
		return files, nil
	}
	parts := strings.SplitN(name, "/", 2)
	if len(parts) == 0 {
		return nil, errors.Wrapf(fs.ErrNotExist, "invalid state path: %s", name)
	}
	var versionStr string
	if parts[0] == "latest" {
		versionStr = o.inventory.GetHead()
	} else {
		versionStr = parts[0]
	}
	path := ""
	if len(parts) > 1 {
		path = parts[1]
	}
	versions := o.inventory.GetVersions()
	version, ok := versions[versionStr]
	if !ok {
		return nil, errors.Wrapf(fs.ErrNotExist, "invalid version: %s", version)
	}
	var entries = []fs.DirEntry{}
	var found bool
	for _, paths := range version.State.State {
		for _, p := range paths {
			if strings.HasPrefix(p, path) {
				sub := strings.TrimPrefix(p, path)
				if sub != "" && sub[0] != '/' && path != "" {
					continue
				}
				if sub == "" {
					return []fs.DirEntry{}, errors.Wrapf(fs.ErrInvalid, "invalid state path: %s - not a directory", name)
				}
				sub = strings.TrimPrefix(sub, "/")
				parts = strings.SplitN(sub, "/", 2)
				switch len(parts) {
				case 0:
					return nil, errors.Wrapf(fs.ErrInvalid, "invalid state path: %s - not a directory", name)
				case 1:
					found = true
					// todo: add file fileinfo and metadata from filesystem extension
					e := NewObjectFileInfoFile(parts[0], time.Time{}, 0)
					de, ok := any(e).(fs.DirEntry)
					if !ok {
						return nil, errors.Wrapf(fs.ErrInvalid, "invalid state path: %T - not a fs.DirEntry", e)
					}
					entries = InsertSortedUniqueFunc(entries, de, func(i, j fs.DirEntry) int {
						return strings.Compare(i.Name(), j.Name())
					})
				case 2:
					found = true
					e := NewObjectFileInfoDir(parts[0], time.Time{})
					de, ok := any(e).(fs.DirEntry)
					if !ok {
						return nil, errors.Wrapf(fs.ErrInvalid, "invalid state path: %T - not a fs.DirEntry", e)
					}
					entries = InsertSortedUniqueFunc(entries, de, func(i, j fs.DirEntry) int {
						return strings.Compare(i.Name(), j.Name())
					})
				}
				if len(entries) >= num && found {
					break
				}
			}
		}
		if len(entries) >= num && found {
			break
		}
	}
	if !found {
		// empty folders don't exist
		return nil, errors.Wrapf(fs.ErrNotExist, "invalid state path: %s", name)
	}
	if num > len(entries) {
		return entries, nil
	}
	return entries[:num], nil
}

func (o *ObjectFS) openManifest(name string) (fs.File, error) {
	realpaths, ok := o.manifest[name]
	if !ok {
		return nil, errors.Wrapf(fs.ErrNotExist, "unknown checksum %s", name)
	}
	if len(realpaths) == 0 {
		return nil, errors.Wrapf(fs.ErrNotExist, "no files found for checksum %s", name)
	}
	fsys := o.object.GetFS()
	return fsys.Open(realpaths[0])
}
func (o *ObjectFS) statManifest(name string) (fs.FileInfo, error) {
	realpaths, ok := o.manifest[name]
	if !ok {
		return nil, errors.Wrapf(fs.ErrNotExist, "unknown checksum %s", name)
	}
	if len(realpaths) == 0 {
		return nil, errors.Wrapf(fs.ErrNotExist, "no files found for checksum %s", name)
	}
	fsys := o.object.GetFS()
	return fs.Stat(fsys, realpaths[0])
}
func (o *ObjectFS) readDirManifest(name string, num int) (files []fs.DirEntry, err error) {
	if name != "" {
		if _, ok := o.manifest[name]; !ok {
			return nil, errors.Wrapf(fs.ErrNotExist, "unknown checksum %s", name)
		} else {
			return nil, errors.Wrapf(fs.ErrInvalid, "%s is a file", name)
		}
	}
	if num == -1 {
		num = math.MaxInt
	}
	files = []fs.DirEntry{}
	var count int
	for checksum := range o.manifest {
		if count >= num {
			break
		}
		// todo: add size and time from filesystem extension
		files = append(files, NewObjectFileInfoFile(checksum, time.Time{}, 0))
		count++
	}
	return files, nil
}

func (o *ObjectFS) openData(name string) (fs.File, error) {
	for _, realpaths := range o.manifest {
		if !slices.Contains(realpaths, name) {
			continue
		}
		fs := o.object.GetFS()
		return fs.Open(realpaths[0])
	}
	return nil, errors.Wrapf(fs.ErrNotExist, "unknown file %s", name)
}
func (o *ObjectFS) statData(name string) (fs.FileInfo, error) {
	for _, realpaths := range o.manifest {
		if !slices.Contains(realpaths, name) {
			continue
		}
		fsys := o.object.GetFS()
		return fs.Stat(fsys, realpaths[0])
	}
	return nil, errors.Wrapf(fs.ErrNotExist, "unknown file %s", name)
}
func (o *ObjectFS) readDirData(name string, num int) (files []fs.DirEntry, err error) {
	if num == -1 {
		num = math.MaxInt
	}
	files = []fs.DirEntry{}
	name = strings.TrimSuffix(name, "/")
	var entries = []fs.DirEntry{}
	var found bool
	for _, paths := range o.manifest {
		for _, p := range paths {
			if strings.HasPrefix(p, name) {
				sub := strings.TrimPrefix(p, name)
				if sub != "" && sub[0] != '/' && name != "" {
					continue
				}
				if sub == "" {
					return []fs.DirEntry{}, errors.Wrapf(fs.ErrInvalid, "invalid state path: %s - not a directory", name)
				}
				sub = strings.TrimPrefix(sub, "/")
				parts := strings.SplitN(sub, "/", 2)
				switch len(parts) {
				case 0:
					return nil, errors.Wrapf(fs.ErrInvalid, "invalid state path: %s - not a directory", name)
				case 1:
					found = true
					// todo: add file fileinfo and metadata from filesystem extension
					e := NewObjectFileInfoFile(parts[0], time.Time{}, 0)
					de, ok := any(e).(fs.DirEntry)
					if !ok {
						return nil, errors.Wrapf(fs.ErrInvalid, "invalid state path: %T - not a fs.DirEntry", e)
					}
					entries = InsertSortedUniqueFunc(entries, de, func(i, j fs.DirEntry) int {
						return strings.Compare(i.Name(), j.Name())
					})
				case 2:
					found = true
					e := NewObjectFileInfoDir(parts[0], time.Time{})
					de, ok := any(e).(fs.DirEntry)
					if !ok {
						return nil, errors.Wrapf(fs.ErrInvalid, "invalid state path: %T - not a fs.DirEntry", e)
					}
					entries = InsertSortedUniqueFunc(entries, de, func(i, j fs.DirEntry) int {
						return strings.Compare(i.Name(), j.Name())
					})
				}
				if len(entries) >= num && found {
					break
				}
			}
		}
		if len(entries) >= num && found {
			break
		}
	}
	if !found {
		// empty folders don't exist
		return nil, errors.Wrapf(fs.ErrNotExist, "invalid state path: %s", name)
	}
	if num > len(entries) {
		return entries, nil
	}
	return entries[:num], nil
}
func (o *ObjectFS) Open(name string) (fs.File, error) {
	name = strings.TrimSuffix(name, "/")
	_, err := o.readDir(name, 0)
	if errors.Is(err, fs.ErrNotExist) {
		if path.Base(name) == "index.html" {
			dir := path.Dir(name)
			files, err := o.readDir(dir, -1)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to read directory %s", dir)
			}
			html := `
<!DOCTYPE html>
<html>
<head>
	<title>Index of ` + dir + `</title>
</head>
<body>
	<h1>Index of ` + dir + `</h1>
	<ul>
`
			if dir != "." {
				html += `<li><a href="../">[up]</a></li>`
			}
			for _, file := range files {
				html += `<li><a href="` + file.Name() + `">` + file.Name() + `</a></li>`
			}
			html += `
	</ul>
</body>
</html>
`
			return NewFileBuffer("index.html", []byte(html)), nil
		}
		// if not exists, file does not exist too
		return nil, errors.Wrapf(err, "failed to read directory %s", name)
	}
	if !errors.Is(err, fs.ErrInvalid) {
		// if not invalid, it's a directory
		return NewObjectFileDir(name, time.Time{}, o), nil
	}

	parts := strings.SplitN(name, "/", 2)
	if len(parts) != 2 {
		return nil, errors.Wrapf(fs.ErrNotExist, "invalid path: %s", name)
	}
	pathStr := parts[1]
	switch parts[0] {
	case DataDir:
		return o.openData(pathStr)
	case ManifestDir:
		return o.openManifest(pathStr)
	case StateDir:
		return o.openState(pathStr)
	}
	return nil, errors.Wrapf(fs.ErrNotExist, "invalid path: %s", name)
}

func (o *ObjectFS) Stat(name string) (fs.FileInfo, error) {
	name = strings.TrimSuffix(name, "/")
	_, err := o.readDir(name, 0)
	if errors.Is(err, fs.ErrNotExist) {
		// if not exists, file does not exist too
		return nil, errors.Wrapf(err, "failed to read directory %s", name)
	}
	if !errors.Is(err, fs.ErrInvalid) {
		// if not invalid, it's a directory
		return NewObjectFileInfoDir(name, time.Time{}), nil
	}
	parts := strings.SplitN(name, "/", 2)
	if len(parts) != 2 {
		return nil, errors.Wrapf(fs.ErrNotExist, "invalid path: %s", name)
	}
	switch parts[0] {
	case DataDir:
		return o.statData(parts[1])
	case ManifestDir:
		return o.statManifest(parts[1])
	case StateDir:
		return o.statState(parts[1])
	}
	return nil, errors.Wrapf(fs.ErrNotExist, "invalid path: %s", name)
}

/*
func (o *ObjectFS) folderState(name string) (folders []string, err error) {
	parts := strings.SplitN(name, "/", 2)
	if len(parts) != 2 {
		return nil, errors.Wrapf(fs.ErrNotExist, "invalid state path: %s", name)
	}
	var versionStr string
	if parts[0] == "latest" {
		versionStr = o.inventory.GetHead()
	} else {
		versionStr = parts[0]
	}
	path := parts[1]
	versions := o.inventory.GetVersions()
	version, ok := versions[versionStr]
	if !ok {
		return nil, errors.Wrapf(fs.ErrNotExist, "invalid version: %s", version)
	}
	for _, paths := range version.State.State {
		if slices.Contains(paths, path) {
			return paths, nil
		}
	}
	return []string{}, nil
}

func (o *ObjectFS) checksumState(name string) (cs string, realpaths []string, err error) {
	parts := strings.SplitN(name, "/", 2)
	if len(parts) != 2 {
		return "", nil, errors.Wrapf(fs.ErrNotExist, "invalid state path: %s", name)
	}
	var versionStr string
	if parts[0] == "latest" {
		versionStr = o.inventory.GetHead()
	} else {
		versionStr = parts[0]
	}
	path := parts[1]
	versions := o.inventory.GetVersions()
	version, ok := versions[versionStr]
	if !ok {
		return "", nil, errors.Wrapf(fs.ErrNotExist, "invalid version: %s", version)
	}
	for checksum, paths := range version.State.State {
		if slices.Contains(paths, path) {
			realpaths, ok = o.manifest[checksum]
			if !ok {
				return "", nil, errors.Wrapf(fs.ErrNotExist, "unknown checksum %s for version %s and path %s", checksum, versionStr, path)
			}
			if len(realpaths) == 0 {
				return "", nil, errors.Wrapf(fs.ErrNotExist, "no files found for checksum %s for version %s and path %s", checksum, versionStr, path)
			}
			return checksum, realpaths, nil
		}
	}
	return "", []string{}, nil
}
*/

var _ fs.FS = &ObjectFS{}

package extension

import (
	"bytes"
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"github.com/atsushinee/go-markdown-generator/doc"
	"github.com/je4/filesystem/v2/pkg/writefs"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
)

const ContentSubPathName = "NNNN-content-subpath"
const ContentSubPathDescription = "prepend a path inside the version content"

func GetContentSubPathParams() []*ocfl.ExtensionExternalParam {
	return []*ocfl.ExtensionExternalParam{
		{
			ExtensionName: ContentSubPathName,
			Functions:     []string{"extract"},
			Param:         "area",
			File:          "area",
			Description:   "subpath for extraction (default: 'content'). 'all' for complete extraction",
		},
	}
}

func NewContentSubPathFS(fsys fs.FS) (*ContentSubPath, error) {
	fp, err := fsys.Open("config.json")
	if err != nil {
		return nil, errors.Wrap(err, "cannot open config.json")
	}
	defer fp.Close()
	data, err := io.ReadAll(fp)
	if err != nil {
		return nil, errors.Wrap(err, "cannot read config.json")
	}

	var config = &ContentSubPathConfig{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal ContentSubPathConfig '%s'", string(data))
	}
	return NewContentSubPath(config)
}
func NewContentSubPath(config *ContentSubPathConfig) (*ContentSubPath, error) {
	sl := &ContentSubPath{
		ContentSubPathConfig: config,
	}
	if config.ExtensionName != sl.GetName() {
		return nil, errors.New(fmt.Sprintf("invalid extension name'%s'for extension %s", config.ExtensionName, sl.GetName()))
	}
	return sl, nil
}

type ContentSubPathEntry struct {
	Path        string `json:"path"`
	Description string `json:"description"`
}

type ContentSubPathConfig struct {
	*ocfl.ExtensionConfig
	Paths map[string]ContentSubPathEntry `json:"subPath"`
}
type ContentSubPath struct {
	*ContentSubPathConfig
	fsys fs.FS
	area string
}

func (sl *ContentSubPath) GetFS() fs.FS {
	return sl.fsys
}

func (sl *ContentSubPath) GetConfig() any {
	return sl.ContentSubPathConfig
}

func (sl *ContentSubPath) IsRegistered() bool {
	return false
}

func (sl *ContentSubPath) SetFS(fsys fs.FS) {
	sl.fsys = fsys
}

func (sl *ContentSubPath) SetParams(params map[string]string) error {
	name := fmt.Sprintf("ext-%s-%s", ContentSubPathName, "area")
	sl.area, _ = params[name]
	if sl.area == "" {
		sl.area = "content"
	}
	return nil
}

func (sl *ContentSubPath) GetName() string { return ContentSubPathName }

func (sl *ContentSubPath) WriteConfig() error {
	if sl.fsys == nil {
		return errors.New("no filesystem set")
	}
	configWriter, err := writefs.Create(sl.fsys, "config.json")
	if err != nil {
		return errors.Wrap(err, "cannot open config.json")
	}
	defer configWriter.Close()
	jenc := json.NewEncoder(configWriter)
	jenc.SetIndent("", "   ")
	if err := jenc.Encode(sl.ContentSubPathConfig); err != nil {
		return errors.Wrapf(err, "cannot encode config to file")
	}

	return nil
}

func (sl *ContentSubPath) BuildObjectManifestPath(object ocfl.Object, originalPath string, area string) (string, error) {
	if area == "" {
		return originalPath, nil
	}
	subpath, ok := sl.Paths[area]
	if !ok {
		return "", errors.Errorf("invalid area '%s'", area)
	}
	path := filepath.ToSlash(filepath.Join(subpath.Path, originalPath))
	return path, nil
}

func (sl *ContentSubPath) UpdateObjectBefore(object ocfl.Object) error {

	return nil
}
func (sl *ContentSubPath) UpdateObjectAfter(object ocfl.Object) error {
	readme := doc.NewMarkDown()
	readme.WriteTitle("Description of folders", doc.LevelTitle).
		WriteLines(2)
	var row int
	for _, entry := range sl.Paths {
		readme.WriteTitle(entry.Path, doc.LevelNormal)
		readme.Write(entry.Description)
		readme.Write("\n\n")
		row++
	}

	buf := bytes.NewBuffer([]byte(readme.String()))
	if err := object.AddReader(io.NopCloser(buf), []string{"README.md"}, "", false, false); err != nil {
		return errors.Wrap(err, "cannot write 'README.md'")
	}
	return nil
}

func (sl *ContentSubPath) BuildObjectStatePath(object ocfl.Object, originalPath string, area string) (string, error) {
	if area == "" {
		area = sl.area
	}
	if area == "full" {
		return originalPath, nil
	}
	subpath, ok := sl.Paths[area]
	if !ok {
		return "", errors.Errorf("invalid area '%s'", area)
	}
	path := filepath.ToSlash(filepath.Join(subpath.Path, originalPath))
	return path, nil
}

func (sl *ContentSubPath) BuildObjectExtractPath(object ocfl.Object, originalPath string, area string) (string, error) {
	if area == "" {
		area = sl.area
	}
	if area == "full" {
		return originalPath, nil
	}
	subpath, ok := sl.Paths[area]
	if !ok {
		return "", errors.Errorf("invalid area '%s'", area)
	}
	originalPath = strings.TrimLeft(originalPath, "/")
	if !strings.HasPrefix(originalPath, subpath.Path) {
		return "", errors.Wrapf(ocfl.ExtensionObjectExtractPathWrongAreaError, "'%s' does not belong to area '%s'", originalPath, area)
	}
	originalPath = strings.TrimLeft(strings.TrimPrefix(originalPath, subpath.Path), "/")
	return originalPath, nil
}

func (sl *ContentSubPath) GetAreaPath(object ocfl.Object, area string) (string, error) {
	subpath, ok := sl.Paths[area]
	if !ok {
		return "", errors.Errorf("invalid area '%s'", sl.area)
	}
	return subpath.Path, nil
}

// check interface satisfaction
var (
	_ ocfl.Extension                  = &ContentSubPath{}
	_ ocfl.ExtensionObjectContentPath = &ContentSubPath{}
	_ ocfl.ExtensionObjectChange      = &ContentSubPath{}
	_ ocfl.ExtensionObjectStatePath   = &ContentSubPath{}
	_ ocfl.ExtensionObjectExtractPath = &ContentSubPath{}
	_ ocfl.ExtensionArea              = &ContentSubPath{}
)

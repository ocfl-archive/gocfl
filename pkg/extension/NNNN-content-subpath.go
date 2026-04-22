package extension

import (
	"encoding/json"
	"fmt"

	"io/fs"
	"path/filepath"
	"strings"

	"emperror.dev/errors"
	"github.com/atsushinee/go-markdown-generator/doc"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/ocfl-archive/gocfl/v3/pkg/appendfs"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"
	extensiontypes "github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

const ContentSubPathName = "NNNN-content-subpath"
const ContentSubPathDescription = "prepend a path inside the version content"

func init() {
	extension.RegisterExtension(ContentSubPathName, NewContentSubPath, GetContentSubPathParams)
}

func GetContentSubPathParams() ([]*extension.ExternalParam, error) {
	return []*extension.ExternalParam{
		{
			ExtensionName: ContentSubPathName,
			Functions:     []string{"extract"},
			Param:         "area",
			//File:          "area",
			Description: "subpath for extraction (default: 'content'). 'all' for complete extraction",
		},
	}, nil
}

func NewContentSubPath() (extensiontypes.Extension, error) {
	var config = &ContentSubPathConfig{
		ExtensionConfig: &extensiontypes.ExtensionConfig{ExtensionName: ContentSubPathName},
		Paths:           map[string]ContentSubPathEntry{},
	}
	sl := &ContentSubPath{
		ContentSubPathConfig: config,
	}
	return sl, nil
}

type ContentSubPathEntry struct {
	Path        string `json:"path"`
	Description string `json:"description"`
}

type ContentSubPathConfig struct {
	*extensiontypes.ExtensionConfig
	Paths map[string]ContentSubPathEntry `json:"subPath"`
}
type ContentSubPath struct {
	*ContentSubPathConfig
	area   string
	logger ocfllogger.OCFLLogger
}

func (sl *ContentSubPath) WithLogger(logger ocfllogger.OCFLLogger) extensiontypes.Extension {
	sl.logger = logger.With("extension", ContentSubPathName)
	return sl
}

func (sl *ContentSubPath) Load(fsys fs.FS) error {
	data, err := fs.ReadFile(fsys, "config.json")
	if err != nil {
		return errors.Wrap(err, "cannot read config.json")
	}
	if err := json.Unmarshal(data, sl.ContentSubPathConfig); err != nil {
		return errors.Wrapf(err, "cannot unmarshal ContentSubPathConfig '%s'", string(data))
	}
	if sl.Paths == nil {
		sl.Paths = map[string]ContentSubPathEntry{}
	}
	return nil
}

func (sl *ContentSubPath) Terminate() error {
	return nil
}

func (sl *ContentSubPath) GetMetadata(fs.FS, object.Object) (map[string]any, error) {
	return map[string]any{"": sl.Paths}, nil
}

func (sl *ContentSubPath) GetConfig() any {
	return sl.ContentSubPathConfig
}

func (sl *ContentSubPath) IsRegistered() bool {
	return false
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

func (sl *ContentSubPath) WriteConfig(fsys appendfs.FS) error {
	configWriter, err := writefs.Create(fsys, "config.json")
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

func (sl *ContentSubPath) BuildObjectManifestPath(originalPath string, area string) (string, error) {
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

func (sl *ContentSubPath) UpdateObjectBefore(object object.VersionWriter) error {

	return nil
}
func (sl *ContentSubPath) UpdateObjectAfter(object object.VersionWriter) error {
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

	//buf := bytes.NewBuffer([]byte(readme.String()))
	//if err := object.AddReader(io.NopCloser(buf), []string{"README.md"}, "", false, false); err != nil {
	if err := object.AddData([]byte(readme.String()), "README.md", true, "full", false, false); err != nil {
		return errors.Wrap(err, "cannot write 'README.md'")
	}
	return nil
}

func (sl *ContentSubPath) BuildObjectStatePath(originalPath string, area string) (string, error) {
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

func (sl *ContentSubPath) BuildObjectExtractPath(originalPath string, area string) (string, error) {
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
		return "", errors.Wrapf(object.ExtensionObjectExtractPathWrongAreaError, "'%s' does not belong to area '%s'", originalPath, area)
	}
	originalPath = strings.TrimLeft(strings.TrimPrefix(originalPath, subpath.Path), "/")
	return originalPath, nil
}

func (sl *ContentSubPath) GetAreaPath(area string) (string, error) {
	subpath, ok := sl.Paths[area]
	if !ok {
		return "", errors.Errorf("invalid area '%s'", sl.area)
	}
	return subpath.Path, nil
}

// check interface satisfaction
var (
	_ extensiontypes.Extension          = &ContentSubPath{}
	_ object.ExtensionObjectContentPath = &ContentSubPath{}
	_ object.ExtensionObjectChange      = &ContentSubPath{}
	_ object.ExtensionObjectStatePath   = &ContentSubPath{}
	_ object.ExtensionObjectExtractPath = &ContentSubPath{}
	_ object.ExtensionArea              = &ContentSubPath{}
	_ object.ExtensionMetadata          = &ContentSubPath{}
)

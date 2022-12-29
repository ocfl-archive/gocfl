package extension

import (
	"bytes"
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"github.com/atsushinee/go-markdown-generator/doc"
	"go.ub.unibas.ch/gocfl/v2/pkg/ocfl"
	"io"
	"path/filepath"
)

const ContentSubPathName = "NNNN-content-subpath"
const ContentSubPathDescription = "prepend a path inside the version content"

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
	fs ocfl.OCFLFS
}

func NewContentSubPathFS(fsys ocfl.OCFLFSRead) (*ContentSubPath, error) {
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

func (sl *ContentSubPath) SetFS(fs ocfl.OCFLFS) {
	sl.fs = fs
}

func (sl *ContentSubPath) GetName() string { return ContentSubPathName }

func (sl *ContentSubPath) GetConfigString() string {
	str, _ := json.MarshalIndent(sl.ContentSubPathConfig, "", "  ")
	return string(str)
}

func (sl *ContentSubPath) WriteConfig() error {
	if sl.fs == nil {
		return errors.New("no filesystem set")
	}
	configWriter, err := sl.fs.Create("config.json")
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

func (sl *ContentSubPath) BuildObjectContentPath(object ocfl.Object, originalPath string, area string) (string, error) {
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
	if err := object.AddReader(io.NopCloser(buf), "README.md", ""); err != nil {
		return errors.Wrap(err, "cannot write 'README.md'")
	}
	return nil
}

func (sl *ContentSubPath) BuildObjectExternalPath(object ocfl.Object, originalPath string, area string) (string, error) {
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

// check interface satisfaction
var (
	_ ocfl.Extension                   = &ContentSubPath{}
	_ ocfl.ExtensionObjectContentPath  = &ContentSubPath{}
	_ ocfl.ExtensionObjectChange       = &ContentSubPath{}
	_ ocfl.ExtensionObjectExternalPath = &ContentSubPath{}
)

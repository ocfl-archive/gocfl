package extension

import (
	"bytes"
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"go.ub.unibas.ch/gocfl/v2/pkg/checksum"
	"go.ub.unibas.ch/gocfl/v2/pkg/ocfl"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const MetaFileName = "NNNN-metafile"
const MetaFileDescription = "adds a file in extension folder"

func GetMetaFileParams() []*ocfl.ExtensionExternalParam {
	return []*ocfl.ExtensionExternalParam{
		{
			ExtensionName: MetaFileName,
			Functions:     []string{"add", "update", "create"},
			Param:         "source",
			File:          "Source",
			Description:   "url with metadata file. $ID will be replaced with object ID i.e. file:///c:/temp/$ID.json",
		},
		{
			ExtensionName: MetaFileName,
			Functions:     []string{"extract", "objectextension"},
			Param:         "target",
			File:          "Target",
			Description:   "url with metadata target folder",
		},
	}
}

type MetaFileConfig struct {
	*ocfl.ExtensionConfig
	Versioned bool `json:"versioned"`
}
type MetaFile struct {
	*MetaFileConfig
	metadataSource *url.URL
	fs             ocfl.OCFLFS
}

func NewMetaFileFS(fsys ocfl.OCFLFSRead) (*MetaFile, error) {
	fp, err := fsys.Open("config.json")
	if err != nil {
		return nil, errors.Wrap(err, "cannot open config.json")
	}
	defer fp.Close()
	data, err := io.ReadAll(fp)
	if err != nil {
		return nil, errors.Wrap(err, "cannot read config.json")
	}

	var config = &MetaFileConfig{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal DirectCleanConfig '%s'", string(data))
	}
	return NewMetaFile(config)
}
func NewMetaFile(config *MetaFileConfig) (*MetaFile, error) {
	sl := &MetaFile{MetaFileConfig: config}
	if config.ExtensionName != sl.GetName() {
		return nil, errors.New(fmt.Sprintf("invalid extension name'%s'for extension %s", config.ExtensionName, sl.GetName()))
	}
	/*
		if sl.metadataSource == nil {
			return nil, errors.Errorf("no metadata-source for extension '%s'", MetaFileName)
		}
	*/
	return sl, nil
}

func (sl *MetaFile) SetParams(params map[string]string) error {
	if params != nil {
		name := fmt.Sprintf("ext-%s-%s", MetaFileName, "source")
		if urlString, ok := params[name]; ok {
			u, err := url.Parse(urlString)
			if err != nil {
				return errors.Wrapf(err, "invalid url parameter '%s' - '%s'", name, urlString)
			}
			sl.metadataSource = u
		}
	}
	return nil
}

func (sl *MetaFile) SetFS(fs ocfl.OCFLFS) {
	sl.fs = fs
}

func (sl *MetaFile) GetName() string { return MetaFileName }

func (sl *MetaFile) GetConfigString() string {
	str, _ := json.MarshalIndent(sl.MetaFileConfig, "", "  ")
	return string(str)
}

func (sl *MetaFile) WriteConfig() error {
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
	if err := jenc.Encode(sl.MetaFileConfig); err != nil {
		return errors.Wrapf(err, "cannot encode config to file")
	}

	return nil
}

func (sl *MetaFile) UpdateObjectBefore(object ocfl.Object) error {
	return nil
}

func downloadFile(u string) ([]byte, error) {
	resp, err := http.Get(u)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot get '%s'", u)
	}
	defer resp.Body.Close()
	metadata, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot read body from '%s'", u)
	}
	return metadata, nil
}

var windowsPathWithDrive = regexp.MustCompile("^/[a-zA-Z]:")

func (sl *MetaFile) UpdateObjectAfter(object ocfl.Object) error {
	var err error
	inventory := object.GetInventory()
	if inventory == nil {
		return errors.New("no inventory available")
	}
	if sl.metadataSource == nil {
		// only a problem, if first version
		if len(inventory.GetVersionStrings()) < 2 {
			return errors.New("no metadata source configured")
		}
		return nil
	}
	if sl.fs == nil {
		return errors.New("no filesystem set")
	}
	var rc io.ReadCloser
	switch sl.metadataSource.Scheme {
	case "http":
		fname := strings.Replace(sl.metadataSource.String(), "$ID", object.GetID(), -1)
		resp, err := http.Get(fname)
		if err != nil {
			return errors.Wrapf(err, "cannot get '%s'", fname)
		}
		rc = resp.Body
	case "https":
		fname := strings.Replace(sl.metadataSource.String(), "$ID", object.GetID(), -1)
		resp, err := http.Get(fname)
		if err != nil {
			return errors.Wrapf(err, "cannot get '%s'", fname)
		}
		rc = resp.Body
	case "file":
		fname := strings.Replace(sl.metadataSource.Path, "$ID", object.GetID(), -1)
		fname = "/" + strings.TrimLeft(fname, "/")
		if windowsPathWithDrive.Match([]byte(fname)) {
			fname = strings.TrimLeft(fname, "/")
		}
		rc, err = os.Open(fname)
		if err != nil {
			return errors.Wrapf(err, "cannot open '%s'", fname)
		}
	case "":
		fname := strings.Replace(sl.metadataSource.Path, "$ID", object.GetID(), -1)
		fname = "/" + strings.TrimLeft(fname, "/")
		rc, err = os.Open(fname)
		if err != nil {
			return errors.Wrapf(err, "cannot open '%s'", fname)
		}
	default:
		return errors.Errorf("url scheme '%s' not supported", sl.metadataSource.Scheme)
	}

	entries, err := sl.fs.ReadDir(".")
	if err != nil {
		return errors.Wrapf(err, "cannot read directory of '%s'", sl.fs)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if entry.Name() == "config.json" {
			continue
		}
		if err := sl.fs.Delete(entry.Name()); err != nil {
			return errors.Wrapf(err, "cannot delete '%s' from '%s'", entry.Name(), sl.fs)
		}
	}

	// complex writes to prevent simultaneous writes on filesystems, which do not support that
	targetBase := filepath.Base(sl.metadataSource.Path)
	w2, err := sl.fs.Create(targetBase)
	if err != nil {
		return errors.Wrapf(err, "cannot create '%s'", targetBase)
	}

	allTargets := []io.Writer{w2}
	var buf = bytes.NewBuffer(nil)
	if sl.Versioned {
		allTargets = append(allTargets, buf)
	}

	mw := io.MultiWriter(allTargets...)
	digests, err := checksum.Copy(mw, rc, []checksum.DigestAlgorithm{inventory.GetDigestAlgorithm()})
	if err != nil {
		w2.Close()
		return errors.Wrap(err, "cannot write data")
	}
	w2.Close()

	digest, ok := digests[inventory.GetDigestAlgorithm()]
	if !ok {
		return errors.Wrapf(err, "digest '%s' not created", inventory.GetDigestAlgorithm())
	}

	targetBaseSidecar := fmt.Sprintf("%s.%s", targetBase, inventory.GetDigestAlgorithm())
	w2Sidecar, err := sl.fs.Create(targetBaseSidecar)
	if err != nil {
		return errors.Wrapf(err, "cannot create '%s'", targetBaseSidecar)
	}
	if _, err := io.WriteString(w2Sidecar, fmt.Sprintf("%s %s", digest, targetBase)); err != nil {
		w2Sidecar.Close()
		return errors.Wrapf(err, "cannot write to sidecar '%s'", targetBaseSidecar)
	}
	w2Sidecar.Close()

	if sl.Versioned {
		targetVersioned := fmt.Sprintf("%s/%s", inventory.GetHead(), targetBase)
		w, err := sl.fs.Create(targetVersioned)
		if err != nil {
			return errors.Wrapf(err, "cannot create '%s'", targetVersioned)
		}
		if _, err := io.Copy(w, buf); err != nil {
			w.Close()
			return errors.Wrapf(err, "cannot write data to '%s'", targetVersioned)
		}
		w.Close()
		targetVersionedSidecar := fmt.Sprintf("%s.%s", targetVersioned, inventory.GetDigestAlgorithm())
		wSidecar, err := sl.fs.Create(targetVersionedSidecar)
		if err != nil {
			return errors.Wrapf(err, "cannot create '%s'", targetVersionedSidecar)
		}
		if _, err := io.WriteString(wSidecar, fmt.Sprintf("%s %s", digest, targetVersioned)); err != nil {
			wSidecar.Close()
			return errors.Wrapf(err, "cannot write to sidecar '%s'", targetVersionedSidecar)
		}
		wSidecar.Close()
	}

	return nil
}

// check interface satisfaction
var (
	_ ocfl.Extension             = &MetaFile{}
	_ ocfl.ExtensionObjectChange = &MetaFile{}
)

package extension

import (
	"bytes"
	"crypto"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"io"
	"io/fs"
	"net/http"
	"strings"

	"emperror.dev/errors"
	"github.com/digitorus/timestamp"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/ocfl-archive/gocfl/v3/pkg/appendfs"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"
	extensiontypes "github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

const TimestampName = "NNNN-timestamp"
const TimestampDescription = "signs ocfl versions"

func init() {
	extension.RegisterExtension(TimestampName, NewTimestamp, GetTimestampParams)
}

func GetTimestampParams() ([]*extension.ExternalParam, error) {
	return nil, nil
}

func NewTimestamp() (extensiontypes.Extension, error) {
	config := &TimestampConfig{
		ExtensionConfig: &extensiontypes.ExtensionConfig{ExtensionName: TimestampName},
	}
	sl := &Timestamp{
		TimestampConfig: config,
	}
	return sl, nil
}

type TimestampConfig struct {
	*extensiontypes.ExtensionConfig
	Authority map[string]string `json:"Authority"` // https://freetsa.org/tsr
	CertChain bool              `json:"CertChain"`
}
type Timestamp struct {
	*TimestampConfig
	fsys   appendfs.FS
	logger ocfllogger.OCFLLogger
}

func (sl *Timestamp) WithLogger(logger ocfllogger.OCFLLogger) extensiontypes.Extension {
	sl.logger = logger.With("extension", TimestampName)
	return sl
}

func (sl *Timestamp) Load(fsys fs.FS) error {
	data, err := fs.ReadFile(fsys, "config.json")
	if err != nil {
		return errors.Wrap(err, "cannot read config.json")
	}
	if err := json.Unmarshal(data, sl.TimestampConfig); err != nil {
		return errors.Wrapf(err, "cannot unmarshal TimestampConfig '%s'", string(data))
	}
	return nil
}

func (sl *Timestamp) trustedTimestamp(object object.Object) error {
	_, checksumString, err := object.GetInventory().Bytes()
	if err != nil {
		return errors.Wrap(err, "cannot marshal inventory")
	}

	var ha crypto.Hash
	switch strings.ToLower(string(object.GetInventory().GetDigestAlgorithm())) {
	case "sha256":
		ha = crypto.SHA256
	case "sha512":
		ha = crypto.SHA512
	case "sha1":
		ha = crypto.SHA1
	case "md5":
		ha = crypto.MD5
	default:
		return errors.Errorf("unsupported hash algorithm '%s'", object.GetInventory().GetDigestAlgorithm())
	}
	/*
		rqst, err := timestamp.CreateRequest(bytes.NewReader(inventory), &timestamp.RequestOptions{
			Hash: ha,
		})
		if err != nil {
			return errors.Wrap(err, "cannot create request")
		}
	*/
	checksumBytes, err := hex.DecodeString(checksumString)
	if err != nil {
		return errors.Wrapf(err, "cannot decode checksum '%s'", checksumString)
	}
	req := &timestamp.Request{
		HashAlgorithm: ha,
		HashedMessage: checksumBytes,
		Certificates:  sl.CertChain,
	}
	rqst, err := req.Marshal()
	if err != nil {
		return errors.Wrap(err, "cannot marshal request")
	}

	sl.logger.Debug().Msgf("Timestamp request: %s", string(rqst))
	/*
		tsr, err := timestamp.ParseRequest(rqst)
		if err != nil {
			return errors.Wrap(err, "cannot parse request")
		}
	*/
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	for name, url := range sl.Authority {
		tsaReq, err := http.NewRequest("POST", url, bytes.NewReader(rqst))
		if err != nil {
			return errors.Wrap(err, "cannot create request")
		}
		tsaReq.Header.Set("content-type", "application/timestamp-query")
		tsaReq.Header.Set("content-length", fmt.Sprintf("%d", len(rqst)))
		// tsaReq.Header.Set("user-agent", "curl/8.5.0")
		// tsaReq.Header.Set("accept", "*/*")

		resp, err := client.Do(tsaReq)
		if err != nil {
			return errors.Wrap(err, "cannot send request")
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "cannot read response")
		}
		if resp.StatusCode != http.StatusOK {
			return errors.Errorf("invalid response from TSA %v: %s", tsaReq, resp.Status)
		}
		ts, err := timestamp.ParseResponse(body)
		if err != nil {
			return errors.Wrapf(err, "cannot parse response '%s'", string(body))
		}
		sl.logger.Debug().Msgf("Timestamp response: %+v", ts)
		//	_ = ts
		sigfile := fmt.Sprintf("data/%s.%s.tsr", name, object.GetInventory().GetHead())
		if _, err := writefs.WriteFile(sl.fsys, sigfile, body); err != nil {
			return errors.Wrapf(err, "cannot write Timestamp file '%s'", sigfile)
		}
		queryfile := fmt.Sprintf("data/%s.%s.tsq", name, object.GetInventory().GetHead())
		if _, err := writefs.WriteFile(sl.fsys, queryfile, rqst); err != nil {
			return errors.Wrapf(err, "cannot write query file '%s'", queryfile)
		}
	}
	return nil
}

func (sl *Timestamp) VersionDone(object object.Object) error {
	if sl.fsys == nil {
		return errors.New("no filesystem set")
	}
	return errors.WithStack(sl.trustedTimestamp(object))
}

func (sl *Timestamp) Terminate() error {
	return nil
}

func (sl *Timestamp) GetMetadata(sourceFS fs.FS, obj object.Object) (map[string]any, error) {
	return map[string]any{"TimestampAuthority": sl.Authority}, nil
}

func (sl *Timestamp) GetConfig() any {
	return sl.TimestampConfig
}

func (sl *Timestamp) IsRegistered() bool {
	return false
}

func (sl *Timestamp) SetFS(fsys fs.FS, create bool) {
	if sfs, ok := fsys.(appendfs.FS); ok {
		sl.fsys = sfs
	}
}

func (sl *Timestamp) SetParams(params map[string]string) error {
	return nil
}

func (sl *Timestamp) GetName() string { return TimestampName }

func (sl *Timestamp) WriteConfig(fsys appendfs.FS) error {
	configWriter, err := writefs.Create(fsys, "config.json")
	if err != nil {
		return errors.Wrap(err, "cannot open config.json")
	}
	defer configWriter.Close()
	jenc := json.NewEncoder(configWriter)
	jenc.SetIndent("", "   ")
	if err := jenc.Encode(sl.TimestampConfig); err != nil {
		return errors.Wrapf(err, "cannot encode config to file")
	}

	return nil
}

// check interface satisfaction
var (
	_ extensiontypes.Extension    = &Timestamp{}
	_ object.ExtensionVersionDone = &Timestamp{}
)

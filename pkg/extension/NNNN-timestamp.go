package extension

import (
	"bytes"
	"crypto"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"emperror.dev/errors"
	"github.com/digitorus/timestamp"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/object"

	"io"
	"io/fs"
	"net/http"
	"strings"
)

const TimestampName = "NNNN-timestamp"
const TimestampDescription = "signs ocfl versions"

func GetTimestampParams() []*extension.ExtensionExternalParam {
	return []*extension.ExtensionExternalParam{}
}

func NewTimestampFS(fsys fs.FS, logger zLogger.ZLogger) (*Timestamp, error) {
	fp, err := fsys.Open("config.json")
	if err != nil {
		return nil, errors.Wrap(err, "cannot open config.json")
	}
	defer fp.Close()
	data, err := io.ReadAll(fp)
	if err != nil {
		return nil, errors.Wrap(err, "cannot read config.json")
	}

	var config = &TimestampConfig{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal TimestampConfig '%s'", string(data))
	}
	return NewTimestamp(config, logger)
}
func NewTimestamp(config *TimestampConfig, logger zLogger.ZLogger) (*Timestamp, error) {
	sl := &Timestamp{
		TimestampConfig: config,
		logger:          logger,
	}
	if config.ExtensionName != sl.GetName() {
		return nil, errors.New(fmt.Sprintf("invalid extension name'%s'for extension %s", config.ExtensionName, sl.GetName()))
	}
	return sl, nil
}

type TimestampConfig struct {
	*extension.ExtensionConfig
	Authority map[string]string `json:"Authority"` // https://freetsa.org/tsr
	CertChain bool              `json:"CertChain"`
}
type Timestamp struct {
	*TimestampConfig
	fsys   fs.FS
	logger zLogger.ZLogger
}

func (sl *Timestamp) trustedTimestamp(object object.Object) error {
	_, checksumString, err := object.GetInventoryContent()
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

func (sl *Timestamp) GetMetadata(object object.Object) (map[string]any, error) {
	return map[string]any{"TimestampAuthority": sl.Authority}, nil
}

func (sl *Timestamp) GetFS() fs.FS {
	return sl.fsys
}

func (sl *Timestamp) GetConfig() any {
	return sl.TimestampConfig
}

func (sl *Timestamp) IsRegistered() bool {
	return false
}

func (sl *Timestamp) SetFS(fsys fs.FS, create bool) {
	sl.fsys = fsys
}

func (sl *Timestamp) SetParams(params map[string]string) error {
	return nil
}

func (sl *Timestamp) GetName() string { return TimestampName }

func (sl *Timestamp) WriteConfig() error {
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
	if err := jenc.Encode(sl.TimestampConfig); err != nil {
		return errors.Wrapf(err, "cannot encode config to file")
	}

	return nil
}

// check interface satisfaction
var (
	_ extension.Extension         = &Timestamp{}
	_ object.ExtensionVersionDone = &Timestamp{}
)

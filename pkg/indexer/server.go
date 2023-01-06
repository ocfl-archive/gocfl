package indexer

import (
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"github.com/op/go-logging"
	"github.com/richardlehane/siegfried"
	"golang.org/x/exp/slices"
	"io"
	"net/http"
	"regexp"
)

type ActionParam struct {
	Url           string            `json:"url"`
	Actions       []string          `json:"actions,omitempty"`
	ForceDownload string            `json:"forcedownload,omitempty"`
	HeaderSize    int64             `json:"headersize,omitempty"`
	Checksums     map[string]string `json:"checksums,omitempty"`
}

type MimeWeightString struct {
	Regexp string
	Weight int
}
type MimeWeight struct {
	regexp *regexp.Regexp
	weight int
}

type Server struct {
	srv           *http.Server
	logger        *logging.Logger
	mimeRelevance []MimeWeight
	sf            *siegfried.Siegfried
}

func NewServer(logger *logging.Logger, signaturePath string, mimeRelevance map[int]MimeWeightString) (*Server, error) {
	var err error
	srv := &Server{
		srv:           nil,
		logger:        logger,
		mimeRelevance: nil,
	}
	srv.sf, err = siegfried.Load(signaturePath)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot load siegfied signature from '%s'", signaturePath)
	}
	mKeys := []int{}
	for key, _ := range mimeRelevance {
		mKeys = append(mKeys, key)
	}
	slices.Sort(mKeys)
	for _, key := range mKeys {
		rexp, err := regexp.Compile(mimeRelevance[key].Regexp)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot compile Regexp %s", key)
		}
		srv.mimeRelevance = append(srv.mimeRelevance, MimeWeight{
			regexp: rexp,
			weight: mimeRelevance[key].Weight,
		})
	}
	return srv, nil
}

func (s *Server) HandleDefault(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1048576))
	r.Body.Close()
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot read body: %v", err), http.StatusInternalServerError)
		return
	}
	param := ActionParam{Checksums: map[string]string{}}
	if err := json.Unmarshal(body, &param); err != nil {
		http.Error(w, fmt.Sprintf("cannot unmarshal json - %s: %v", string(body), err), http.StatusBadRequest)
		return
	}
	// if no action is given, just use all
	if len(param.Actions) == 0 {
		param.Actions = append(param.Actions, "siegfried")
	}
	if len(param.Actions) > 1 || param.Actions[0] != "siegfried" {
		http.Error(w, fmt.Sprintf("invalid action %v. only 'siegfried' is supported", param.Actions), http.StatusBadRequest)
	}

	result, err := s.doIndex(param)
	if err != nil {
		result = map[string]interface{}{}
		errors := map[string]string{}
		errors["index"] = err.Error()
		result["errors"] = errors
		s.logger.Errorf("error on indexing: %v", err)
	}

	js, err := json.Marshal(result)
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot marshal result %v: %v", result, err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)

}

func (s *Server) doIndex(param ActionParam) (map[string]interface{}, error) {
	return nil, nil
}

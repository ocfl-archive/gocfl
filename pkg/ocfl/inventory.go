package ocfl

import (
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"github.com/op/go-logging"
	"go.ub.unibas.ch/gocfl/v2/pkg/checksum"
	"time"
)

type OCFLTime struct{ time.Time }

func (t *OCFLTime) MarshalJSON() ([]byte, error) {
	tstr := t.Format(time.RFC3339)
	return json.Marshal(tstr)
}

func (t *OCFLTime) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return errors.Wrapf(err, "cannot unmarshal string of %s", string(data))
	}
	tt, err := time.Parse(time.RFC3339, str)
	if err != nil {
		return errors.Wrapf(err, "cannot parse %s", string(data))
	}
	t.Time = tt
	return nil
}

type User struct {
	Address string `json:"address"`
	Name    string `json:"name"`
}

type Version struct {
	Created OCFLTime            `json:"created"`
	Message string              `json:"message"`
	State   map[string][]string `json:"state"`
	User    User                `json:"user"`
}

type Inventory interface {
	Init() error
	GetID() string

	DeleteFile(virtualFilename string) error
	Rename(oldVirtualFilename, newVirtualFilename string) error
	AddFile(virtualFilename string, realFilename string, checksum string) error

	//GetContentDirectory() string
	GetVersion() string
	GetVersions() []string
	GetDigestAlgorithm() checksum.DigestAlgorithm
	IsWriteable() bool
	//	IsModified() bool
	BuildRealname(virtualFilename string) string
	NewVersion(msg, UserName, UserAddress string) error
	IsDuplicate(checksum string) bool
	AlreadyExists(virtualFilename, checksum string) (bool, error)
	//	IsUpdate(virtualFilename, checksum string) (bool, error)
	Clean() error
}

func NewInventory(object Object, id string, version OCFLVersion, logger *logging.Logger) (Inventory, error) {
	switch version {
	case Version1_0:
		sr, err := NewInventoryV1_0(object, id, logger)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return sr, nil
	case Version1_1:
		sr, err := NewInventoryV1_1(object, id, logger)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return sr, nil
	default:
		return nil, errors.New(fmt.Sprintf("Inventory Version %s not supported", version))
	}
}

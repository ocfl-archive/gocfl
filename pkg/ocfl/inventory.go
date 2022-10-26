package ocfl

import (
	"context"
	"emperror.dev/errors"
	"encoding/json"
	"github.com/op/go-logging"
	"go.ub.unibas.ch/gocfl/v2/pkg/checksum"
	"time"
)

// type OCFLTime struct{ time.Time }
type OCFLCreated struct{ time.Time }

var InvalidOCFLCreatedFormatError = errors.NewPlain("invalid OCFL Time Format")

func (t *OCFLCreated) MarshalJSON() ([]byte, error) {
	tstr := t.Format(time.RFC3339)
	return json.Marshal(tstr)
}

func (t *OCFLCreated) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return errors.Wrapf(err, "cannot unmarshal string of %s", string(data))
	}
	tt, err := time.Parse(time.RFC3339, str)
	if err != nil {
		return errors.Wrapf(errors.Wrap(InvalidOCFLCreatedFormatError, err.Error()), "cannot parse %s", string(data))
	}
	t.Time = tt
	return nil
}

type User struct {
	Address string `json:"address"`
	Name    string `json:"name"`
}

type Version struct {
	Created OCFLCreated         `json:"created"`
	Message string              `json:"message"`
	State   map[string][]string `json:"state"`
	User    User                `json:"user"`
}

type Inventory interface {
	Init() error
	GetID() string
	GetContentDir() string
	GetHead() string

	DeleteFile(virtualFilename string) error
	Rename(oldVirtualFilename, newVirtualFilename string) error
	AddFile(virtualFilename string, realFilename string, checksum string) error

	//GetContentDirectory() string
	GetVersion() string
	GetVersions() []string
	GetFiles() map[string][]string
	GetFilesFlat() []string
	GetDigestAlgorithm() checksum.DigestAlgorithm
	IsWriteable() bool
	//	IsModified() bool
	BuildRealname(virtualFilename string) string
	NewVersion(msg, UserName, UserAddress string) error
	IsDuplicate(checksum string) bool
	AlreadyExists(virtualFilename, checksum string) (bool, error)
	//	IsUpdate(virtualFilename, checksum string) (bool, error)
	Clean() error

	VersionLessOrEqual(v1, v2 string) bool
}

func NewInventory(ctx context.Context, object Object, id string, version OCFLVersion, logger *logging.Logger) (Inventory, error) {
	switch version {
	case Version1_1:
		sr, err := NewInventoryV1_1(ctx, object, id, logger)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return sr, nil
	default:
		//case Version1_0:
		sr, err := NewInventoryV1_0(ctx, object, id, logger)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return sr, nil
		//		return nil, errors.New(fmt.Sprintf("Inventory Version %s not supported", version))
	}
}

func LoadInventory(ctx context.Context, object Object, data []byte, version OCFLVersion, logger *logging.Logger) (Inventory, error) {
	inventory, err := NewInventory(ctx, object, "", version, logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create empty inventory")
	}
	if err := json.Unmarshal(data, inventory); err != nil {
		// now lets try it again
		jsonMap := map[string]any{}
		// check for json format error
		if err2 := json.Unmarshal(data, &jsonMap); err2 != nil {
			addValidationErrors(ctx, GetValidationError(version, E033).AppendDescription("json syntax error: %v", err2))
			addValidationErrors(ctx, GetValidationError(version, E034).AppendDescription("json syntax error: %v", err2))
		} else {
			if _, ok := jsonMap["head"].(string); !ok {
				addValidationErrors(ctx, GetValidationError(version, E040).AppendDescription("head is not of string type: %v", jsonMap["head"]))
			}
		}
		errCause := errors.Cause(err)
		if errCause == InvalidOCFLCreatedFormatError {
			addValidationErrors(ctx, GetValidationError(version, E049).AppendDescription("invalid created field"))
		}
		return nil, errors.Wrapf(err, "cannot marshal data - %s", string(data))
	}
	return inventory, nil
}

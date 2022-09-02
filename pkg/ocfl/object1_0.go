package ocfl

import (
	"emperror.dev/errors"
	"github.com/op/go-logging"
)

const ObjectV10Version = "1.0"

type ObjectV1_0 struct {
	*ObjectBase
}

func NewObjectV1_0(fs OCFLFS, pathPrefix, id string, logger *logging.Logger) (*ObjectV1_0, error) {
	ob, err := NewObjectBase(fs, pathPrefix, ObjectV10Version, id, logger)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	obv10 := &ObjectV1_0{ObjectBase: ob}
	return obv10, nil
}

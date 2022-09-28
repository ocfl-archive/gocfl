package ocfl

import (
	"emperror.dev/errors"
	"github.com/op/go-logging"
)

const ObjectV11Version = "1.1"

type ObjectV1_1 struct {
	*ObjectBase
}

func NewObjectV1_1(fs OCFLFS, id string, logger *logging.Logger) (*ObjectV1_1, error) {
	ob, err := NewObjectBase(fs, Version1_1, id, logger)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	obv11 := &ObjectV1_1{ObjectBase: ob}
	return obv11, nil
}
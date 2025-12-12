package ocflerrors

import "emperror.dev/errors"

var ErrVersionMultiple = errors.New("multiple version files found")
var ErrVersionNone = errors.New("no version file found")
var ErrInvalidContent = errors.New("content of version declaration does not equal filename")

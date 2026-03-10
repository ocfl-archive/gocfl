package extension

import (
	"emperror.dev/errors"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfllogger"
)

type registrationStruct struct {
	builder           BuilderFunc
	externalParamFunc ExternalParamFunc
}

var registration = map[string]*registrationStruct{}

// RegisterExtension registers an extension with a given name, builder function, and external parameter function. if func is nil, it is ignored.
func RegisterExtension(name string, builder BuilderFunc, externalParamFunc ExternalParamFunc) {
	reg, ok := registration[name]
	if !ok {
		reg = &registrationStruct{}
		registration[name] = reg
	}
	if builder != nil {
		reg.builder = builder
	}
	if externalParamFunc != nil {
		reg.externalParamFunc = externalParamFunc
	}
}

// RegisterWithFactory registers all extensions with the provided factory if they have a valid builder function.
func RegisterWithFactory(fact Factory, logger ocfllogger.OCFLLogger) {
	for name, reg := range registration {
		if reg == nil || reg.builder == nil {
			logger.Debug().Msgf("extension %s has no builder func", name)
			continue
		}
		fact.RegisterExtension(name, reg.builder)
	}
}

func GetExtensionParamValues(command string, callback GetParamsFunc) error {
	extParams, err := GetExternalParams()
	if err != nil {
		return errors.Wrap(err, "GetExternalParams")
	}
	for _, param := range extParams {
		param.GetParam(command, callback)
	}
	return nil
}

func GetExternalParams() ([]*ExternalParam, error) {
	result := make([]*ExternalParam, len(registration))
	for name, reg := range registration {
		if reg == nil || reg.externalParamFunc == nil {
			continue
		}
		ps, err := reg.externalParamFunc()
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get external params for extension %s", name)
		}
		result = append(result, ps...)
	}
	return result, nil
}

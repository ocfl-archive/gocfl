package extension

import (
	"emperror.dev/errors"
	"github.com/ocfl-archive/gocfl/v2/config"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfllogger"
	"github.com/spf13/cobra"
)

type registrationStruct struct {
	builder           BuilderFunc
	externalParamFunc ExternalParamFunc
}

var registration = map[string]*registrationStruct{}

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

func RegisterWithFactory(fact Factory, logger ocfllogger.OCFLLogger) {
	for name, reg := range registration {
		if reg == nil || (reg.builder == nil && reg.externalParamFunc == nil) {
			logger.Debug().Msgf("extension %s has no builder or external param func", name)
			continue
		}
		fact.RegisterExtension(name, reg.builder)
	}
}

// todo: get rid of cobra.Command and config.GOCFLConfig
func GetExtensionParamValues(cmd *cobra.Command, conf *config.GOCFLConfig) (map[string]string, error) {
	var result = map[string]string{}
	extParams, err := GetExternalParams()
	if err != nil {
		return nil, errors.Wrap(err, "GetExternalParams")
	}
	for _, param := range extParams {
		name, value := param.GetParam(cmd, conf)
		if name != "" {
			result[name] = value
		}
	}
	return result, nil
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

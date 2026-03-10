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

// todo: get rid of cobra.Command and config.GOCFLConfig
func GetExtensionParamValuesOld(cmd *cobra.Command, conf *config.GOCFLConfig) (map[string]string, error) {
	var result = map[string]string{}
	extParams, err := GetExternalParams()
	if err != nil {
		return nil, errors.Wrap(err, "GetExternalParams")
	}
	for _, param := range extParams {
		param.GetParam(cmd.Name(), func(name, extensionName, param, defaultValue string) {
			if name == "" {
				return
			}
			var value string
			confExt, ok := conf.Extension[extensionName]
			if ok {
				if str, ok := confExt[param]; ok {
					if str != "" {
						value = str
					}
				}
			}
			_value, _ := cmd.Flags().GetString(name)
			// if cmdline value ist set or config value is empty
			if (_value != "" && _value != defaultValue) || value == "" {
				value = _value
			}
			result[name] = value
		})
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

package extension

import (
	"reflect"
	"strings"

	"emperror.dev/errors"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

type registrationStruct struct {
	builder           BuilderFunc
	externalParamFunc ExternalParamFunc
	documentation     *string
}

var registration = map[string]*registrationStruct{}

// RegisterExtensionStorageRoot registers an extension for storage roots with a given name, builder function, and external parameter function.
func RegisterExtensionStorageRoot(name string, builder BuilderFunc, externalParamFunc ExternalParamFunc, doc *string) {
	reg := &registrationStruct{
		documentation:     doc,
		builder:           builder,
		externalParamFunc: externalParamFunc,
	}
	registration["storageroot.ExtensionManager."+name] = reg
}

// RegisterExtensionObject registers an extension for objects with a given name, builder function, and external parameter function.
func RegisterExtensionObject(name string, builder BuilderFunc, externalParamFunc ExternalParamFunc, doc *string) {
	reg := &registrationStruct{
		documentation:     doc,
		builder:           builder,
		externalParamFunc: externalParamFunc,
	}
	registration["object.ExtensionManager."+name] = reg
}

// RegisterWithFactory registers all extensions with the provided factory if they have a valid builder function and match the factory's manager type.
func RegisterWithFactory[T ManagerCore[T]](fact Factory[T], logger ocfllogger.OCFLLogger) {
	typeName := reflect.TypeOf((*T)(nil)).Elem().String() + "."
	for name, reg := range registration {
		if !strings.HasPrefix(name, typeName) {
			continue
		}
		if reg == nil || reg.builder == nil {
			logger.Debug().Msgf("extension %s has no builder func", name)
			continue
		}
		fact.RegisterExtension(name[len(typeName):], reg.builder, reg.documentation)
	}
}

// GetExtensionParamValues retrieves all external parameter values for a given command.
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

// GetExternalParams returns all external parameters for all registered extensions.
func GetExternalParams() ([]*ExternalParam, error) {
	result := []*ExternalParam{}
	for name, reg := range registration {
		if reg == nil || reg.externalParamFunc == nil {
			continue
		}
		ps, err := reg.externalParamFunc()
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get external params for extension %s", name)
		}
		for _, p := range ps {
			if p != nil {
				result = append(result, p)
			}
		}
	}
	return result, nil
}

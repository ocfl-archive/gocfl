package extension

import (
	"fmt"
	"slices"
)

type ExternalParam struct {
	ExtensionName string
	Functions     []string
	Param         string
	//	File          string
	Description string
	Default     string
}

type SetParamsFunc func(name string, defaultValue string, description string)
type GetParamsFunc func(name, extensionName, param, defaultValue string)

func (eep *ExternalParam) SetParam(command string, callback SetParamsFunc) {
	if !slices.Contains(eep.Functions, command) {
		return
	}
	callback(eep.GetParamName(), eep.Default, eep.Description)
}

func (eep *ExternalParam) GetParam(command string, callback GetParamsFunc) {
	if !slices.Contains(eep.Functions, command) {
		return
	}
	callback(eep.GetParamName(), eep.ExtensionName, eep.Param, eep.Default)
}

func (eep *ExternalParam) GetParamName() string {
	flagName := fmt.Sprintf("ext-%s-%s", eep.ExtensionName, eep.Param)
	return flagName
}

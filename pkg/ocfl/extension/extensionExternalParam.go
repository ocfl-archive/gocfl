package extension

import (
	"fmt"
	"slices"
)

// ExternalParam represents a parameter for an extension that can be set externally (e.g., via CLI).
type ExternalParam struct {
	ExtensionName string   // Name of the extension this parameter belongs to.
	Functions     []string // List of commands/functions where this parameter is applicable.
	Param         string   // Name of the parameter itself.
	//	File          string
	Description string // Human-readable description of the parameter.
	Default     string // Default value if none is provided.
}

// SetParamsFunc is a callback type for registering a parameter.
type SetParamsFunc func(name string, defaultValue string, description string)

// GetParamsFunc is a callback type for retrieving a parameter value.
type GetParamsFunc func(name, extensionName, param, defaultValue string)

// SetParam checks if the parameter is applicable to the given command and, if so, calls the provided callback.
func (eep *ExternalParam) SetParam(command string, callback SetParamsFunc) {
	if !slices.Contains(eep.Functions, command) {
		return
	}
	callback(eep.GetParamName(), eep.Default, eep.Description)
}

// GetParam checks if the parameter is applicable to the given command and, if so, calls the provided callback.
func (eep *ExternalParam) GetParam(command string, callback GetParamsFunc) {
	if !slices.Contains(eep.Functions, command) {
		return
	}
	callback(eep.GetParamName(), eep.ExtensionName, eep.Param, eep.Default)
}

// GetParamName returns the formatted name of the parameter used in external interfaces (e.g., ext-NAME-PARAM).
func (eep *ExternalParam) GetParamName() string {
	flagName := fmt.Sprintf("ext-%s-%s", eep.ExtensionName, eep.Param)
	return flagName
}

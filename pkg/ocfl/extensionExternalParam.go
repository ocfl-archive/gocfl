package ocfl

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/exp/slices"
)

type ExtensionExternalParam struct {
	ExtensionName string
	Functions     []string
	Param         string
	File          string
	Description   string
	Default       string
}

func (eep *ExtensionExternalParam) SetParam(cmd *cobra.Command) {
	if !slices.Contains(eep.Functions, cmd.Name()) {
		return
	}
	name := eep.GetCobraName()
	cmd.Flags().String(name, eep.Default, eep.Description)
	if eep.File != "" {
		viper.BindPFlag(eep.GetViperName(cmd.Name()), cmd.Flags().Lookup(name))
	}
}

func (eep *ExtensionExternalParam) GetParam(cmd *cobra.Command) (name, value string) {
	if !slices.Contains(eep.Functions, cmd.Name()) {
		return
	}
	name = eep.GetCobraName()
	if eep.File != "" {
		value = viper.GetString(eep.GetViperName(cmd.Name()))
	} else {
		value, _ = cmd.Flags().GetString(name)
	}
	return
}

func (eep *ExtensionExternalParam) GetCobraName() string {
	flagName := fmt.Sprintf("ext-%s-%s", eep.ExtensionName, eep.Param)
	return flagName
}

func (eep *ExtensionExternalParam) GetViperName(action string) string {
	cfgName := fmt.Sprintf("%s.ext.%s.%s", action, eep.ExtensionName, eep.File)
	return cfgName
}

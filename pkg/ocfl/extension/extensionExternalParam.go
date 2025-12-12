package extension

import (
	"fmt"

	"github.com/ocfl-archive/gocfl/v2/config"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

type ExtensionExternalParam struct {
	ExtensionName string
	Functions     []string
	Param         string
	//	File          string
	Description string
	Default     string
}

func (eep *ExtensionExternalParam) SetParam(cmd *cobra.Command) {
	if !slices.Contains(eep.Functions, cmd.Name()) {
		return
	}
	name := eep.GetCobraName()
	cmd.Flags().String(name, eep.Default, eep.Description)
}

func (eep *ExtensionExternalParam) GetParam(cmd *cobra.Command, conf *config.GOCFLConfig) (name, value string) {
	if !slices.Contains(eep.Functions, cmd.Name()) {
		return
	}
	name = eep.GetCobraName()
	value, _ = cmd.Flags().GetString(name)
	confExt, ok := conf.Extension[eep.ExtensionName]
	if ok {
		if str, ok := confExt[eep.Param]; ok {
			if str != "" {
				value = str
			}
		}
	}
	return
}

func (eep *ExtensionExternalParam) GetCobraName() string {
	flagName := fmt.Sprintf("ext-%s-%s", eep.ExtensionName, eep.Param)
	return flagName
}

package ocfl

import (
	"emperror.dev/errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.ub.unibas.ch/gocfl/v2/pkg/checksum"
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
		viper.BindPFlag(eep.GetViperName(cmd.Name()), cmd.Flags().Lookup("name"))
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

type ExtensionConfig struct {
	ExtensionName string `json:"extensionName"`
}

type Extension interface {
	GetName() string
	SetFS(fs OCFLFS)
	SetParams(params map[string]string) error
	WriteConfig() error
	GetConfigString() string
}

const (
	ExtensionStorageRootPathName    = "StorageRootPath"
	ExtensionObjectContentPathName  = "ObjectContentPath"
	ExtensionObjectExtractPathName  = "ObjectExtractPath"
	ExtensionObjectExternalPathName = "ObjectExternalPath"
	ExtensionContentChangeName      = "ContentChange"
	ExtensionObjectChangeName       = "ObjectChange"
	ExtensionFixityDigestName       = "FixityDigest"
)

type ExtensionStorageRootPath interface {
	Extension
	WriteLayout(fs OCFLFS) error
	BuildStorageRootPath(storageRoot StorageRoot, id string) (string, error)
}

type ExtensionObjectContentPath interface {
	Extension
	BuildObjectContentPath(object Object, originalPath string, area string) (string, error)
}

var ExtensionObjectExtractPathWrongAreaError = errors.New("invalid area")

type ExtensionObjectExtractPath interface {
	Extension
	BuildObjectExtractPath(object Object, originalPath string) (string, error)
}

type ExtensionObjectExternalPath interface {
	Extension
	BuildObjectExternalPath(object Object, originalPath string, area string) (string, error)
}

type ExtensionContentChange interface {
	Extension
	AddFileBefore(object Object, source, dest string) error
	UpdateFileBefore(object Object, source, dest string) error
	DeleteFileBefore(object Object, dest string) error
	AddFileAfter(object Object, source, dest string) error
	UpdateFileAfter(object Object, source, dest string) error
	DeleteFileAfter(object Object, dest string) error
}

type ExtensionObjectChange interface {
	Extension
	UpdateObjectBefore(object Object) error
	UpdateObjectAfter(object Object) error
}

type ExtensionFixityDigest interface {
	Extension
	GetFixityDigests() []checksum.DigestAlgorithm
}

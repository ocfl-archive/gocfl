package config

import (
	"emperror.dev/errors"
	"github.com/BurntSushi/toml"
	"github.com/je4/indexer/v2/pkg/indexer"
	"github.com/je4/utils/v2/pkg/checksum"
)

type InitConfig struct {
	OCFLVersion                string
	StorageRootExtensionFolder string `toml:"storagerootextensions"`
}

type AddConfig struct {
	Deduplicate           bool
	NoCompress            bool
	ObjectExtensionFolder string `toml:"objectextensions"`
}

type AESConfig struct {
	Enable bool
	Key    string
	IV     string
}

type DisplayConfig struct {
	Addr     string
	AddrExt  string
	CertFile string
	KeyFile  string
}

type ExtractConfig struct {
	Manifest bool
	Version  string
}

type ExtractMetaConfig struct {
	Version string
	Format  string
	Output  string
}

type StatConfig struct {
	Info []string
}

type UpdateConfig struct {
	Deduplicate bool
	NoCompress  bool
	Echo        bool
}

type DefaultUserConfig struct {
	UserName    string
	UserAddress string
}

type GOCFLConfig struct {
	ErrorTemplate  string
	Logfile        string
	Loglevel       string
	LogFormat      string
	AccessLog      string
	Digest         checksum.DigestAlgorithm
	Indexer        *indexer.IndexerConfig
	AES            *AESConfig
	Init           *InitConfig
	Add            *AddConfig
	Display        *DisplayConfig
	Extract        *ExtractConfig
	ExtractMeta    *ExtractMetaConfig
	Stat           *StatConfig
	DefaultUser    *DefaultUserConfig
	DefaultMessage string
}

func LoadGOCFLConfig(data string) (*GOCFLConfig, error) {
	var conf = &GOCFLConfig{
		LogFormat: `%{time:2006-01-02T15:04:05.000} %{shortpkg}::%{longfunc} [%{shortfile}] > %{level:.5s} - %{message}`,
		Loglevel:  "ERROR",
		Indexer:   indexer.GetDefaultConfig(),
		AES:       &AESConfig{},
		Init:      &InitConfig{},
		Add: &AddConfig{
			Deduplicate:           false,
			NoCompress:            true,
			ObjectExtensionFolder: "",
		},
		Display: &DisplayConfig{
			Addr:    "localhost:80",
			AddrExt: "http://localhost:80/",
		},
		Extract: &ExtractConfig{
			Manifest: false,
			Version:  "latest",
		},
		ExtractMeta: &ExtractMetaConfig{
			Version: "latest",
			Format:  "json",
		},
		Stat: &StatConfig{
			Info: []string{
				"ExtensionConfigs",
				"Objects",
				"ObjectVersionState",
				"ObjectManifest",
				"ObjectFolders",
				"Extension",
				"ObjectVersions",
				"ObjectExtension",
				"ObjectExtensionConfigs",
			},
		},
		DefaultUser:    &DefaultUserConfig{},
		DefaultMessage: "initial add",
	}

	if _, err := toml.Decode(data, conf); err != nil {
		return nil, errors.Wrap(err, "Error on loading config")
	}

	return conf, nil
}

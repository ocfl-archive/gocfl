package config

import (
	"emperror.dev/errors"
	"github.com/BurntSushi/toml"
	"github.com/je4/indexer/v2/pkg/indexer"
	"github.com/je4/utils/v2/pkg/checksum"
	configutil "github.com/je4/utils/v2/pkg/config"
)

type InitConfig struct {
	OCFLVersion                string
	StorageRootExtensionFolder string `toml:"storagerootextensions"`
	Digest                     checksum.DigestAlgorithm
}

type AddConfig struct {
	Deduplicate           bool
	NoCompress            bool
	ObjectExtensionFolder string `toml:"objectextensions"`
	User                  *UserConfig
	Digest                checksum.DigestAlgorithm
	Fixity                []string
	Message               string
}

type AESConfig struct {
	Enable       bool
	KeepassFile  configutil.EnvString
	KeepassEntry configutil.EnvString
	KeepassKey   configutil.EnvString
	IV           configutil.EnvString
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

type UserConfig struct {
	Name    string
	Address string
}

type S3Config struct {
	Endpoint    configutil.EnvString
	AccessKeyID configutil.EnvString
	AccessKey   configutil.EnvString
	Region      configutil.EnvString
}

type GOCFLConfig struct {
	ErrorTemplate  string
	Logfile        string
	Loglevel       string
	LogFormat      string
	AccessLog      string
	Indexer        *indexer.IndexerConfig
	AES            *AESConfig
	Init           *InitConfig
	Add            *AddConfig
	Display        *DisplayConfig
	Extract        *ExtractConfig
	ExtractMeta    *ExtractMetaConfig
	Stat           *StatConfig
	S3             *S3Config
	DefaultMessage string
}

func LoadGOCFLConfig(data string) (*GOCFLConfig, error) {
	var conf = &GOCFLConfig{
		LogFormat: `%{time:2006-01-02T15:04:05.000} %{shortpkg}::%{longfunc} [%{shortfile}] > %{level:.5s} - %{message}`,
		Loglevel:  "ERROR",
		Indexer:   indexer.GetDefaultConfig(),
		AES:       &AESConfig{},
		Add: &AddConfig{
			Deduplicate:           false,
			NoCompress:            true,
			ObjectExtensionFolder: "",
			User:                  &UserConfig{},
			Fixity:                []string{},
			Message:               "",
			Digest:                "",
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
		Init: &InitConfig{
			OCFLVersion:                "1.1",
			StorageRootExtensionFolder: "",
		},
		S3:             &S3Config{},
		DefaultMessage: "initial add",
	}

	if _, err := toml.Decode(data, conf); err != nil {
		return nil, errors.Wrap(err, "Error on loading config")
	}

	return conf, nil
}

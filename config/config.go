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

type UpdateConfig struct {
	Deduplicate bool
	NoCompress  bool
	User        *UserConfig
	Echo        bool
	Message     string
	Digest      checksum.DigestAlgorithm
}

type AESConfig struct {
	Enable       bool
	KeepassFile  configutil.EnvString
	KeepassEntry configutil.EnvString
	KeepassKey   configutil.EnvString
	IV           configutil.EnvString
}

type DisplayConfig struct {
	Addr      string
	AddrExt   string
	CertFile  string
	KeyFile   string
	Templates string
}

type ExtractConfig struct {
	Manifest   bool
	Version    string
	ObjectPath string
	ObjectID   string
	Area       string
}

type ValidateConfig struct {
	ObjectPath string
	ObjectID   string
}

type ExtractMetaConfig struct {
	Version    string
	Format     string
	Output     string
	ObjectPath string
	ObjectID   string
}

type StatConfig struct {
	Info       []string
	ObjectPath string
	ObjectID   string
}

type UserConfig struct {
	Name    string
	Address string
}

type ThumbnailFunction struct {
	ID      string
	Title   string
	Command string
	Timeout configutil.Duration
	Pronoms []string
	Mime    []string
}

type Thumbnail struct {
	Enabled    bool
	Background string
	Function   map[string]*ThumbnailFunction
}

type MigrationFunction struct {
	ID                  string
	Title               string
	Command             string
	Strategy            string
	FilenameRegexp      string
	FilenameReplacement string
	Timeout             configutil.Duration
	Pronoms             []string
}

type Migration struct {
	Enabled  bool
	Function map[string]*MigrationFunction
}

type S3Config struct {
	Endpoint    configutil.EnvString
	AccessKeyID configutil.EnvString
	AccessKey   configutil.EnvString
	Region      configutil.EnvString
}

type GOCFLConfig struct {
	ErrorTemplate string
	Logfile       string
	LogLevel      string
	LogFormat     string
	AccessLog     string
	Extension     map[string]map[string]string
	Indexer       *indexer.IndexerConfig
	Thumbnail     *Thumbnail
	Migration     *Migration
	AES           *AESConfig
	Init          *InitConfig
	Add           *AddConfig
	Update        *UpdateConfig
	Display       *DisplayConfig
	Extract       *ExtractConfig
	ExtractMeta   *ExtractMetaConfig
	Stat          *StatConfig
	Validate      *ValidateConfig
	S3            *S3Config
	DefaultArea   string
}

func LoadGOCFLConfig(data string) (*GOCFLConfig, error) {
	var conf = &GOCFLConfig{
		LogFormat:   `%{time:2006-01-02T15:04:05.000} %{shortpkg}::%{longfunc} [%{shortfile}] > %{level:.5s} - %{message}`,
		LogLevel:    "ERROR",
		DefaultArea: "content",
		Extension:   map[string]map[string]string{},
		Indexer:     indexer.GetDefaultConfig(),
		Thumbnail: &Thumbnail{
			Enabled:    false,
			Background: "",
			Function:   map[string]*ThumbnailFunction{},
		},
		Migration: &Migration{
			Enabled:  false,
			Function: map[string]*MigrationFunction{},
		},
		AES: &AESConfig{},
		Add: &AddConfig{
			Deduplicate:           false,
			NoCompress:            true,
			ObjectExtensionFolder: "",
			User:                  &UserConfig{},
			Fixity:                []string{},
			Message:               "initial add",
			Digest:                "sha512",
		},
		Update: &UpdateConfig{
			Deduplicate: true,
			NoCompress:  false,
			User:        &UserConfig{},
			Echo:        false,
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
		Validate: &ValidateConfig{},
		Init: &InitConfig{
			OCFLVersion:                "1.1",
			StorageRootExtensionFolder: "",
		},
		S3: &S3Config{},
	}

	if _, err := toml.Decode(data, conf); err != nil {
		return nil, errors.Wrap(err, "Error on loading config")
	}

	return conf, nil
}

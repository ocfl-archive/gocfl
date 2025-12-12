package config

import (
	"emperror.dev/errors"
	"github.com/BurntSushi/toml"
	"github.com/je4/utils/v2/pkg/checksum"
	configutil "github.com/je4/utils/v2/pkg/config"
	"github.com/je4/utils/v2/pkg/stashconfig"
	"github.com/ocfl-archive/indexer/v3/pkg/indexer"
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
	Addr      string `toml:"addr"`
	AddrExt   string `toml:"addrext"`
	CertFile  string `toml:"certfile"`
	KeyFile   string `toml:"keyfile"`
	Templates string `toml:"templates"`
	Obfuscate bool   `toml:"obfuscate"`
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
	Obfuscate  bool
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
	ErrorTemplate string                       `toml:"errortemplate"`
	ErrorConfig   string                       `toml:"errorconfig"`
	AccessLog     string                       `toml:"accesslog"`
	Extension     map[string]map[string]string `json:"extension"`
	Indexer       *indexer.IndexerConfig       `toml:"indexer"`
	Thumbnail     Thumbnail                    `toml:"thumbnail"`
	Migration     Migration                    `toml:"migration"`
	AES           AESConfig                    `toml:"aes"`
	Init          InitConfig                   `toml:"init"`
	Add           AddConfig                    `toml:"add"`
	Update        UpdateConfig                 `toml:"update"`
	Display       DisplayConfig                `toml:"display"`
	Extract       ExtractConfig                `toml:"extract"`
	ExtractMeta   ExtractMetaConfig            `toml:"extractmeta"`
	Stat          StatConfig                   `toml:"stat"`
	Validate      ValidateConfig               `toml:"validate"`
	S3            S3Config                     `toml:"s3"`
	DefaultArea   string                       `toml:"defaultarea"`
	Log           stashconfig.Config           `toml:"log"`
}

func LoadGOCFLConfig(filename string) (*GOCFLConfig, error) {
	var conf = &GOCFLConfig{
		Indexer: indexer.GetDefaultConfig(),
	}
	if _, err := toml.Decode(defaultConfig, conf); err != nil {
		return nil, errors.Wrap(err, "error decoding GOCFL default configuration")
	}
	if filename == "" {
		return conf, nil
	}
	if _, err := toml.DecodeFile(filename, conf); err != nil {
		return nil, errors.Wrapf(err, "error decoding configuration file %s", filename)
	}
	return conf, nil
}

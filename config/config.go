package config

import (
	"os"

	"emperror.dev/errors"
	"github.com/BurntSushi/toml"
	"github.com/je4/filesystem/v3/pkg/vfsrw"
	"github.com/je4/utils/v2/pkg/checksum"
	configutil "github.com/je4/utils/v2/pkg/config"
	"github.com/je4/utils/v2/pkg/stashconfig"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/util"
	"github.com/ocfl-archive/indexer/v3/pkg/indexer"
	indexerutil "github.com/ocfl-archive/indexer/v3/pkg/util"
)

const DefaultPath = "~/gocfl/gocfl.toml"

type InitConfig struct {
	OCFLVersion                string                   `toml:"ocflversion"`
	StorageRootExtensionFolder string                   `toml:"storagerootextensions"`
	Digest                     checksum.DigestAlgorithm `toml:"digest"`
}

type AddConfig struct {
	Deduplicate           bool                     `toml:"deduplicate"`
	NoCompress            bool                     `toml:"nocompress"`
	ObjectExtensionFolder string                   `toml:"objectextensions"`
	User                  *UserConfig              `toml:"user"`
	Digest                checksum.DigestAlgorithm `toml:"digest"`
	Fixity                []string                 `toml:"fixity"`
	Message               string                   `toml:"message"`
}

type UpdateConfig struct {
	Deduplicate bool                     `toml:"deduplicate"`
	NoCompress  bool                     `toml:"nocompress"`
	User        *UserConfig              `toml:"user"`
	Echo        bool                     `toml:"echo"`
	Message     string                   `toml:"message"`
	Digest      checksum.DigestAlgorithm `toml:"digest"`
}

type AESConfig struct {
	Enable       bool                 `toml:"enable"`
	KeepassFile  configutil.EnvString `toml:"keepassfile"`
	KeepassEntry configutil.EnvString `toml:"keepassentry"`
	KeepassKey   configutil.EnvString `toml:"keepasskey"`
	IV           configutil.EnvString `toml:"iv"`
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
	Manifest   bool   `toml:"manifest"`
	Version    string `toml:"version"`
	ObjectPath string `toml:"objectpath"`
	ObjectID   string `toml:"objectid"`
	Area       string `toml:"area"`
}

type ValidateConfig struct {
	ObjectPath string `toml:"objectpath"`
	ObjectID   string `toml:"objectid"`
}

type ExtractMetaConfig struct {
	Version    string `toml:"version"`
	Format     string `toml:"format"`
	Output     string `toml:"output"`
	ObjectPath string `toml:"objectpath"`
	ObjectID   string `toml:"objectid"`
	Obfuscate  bool   `toml:"obfuscate"`
}

type StatConfig struct {
	Info       []string `toml:"info"`
	ObjectPath string   `toml:"objectpath"`
	ObjectID   string   `toml:"objectid"`
}

type TestConfig struct {
	FixturePath string `toml:"fixturepath"`
	ObjectPath  string `toml:"objectpath"`
}

type UserConfig struct {
	Name    string `toml:"name"`
	Address string `toml:"address"`
}

type ThumbnailFunction struct {
	ID      string              `toml:"id"`
	Title   string              `toml:"title"`
	Command string              `toml:"command"`
	Timeout configutil.Duration `toml:"timeout"`
	Pronoms []string            `toml:"pronoms"`
	Mime    []string            `toml:"mime"`
}

type Thumbnail struct {
	Enabled    bool                          `toml:"enabled"`
	Background string                        `toml:"background"`
	Function   map[string]*ThumbnailFunction `toml:"function"`
}

type MigrationFunction struct {
	ID                  string              `toml:"id"`
	Title               string              `toml:"title"`
	Command             string              `toml:"command"`
	Strategy            string              `toml:"strategy"`
	FilenameRegexp      string              `toml:"filenameregexp"`
	FilenameReplacement string              `toml:"filenamereplacement"`
	Timeout             configutil.Duration `toml:"timeout"`
	Pronoms             []string            `toml:"pronoms"`
}

type Migration struct {
	Enabled  bool                          `toml:"enabled"`
	Function map[string]*MigrationFunction `toml:"function"`
}

type S3Config struct {
	Endpoint    configutil.EnvString `toml:"endpoint"`
	AccessKeyID configutil.EnvString `toml:"accesskeyid"`
	AccessKey   configutil.EnvString `toml:"accesskey"`
	Region      configutil.EnvString `toml:"region"`
}

type StoreConfig struct {
	ConfigFolder    string `toml:"configfolder"`
	TOMLFile        string `toml:"tomlfile"`
	ExtensionFolder string `toml:"extensionfolder"`
	ScriptFolder    string `toml:"scriptfolder"`
	FullConfig      bool   `toml:"fullconfig"`
	Extensions      bool   `toml:"extensions"`
	Scripts         bool   `toml:"scripts"`
}

type GOCFLConfig struct {
	ErrorTemplate string                       `toml:"errortemplate"`
	ErrorConfig   string                       `toml:"errorconfig"`
	AccessLog     string                       `toml:"accesslog"`
	Extension     map[string]map[string]string `toml:"extension"`
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
	Test          TestConfig                   `toml:"test"`
	Validate      ValidateConfig               `toml:"validate"`
	StoreConfig   StoreConfig                  `toml:"storeconfig"`
	S3            S3Config                     `toml:"s3"`
	DefaultArea   string                       `toml:"defaultarea"`
	VFS           vfsrw.Config                 `toml:"vfs"`
	Log           stashconfig.Config           `toml:"log"`
}

func LoadGOCFLConfig(filename string) (*GOCFLConfig, error) {
	var err error
	var conf = &GOCFLConfig{
		Indexer: indexer.GetDefaultConfig(),
	}
	if _, err := toml.Decode(defaultConfig, conf); err != nil {
		return nil, errors.Wrap(err, "error decoding GOCFL default configuration")
	}
	if filename == "" {
		filename, err = util.Fullpath(DefaultPath)
		if err != nil {
			return nil, errors.Wrapf(err, "error getting '~/gocfl/gocfl.toml' file path")
		}
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			filename = ""
		}
	}
	if filename != "" {
		if _, err := toml.DecodeFile(filename, conf); err != nil {
			return nil, errors.Wrapf(err, "error decoding configuration file %s", filename)
		}
	}
	if conf.Indexer.Optimize {
		err := indexerutil.OptimizeConfig(conf.Indexer, nil)
		if err != nil {
			return nil, errors.Wrap(err, "error optimizing indexer")
		}
	}
	return conf, nil
}

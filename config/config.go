package config

import (
	"emperror.dev/errors"
	"github.com/BurntSushi/toml"
	"github.com/je4/utils/v2/pkg/checksum"
	configutil "github.com/je4/utils/v2/pkg/config"
	"github.com/je4/utils/v2/pkg/stashconfig"
	"github.com/ocfl-archive/gocfl/v2/docs"
	"github.com/ocfl-archive/indexer/v3/pkg/indexer"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"os"
	"strings"
)

type InitConfig struct {
	OCFLVersion                string                   `toml:"ocflversion"`
	StorageRootExtensionFolder string                   `toml:"storagerootextensionfolder"`
	Digest                     checksum.DigestAlgorithm `toml:"digest"`
	Documentation              string                   `toml:"documentation"`
}

type AddConfig struct {
	Deduplicate           bool                     `toml:"deduplicate"`
	NoCompress            bool                     `toml:"nocompress"`
	ObjectExtensionFolder string                   `toml:"objectextensionfolder"`
	User                  *UserConfig              `toml:"User"`
	Digest                checksum.DigestAlgorithm `toml:"digest"`
	Fixity                []string                 `toml:"fixity"`
	Message               string                   `toml:"message"`
}

type UpdateConfig struct {
	Deduplicate bool                     `toml:"deduplicate"`
	NoCompress  bool                     `toml:"nocompress"`
	User        *UserConfig              `toml:"User"`
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
	Key          configutil.EnvString `toml:"key"`
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
	Types   []string            `toml:"types"`
}

type Thumbnail struct {
	Enabled    bool                          `toml:"enabled"`
	Background string                        `toml:"background"`
	Function   map[string]*ThumbnailFunction `toml:"Function"`
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
	Function map[string]*MigrationFunction `toml:"Function"`
}

type S3Config struct {
	Endpoint    configutil.EnvString `toml:"endpoint"`
	AccessKeyID configutil.EnvString `toml:"accesskeyid"`
	AccessKey   configutil.EnvString `toml:"accesskey"`
	Region      configutil.EnvString `toml:"region"`
}

type GOCFLConfig struct {
	ErrorTemplate string                       `toml:"errortemplate"`
	ErrorConfig   string                       `toml:"errorconfig"`
	AccessLog     string                       `toml:"accesslog"`
	Extension     map[string]map[string]string `toml:"extension"`
	Indexer       *indexer.IndexerConfig       `toml:"Indexer"`
	Thumbnail     *Thumbnail                   `toml:"Thumbnail"`
	Migration     *Migration                   `toml:"Migration"`
	AES           *AESConfig                   `toml:"AES"`
	Init          *InitConfig                  `toml:"Init"`
	Add           *AddConfig                   `toml:"Add"`
	Update        *UpdateConfig                `toml:"Update"`
	Display       *DisplayConfig               `toml:"Display"`
	Extract       *ExtractConfig               `toml:"Extract"`
	ExtractMeta   *ExtractMetaConfig           `toml:"Extractmeta"`
	Stat          *StatConfig                  `toml:"Stat"`
	Validate      *ValidateConfig              `toml:"Validate"`
	S3            *S3Config                    `toml:"S3"`
	DefaultArea   string                       `toml:"defaultarea"`
	Log           stashconfig.Config           `toml:"Log"`
	TempDir       string                       `toml:"tempdir"`
}

func LoadGOCFLConfig(data string) (*GOCFLConfig, error) {
	var conf = &GOCFLConfig{
		Log: stashconfig.Config{
			Level: "ERROR",
		},
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
			NoCompress:  true,
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
			Documentation:              "ocfl",
		},
		S3:      &S3Config{},
		TempDir: os.TempDir(),
	}

	if _, err := toml.Decode(data, conf); err != nil {
		return nil, errors.Wrap(err, "Error on loading config")
	}
	conf.Init.Documentation = strings.ToLower(conf.Init.Documentation)
	if conf.Init.Documentation != "" {
		if !slices.Contains(maps.Keys(docs.Documentations), conf.Init.Documentation) {
			return nil, errors.Errorf("unknown documentation '%s' please use %v", conf.Init.Documentation, maps.Keys(docs.Documentations))
		}
	}
	return conf, nil
}

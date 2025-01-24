package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"emperror.dev/errors"
	configutil "github.com/je4/utils/v2/pkg/config"
	archiveerror "github.com/ocfl-archive/error/pkg/error"
	"github.com/ocfl-archive/gocfl/v2/config"
	"github.com/ocfl-archive/gocfl/v2/internal"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl"
	"github.com/ocfl-archive/gocfl/v2/version"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const LOGFORMAT = `%{time:2006-01-02T15:04:05.000} %{shortpkg}::%{longfunc} [%{shortfile}] > %{level:.5s} - %{message}`

type LogLevelFlag uint

const (
	LOGLEVELCRITICAL LogLevelFlag = iota
	LOGLEVELERROR
	LOGLEVELWARNING
	LOGLEVELNOTICE
	LOGLEVELINFO
	LOGLEVELDEBUG
)

var LogLevelIds = map[LogLevelFlag][]string{
	LOGLEVELDEBUG:    {"DEBUG"},
	LOGLEVELINFO:     {"INFO"},
	LOGLEVELNOTICE:   {"NOTICE"},
	LOGLEVELWARNING:  {"WARNING"},
	LOGLEVELERROR:    {"ERROR"},
	LOGLEVELCRITICAL: {"CRITICAL"},
}

type VersionFlag uint

const (
	VERSION1_1 VersionFlag = iota
	VERSION1_0
)

var VersionIds = map[VersionFlag][]string{
	VERSION1_1: {"1.1", "v1.1"},
	VERSION1_0: {"1.0", "v1.0"},
}

var VersionIdsVersion = map[VersionFlag]ocfl.OCFLVersion{
	VERSION1_1: ocfl.Version1_1,
	VERSION1_0: ocfl.Version1_0,
}

type DigestFlag uint

const (
	DIGESTSHA512 DigestFlag = iota
	DIGESTSHA256
	DIGESTMD5
	DIGESTSHA1
	DIGESTBlake2b160
	DIGESTBlake2b256
	DIGESTBlake2b384
	DIGESTBlake2b512
)

var DigestIds = map[DigestFlag][]string{
	DIGESTSHA512:     {"sha512"},
	DIGESTSHA256:     {"sha256"},
	DIGESTMD5:        {"md5"},
	DIGESTSHA1:       {"sha1"},
	DIGESTBlake2b160: {"blake2b160"},
	DIGESTBlake2b256: {"blake2b256"},
	DIGESTBlake2b384: {"blake2b384"},
	DIGESTBlake2b512: {"blake2b512"},
}

// all possible flags of all modules go here
var persistentFlagConfigFile string

var persistentFlagErrorConfig string

var persistentFlagLogfile string
var persistentFlagLoglevel string

var persistenFlagS3Endpoint string
var persistenFlagS3AccessKeyID string
var persistenFlagS3SecretAccessKey string
var persistentFlagS3Region string

var flagObjectID string
var flagStatInfo = []string{}

var conf *config.GOCFLConfig
var ErrorFactory = archiveerror.NewFactory("gocfl")

var areaPathRegexp = regexp.MustCompile("^([a-z]{2,}):(.*)$")

var appname = "gocfl"

var rootCmd = &cobra.Command{
	Use:   appname,
	Short: "gocfl is a fast ocfl creator/extractor/validator with focus on zip containers",
	Long: `A fast and reliable OCFL creator, extractor and validator.
source code is available at: https://github.com/ocfl-archive/gocfl

by Jürgen Enge (University Library Basel, juergen@info-age.net)`,
	Version: fmt.Sprintf("%s '%s' (%s)", version.Version, version.ShortCommit(), version.Date),
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func getFlagString(cmd *cobra.Command, flag string) string {
	str, err := cmd.Flags().GetString(flag)
	if err != nil {
		_ = cmd.Help()
		cobra.CheckErr(errors.Errorf("canot get flag %s: %v", flag, err))
	}
	return str
}

func getFlagBool(cmd *cobra.Command, flag string) bool {
	b, err := cmd.Flags().GetBool(flag)
	if err != nil {
		_ = cmd.Help()
		cobra.CheckErr(errors.Errorf("canot get flag %s: %v", flag, err))
	}
	return b
}

func configErrorFactory() {
	var archiveErrs []*archiveerror.Error
	if conf.ErrorConfig != "" {
		errorExt := filepath.Ext(conf.ErrorConfig)
		var err error
		switch errorExt {
		case ".toml":
			archiveErrs, err = archiveerror.LoadTOMLFile(conf.ErrorConfig)
		case ".yaml":
			archiveErrs, err = archiveerror.LoadYAMLFile(conf.ErrorConfig)
		default:
			err = errors.Errorf("unknown error config file extension %s", errorExt)
		}
		if err != nil {
			log.Fatal().Err(err).Msgf("cannot load error config file %s", conf.ErrorConfig)
		}
	} else {
		var err error
		const errorsEmbedToml string = "errors.toml"
		archiveErrs, err = archiveerror.LoadTOMLFileFS(internal.InternalFS, errorsEmbedToml)
		if err != nil {
			log.Fatal().Err(err).Msg("cannot load error config file")
		}
	}
	if err := ErrorFactory.RegisterErrors(archiveErrs); err != nil {
		log.Fatal().Err(err).Msg("cannot register errors")
	}
}

func initConfig() {

	// load config file
	if persistentFlagConfigFile != "" {
		var err error
		persistentFlagConfigFile, err = ocfl.Fullpath(persistentFlagConfigFile)
		if err != nil {
			cobra.CheckErr(errors.Errorf("cannot convert '%s' to absolute path: %v", persistentFlagConfigFile, err))
			return
		}
		log.Info().Msgf("loading configuration from %s", persistentFlagLogfile)
		data, err := os.ReadFile(persistentFlagConfigFile)
		if err != nil {
			_ = rootCmd.Help()
			log.Error().Msgf("error reading config file %s: %v\n", persistentFlagConfigFile, err)
			os.Exit(1)
		}
		conf, err = config.LoadGOCFLConfig(string(data))
		if err != nil {
			_ = rootCmd.Help()
			log.Error().Msgf("error loading config file %s: %v\n", persistentFlagConfigFile, err)
			os.Exit(1)
		}
	} else {
		var err error
		conf, err = config.LoadGOCFLConfig(string(config.DefaultConfig))
		if err != nil {
			_ = rootCmd.Help()
			log.Error().Msgf("error loading config file %s: %v\n", persistentFlagConfigFile, err)
			os.Exit(1)
		}
	}

	// overwrite config file with command line data
	if persistentFlagErrorConfig != "" {
		conf.ErrorConfig = persistentFlagErrorConfig
	}
	if persistentFlagLogfile != "" {
		conf.Log.File = persistentFlagLogfile
	}
	if persistentFlagLoglevel != "" {
		conf.Log.Level = persistentFlagLoglevel
	}
	if persistenFlagS3Endpoint != "" {
		conf.S3.Endpoint = configutil.EnvString(persistenFlagS3Endpoint)
	}
	if persistentFlagS3Region != "" {
		conf.S3.Region = configutil.EnvString(persistentFlagS3Region)
	}
	if persistenFlagS3AccessKeyID != "" {
		conf.S3.AccessKeyID = configutil.EnvString(persistenFlagS3AccessKeyID)
	}
	if persistenFlagS3SecretAccessKey != "" {
		conf.S3.AccessKey = configutil.EnvString(persistenFlagS3SecretAccessKey)
	}

	configErrorFactory()
	return
}

func setExtensionFlags(commands ...*cobra.Command) {
	extensionParams := GetExtensionParams()
	for _, cmd := range commands {
		for _, param := range extensionParams {
			param.SetParam(cmd)
		}
	}
}

func init() {

	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&persistentFlagConfigFile, "config", "", "config file (default is embedded)")
	rootCmd.PersistentFlags().StringVar(&persistentFlagErrorConfig, "error-config", "", "error config file (default is embedded)")
	rootCmd.PersistentFlags().StringVar(&persistentFlagLogfile, "log-file", "", "log output file (default is console)")
	rootCmd.PersistentFlags().StringVar(&persistentFlagLoglevel, "log-level", "", "log level (CRITICAL|ERROR|WARNING|NOTICE|INFO|DEBUG)")
	rootCmd.PersistentFlags().StringVar(&persistenFlagS3Endpoint, "s3-endpoint", "", "Endpoint for S3 Buckets")
	rootCmd.PersistentFlags().StringVar(&persistenFlagS3AccessKeyID, "s3-access-key-id", "", "Access Key ID for S3 Buckets")
	rootCmd.PersistentFlags().StringVar(&persistenFlagS3SecretAccessKey, "s3-secret-access-key", "", "Secret Access Key for S3 Buckets")
	rootCmd.PersistentFlags().StringVar(&persistentFlagS3Region, "s3-region", "", "Region for S3 Access")

	initValidate()
	initInit()
	initCreate()
	initAdd()
	initUpdate()
	initStat()
	initExtract()
	initExtractMeta()
	initDisplay()

	setExtensionFlags(validateCmd, initCmd, createCmd, addCmd, updateCmd, statCmd, extractCmd, extractMetaCmd, displayCmd)
	rootCmd.AddCommand(validateCmd, initCmd, createCmd, addCmd, updateCmd, statCmd, extractCmd, extractMetaCmd, displayCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

package cmd

import (
	"emperror.dev/errors"
	"fmt"
	"github.com/google/martian/log"
	"github.com/je4/gocfl/v2/config"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	configutil "github.com/je4/utils/v2/pkg/config"
	"github.com/spf13/cobra"
	"os"
	"regexp"
)

const VERSION = "v1.0-beta.9"

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

var persistentFlagLogfile string
var persistentFlagLoglevel string

var persistenFlagS3Endpoint string
var persistenFlagS3AccessKeyID string
var persistenFlagS3SecretAccessKey string
var persistentFlagS3Region string

//var flagDigest DigestFlag

// var flagExtensionFolder string
// var flagVersion VersionFlag
var flagObjectID string

var flagStatInfo = []string{}

// var flagMessage string
// var flagUserName string
// var flagUserAddress string
// var flagFixity string
// var flagDigestSHA256, flagDigestSHA512 bool

var conf *config.GOCFLConfig

var areaPathRegexp = regexp.MustCompile("^([a-z]{2,}):(.*)$")

var rootCmd = &cobra.Command{
	Use:   "gocfl",
	Short: "gocfl is a fast ocfl creator/extractor/validator with focus on zip containers",
	Long: fmt.Sprintf(`A fast and reliable OCFL creator, extractor and validator.
https://github.com/je4/gocfl
Jürgen Enge (University Library Basel, juergen@info-age.net)
Version %s`, VERSION),
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

func initConfig() {

	// load config file
	if persistentFlagConfigFile != "" {
		data, err := os.ReadFile(persistentFlagConfigFile)
		if err != nil {
			_ = rootCmd.Help()
			log.Errorf("error reading config file %s: %v\n", persistentFlagConfigFile, err)
			os.Exit(1)
		}
		conf, err = config.LoadGOCFLConfig(string(data))
		if err != nil {
			_ = rootCmd.Help()
			log.Errorf("error loading config file %s: %v\n", persistentFlagConfigFile, err)
			os.Exit(1)
		}
	} else {
		var err error
		conf, err = config.LoadGOCFLConfig(string(config.DefaultConfig))
		if err != nil {
			_ = rootCmd.Help()
			log.Errorf("error loading config file %s: %v\n", persistentFlagConfigFile, err)
			os.Exit(1)
		}
	}

	// overwrite config file with command line data
	if persistentFlagLogfile != "" {
		conf.Logfile = persistentFlagLogfile
	}
	if persistentFlagLoglevel != "" {
		conf.Loglevel = persistentFlagLoglevel
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

	rootCmd.PersistentFlags().StringVar(&persistentFlagConfigFile, "config", "", "config file (default is $HOME/.gocfl.toml)")

	rootCmd.PersistentFlags().StringVar(&persistentFlagLogfile, "log-file", "", "log output file (default is console)")
	//emperror.Panic(viper.BindPFlag("LogFile", rootCmd.PersistentFlags().Lookup("log-file")))

	rootCmd.PersistentFlags().StringVar(&persistentFlagLoglevel, "log-level", "ERROR", "log level (CRITICAL|ERROR|WARNING|NOTICE|INFO|DEBUG)")
	//emperror.Panic(viper.BindPFlag("LogLevel", rootCmd.PersistentFlags().Lookup("log-level")))

	rootCmd.PersistentFlags().StringVar(&persistenFlagS3Endpoint, "s3-endpoint", "", "Endpoint for S3 Buckets")
	//emperror.Panic(viper.BindPFlag("S3Endpoint", rootCmd.PersistentFlags().Lookup("s3-endpoint")))

	rootCmd.PersistentFlags().StringVar(&persistenFlagS3AccessKeyID, "s3-access-key-id", "", "Access Key ID for S3 Buckets")
	//emperror.Panic(viper.BindPFlag("S3AccessKeyID", rootCmd.PersistentFlags().Lookup("s3-access-key-id")))

	rootCmd.PersistentFlags().StringVar(&persistenFlagS3SecretAccessKey, "s3-secret-access-key", "", "Secret Access Key for S3 Buckets")
	//emperror.Panic(viper.BindPFlag("S3SecretAccessKey", rootCmd.PersistentFlags().Lookup("s3-secret-access-key")))

	rootCmd.PersistentFlags().StringVar(&persistentFlagS3Region, "s3-region", "", "Region for S3 Access")
	//emperror.Panic(viper.BindPFlag("S3Region", rootCmd.PersistentFlags().Lookup("s3-region")))

	//	rootCmd.PersistentFlags().Bool("with-indexer", false, "starts indexer as a local service")
	//emperror.Panic(viper.BindPFlag("Indexer.Enable", rootCmd.PersistentFlags().Lookup("with-indexer")))

	//	rootCmd.PersistentFlags().StringVar(&flagExtensionFolder, "extensions", "", "folder with default extension configurations")
	//	emperror.Panic(viper.BindPFlag("Extensions", rootCmd.PersistentFlags().Lookup("extensions"))

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

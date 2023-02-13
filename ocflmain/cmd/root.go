package cmd

import (
	"emperror.dev/errors"
	"fmt"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"regexp"
)

const VERSION = "v1.0-beta.1"

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

// var persistentFlagLogfile string
var intPersistentFlagLoglevel LogLevelFlag

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

var areaPathRegexp = regexp.MustCompile("^([a-z]+):([^/].*)$")

var rootCmd = &cobra.Command{
	Use:   "gocfl",
	Short: "gocfl is a fast ocfl creator/extractor/validator with focus on zip containers",
	Long: fmt.Sprintf(`A fast and reliable OCFL creator, extractor and validator.
https://github.com/je4/gocfl
Jürgen Enge (University Library Basel, juergen@info-age.net)
Version %s`, VERSION),
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func initConfig() {
	if persistentFlagConfigFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(persistentFlagConfigFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".gocfl" (without extension).
		viper.AddConfigPath(".")
		viper.AddConfigPath(home)
		viper.SetConfigType("toml")
		viper.SetConfigName(".gocfl")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
		// fmt.Println(viper.AllSettings())
	} else {
		fmt.Printf("error reading config file %s: %v\n", viper.ConfigFileUsed(), err)
	}
	persistentFlagLoglevel := viper.GetInt64("LogLevel")
	if _, ok := LogLevelIds[LogLevelFlag(persistentFlagLoglevel)]; !ok {
		cobra.CheckErr(errors.Errorf("invalid Loglevel ID %v", persistentFlagLoglevel))
	}

}

func setExtensionFlags(commands ...*cobra.Command) {
	extensionParams := GetExtensionParams()
	for _, cmd := range commands {
		for _, param := range extensionParams {
			param.SetParam(cmd)
		}
	}
}

func getExtensionFlags(command *cobra.Command) map[string]string {
	var result = map[string]string{}
	extensionParams := GetExtensionParams()
	for _, param := range extensionParams {
		name, value := param.GetParam(command)
		if name != "" {
			result[name] = value
		}
	}
	return result
}

func init() {

	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&persistentFlagConfigFile, "config", "", "config file (default is $HOME/.gocfl.toml)")

	rootCmd.PersistentFlags().String("log-file", "", "log output file (default is console)")
	viper.BindPFlag("LogFile", rootCmd.PersistentFlags().Lookup("log-file"))

	rootCmd.PersistentFlags().String("log-level", "ERROR", "log level (CRITICAL|ERROR|WARNING|NOTICE|INFO|DEBUG)")
	viper.BindPFlag("LogLevel", rootCmd.PersistentFlags().Lookup("log-level"))

	rootCmd.PersistentFlags().String("s3-endpoint", "", "Endpoint for S3 Buckets")
	viper.BindPFlag("S3Endpoint", rootCmd.PersistentFlags().Lookup("s3-endpoint"))

	rootCmd.PersistentFlags().String("s3-access-key-id", "", "Access Key ID for S3 Buckets")
	viper.BindPFlag("S3AccessKeyID", rootCmd.PersistentFlags().Lookup("s3-access-key-id"))

	rootCmd.PersistentFlags().String("s3-secret-access-key", "", "Secret Access Key for S3 Buckets")
	viper.BindPFlag("S3SecretAccessKey", rootCmd.PersistentFlags().Lookup("s3-secret-access-key"))

	rootCmd.PersistentFlags().String("s3-region", "", "Region for S3 Access")
	viper.BindPFlag("S3Region", rootCmd.PersistentFlags().Lookup("s3-region"))

	createCmd.PersistentFlags().Bool("with-indexer", false, "starts indexer as a local service")
	viper.BindPFlag("Indexer.Local", createCmd.PersistentFlags().Lookup("with-indexer"))

	//	rootCmd.PersistentFlags().StringVar(&flagExtensionFolder, "extensions", "", "folder with default extension configurations")
	//	viper.BindPFlag("Extensions", rootCmd.PersistentFlags().Lookup("extensions"))

	initValidate()
	initInit()
	initCreate()
	initAdd()
	initUpdate()
	initStat()
	initExtract()
	initExtractMeta()

	setExtensionFlags(validateCmd, initCmd, createCmd, addCmd, updateCmd, statCmd, extractCmd, extractMetaCmd)
	rootCmd.AddCommand(validateCmd, initCmd, createCmd, addCmd, updateCmd, statCmd, extractCmd, extractMetaCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

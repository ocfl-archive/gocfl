package cmd

import (
	"emperror.dev/errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/thediveo/enumflag"
	"go.ub.unibas.ch/gocfl/v2/pkg/ocfl"
	"golang.org/x/exp/slices"
	"os"
	"regexp"
)

const LOGFORMAT = `%{time:2006-01-02T15:04:05.000} %{shortpkg}::%{longfunc} [%{shortfile}] > %{level:.5s} - %{message}`

type LogLevelFlag enumflag.Flag

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

type VersionFlag enumflag.Flag

const (
	VERSION1_1 = iota
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

type DigestFlag enumflag.Flag

const (
	DIGESTSHA512 = iota
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
var flagDigest DigestFlag

// var flagExtensionFolder string
// var flagVersion VersionFlag
var flagObjectID string

// var flagMessage string
// var flagUserName string
// var flagUserAddress string
// var flagFixity string
// var flagDigestSHA256, flagDigestSHA512 bool

var areaPathRegexp = regexp.MustCompile("^([a-z]+):([^/].*)$")

var rootCmd = &cobra.Command{
	Use:   "gocfl",
	Short: "gocfl is a fast ocfl creator/validator with focus on zip containers",
	Long: `A fast and reliable OCFL creator and validator created by University Library Basel.
                https://go.ub.unibas.ch/gocfl`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
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
		fmt.Printf("error reading config file %s: %v", viper.ConfigFileUsed(), err)
	}
	persistentFlagLoglevel := viper.GetInt64("LogLevel")
	if _, ok := LogLevelIds[LogLevelFlag(persistentFlagLoglevel)]; !ok {
		cobra.CheckErr(errors.Errorf("invalid Loglevel ID %v", persistentFlagLoglevel))
	}

}

func setExtensionFlags(commands ...*cobra.Command) {
	extensionParams := GetExtensionParams()
	for _, command := range commands {
		for ext, params := range extensionParams {
			for _, param := range params {
				if !slices.Contains(param.Functions, command.Name()) {
					continue
				}
				flagName := fmt.Sprintf("ext-%s-%s", ext, param.Param)
				if command.Flags().Lookup(flagName) == nil {
					command.Flags().String(flagName, param.Default, param.Description)
					if param.File != "" {
						cfgName := fmt.Sprintf("%s.ext.%s.%s", command.Name(), ext, param.File)
						viper.BindPFlag(cfgName, command.Flags().Lookup(flagName))
					}
				}
			}
		}
	}
}

func getExtensionFlags(command *cobra.Command) (map[string]map[string]string, error) {
	var err error
	var result = map[string]map[string]string{}
	extensionParams := GetExtensionParams()
	for ext, params := range extensionParams {
		for _, param := range params {
			if !slices.Contains(param.Functions, command.Name()) {
				continue
			}
			flagName := fmt.Sprintf("ext-%s-%s", ext, param.Param)
			if _, ok := result[ext]; !ok {
				result[ext] = map[string]string{}
			}
			if param.File != "" {
				cfgName := fmt.Sprintf("%s.ext.%s.%s", command.Name(), ext, param.File)
				result[ext][param.Param] = viper.GetString(cfgName)
			} else {
				if result[ext][param.Param], err = command.Flags().GetString(flagName); err != nil {
					return nil, errors.Wrapf(err, "cannot get flag '%s'", flagName)
				}
			}
		}
	}
	return result, nil
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

	//	rootCmd.PersistentFlags().StringVar(&flagExtensionFolder, "extensions", "", "folder with default extension configurations")
	//	viper.BindPFlag("Extensions", rootCmd.PersistentFlags().Lookup("extensions"))

	initValidate()
	initInit()
	initCreate()
	initAdd()
	initUpdate()
	setExtensionFlags(validateCmd, initCmd, createCmd, addCmd, updateCmd)
	rootCmd.AddCommand(validateCmd, initCmd, createCmd, addCmd, updateCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

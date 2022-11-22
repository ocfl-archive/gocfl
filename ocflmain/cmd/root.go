package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/thediveo/enumflag"
	"os"
)

type LogLevel enumflag.Flag

const (
	LOGLEVELDEBUG = iota
	LOGLEVELINFO
	LOGLEVELNOTICE
	LOGLEVELWARNING
	LOGLEVELERROR
	LOGLEVELCRITICAL
)

var LogLevelIds = map[LogLevel][]string{
	LOGLEVELDEBUG:    {"DEBUG"},
	LOGLEVELINFO:     {"INFO"},
	LOGLEVELNOTICE:   {"NOTICE"},
	LOGLEVELWARNING:  {"WARNING"},
	LOGLEVELERROR:    {"ERROR"},
	LOGLEVELCRITICAL: {"CRITICAL"},
}

var logfile string
var loglevel LogLevel
var extensionFolder string

var rootCmd = &cobra.Command{
	Use:   "gocfl",
	Short: "gocfl is a fast ocfl creator/validator with focus on zip containers",
	Long: `A fast and reliable OCFL creator and validator created by University Library Basel.
                https://go.ub.unibas.ch/gocfl`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}

func init() {

	rootCmd.PersistentFlags().StringVar(&logfile, "log-file", "", "log output file (default is console)")
	rootCmd.PersistentFlags().Var(
		enumflag.New(&loglevel, "log-level", LogLevelIds, enumflag.EnumCaseInsensitive),
		"log-level", "log level (CRITICAL|ERROR|WARNING|NOTICE|INFO|DEBUG)")
	rootCmd.PersistentFlags().StringVar(&extensionFolder, "extensions", "", "folder with default extension configurations")

	rootCmd.AddCommand(validateCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

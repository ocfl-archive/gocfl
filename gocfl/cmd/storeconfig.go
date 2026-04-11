package cmd

import (
	"context"
	"crypto/tls"
	"io"
	"log"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/util"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfllogger"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/cobra"
	ublogger "gitlab.switch.ch/ub-unibas/go-ublogger/v2"
	"go.ub.unibas.ch/cloud/certloader/v2/pkg/loader"
)

var storeConfigCmd = &cobra.Command{
	Use:     "storeconfig [path to config file]",
	Aliases: []string{},
	Short:   "store configuration of gocfl in toml format",
	//Long:    "an utterly useless command for testing",
	Example: "gocfl storeconfig ./gocfl.toml",
	Args:    cobra.ExactArgs(1),
	Run:     doStoreConfig,
}

func initStoreConfig() {
}

func doStoreConfigConf(cmd *cobra.Command) {
}

func doStoreConfig(cmd *cobra.Command, args []string) {
	configPath, err := util.Fullpath(args[0])
	if err != nil {
		cobra.CheckErr(err)
		return
	}

	// create logger instance
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("cannot get hostname: %v", err)
	}

	var loggerTLSConfig *tls.Config
	var loggerLoader io.Closer
	if conf.Log.Stash.TLS != nil {
		loggerTLSConfig, loggerLoader, err = loader.CreateClientLoader(conf.Log.Stash.TLS, nil)
		if err != nil {
			log.Fatalf("cannot create client loader: %v", err)
		}
		defer loggerLoader.Close()
	}

	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	_logger, _logstash, _logfile, err := ublogger.CreateUbMultiLoggerTLS(conf.Log.Level, conf.Log.File,
		ublogger.SetDataset(conf.Log.Stash.Dataset),
		ublogger.SetLogStash(conf.Log.Stash.LogstashHost, conf.Log.Stash.LogstashPort, conf.Log.Stash.Namespace, conf.Log.Stash.LogstashTraceLevel),
		ublogger.SetTLS(conf.Log.Stash.TLS != nil),
		ublogger.SetTLSConfig(loggerTLSConfig),
	)
	if err != nil {
		log.Fatalf("cannot create logger: %v", err)
	}
	if _logstash != nil {
		defer _logstash.Close()
	}

	if _logfile != nil {
		defer _logfile.Close()
	}

	l2 := _logger.With().Timestamp().Str("host", hostname).Logger() //.Output(output)
	ctx := context.TODO()
	var logger = ocfllogger.NewOCFLLogger(ctx, &l2, nil, version.Default, nil)

	doStoreConfigConf(cmd)

	fp, err := os.Create(configPath)
	if err != nil {
		log.Fatalf("cannot create config file: %v", err)
	}
	defer func(fp *os.File) {
		err := fp.Close()
		if err != nil {
			logger.Error().Msgf("cannot close config file: %v", err)
		}
	}(fp)
	tenc := toml.NewEncoder(fp)
	tenc.Indent = "  "
	if err := tenc.Encode(conf); err != nil {
		logger.Error().Msgf("cannot encode config: %v", err)
	}
}

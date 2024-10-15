package cmd

import (
	"context"
	"crypto/tls"
	"emperror.dev/emperror"
	"emperror.dev/errors"
	"fmt"
	"github.com/je4/certloader/v2/pkg/loader"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/cobra"
	ublogger "gitlab.switch.ch/ub-unibas/go-ublogger/v2"
	"io"
	"log"
	"os"
	"strings"
)

var statCmd = &cobra.Command{
	Use:     "stat [path to ocfl structure]",
	Aliases: []string{"info"},
	Short:   "statistics of an ocfl structure",
	//Long:    "an utterly useless command for testing",
	Example: "gocfl stat ./archive.zip",
	Args:    cobra.MinimumNArgs(1),
	Run:     doStat,
}

func initStat() {
	statCmd.Flags().StringP("object-path", "p", "", "object path to show statistics for")
	statCmd.Flags().StringP("object-id", "i", "", "object id to show statistics for")

	infos := []string{}
	for info, _ := range ocfl.StatInfoString {
		infos = append(infos, info)
	}
	statCmd.Flags().String("stat-info", "", fmt.Sprintf("comma separated list of info fields to show [%s]", strings.Join(infos, ",")))
}

func doStatConf(cmd *cobra.Command) {
	if str := getFlagString(cmd, "object-path"); str != "" {
		conf.Stat.ObjectPath = str
	}
	if str := getFlagString(cmd, "object-id"); str != "" {
		conf.Stat.ObjectID = str
	}
	if str := getFlagString(cmd, "stat-info"); str != "" {
		conf.Stat.Info = []string{}
		for _, s := range strings.Split(str, ",") {
			conf.Stat.Info = append(conf.Stat.Info, strings.ToLower(strings.TrimSpace(s)))
		}
	}
}

func doStat(cmd *cobra.Command, args []string) {
	ocflPath, err := ocfl.Fullpath(args[0])
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
	var logger zLogger.ZLogger = &l2

	doStatConf(cmd)

	oPath := conf.Stat.ObjectPath
	oID := conf.Stat.ObjectID
	if oPath != "" && oID != "" {
		emperror.Panic(cmd.Help())
		cobra.CheckErr(errors.New("do not use object-path AND object-id at the same time"))
		return
	}

	statInfo := []ocfl.StatInfo{}
	for _, statInfoString := range conf.Stat.Info {
		statInfoString = strings.ToLower(strings.TrimSpace(statInfoString))
		var found bool
		for str, info := range ocfl.StatInfoString {
			if strings.ToLower(str) == statInfoString {
				found = true
				statInfo = append(statInfo, info)
			}
		}
		if !found {
			emperror.Panic(cmd.Help())
			cobra.CheckErr(errors.Errorf("--stat-info invalid value '%s' ", statInfoString))
		}
	}

	t := startTimer()
	defer func() { logger.Info().Msgf("Duration: %s", t.String()) }()

	logger.Info().Msgf("opening '%s'", ocflPath)

	fsFactory, err := initializeFSFactory(nil, nil, nil, true, false, logger)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("cannot create filesystem factory")
		return
	}

	destFS, err := fsFactory.Get(ocflPath)
	if err != nil {
		logger.Error().Stack().Err(err).Msgf("cannot get filesystem for '%s'", ocflPath)
		return
	}
	defer func() {
		if err := writefs.Close(destFS); err != nil {
			logger.Error().Stack().Err(err).Msgf("cannot close filesystem '%s'", destFS)
		}
	}()

	extensionParams := GetExtensionParamValues(cmd, conf)
	extensionFactory, err := InitExtensionFactory(extensionParams, "", false, nil, nil, nil, nil, (logger))
	if err != nil {
		logger.Error().Stack().Err(err).Msgf("cannot initialize extension factory")
		return
	}

	ctx := ocfl.NewContextValidation(context.TODO())
	if !writefs.HasContent(destFS) {

	}
	storageRoot, err := ocfl.LoadStorageRoot(ctx, destFS, extensionFactory, (logger))
	if err != nil {
		logger.Error().Stack().Err(err).Msg("cannot load storage root")
		return
	}

	if err := storageRoot.Stat(os.Stdout, oPath, oID, statInfo); err != nil {
		logger.Error().Stack().Err(err).Msg("cannot get statistics")
		return
	}
	_ = showStatus(ctx, logger)
}

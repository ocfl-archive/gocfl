package cmd

import (
	"context"
	"crypto/tls"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/cobra"
	ublogger "gitlab.switch.ch/ub-unibas/go-ublogger/v2"
	"go.ub.unibas.ch/cloud/certloader/v2/pkg/loader"
	"io"
	"log"
	"os"
)

var validateCmd = &cobra.Command{
	Use:     "validate [path to ocfl structure]",
	Aliases: []string{"check"},
	Short:   "validates an ocfl structure",
	//Long:    "an utterly useless command for testing",
	Example: "gocfl validate ./archive.zip",
	Args:    cobra.ExactArgs(1),
	Run:     validate,
}

func initValidate() {
	validateCmd.Flags().StringP("object-path", "o", "", "validate only the object at the specified path in storage root")
	validateCmd.Flags().String("object-id", "", "validate only the object with the specified id in storage root")
}

func doValidateConf(cmd *cobra.Command) {
	if str := getFlagString(cmd, "object-path"); str != "" {
		conf.Validate.ObjectPath = str
	}
	if str := getFlagString(cmd, "object-id"); str != "" {
		conf.Validate.ObjectID = str
	}
}

func validate(cmd *cobra.Command, args []string) {
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

	t := startTimer()
	defer func() { logger.Info().Msgf("Duration: %s", t.String()) }()

	doValidateConf(cmd)

	logger.Info().Msgf("validating '%s'", ocflPath)

	extensionParams := GetExtensionParamValues(cmd, conf)
	extensionFactory, err := InitExtensionFactory(extensionParams, "", false, nil, nil, nil, nil, logger, conf.TempDir)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("cannot initialize extension factory")
		return
	}

	fsFactory, err := initializeFSFactory(nil, nil, nil, true, true, logger)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("cannot create filesystem factory")
		return
	}

	destFS, err := fsFactory.Get(ocflPath, true)
	if err != nil {
		logger.Error().Stack().Err(err).Msgf("cannot get filesystem for '%s'", ocflPath)
		return
	}
	defer func() {
		if err := writefs.Close(destFS); err != nil {
			logger.Error().Stack().Err(err).Msgf("cannot close filesystem for '%s'", destFS)
		}
	}()

	ctx := ocfl.NewContextValidation(context.TODO())
	storageRoot, err := ocfl.LoadStorageRoot(ctx, destFS, extensionFactory, logger, ErrorFactory, conf.Init.Documentation)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("cannot load storageroot")
		return
	}
	objectID := conf.Validate.ObjectID
	objectPath := conf.Validate.ObjectPath
	if objectID != "" && objectPath != "" {
		logger.Error().Msg("do not use object-path AND object-id at the same time")
		return
	}
	if objectID == "" && objectPath == "" {
		if err := storageRoot.Check(); err != nil {
			logger.Error().Stack().Err(err).Msg("ocfl not valid")
			return
		}
	} else {
		if objectID != "" {
			if err := storageRoot.CheckObjectByID(objectID); err != nil {
				logger.Error().Stack().Err(err).Msgf("ocfl object '%s' not valid", objectID)
				return
			}
		} else {
			if err := storageRoot.CheckObjectByFolder(objectPath); err != nil {
				logger.Error().Stack().Err(err).Msgf("ocfl object '%s' not valid", objectPath)
				return
			}
		}
	}
	_ = showStatus(ctx, logger)
}

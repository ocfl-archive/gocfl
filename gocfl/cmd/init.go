package cmd

import (
	"context"
	"crypto/tls"
	"emperror.dev/errors"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	"github.com/je4/trustutil/v2/pkg/loader"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/cobra"
	ublogger "gitlab.switch.ch/ub-unibas/go-ublogger"
	"io"
	"log"
	"os"
)

var initCmd = &cobra.Command{
	Use:     "init [path to ocfl structure]",
	Aliases: []string{},
	Short:   "initializes an empty ocfl structure",
	Long:    "initializes an empty ocfl structure",
	Example: "gocfl init ./archive.zip",
	Args:    cobra.ExactArgs(1),
	Run:     doInit,
}

func initInit() {
	initCmd.Flags().String("default-storageroot-extensions", "", "folder with initial extension configurations for new OCFL Storage Root")
	initCmd.Flags().String("ocfl-version", "", "ocfl version for new storage root")
	initCmd.Flags().StringP("digest", "d", "", "digest to use for ocfl checksum")
	initCmd.Flags().Bool("no-compress", false, "do not compress data in zip file")
}

func doInitConf(cmd *cobra.Command) {
	if str := getFlagString(cmd, "default-storageroot-extensions"); str != "" {
		conf.Init.StorageRootExtensionFolder = str
	}

	if str := getFlagString(cmd, "ocfl-version"); str != "" {
		conf.Init.OCFLVersion = str
	}

	if str := getFlagString(cmd, "digest"); str != "" {
		conf.Init.Digest = checksum.DigestAlgorithm(str)
	}
	if _, err := checksum.GetHash(conf.Init.Digest); err != nil {
		_ = cmd.Help()
		cobra.CheckErr(errors.Errorf("invalid digest '%s' for flag 'digest' or 'Init.DigestAlgorithm' config file entry", conf.Init.Digest))
	}

}

func doInit(cmd *cobra.Command, args []string) {
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
	_logger, _logstash, _logfile := ublogger.CreateUbMultiLoggerTLS(conf.Log.Level, conf.Log.File,
		ublogger.SetDataset(conf.Log.Stash.Dataset),
		ublogger.SetLogStash(conf.Log.Stash.LogstashHost, conf.Log.Stash.LogstashPort, conf.Log.Stash.Namespace, conf.Log.Stash.LogstashTraceLevel),
		ublogger.SetTLS(conf.Log.Stash.TLS != nil),
		ublogger.SetTLSConfig(loggerTLSConfig),
	)
	if _logstash != nil {
		defer _logstash.Close()
	}

	if _logfile != nil {
		defer _logfile.Close()
	}

	l2 := _logger.With().Timestamp().Str("host", hostname).Logger() //.Output(output)
	var logger zLogger.ZLogger = &l2

	doInitConf(cmd)

	logger.Info().Msgf("creating '%s'", ocflPath)
	t := startTimer()
	defer func() { logger.Info().Msgf("Duration: %s", t.String()) }()

	fsFactory, err := initializeFSFactory([]checksum.DigestAlgorithm{conf.Init.Digest}, conf.AES, conf.S3, true, false, logger)
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
	extensionFactory, err := InitExtensionFactory(
		extensionParams,
		"",
		false,
		nil,
		nil,
		nil,
		nil,
		(logger),
	)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("cannot create extension factory")
		return
	}
	storageRootExtensions, _, err := initDefaultExtensions(
		extensionFactory,
		conf.Init.StorageRootExtensionFolder,
		"",
		logger,
	)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("cannot initialize default extensions")
		return
	}

	ctx := ocfl.NewContextValidation(context.TODO())
	if _, err := ocfl.CreateStorageRoot(
		ctx,
		destFS,
		ocfl.OCFLVersion(conf.Init.OCFLVersion),
		extensionFactory, storageRootExtensions,
		conf.Init.Digest,
		(logger),
	); err != nil {
		if err := writefs.Close(destFS); err != nil {
			logger.Error().Stack().Err(err).Msgf("cannot close filesystem '%s'", destFS)
		}
		logger.Error().Stack().Err(err).Msgf("cannot create new storageroot")
		return
	}

	_ = showStatus(ctx, logger)
}

package cmd

import (
	"context"
	"crypto/tls"
	"emperror.dev/errors"
	"fmt"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	"github.com/je4/trustutil/v2/pkg/loader"
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/cobra"
	ublogger "gitlab.switch.ch/ub-unibas/go-ublogger"
	"io"
	"io/fs"
	"log"
	"os"
)

var extractCmd = &cobra.Command{
	Use:     "extract [path to ocfl structure] [path to target folder]",
	Aliases: []string{},
	Short:   "extract version of ocfl content",
	//Long:    "an utterly useless command for testing",
	Example: "gocfl extract ./archive.zip /tmp/archive",
	Args:    cobra.MinimumNArgs(2),
	Run:     doExtract,
}

func initExtract() {
	extractCmd.Flags().StringP("object-path", "p", "", "object path to extract")
	extractCmd.Flags().StringP("object-id", "i", "", "object id to extract")
	extractCmd.Flags().Bool("with-manifest", false, "generate manifest file in object extraction folder")
	extractCmd.Flags().String("version", "", "version to extract")
	extractCmd.Flags().String("area", "content", "data area to extract")
}
func doExtractConf(cmd *cobra.Command) {
	if str := getFlagString(cmd, "object-path"); str != "" {
		conf.Extract.ObjectPath = str
	}
	if str := getFlagString(cmd, "object-id"); str != "" {
		conf.Extract.ObjectID = str
	}
	if b := getFlagBool(cmd, "with-manifest"); b {
		conf.Extract.Manifest = b
	}
	if str := getFlagString(cmd, "version"); str != "" {
		conf.Extract.Version = str
	}
	if str := getFlagString(cmd, "area"); str != "" {
		conf.Extract.Area = str
	}
	if conf.Extract.Version == "" {
		conf.Extract.Version = "latest"
	}
}

func doExtract(cmd *cobra.Command, args []string) {
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

	t := startTimer()
	defer func() { logger.Info().Msgf("Duration: %s", t.String()) }()

	ocflPath, err := ocfl.Fullpath(args[0])
	if err != nil {
		cobra.CheckErr(err)
		return
	}
	destPath, err := ocfl.Fullpath(args[1])
	if err != nil {
		cobra.CheckErr(err)
		return
	}

	doExtractConf(cmd)

	oPath := conf.Extract.ObjectPath
	oID := conf.Extract.ObjectID
	if oPath != "" && oID != "" {
		cmd.Help()
		cobra.CheckErr(errors.New("do not use object-path AND object-id at the same time"))
		return
	}

	logger.Info().Msgf("extracting '%s'", ocflPath)

	fsFactory, err := initializeFSFactory(nil, nil, nil, true, true, logger)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("cannot create filesystem factory")
		return
	}

	ocflFS, err := fsFactory.Get(ocflPath)
	if err != nil {
		logger.Error().Stack().Err(err).Msgf("cannot get filesystem for '%s'", ocflPath)
		return
	}

	destFS, err := fsFactory.Get(destPath)
	if err != nil {
		logger.Error().Stack().Err(err).Msgf("cannot get filesystem for '%s'", destPath)
		return
	}
	defer func() {
		if err := writefs.Close(destFS); err != nil {
			logger.Error().Err(err).Msgf("cannot close filesystem: %v", destFS)
		}
	}()

	extensionParams := GetExtensionParamValues(cmd, conf)
	extensionFactory, err := InitExtensionFactory(extensionParams, "", false, nil, nil, nil, nil, (logger))
	if err != nil {
		logger.Error().Stack().Err(err).Msgf("cannot initialize extension factory")
		return
	}

	ctx := ocfl.NewContextValidation(context.TODO())
	storageRoot, err := ocfl.LoadStorageRoot(ctx, ocflFS, extensionFactory, (logger))
	if err != nil {
		logger.Error().Stack().Err(err).Msg("cannot load storage root")
		return
	}

	dirs, err := fs.ReadDir(destFS, ".")
	if err != nil {
		logger.Error().Stack().Err(err).Msgf("cannot read target folder '%v'", destFS)
		return
	}
	if len(dirs) > 0 {
		fmt.Printf("target folder '%s' is not empty\n", destFS)
		logger.Debug().Msgf("target folder '%s' is not empty", destFS)
		return
	}

	if err := storageRoot.Extract(destFS, oPath, oID, conf.Extract.Version, conf.Extract.Manifest, conf.Extract.Area); err != nil {
		fmt.Printf("cannot extract storage root: %v\n", err)
		logger.Error().Stack().Err(err).Msg("cannot extract storage root")
		return
	}
	fmt.Printf("extraction done without errors\n")
	_ = showStatus(ctx, logger)
}

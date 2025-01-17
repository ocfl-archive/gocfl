package cmd

import (
	"context"
	"crypto/tls"
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
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
	"strings"
)

var extractMetaCmd = &cobra.Command{
	Use:     "extractmeta [path to ocfl structure]",
	Aliases: []string{},
	Short:   "extract metadata from ocfl structure",
	//Long:    "an utterly useless command for testing",
	Example: "gocfl extractmeta ./archive.zip --output-json ./archive_meta.json",
	Args:    cobra.ExactArgs(1),
	Run:     doExtractMeta,
}

func initExtractMeta() {
	extractMetaCmd.Flags().StringP("object-path", "p", "", "object path to extract")
	extractMetaCmd.Flags().StringP("object-id", "i", "", "object id to extract")
	extractMetaCmd.Flags().String("version", "latest", "version to extract")
	extractMetaCmd.Flags().String("format", "json", "output format (json)")
	extractMetaCmd.Flags().String("output", "", "output file (default stdout)")
}

func doExtractMetaConf(cmd *cobra.Command) {
	if str := getFlagString(cmd, "object-path"); str != "" {
		conf.ExtractMeta.ObjectPath = str
	}
	if str := getFlagString(cmd, "object-id"); str != "" {
		conf.ExtractMeta.ObjectID = str
	}
	if str := getFlagString(cmd, "version"); str != "" {
		conf.ExtractMeta.Version = str
	}
	if conf.ExtractMeta.Version == "" {
		conf.ExtractMeta.Version = "latest"
	}
	if str := getFlagString(cmd, "format"); str != "" {
		conf.ExtractMeta.Format = str
	}
	if str := getFlagString(cmd, "output"); str != "" {
		conf.ExtractMeta.Output = str
	}
}

func doExtractMeta(cmd *cobra.Command, args []string) {
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

	doExtractMetaConf(cmd)

	oPath := conf.ExtractMeta.ObjectPath
	oID := conf.ExtractMeta.ObjectID
	if oPath != "" && oID != "" {
		cmd.Help()
		cobra.CheckErr(errors.New("do not use object-path AND object-id at the same time"))
		return
	}
	format := strings.ToLower(conf.ExtractMeta.Format)
	if format != "json" {
		cmd.Help()
		cobra.CheckErr(errors.Errorf("invalid format '%s' for flag 'format' or 'Format' config file entry", format))
		return
	}
	output := conf.ExtractMeta.Output

	logger.Info().Msgf("extracting metadata from '%s'", ocflPath)

	fsFactory, err := initializeFSFactory(nil, nil, nil, true, true, logger)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("cannot create filesystem factory")
		return
	}

	ocflFS, err := fsFactory.Get(ocflPath, true)
	if err != nil {
		logger.Error().Stack().Err(err).Msgf("cannot get filesystem for '%s'", ocflPath)
		return
	}
	defer func() {
		if err := writefs.Close(ocflFS); err != nil {
			logger.Error().Stack().Err(err).Msgf("cannot close filesystem for '%s'", ocflFS)
		}
	}()

	extensionParams := GetExtensionParamValues(cmd, conf)
	extensionFactory, err := InitExtensionFactory(extensionParams, "", false, nil, nil, nil, nil, (logger))
	if err != nil {
		logger.Error().Stack().Err(err).Msg("cannot initialize extension factory")
		return
	}

	ctx := ocfl.NewContextValidation(context.TODO())
	storageRoot, err := ocfl.LoadStorageRoot(ctx, ocflFS, extensionFactory, (logger))
	if err != nil {
		logger.Error().Stack().Err(err).Msg("cannot load storage root")
		return
	}

	metadata, err := storageRoot.ExtractMeta(oPath, oID)
	if err != nil {
		fmt.Printf("cannot extract metadata from storage root: %v\n", err)
		logger.Error().Stack().Err(err).Msg("cannot extract metadata from storage root")
		return
	}

	jsonBytes, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		fmt.Printf("cannot marshal metadata")
		logger.Error().Stack().Err(err).Msg("cannot marshal metadata")
		return
	}
	if output != "" {
		if err := os.WriteFile(output, jsonBytes, 0644); err != nil {
			fmt.Printf("cannot write json to file")
			logger.Error().Stack().Err(err).Msgf("cannot write json to file '%s'", output)
			return
		}
	} else {
		if _, err := os.Stdout.Write(jsonBytes); err != nil {
			fmt.Printf("cannot write json to file")
			logger.Error().Stack().Err(err).Msg("cannot write json to file standard output")
			return
		}
		fmt.Print("\n")
	}
	fmt.Printf("metadata extraction done without errors\n")
	_ = showStatus(ctx, logger)
}

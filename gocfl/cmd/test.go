package cmd

import (
	"context"
	"crypto/tls"
	"io"
	"log"
	"os"

	"github.com/je4/filesystem/v3/pkg/vfsrw"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/ocfl-archive/gocfl/v2/internal"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension/extensionimpl"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/functions"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/util"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfllogger"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/cobra"
	ublogger "gitlab.switch.ch/ub-unibas/go-ublogger/v2"
	"go.ub.unibas.ch/cloud/certloader/v2/pkg/loader"
)

var testCmd = &cobra.Command{
	Use:     "test [path to ocfl structure]",
	Aliases: []string{},
	Short:   "test for object without objectroot",
	Long:    "an utterly useless command for testing",
	Example: "gocfl test ./archive.zip/<path to ocfl object>",
	Args:    cobra.ExactArgs(1),
	Run:     doTest,
}

func initTest() {
}

func doTestConf(cmd *cobra.Command) {
}

func doTest(cmd *cobra.Command, args []string) {
	ocflObjectPath, err := util.Fullpath(args[0])
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
	ctx := validation.NewContextValidation(context.TODO())
	var logger = ocfllogger.NewOCFLLogger(ctx, &l2, nil, version.Default)

	doTestConf(cmd)

	if conf.VFS == nil {
		conf.VFS = vfsrw.Config{}
	}
	for name, val := range getLocalFSConfig() {
		conf.VFS[name] = val
	}
	vfs, err := vfsrw.NewFS(conf.VFS, logger.Logger())
	if err != nil {
		logger.Panic().Err(err).Msg("cannot create vfs")
	}
	defer func() {
		if err := vfs.Close(); err != nil {
			logger.Error().Err(err).Msg("cannot close vfs")
		}
	}()
	vfs.AddFS("internal", internal.InternalFS)

	ocflObjectPath, err = path2vfs(ocflObjectPath)
	if err != nil {
		logger.Error().Err(err).Msg("cannot create ocfl path")
		return
	}
	logger.Info().Msgf("vfs created : %v", vfs)

	t := startTimer()
	defer func() { logger.Info().Msgf("Duration: %s", t.String()) }()

	logger.Info().Msgf("opening '%s'", ocflObjectPath)

	extensionParams, err := getExtensionParams(cmd)
	if err != nil {
		logger.Error().Err(err).Msg("cannot get extension params")
		return
	}

	objFsys, err := writefs.Sub(vfs, ocflObjectPath)
	if err != nil {
		logger.Error().Err(err).Msgf("cannot open ocfl filesystem at '%s'", ocflObjectPath)
		return
	}

	extensionFactory, err := extensionimpl.NewFactory(extensionParams, logger)
	if err != nil {
		logger.Error().Err(err).Msg("cannot create extension factory")
		return
	}

	obj, err := functions.LoadObject(ctx, objFsys, extensionFactory, logger)
	if err != nil {
		logger.Error().Err(err).Msgf("cannot load object '%v'", objFsys)
		return
	}

	checker := obj.GetChecker(objFsys)
	if err := checker.Check(); err != nil {
		logger.Error().Err(err).Msgf("cannot stat object '%v'", objFsys)
		return
	}

	_ = showStatus(ctx, logger)
}

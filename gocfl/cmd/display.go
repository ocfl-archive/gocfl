package cmd

import (
	"context"
	"crypto/tls"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v2/data/displaydata"
	"github.com/ocfl-archive/gocfl/v2/gocfl/cmd/display"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/cobra"
	ublogger "gitlab.switch.ch/ub-unibas/go-ublogger/v2"
	"go.ub.unibas.ch/cloud/certloader/v2/pkg/loader"
	"io"
	"io/fs"
	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var displayCmd = &cobra.Command{
	Use:     "display [path to ocfl structure]",
	Aliases: []string{"viewer"},
	Short:   "show content of ocfl object in webbrowser",
	//Long:    "an utterly useless command for testing",
	Example: "gocfl display ./archive.zip",
	Args:    cobra.MinimumNArgs(1),
	Run:     doDisplay,
}

/*
[Display]
# --display-addr
Addr = "localhost:8080"
# --display-external-addr
ExternalAddr = "http://localhost:8080"
# --display-templates
Templates = "./data/displaydata/templates"
*/

func initDisplay() {
	displayCmd.Flags().StringP("display-addr", "a", "localhost:8080", "address to listen on")
	displayCmd.Flags().StringP("display-external-addr", "e", "http://localhost:8080", "external address to access the server")
	displayCmd.Flags().StringP("display-templates", "t", "", "path to templates")
	displayCmd.Flags().StringP("display-tls-cert", "c", "", "path to tls certificate")
	displayCmd.Flags().StringP("display-tls-key", "k", "", "path to tls certificate key")
}

func doDisplayConf(cmd *cobra.Command) {
	if str := getFlagString(cmd, "display-addr"); str != "" {
		conf.Display.Addr = str
	}
	if str := getFlagString(cmd, "display-external-addr"); str != "" {
		conf.Display.AddrExt = str
	}
	if str := getFlagString(cmd, "display-templates"); str != "" {
		conf.Display.Templates = str
	}
	if str := getFlagString(cmd, "display-tls-cert"); str != "" {
		conf.Display.CertFile = str
	}
	if str := getFlagString(cmd, "display-tls-key"); str != "" {
		conf.Display.KeyFile = str
	}
}

func doDisplay(cmd *cobra.Command, args []string) {
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

	doDisplayConf(cmd)

	logger.Info().Msgf("opening '%s'", ocflPath)

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

	urlC, _ := url.Parse(conf.Display.AddrExt)
	var templateFS fs.FS
	if conf.Display.Templates == "" {
		templateFS, err = writefs.Sub(displaydata.TemplateRoot, "templates")
		if err != nil {
			logger.Error().Stack().Err(err).Msg("cannot get templates")
			return
		}
	} else {
		templateFS = os.DirFS(conf.Display.Templates)
	}
	srv, err := display.NewServer(storageRoot, "gocfl", conf.Display.Addr, urlC, displaydata.WebRoot, templateFS, (logger), io.Discard)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("cannot create server")
		return
	}

	go func() {
		if err := srv.ListenAndServe("", ""); err != nil {
			logger.Error().Stack().Err(err).Msgf("cannot start server")
			return
		}
	}()

	end := make(chan bool, 1)

	// process waiting for interrupt signal (TERM or KILL)
	go func() {
		sigint := make(chan os.Signal, 1)

		// interrupt signal sent from terminal
		signal.Notify(sigint, os.Interrupt)

		signal.Notify(sigint, syscall.SIGTERM)
		signal.Notify(sigint, syscall.SIGKILL)

		<-sigint

		// We received an interrupt signal, shut down.
		logger.Info().Msg("interrupt signal received")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		srv.Shutdown(ctx)

		end <- true
	}()

	<-end
	logger.Info().Msg("server stopped")

}

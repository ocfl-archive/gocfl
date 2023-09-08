package cmd

import (
	"context"
	"github.com/je4/filesystem/v2/pkg/writefs"
	"github.com/je4/gocfl/v2/data/displaydata"
	"github.com/je4/gocfl/v2/gocfl/cmd/display"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	lm "github.com/je4/utils/v2/pkg/logger"
	"github.com/spf13/cobra"
	"io"
	"io/fs"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
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
	ocflPath := filepath.ToSlash(args[0])

	daLogger, lf := lm.CreateLogger("ocfl", persistentFlagLogfile, nil, conf.LogLevel, conf.LogFormat)
	defer lf.Close()
	t := startTimer()
	defer func() { daLogger.Infof("Duration: %s", t.String()) }()

	doDisplayConf(cmd)

	daLogger.Infof("opening '%s'", ocflPath)

	fsFactory, err := initializeFSFactory(nil, nil, nil, true, true, daLogger)
	if err != nil {
		daLogger.Errorf("cannot create filesystem factory: %v", err)
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	destFS, err := fsFactory.Get(ocflPath)
	if err != nil {
		daLogger.Errorf("cannot get filesystem for '%s': %v", ocflPath, err)
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}
	defer func() {
		if err := writefs.Close(destFS); err != nil {
			daLogger.Errorf("cannot close filesystem: %v", err)
			daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		}
	}()

	extensionParams := GetExtensionParamValues(cmd, conf)
	extensionFactory, err := initExtensionFactory(extensionParams, "", false, nil, nil, nil, nil, daLogger)
	if err != nil {
		daLogger.Errorf("cannot initialize extension factory: %v", err)
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	ctx := ocfl.NewContextValidation(context.TODO())
	if !writefs.HasContent(destFS) {

	}
	storageRoot, err := ocfl.LoadStorageRoot(ctx, destFS, extensionFactory, daLogger)
	if err != nil {
		daLogger.Errorf("cannot open storage root: %v", err)
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	urlC, _ := url.Parse(conf.Display.AddrExt)
	var templateFS fs.FS
	if conf.Display.Templates == "" {
		templateFS, err = fs.Sub(displaydata.TemplateRoot, "templates")
		if err != nil {
			daLogger.Errorf("cannot get templates: %v", err)
			daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
			return
		}
	} else {
		templateFS = os.DirFS(conf.Display.Templates)
	}
	srv, err := display.NewServer(storageRoot, "gocfl", conf.Display.Addr, urlC, displaydata.WebRoot, templateFS, daLogger, io.Discard)
	if err != nil {
		daLogger.Errorf("cannot create server: %v", err)
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	go func() {
		if err := srv.ListenAndServe("", ""); err != nil {
			daLogger.Errorf("cannot start server: %v", err)
			daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
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
		daLogger.Infof("shutdown requested")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		srv.Shutdown(ctx)

		end <- true
	}()

	<-end
	daLogger.Info("server stopped")

}

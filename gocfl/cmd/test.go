package cmd

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path"
	"regexp"

	"emperror.dev/errors"
	"github.com/je4/filesystem/v3/pkg/vfsrw"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/ocfl-archive/gocfl/v2/internal"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension/extensionimpl"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/functions"
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
	Use:     "test [path to folder with test fixtures]",
	Aliases: []string{"fixtures"},
	Short:   "check ocfl fixtures",
	Long:    "check gocfl against folder with test fixtures. Every folder contains one fixture object. If folder name starts with validation codes it's checked, whether they are found.",
	Example: "gocfl test <path to ocfl test fixtures>",
	Args:    cobra.MaximumNArgs(1),
	Run:     doTest,
}

func initTest() {
}

func doTestConf(cmd *cobra.Command) {
}

func doTest(cmd *cobra.Command, args []string) {
	if len(args) > 0 && len(args[0]) > 0 {
		conf.Test.FixturePath = args[0]
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

	fixturePath := conf.Test.FixturePath

	fixturePath, err = path2vfs(fixturePath)
	if err != nil {
		logger.Error().Err(err).Msg("cannot create ocfl path")
		return
	}
	logger.Info().Msgf("vfs created : %v", vfs)

	t := startTimer()
	defer func() { logger.Info().Msgf("Duration: %s", t.String()) }()

	logger.Info().Msgf("opening '%s'", fixturePath)

	extensionParams, err := getExtensionParams(cmd)
	if err != nil {
		logger.Error().Err(err).Msg("cannot get extension params")
		return
	}

	dirs, err := fs.ReadDir(vfs, fixturePath)
	if err != nil {
		logger.Error().Err(err).Msgf("cannot read dir '%s'", fixturePath)
		return
	}
	for _, dir := range dirs {
		folderName := dir.Name()
		logger.Info().Msgf("dir: %s", folderName)
		if err := func() error {

			objFsys, err := writefs.Sub(vfs, path.Join(fixturePath, folderName))
			if err != nil {
				return errors.Wrapf(err, "cannot open ocfl filesystem '%s'", fixturePath)
			}

			extensionFactory, err := extensionimpl.NewFactory(extensionParams, logger)
			if err != nil {
				return errors.Wrapf(err, "cannot create extension factory '%s'", fixturePath)
			}

			obj, err := functions.LoadObject(ctx, objFsys, extensionFactory, logger)
			if err != nil {
				return errors.Wrapf(err, "cannot load object '%v'", objFsys)
			}

			checker := obj.GetChecker(objFsys)
			if err := checker.Check(); err != nil {
				return errors.Wrapf(err, "cannot check object '%v'", objFsys)
			}
			return nil
		}(); err != nil {
			logger.Error().Err(err).Msgf("cannot validate object '%v'", folderName)
		}
		status, err := validation.GetValidationStatus(ctx)
		if err != nil {
			logger.Error().Err(err).Msg("cannot get validation status")
			continue
		}
		status.Compact()
		contextString := ""
		errs := 0
		for _, err := range status.Errors {
			if err.Code[0] == 'E' {
				errs++
			}
			if err.Context != contextString {
				fmt.Printf("[%s] [%s]\n", folderName, err.Context)
				contextString = err.Context
			}
			fmt.Printf("   #%s - %s\n", err.Code, err.Description)
		}
		if errs > 0 {
			fmt.Printf("\n%d errors found\n", errs)
		} else {
			fmt.Printf("\nno errors found\n")
		}
		errorList := errorsFromFolder(folderName)
		errorNotFound := []string{}
		for _, errNo := range errorList {
			var found = false
			var allErrors = map[validation.ErrorCode]string{}
			for _, err := range status.Errors {
				allErrors[err.Code] = err.Description
			}
			for code, desc := range allErrors {
				if code == validation.ErrorCode(errNo) {
					fmt.Printf("Error found:   #%s - %s\n", code, desc)
					found = true
					continue
				}
			}
			if !found {
				errorNotFound = append(errorNotFound, errNo)
			}
		}
		if len(errorNotFound) > 0 {
			fmt.Printf("[%s] Errors not found: %v\n", folderName, errorNotFound)
		} else if len(errorList) == 0 && len(status.Errors) > 0 {
			fmt.Printf("[%s] Errors found, but object should be valid\n", folderName)
		} else {
			fmt.Printf("[%s] All errors found\n", folderName)
		}
	}
}

var folderErrorRegexp = regexp.MustCompile(`^((?:[EW]\d{3}_)+)`)
var errorCodeRegexp = regexp.MustCompile(`([EW]\d{3})`)

func errorsFromFolder(folder string) []string {
	matches := folderErrorRegexp.FindStringSubmatch(folder)
	if len(matches) < 2 {
		return nil
	}
	errorMatches := errorCodeRegexp.FindAllString(matches[1], -1)
	return errorMatches
}

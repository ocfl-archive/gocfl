package cmd

import (
	"context"
	"emperror.dev/errors"
	"fmt"
	lm "github.com/je4/utils/v2/pkg/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.ub.unibas.ch/gocfl/v2/pkg/ocfl"
	"golang.org/x/exp/slices"
	"path/filepath"
	"strings"
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
	validateCmd.Flags().StringVarP(&objectPath, "object-path", "o", "", "validate only the selected object in storage root")
}

var objectPath string

func validate(cmd *cobra.Command, args []string) {
	ocflPath := filepath.ToSlash(filepath.Clean(args[0]))
	persistentFlagLogfile := viper.GetString("LogFile")
	persistentFlagLoglevel := strings.ToUpper(viper.GetString("LogLevel"))
	if !slices.Contains([]string{"DEBUG", "ERROR", "WARNING", "INFO", "CRITICAL"}, persistentFlagLoglevel) {
		cmd.Help()
		cobra.CheckErr(errors.Errorf("invalid log level '%s' for flag 'log-level' or 'LogLevel' config file entry", persistentFlagLoglevel))
	}

	logger, lf := lm.CreateLogger("ocfl", persistentFlagLogfile, nil, persistentFlagLoglevel, LOGFORMAT)
	defer lf.Close()

	extensionFlags, err := getExtensionFlags(cmd)
	if err != nil {
		logger.Errorf("cannot get extension flags: %v", err)
		logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	fmt.Printf("validating '%s'\n", ocflPath)

	extensionFactory, err := ocfl.NewExtensionFactory(logger)
	if err != nil {
		logger.Errorf("cannot instantiate extension factory: %v", err)
		logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}
	if err := initExtensionFactory(extensionFactory, extensionFlags); err != nil {
		logger.Errorf("cannot initialize extension factory: %v", err)
		logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}
	/*
		storageRootExtensions, objectExtensions, err := initDefaultExtensions(extensionFactory, flagExtensionFolder, logger)
		if err != nil {
			logger.Errorf("cannot initialize default extensions: %v", err)
			return
		}
	*/

	ocfs, err := OpenRO(ocflPath, logger)
	if err != nil {
		logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	ctx := ocfl.NewContextValidation(context.TODO())
	defer showStatus(ctx)
	storageRoot, err := ocfl.LoadStorageRoot(ctx, ocfs, extensionFactory, logger)
	if err != nil {
		logger.Errorf("cannot load storageroot: %v", err)
		logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}
	if objectPath == "" {
		if err := storageRoot.Check(); err != nil {
			logger.Errorf("ocfl not valid: %v", err)
			logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
			return
		}
	} else {
		if err := storageRoot.CheckObject(objectPath); err != nil {
			logger.Errorf("ocfl object '%s' not valid: %v", objectPath, err)
			logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
			return
		}
	}
}

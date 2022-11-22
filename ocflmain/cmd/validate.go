package cmd

import (
	"context"
	lm "github.com/je4/utils/v2/pkg/logger"
	"github.com/spf13/cobra"
	"go.ub.unibas.ch/gocfl/v2/pkg/ocfl"
)

const LOGFORMAT = `%{time:2006-01-02T15:04:05.000} %{shortpkg}::%{longfunc} [%{shortfile}] > %{level:.5s} - %{message}`

var validateCmd = &cobra.Command{
	Use:     "validate [path to ocfl structure]",
	Aliases: []string{"check"},
	Short:   "validates an ocfl structure",
	//Long:    "an utterly useless command for testing",
	Example: "gocfl validate ./archive.zip",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		validate(cmd, args)
	},
}

func validate(cmd *cobra.Command, args []string) {
	ocflPath := args[0]

	logger, lf := lm.CreateLogger("ocfl", logfile, nil, LogLevelIds[loglevel][0], LOGFORMAT)
	defer lf.Close()

	extensionFactory, err := ocfl.NewExtensionFactory(logger)
	if err != nil {
		logger.Errorf("cannot instantiate extension factory: %v", err)
		return
	}
	if err := initExtensionFactory(extensionFactory); err != nil {
		logger.Errorf("cannot initialize extension factory: %v", err)
		return
	}
	/*
		storageRootExtensions, objectExtensions, err := initDefaultExtensions(extensionFactory, extensionFolder, logger)
		if err != nil {
			logger.Errorf("cannot initialize default extensions: %v", err)
			return
		}
	*/

	ocfs, err := OpenRO(ocflPath, logger)
	if err != nil {
		logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	ctx := ocfl.NewContextValidation(context.TODO())
	defer showStatus(ctx)
	storageRoot, err := ocfl.LoadStorageRoot(ctx, ocfs, extensionFactory, logger)

	if err != nil {
		logger.Errorf("cannot create new storageroot: %v", err)
		return
	}
	if err := storageRoot.Check(); err != nil {
		logger.Errorf("ocfl not valid: %v", err)
		return
	}

}

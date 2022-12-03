package cmd

import (
	"context"
	"fmt"
	lm "github.com/je4/utils/v2/pkg/logger"
	"github.com/spf13/cobra"
	"github.com/thediveo/enumflag"
	"go.ub.unibas.ch/gocfl/v2/pkg/ocfl"
	"os"
	"path/filepath"
	"strings"
)

var initCmd = &cobra.Command{
	Use:     "init [path to ocfl structure]",
	Aliases: []string{"check"},
	Short:   "initializes an empty ocfl structure",
	Long:    "initializes an empty ocfl structure",
	Example: "gocfl init ./archive.zip /tmp/testdata",
	Args:    cobra.ExactArgs(1),
	Run:     doInit,
}

func initInit() {
	initCmd.PersistentFlags().StringVarP(&flagExtensionFolder, "extensions", "e", "", "folder with extension configurations")
	initCmd.PersistentFlags().VarP(
		enumflag.New(&flagVersion, "ocfl-version", VersionIds, enumflag.EnumCaseInsensitive),
		"ocfl-version", "v", "ocfl version for new storage root")
}

func doInit(cmd *cobra.Command, args []string) {
	ocflPath := filepath.ToSlash(filepath.Clean(args[0]))

	fmt.Printf("creating '%s'\n", ocflPath)

	logger, lf := lm.CreateLogger("ocfl", persistentFlagLogfile, nil, LogLevelIds[persistentFlagLoglevel][0], LOGFORMAT)
	defer lf.Close()
	logger.Infof("creating '%s'", ocflPath)

	finfo, err := os.Stat(ocflPath)
	if err != nil {
		if !(os.IsNotExist(err) && strings.HasSuffix(strings.ToLower(ocflPath), ".zip")) {
			logger.Errorf("cannot stat '%s': %v", ocflPath, err)
			logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
			return
		}
	} else {
		if strings.HasSuffix(strings.ToLower(ocflPath), ".zip") {
			logger.Errorf("path '%s' already exists", ocflPath)
			fmt.Printf("path '%s' already exists\n", ocflPath)
			return
		}
		if !finfo.IsDir() {
			logger.Errorf("'%s' is not a directory", ocflPath)
			logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
			return
		}
	}

	extensionFactory, err := ocfl.NewExtensionFactory(logger)
	if err != nil {
		logger.Errorf("cannot instantiate extension factory: %v", err)
		logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}
	if err := initExtensionFactory(extensionFactory); err != nil {
		logger.Errorf("cannot initialize extension factory: %v", err)
		logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}
	storageRootExtensions, _, err := initDefaultExtensions(extensionFactory, flagExtensionFolder, logger)
	if err != nil {
		logger.Errorf("cannot initialize default extensions: %v", err)
		logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	tempFile := fmt.Sprintf("%s.tmp", ocflPath)
	reader, writer, ocfs, err := OpenRW(ocflPath, tempFile, logger)
	if err != nil {
		logger.Errorf("cannot create target filesystem: %v", err)
		logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	ctx := ocfl.NewContextValidation(context.TODO())
	defer showStatus(ctx)
	if _, err = ocfl.CreateStorageRoot(ctx, ocfs, VersionIdsVersion[flagVersion], extensionFactory, storageRootExtensions, "", logger); err != nil {
		ocfs.Discard()
		logger.Errorf("cannot create new storageroot: %v", err)
		logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}
	if err := ocfs.Close(); err != nil {
		logger.Errorf("error closing filesystem '%s': %v", ocfs, err)
		logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
	} else {
		if err := reader.Close(); err != nil {
			logger.Errorf("error closing reader: %v", err)
			logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		}
		if err := writer.Close(); err != nil {
			logger.Errorf("error closing writer: %v", err)
			logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		}
		if err := os.Rename(tempFile, ocflPath); err != nil {
			logger.Errorf("cannot rename '%s' -> '%s': %v", tempFile, ocflPath, err)
		}
	}

}

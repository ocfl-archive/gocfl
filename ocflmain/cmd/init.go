package cmd

import (
	"context"
	"emperror.dev/errors"
	"fmt"
	lm "github.com/je4/utils/v2/pkg/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.ub.unibas.ch/gocfl/v2/pkg/checksum"
	"go.ub.unibas.ch/gocfl/v2/pkg/ocfl"
	"golang.org/x/exp/slices"
	"os"
	"path/filepath"
	"strings"
)

var initCmd = &cobra.Command{
	Use:     "init [path to ocfl structure]",
	Aliases: []string{"check"},
	Short:   "initializes an empty ocfl structure",
	Long:    "initializes an empty ocfl structure",
	Example: "gocfl init ./archive.zip",
	Args:    cobra.ExactArgs(1),
	Run:     doInit,
}

func initInit() {
	initCmd.Flags().String("default-storageroot-extensions", "", "folder with initial extension configurations for new OCFL Storage Root")
	viper.BindPFlag("Init.StorageRootExtensions", initCmd.Flags().Lookup("default-storageroot-extensions"))

	initCmd.Flags().String("ocfl-version", "v", "ocfl version for new storage root")
	viper.BindPFlag("Init.OCFLVersion", initCmd.Flags().Lookup("ocfl-version"))

	initCmd.Flags().StringP("digest", "d", "", "digest to use for ocfl checksum")
	viper.BindPFlag("Init.DigestAlgorithm", initCmd.Flags().Lookup("digest"))
}

func doInit(cmd *cobra.Command, args []string) {
	ocflPath := filepath.ToSlash(filepath.Clean(args[0]))
	persistentFlagLogfile := viper.GetString("LogFile")

	persistentFlagLoglevel := strings.ToUpper(viper.GetString("LogLevel"))
	if !slices.Contains([]string{"DEBUG", "ERROR", "WARNING", "INFO", "CRITICAL"}, persistentFlagLoglevel) {
		cmd.Help()
		cobra.CheckErr(errors.Errorf("invalid log level '%s' for flag 'log-level' or 'LogLevel' config file entry", persistentFlagLoglevel))
	}

	flagStorageRootExtensionFolder := viper.GetString("Init.StorageRootExtensions")

	flagVersion := viper.GetString("Init.OCFLVersion")
	if !ocfl.ValidVersion(ocfl.OCFLVersion(flagVersion)) {
		cmd.Help()
		cobra.CheckErr(errors.Errorf("invalid version '%s' for flag 'ocfl-version' or 'Init.OCFLVersion' config file entry", flagVersion))
	}

	flagInitDigest := viper.GetString("Init.DigestAlgorithm")
	if _, err := checksum.GetHash(checksum.DigestAlgorithm(flagInitDigest)); err != nil {
		cmd.Help()
		cobra.CheckErr(errors.Errorf("invalid digest '%s' for flag 'digest' or 'Init.DigestAlgorithm' config file entry", flagInitDigest))
	}

	fmt.Printf("creating '%s'\n", ocflPath)

	logger, lf := lm.CreateLogger("ocfl", persistentFlagLogfile, nil, persistentFlagLoglevel, LOGFORMAT)
	defer lf.Close()
	logger.Infof("creating '%s'", ocflPath)

	extensionFlags, err := getExtensionFlags(cmd)
	if err != nil {
		logger.Errorf("cannot get extension flags: %v", err)
		logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

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
	if err := initExtensionFactory(extensionFactory, extensionFlags); err != nil {
		logger.Errorf("cannot initialize extension factory: %v", err)
		logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}
	storageRootExtensions, _, err := initDefaultExtensions(extensionFactory, flagStorageRootExtensionFolder, "", logger)
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
	if _, err = ocfl.CreateStorageRoot(ctx, ocfs, ocfl.OCFLVersion(flagVersion), extensionFactory, storageRootExtensions, checksum.DigestAlgorithm(flagInitDigest), logger); err != nil {
		ocfs.Discard()
		logger.Errorf("cannot create new storageroot: %v", err)
		logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}
	if err := ocfs.Close(); err != nil {
		logger.Errorf("error closing filesystem '%s': %v", ocfs, err)
		logger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
	} else {
		if reader != (*os.File)(nil) {
			if err := reader.Close(); err != nil {
				logger.Errorf("error closing reader: %v", err)
			}
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

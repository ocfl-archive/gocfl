package cmd

import (
	"context"
	"emperror.dev/errors"
	"github.com/je4/filesystem/v2/pkg/writefs"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	"github.com/je4/utils/v2/pkg/checksum"
	lm "github.com/je4/utils/v2/pkg/logger"
	"github.com/spf13/cobra"
	"path/filepath"
)

var initCmd = &cobra.Command{
	Use:     "init [path to ocfl structure]",
	Aliases: []string{},
	Short:   "initializes an empty ocfl structure",
	Long:    "initializes an empty ocfl structure",
	Example: "gocfl init ./archive.zip",
	Args:    cobra.ExactArgs(1),
	Run:     doInit,
}

func initInit() {
	initCmd.Flags().String("default-storageroot-extensions", "", "folder with initial extension configurations for new OCFL Storage Root")
	initCmd.Flags().String("ocfl-version", "", "ocfl version for new storage root")
	initCmd.Flags().StringP("digest", "d", "", "digest to use for ocfl checksum")
	initCmd.Flags().Bool("no-compress", false, "do not compress data in zip file")
}

func doInitConf(cmd *cobra.Command) {
	if str := getFlagString(cmd, "default-storageroot-extensions"); str != "" {
		conf.Init.StorageRootExtensionFolder = str
	}

	if str := getFlagString(cmd, "ocfl-version"); str != "" {
		conf.Init.OCFLVersion = str
	}

	if str := getFlagString(cmd, "digest"); str != "" {
		conf.Init.Digest = checksum.DigestAlgorithm(str)
	}
	if _, err := checksum.GetHash(conf.Init.Digest); err != nil {
		_ = cmd.Help()
		cobra.CheckErr(errors.Errorf("invalid digest '%s' for flag 'digest' or 'Init.DigestAlgorithm' config file entry", conf.Init.Digest))
	}

}

func doInit(cmd *cobra.Command, args []string) {
	//ocflPath := filepath.ToSlash(filepath.Clean(args[0]))
	ocflPath := filepath.ToSlash(args[0])

	daLogger, lf := lm.CreateLogger("ocfl", conf.Logfile, nil, conf.LogLevel, LOGFORMAT)
	defer lf.Close()

	doInitConf(cmd)

	daLogger.Infof("creating '%s'", ocflPath)
	t := startTimer()
	defer func() { daLogger.Infof("Duration: %s", t.String()) }()

	fsFactory, err := initializeFSFactory([]checksum.DigestAlgorithm{conf.Init.Digest}, conf.AES, conf.S3, true, false, daLogger)
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

	extensionParams := GetExtensionParamValues(cmd)
	extensionFactory, err := initExtensionFactory(
		extensionParams,
		"",
		false,
		nil,
		nil,
		nil,
		nil,
		daLogger,
	)
	if err != nil {
		daLogger.Errorf("cannot initialize extension factory: %v", err)
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}
	storageRootExtensions, _, err := initDefaultExtensions(
		extensionFactory,
		conf.Init.StorageRootExtensionFolder,
		"",
	)
	if err != nil {
		daLogger.Errorf("cannot initialize default extensions: %v", err)
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	ctx := ocfl.NewContextValidation(context.TODO())
	if _, err := ocfl.CreateStorageRoot(
		ctx,
		destFS,
		ocfl.OCFLVersion(conf.Init.OCFLVersion),
		extensionFactory, storageRootExtensions,
		conf.Init.Digest,
		daLogger,
	); err != nil {
		if err := writefs.Close(destFS); err != nil {
			daLogger.Errorf("cannot discard filesystem '%s': %v", destFS, err)
		}
		daLogger.Errorf("cannot create new storageroot: %v", err)
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	_ = showStatus(ctx)
}

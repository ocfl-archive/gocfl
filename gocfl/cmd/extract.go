package cmd

import (
	"context"
	"emperror.dev/emperror"
	"emperror.dev/errors"
	"fmt"
	"github.com/je4/filesystem/v2/pkg/writefs"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	lm "github.com/je4/utils/v2/pkg/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/exp/slices"
	"io/fs"
	"path/filepath"
	"strings"
)

var extractCmd = &cobra.Command{
	Use:     "extract [path to ocfl structure] [path to target folder]",
	Aliases: []string{},
	Short:   "extract version of ocfl content",
	//Long:    "an utterly useless command for testing",
	Example: "gocfl extract ./archive.zip /tmp/archive",
	Args:    cobra.MinimumNArgs(2),
	Run:     doExtract,
}

func initExtract() {
	extractCmd.Flags().StringP("object-path", "p", "", "object path to extract")

	extractCmd.Flags().StringP("object-id", "i", "", "object id to extract")

	extractCmd.Flags().Bool("with-manifest", false, "generate manifest file in object extraction folder")
	emperror.Panic(viper.BindPFlag("Extract.Manifest", extractCmd.Flags().Lookup("with-manifest")))

	extractCmd.Flags().String("version", "latest", "version to extract")
	emperror.Panic(viper.BindPFlag("Extract.Version", extractCmd.Flags().Lookup("version")))
}

func doExtract(cmd *cobra.Command, args []string) {
	ocflPath := filepath.ToSlash(args[0])
	destPath := filepath.ToSlash(args[1])

	persistentFlagLogfile := viper.GetString("LogFile")
	persistentFlagLoglevel := strings.ToUpper(viper.GetString("LogLevel"))
	if !slices.Contains([]string{"DEBUG", "ERROR", "WARNING", "INFO", "CRITICAL"}, persistentFlagLoglevel) {
		cmd.Help()
		cobra.CheckErr(errors.Errorf("invalid log level '%s' for flag 'log-level' or 'LogLevel' config file entry", persistentFlagLoglevel))
	}

	withManifest := viper.GetBool("Extract.Manifest")
	version := viper.GetString("Extract.Version")
	if version == "" {
		version = "latest"
	}

	oPath, _ := cmd.Flags().GetString("object-path")
	oID, _ := cmd.Flags().GetString("object-id")
	if oPath != "" && oID != "" {
		cmd.Help()
		cobra.CheckErr(errors.New("do not use object-path AND object-id at the same time"))
		return
	}

	daLogger, lf := lm.CreateLogger("ocfl", persistentFlagLogfile, nil, persistentFlagLoglevel, LOGFORMAT)
	defer lf.Close()
	t := startTimer()
	defer func() { daLogger.Infof("Duration: %s", t.String()) }()

	daLogger.Infof("extracting '%s'", ocflPath)

	fsFactory, err := initializeFSFactory("Extract", cmd, nil, true, daLogger)
	if err != nil {
		daLogger.Errorf("cannot create filesystem factory: %v", err)
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	ocflFS, err := fsFactory.Get(ocflPath)
	if err != nil {
		daLogger.Errorf("cannot get filesystem for '%s': %v", ocflPath, err)
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	destFS, err := fsFactory.Get(destPath)
	if err != nil {
		daLogger.Errorf("cannot get filesystem for '%s': %v", destPath, err)
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}
	defer func() {
		if err := writefs.Close(destFS); err != nil {
			daLogger.Errorf("cannot close filesystem: %v", err)
			daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		}
	}()

	extensionParams := GetExtensionParamValues(cmd)
	extensionFactory, err := initExtensionFactory(extensionParams, "", false, nil, nil, nil, nil, daLogger)
	if err != nil {
		daLogger.Errorf("cannot initialize extension factory: %v", err)
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	ctx := ocfl.NewContextValidation(context.TODO())
	storageRoot, err := ocfl.LoadStorageRoot(ctx, ocflFS, extensionFactory, daLogger)
	if err != nil {
		daLogger.Errorf("cannot open storage root: %v", err)
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	dirs, err := fs.ReadDir(destFS, ".")
	if err != nil {
		daLogger.Errorf("cannot read target folder '%v': %v", destFS, err)
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}
	if len(dirs) > 0 {
		fmt.Printf("target folder '%s' is not empty\n", destFS)
		daLogger.Debugf("target folder '%s' is not empty", destFS)
		return
	}

	if err := storageRoot.Extract(destFS, oPath, oID, version, withManifest); err != nil {
		fmt.Printf("cannot extract storage root: %v\n", err)
		daLogger.Errorf("cannot extract storage root: %v\n", err)
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}
	fmt.Printf("extraction done without errors\n")
	showStatus(ctx)
}

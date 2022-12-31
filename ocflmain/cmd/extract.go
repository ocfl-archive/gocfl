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
	extractCmd.Flags().StringP("object-path", "p", "", "object path to show statistics for")

	extractCmd.Flags().StringP("object-id", "i", "", "object id to show statistics for")

	extractCmd.Flags().Bool("with-manifest", false, "generate manifest file in object extraction folder")
	viper.BindPFlag("Extract.Manifest", extractCmd.Flags().Lookup("with-manifest"))

	extractCmd.Flags().String("version", "latest", "version to extract")
	viper.BindPFlag("Extract.Version", extractCmd.Flags().Lookup("version"))
}

func doExtract(cmd *cobra.Command, args []string) {

	ocflPath := filepath.ToSlash(filepath.Clean(args[0]))
	destPath := filepath.ToSlash(filepath.Clean(args[1]))

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
	daLogger.Infof("creating '%s'", ocflPath)

	fsFactory, err := initializeFSFactory(daLogger)
	if err != nil {
		daLogger.Errorf("cannot create filesystem factory: %v", err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	ocflFS, err := fsFactory.GetFS(ocflPath)
	if err != nil {
		daLogger.Errorf("cannot get filesystem for '%s': %v", ocflPath, err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	destFS, err := fsFactory.GetFSRW(destPath)
	if err != nil {
		daLogger.Errorf("cannot get filesystem for '%s': %v", destPath, err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	extensionParams := GetExtensionParamValues(cmd)
	extensionFactory, err := initExtensionFactory(extensionParams, daLogger)
	if err != nil {
		daLogger.Errorf("cannot initialize extension factory: %v", err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	ctx := ocfl.NewContextValidation(context.TODO())
	defer showStatus(ctx)
	if !ocflFS.HasContent() {

	}
	storageRoot, err := ocfl.LoadStorageRoot(ctx, ocflFS, extensionFactory, daLogger)
	if err != nil {
		daLogger.Errorf("cannot open storage root: %v", err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	if destFS.HasContent() {
		fmt.Printf("target folder '%s' is not empty\n", destFS)
		daLogger.Errorf("target folder '%s' is not empty", destFS)
		return
	}

	if err := storageRoot.Extract(destFS, oPath, oID, version, withManifest); err != nil {
		fmt.Printf("cannot extract storage root: %v\n", err)
		daLogger.Errorf("cannot extract storage root: %v\n", err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}
	fmt.Printf("extraction done without errors\n")
}

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
	"os"
	"path/filepath"
	"strings"
)

var statCmd = &cobra.Command{
	Use:     "stat [path to ocfl structure]",
	Aliases: []string{"info"},
	Short:   "statistics of an ocfl structure",
	//Long:    "an utterly useless command for testing",
	Example: "gocfl stat ./archive.zip",
	Args:    cobra.MinimumNArgs(1),
	Run:     doStat,
}

func initStat() {
	statCmd.Flags().StringP("object-path", "p", "", "object path to show statistics for")

	statCmd.Flags().StringP("object-id", "i", "", "object id to show statistics for")

	infos := []string{}
	for info, _ := range ocfl.StatInfoString {
		infos = append(infos, info)
	}
	statCmd.Flags().StringArray("stat-info", []string{}, fmt.Sprintf("info field to show. multiple use [%s]", strings.Join(infos, ",")))
	viper.BindPFlag("Stat.Info", statCmd.Flags().Lookup("stat-info"))
}

func doStat(cmd *cobra.Command, args []string) {
	ocflPath := filepath.ToSlash(filepath.Clean(args[0]))

	persistentFlagLogfile := viper.GetString("LogFile")
	persistentFlagLoglevel := strings.ToUpper(viper.GetString("LogLevel"))
	if !slices.Contains([]string{"DEBUG", "ERROR", "WARNING", "INFO", "CRITICAL"}, persistentFlagLoglevel) {
		cmd.Help()
		cobra.CheckErr(errors.Errorf("invalid log level '%s' for flag 'log-level' or 'LogLevel' config file entry", persistentFlagLoglevel))
	}

	oPath, _ := cmd.Flags().GetString("object-path")
	oID, _ := cmd.Flags().GetString("object-id")
	if oPath != "" && oID != "" {
		cmd.Help()
		cobra.CheckErr(errors.New("do not use object-path AND object-id at the same time"))
		return
	}

	statInfoStrings := viper.GetStringSlice("Stat.Info")
	statInfo := []ocfl.StatInfo{}
	for _, statInfoString := range statInfoStrings {
		statInfoString = strings.ToLower(statInfoString)
		var found bool
		for str, info := range ocfl.StatInfoString {
			if strings.ToLower(str) == statInfoString {
				found = true
				statInfo = append(statInfo, info)
			}
		}
		if !found {
			cmd.Help()
			cobra.CheckErr(errors.Errorf("--stat-info invalid value '%s' ", statInfoString))
		}
	}

	daLogger, lf := lm.CreateLogger("ocfl", persistentFlagLogfile, nil, persistentFlagLoglevel, LOGFORMAT)
	defer lf.Close()
	daLogger.Infof("creating '%s'", ocflPath)

	extensionFlags, err := getExtensionFlags(cmd)
	if err != nil {
		daLogger.Errorf("cannot get extension flags: %v", err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	fsFactory, err := initializeFSFactory(daLogger)
	if err != nil {
		daLogger.Errorf("cannot create filesystem factory: %v", err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	destFS, err := fsFactory.GetFS(ocflPath)
	if err != nil {
		daLogger.Errorf("cannot get filesystem for '%s': %v", ocflPath, err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	extensionFactory, err := initExtensionFactory(daLogger, extensionFlags)
	if err != nil {
		daLogger.Errorf("cannot initialize extension factory: %v", err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	ctx := ocfl.NewContextValidation(context.TODO())
	defer showStatus(ctx)
	if !destFS.HasContent() {

	}
	storageRoot, err := ocfl.LoadStorageRoot(ctx, destFS, extensionFactory, daLogger)
	if err != nil {
		daLogger.Errorf("cannot open storage root: %v", err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	if err := storageRoot.Stat(os.Stdout, oPath, oID, statInfo); err != nil {
		daLogger.Errorf("cannot get statistics: %v", err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

}

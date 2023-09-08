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
	statCmd.Flags().String("stat-info", "", fmt.Sprintf("comma separated list of info fields to show [%s]", strings.Join(infos, ",")))
}

func doStatConf(cmd *cobra.Command) {
	if str := getFlagString(cmd, "object-path"); str != "" {
		conf.Stat.ObjectPath = str
	}
	if str := getFlagString(cmd, "object-id"); str != "" {
		conf.Stat.ObjectID = str
	}
	if str := getFlagString(cmd, "stat-info"); str != "" {
		conf.Stat.Info = []string{}
		for _, s := range strings.Split(str, ",") {
			conf.Stat.Info = append(conf.Stat.Info, strings.ToLower(strings.TrimSpace(s)))
		}
	}
}

func doStat(cmd *cobra.Command, args []string) {
	ocflPath := filepath.ToSlash(args[0])

	daLogger, lf := lm.CreateLogger("ocfl", persistentFlagLogfile, nil, conf.LogLevel, conf.LogFormat)
	defer lf.Close()

	doStatConf(cmd)

	oPath := conf.Stat.ObjectPath
	oID := conf.Stat.ObjectID
	if oPath != "" && oID != "" {
		emperror.Panic(cmd.Help())
		cobra.CheckErr(errors.New("do not use object-path AND object-id at the same time"))
		return
	}

	statInfo := []ocfl.StatInfo{}
	for _, statInfoString := range conf.Stat.Info {
		statInfoString = strings.ToLower(strings.TrimSpace(statInfoString))
		var found bool
		for str, info := range ocfl.StatInfoString {
			if strings.ToLower(str) == statInfoString {
				found = true
				statInfo = append(statInfo, info)
			}
		}
		if !found {
			emperror.Panic(cmd.Help())
			cobra.CheckErr(errors.Errorf("--stat-info invalid value '%s' ", statInfoString))
		}
	}

	t := startTimer()
	defer func() { daLogger.Infof("Duration: %s", t.String()) }()

	daLogger.Infof("opening '%s'", ocflPath)

	fsFactory, err := initializeFSFactory(nil, nil, nil, true, false, daLogger)
	if err != nil {
		daLogger.Errorf("cannot create filesystem factory: %v", err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	destFS, err := fsFactory.Get(ocflPath)
	if err != nil {
		daLogger.Errorf("cannot get filesystem for '%s': %v", ocflPath, err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}
	defer func() {
		if err := writefs.Close(destFS); err != nil {
			daLogger.Errorf("cannot close filesystem: %v", err)
			daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		}
	}()

	extensionParams := GetExtensionParamValues(cmd, conf)
	extensionFactory, err := initExtensionFactory(extensionParams, "", false, nil, nil, nil, nil, daLogger)
	if err != nil {
		daLogger.Errorf("cannot initialize extension factory: %v", err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	ctx := ocfl.NewContextValidation(context.TODO())
	if !writefs.HasContent(destFS) {

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
	showStatus(ctx)
}

package cmd

import (
	"context"
	"emperror.dev/errors"
	"fmt"
	"github.com/je4/filesystem/v2/pkg/writefs"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	lm "github.com/je4/utils/v2/pkg/logger"
	"github.com/spf13/cobra"
	"io/fs"
	"path/filepath"
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
	extractCmd.Flags().String("version", "", "version to extract")
	extractCmd.Flags().String("area", "content", "data area to extract")
}
func doExtractConf(cmd *cobra.Command) {
	if str := getFlagString(cmd, "object-path"); str != "" {
		conf.Extract.ObjectPath = str
	}
	if str := getFlagString(cmd, "object-id"); str != "" {
		conf.Extract.ObjectID = str
	}
	if b := getFlagBool(cmd, "with-manifest"); b {
		conf.Extract.Manifest = b
	}
	if str := getFlagString(cmd, "version"); str != "" {
		conf.Extract.Version = str
	}
	if str := getFlagString(cmd, "area"); str != "" {
		conf.Extract.Area = str
	}
	if conf.Extract.Version == "" {
		conf.Extract.Version = "latest"
	}
}

func doExtract(cmd *cobra.Command, args []string) {
	daLogger, lf := lm.CreateLogger("ocfl", persistentFlagLogfile, nil, conf.LogLevel, conf.LogFormat)
	defer lf.Close()
	t := startTimer()
	defer func() { daLogger.Infof("Duration: %s", t.String()) }()

	ocflPath := filepath.ToSlash(args[0])
	destPath := filepath.ToSlash(args[1])

	doExtractConf(cmd)

	oPath := conf.Extract.ObjectPath
	oID := conf.Extract.ObjectID
	if oPath != "" && oID != "" {
		cmd.Help()
		cobra.CheckErr(errors.New("do not use object-path AND object-id at the same time"))
		return
	}

	daLogger.Infof("extracting '%s'", ocflPath)

	fsFactory, err := initializeFSFactory(nil, nil, nil, true, true, daLogger)
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

	extensionParams := GetExtensionParamValues(cmd, conf)
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

	if err := storageRoot.Extract(destFS, oPath, oID, conf.Extract.Version, conf.Extract.Manifest, conf.Extract.Area); err != nil {
		fmt.Printf("cannot extract storage root: %v\n", err)
		daLogger.Errorf("cannot extract storage root: %v\n", err)
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}
	fmt.Printf("extraction done without errors\n")
	showStatus(ctx)
}

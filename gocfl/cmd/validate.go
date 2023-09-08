package cmd

import (
	"context"
	"github.com/je4/filesystem/v2/pkg/writefs"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	lm "github.com/je4/utils/v2/pkg/logger"
	"github.com/spf13/cobra"
	"path/filepath"
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
	validateCmd.Flags().StringP("object-path", "o", "", "validate only the object at the specified path in storage root")
	validateCmd.Flags().String("object-id", "", "validate only the object with the specified id in storage root")
}

func doValidateConf(cmd *cobra.Command) {
	if str := getFlagString(cmd, "object-path"); str != "" {
		conf.Validate.ObjectPath = str
	}
	if str := getFlagString(cmd, "object-id"); str != "" {
		conf.Validate.ObjectID = str
	}
}

func validate(cmd *cobra.Command, args []string) {
	//	ocflPath := filepath.ToSlash(filepath.Clean(args[0]))
	ocflPath := filepath.ToSlash(args[0])

	daLogger, lf := lm.CreateLogger("ocfl", persistentFlagLogfile, nil, conf.LogLevel, conf.LogFormat)
	defer lf.Close()
	t := startTimer()
	defer func() { daLogger.Infof("Duration: %s", t.String()) }()

	doValidateConf(cmd)

	daLogger.Infof("validating '%s'", ocflPath)

	extensionParams := GetExtensionParamValues(cmd)
	extensionFactory, err := initExtensionFactory(extensionParams, "", false, nil, nil, nil, nil, daLogger)
	if err != nil {
		daLogger.Errorf("cannot initialize extension factory: %v", err)
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	fsFactory, err := initializeFSFactory(nil, nil, nil, true, false, daLogger)
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

	ctx := ocfl.NewContextValidation(context.TODO())
	storageRoot, err := ocfl.LoadStorageRoot(ctx, destFS, extensionFactory, daLogger)
	if err != nil {
		daLogger.Errorf("cannot load storageroot: %v", err)
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}
	objectID := conf.Validate.ObjectID
	objectPath := conf.Validate.ObjectPath
	if objectID != "" && objectPath != "" {
		daLogger.Errorf("cannot specify both --object-id and --object-path")
		return
	}
	if objectID == "" && objectPath == "" {
		if err := storageRoot.Check(); err != nil {
			daLogger.Errorf("ocfl not valid: %v", err)
			daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
			return
		}
	} else {
		if objectID != "" {
			if err := storageRoot.CheckObjectByID(objectID); err != nil {
				daLogger.Errorf("ocfl object '%s' not valid: %v", objectID, err)
				daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
				return
			}
		} else {
			if err := storageRoot.CheckObjectByFolder(objectPath); err != nil {
				daLogger.Errorf("ocfl object '%s' not va§lid: %v", objectPath, err)
				daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
				return
			}
		}
	}
	showStatus(ctx)
}

package cmd

import (
	"context"
	"emperror.dev/errors"
	"fmt"
	"github.com/je4/filesystem/v2/pkg/writefs"
	"github.com/je4/gocfl/v2/internal"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	"github.com/je4/gocfl/v2/pkg/subsystem/migration"
	"github.com/je4/gocfl/v2/pkg/subsystem/thumbnail"
	ironmaiden "github.com/je4/indexer/v2/pkg/indexer"
	"github.com/je4/utils/v2/pkg/checksum"
	lm "github.com/je4/utils/v2/pkg/logger"
	"github.com/spf13/cobra"
	"io/fs"
	"os"
)

var updateCmd = &cobra.Command{
	Use:     "update [path to ocfl structure]",
	Aliases: []string{},
	Short:   "update object in existing ocfl structure",
	Long:    "opens an existing ocfl structure and updates an object. if an object with the given id does not exist, an error is produced",
	Example: "gocfl update ./archive.zip /tmp/testdata -u 'Jane Doe' -a 'mailto:user@domain' -m 'initial add' -object-id 'id:abc123'",
	Args:    cobra.MinimumNArgs(2),
	Run:     doUpdate,
}

func initUpdate() {
	updateCmd.Flags().StringVarP(&flagObjectID, "object-id", "i", "", "object id to update (required)")
	updateCmd.Flags().StringP("message", "m", "", "message for new object version (required)")
	updateCmd.Flags().StringP("user-name", "u", "", "user name for new object version (required)")
	updateCmd.Flags().StringP("user-address", "a", "", "user address for new object version (required)")
	updateCmd.Flags().StringP("digest", "d", "", "digest to use for zip file checksum")
	updateCmd.Flags().Bool("no-deduplicate", false, "disable deduplication (faster)")
	updateCmd.Flags().Bool("echo", false, "update strategy 'echo' (reflects deletions). if not set, update strategy is 'contribute'")
	updateCmd.Flags().Bool("no-compress", false, "do not compress data in zip file")
	updateCmd.Flags().Bool("encrypt-aes", false, "set flag to create encrypted container (only for container target)")
	updateCmd.Flags().String("aes-key", "", "key to use for encrypted container in hex format (64 chars, empty: generate random key")
	updateCmd.Flags().String("aes-iv", "", "initialisation vector to use for encrypted container in hex format (32 charsempty: generate random vector")
}

func doUpdateConf(cmd *cobra.Command) {
	if str := getFlagString(cmd, "user-name"); str != "" {
		conf.Update.User.Name = str
	}
	if str := getFlagString(cmd, "user-address"); str != "" {
		conf.Update.User.Address = str
	}
	if str := getFlagString(cmd, "message"); str != "" {
		conf.Update.Message = str
	}
	if str := getFlagString(cmd, "digest"); str != "" {
		conf.Update.Digest = checksum.DigestAlgorithm(str)
	}
	if conf.Update.Digest == "" {
		conf.Update.Digest = checksum.DigestSHA512
	}
	if _, err := checksum.GetHash(conf.Update.Digest); err != nil {
		_ = cmd.Help()
		cobra.CheckErr(errors.Errorf("invalid digest '%s' for flag 'digest' or 'Init.DigestAlgorithm' config file entry", conf.Add.Digest))
	}
	if b := getFlagBool(cmd, "no-deduplicate"); b {
		conf.Update.Deduplicate = !b
	}
	if b := getFlagBool(cmd, "no-compress"); b {
		conf.Update.NoCompress = b
	}
	if b := getFlagBool(cmd, "echo"); b {
		conf.Update.Echo = b
	}

}

func doUpdate(cmd *cobra.Command, args []string) {
	var err error

	ocflPath, err := ocfl.Fullpath(args[0])
	if err != nil {
		cobra.CheckErr(err)
		return
	}
	srcPath, err := ocfl.Fullpath(args[1])
	if err != nil {
		cobra.CheckErr(err)
		return
	}

	daLogger, lf := lm.CreateLogger("ocfl", persistentFlagLogfile, nil, conf.LogLevel, conf.LogFormat)
	defer lf.Close()

	doUpdateConf(cmd)

	var addr string
	var localCache bool
	var fss = map[string]fs.FS{"internal": internal.InternalFS}

	indexerActions, err := ironmaiden.InitActionDispatcher(fss, *conf.Indexer, daLogger)
	if err != nil {
		daLogger.Panicf("cannot init indexer: %v", err)
	}

	t := startTimer()
	defer func() { daLogger.Infof("Duration: %s", t.String()) }()

	fmt.Printf("opening '%s'\n", ocflPath)
	daLogger.Infof("opening '%s'", ocflPath)

	if _, err := os.Stat(srcPath); err != nil {
		daLogger.Panicf("cannot stat '%s': %v", srcPath, err)
	}

	fsFactory, err := initializeFSFactory([]checksum.DigestAlgorithm{conf.Update.Digest}, nil, nil, true, false, daLogger)
	if err != nil {
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		daLogger.Panicf("cannot create filesystem factory: %v", err)
	}

	sourceFS, err := fsFactory.Get(srcPath)
	if err != nil {
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		daLogger.Panicf("cannot get filesystem for '%s': %v", srcPath, err)
	}
	destFS, err := fsFactory.Get(ocflPath)
	if err != nil {
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		daLogger.Panicf("cannot get filesystem for '%s': %v", ocflPath, err)
	}
	var doNotClose = false
	defer func() {
		if doNotClose {
			daLogger.Panicf("filesystem '%s' not closed", destFS)
		} else {
			if err := writefs.Close(destFS); err != nil {
				daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
				daLogger.Panicf("error closing filesystem '%s': %v", destFS, err)
			}
		}
	}()

	area := conf.DefaultArea
	if area == "" {
		area = "content"
	}
	var areaPaths = map[string]fs.FS{}
	for i := 2; i < len(args); i++ {
		matches := areaPathRegexp.FindStringSubmatch(args[i])
		if matches == nil {
			daLogger.Errorf("no area given in areapath '%s'", args[i])
			continue
		}
		areaPaths[matches[1]], err = fsFactory.Get(matches[2])
		if err != nil {
			doNotClose = true
			daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
			daLogger.Panicf("cannot get filesystem for '%s': %v", args[i], err)
		}
	}

	mig, err := migration.GetMigrations(conf)
	if err != nil {
		daLogger.Errorf("cannot get migrations: %v", err)
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		doNotClose = true
		return
	}
	mig.SetSourceFS(sourceFS)

	thumb, err := thumbnail.GetThumbnails(conf)
	if err != nil {
		daLogger.Errorf("cannot get thumbnails: %v", err)
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		doNotClose = true
		return
	}
	thumb.SetSourceFS(sourceFS)

	extensionParams := GetExtensionParamValues(cmd, conf)
	extensionFactory, err := InitExtensionFactory(
		extensionParams,
		addr,
		localCache,
		indexerActions,
		mig,
		thumb,
		sourceFS,
		daLogger,
	)
	if err != nil {
		daLogger.Errorf("cannot initialize extension factory: %v", err)
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		doNotClose = true
		return
	}
	_, objectExtensions, err := initDefaultExtensions(
		extensionFactory,
		"",
		"",
	)
	if err != nil {
		daLogger.Errorf("cannot initialize default extensions: %v", err)
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		doNotClose = true
		return
	}

	ctx := ocfl.NewContextValidation(context.TODO())
	if !writefs.HasContent(destFS) {

	}
	storageRoot, err := ocfl.LoadStorageRoot(ctx, destFS, extensionFactory, daLogger)
	if err != nil {
		daLogger.Errorf("cannot open storage root: %v", err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		doNotClose = true
		return
	}

	exists, err := storageRoot.ObjectExists(flagObjectID)
	if err != nil {
		daLogger.Errorf("cannot check for object: %v", err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		doNotClose = true
		return
	}
	if !exists {
		fmt.Printf("Object '%s' does not exists, exiting", flagObjectID)
		doNotClose = true
		return
	}

	_, err = addObjectByPath(
		storageRoot,
		nil,
		objectExtensions,
		conf.Update.Deduplicate,
		flagObjectID,
		conf.Update.User.Name,
		conf.Update.User.Address,
		conf.Update.Message,
		sourceFS,
		area,
		areaPaths,
		conf.Update.Echo)
	if err != nil {
		daLogger.Errorf("error adding content to storageroot filesystem '%s': %v", destFS, err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		doNotClose = true
	}
	_ = showStatus(ctx)

}

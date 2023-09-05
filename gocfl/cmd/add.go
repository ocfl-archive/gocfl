package cmd

import (
	"context"
	"emperror.dev/errors"
	"fmt"
	"github.com/je4/filesystem/v2/pkg/writefs"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	"github.com/je4/gocfl/v2/pkg/subsystem/migration"
	"github.com/je4/gocfl/v2/pkg/subsystem/thumbnail"
	ironmaiden "github.com/je4/indexer/v2/pkg/indexer"
	"github.com/je4/utils/v2/pkg/checksum"
	lm "github.com/je4/utils/v2/pkg/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/exp/slices"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var addCmd = &cobra.Command{
	Use:     "add [path to ocfl structure]",
	Aliases: []string{},
	Short:   "adds new object to existing ocfl structure",
	Long:    "opens an existing ocfl structure and adds a new object. if an object with the given id already exists, an error is produced",
	Example: "gocfl add ./archive.zip /tmp/testdata -u 'Jane Doe' -a 'mailto:user@domain' -m 'initial add' -object-id 'id:abc123'",
	Args:    cobra.MinimumNArgs(2),
	Run:     doAdd,
}

// initAdd initializes the gocfl add command
func initAdd() {
	addCmd.Flags().StringVarP(&flagObjectID, "object-id", "i", "", "object id to update (required)")
	addCmd.MarkFlagRequired("object-id")
	addCmd.Flags().String("default-object-extensions", "", "folder with initial extension configurations for new OCFL objects")
	addCmd.Flags().StringP("message", "m", "", "message for new object version (required)")
	addCmd.Flags().StringP("user-name", "u", "", "user name for new object version (required)")
	addCmd.Flags().StringP("user-address", "a", "", "user address for new object version (required)")
	addCmd.Flags().StringP("fixity", "f", "", "comma separated list of digest algorithms for fixity")
	addCmd.Flags().StringP("digest", "d", "", "digest to use for ocfl checksum")
	addCmd.Flags().Bool("deduplicate", false, "force deduplication (slower)")
	addCmd.Flags().Bool("no-compress", false, "do not compress data in zip file")
}

// initAdd executes the gocfl add command
func doAdd(cmd *cobra.Command, args []string) {
	var err error

	ocflPath := filepath.ToSlash(args[0])
	srcPath := filepath.ToSlash(args[1])
	if !slices.Contains([]string{"DEBUG", "ERROR", "WARNING", "INFO", "CRITICAL"}, conf.Loglevel) {
		_ = cmd.Help()
		cobra.CheckErr(errors.Errorf("invalid log level '%s' for flag 'log-level' or 'LogLevel' config file entry", persistentFlagLoglevel))
	}
	daLogger, lf := lm.CreateLogger("ocfl", conf.Logfile, nil, conf.Loglevel, LOGFORMAT)
	defer lf.Close()

	if str := getFlagString(cmd, "fixity"); str != "" {
		parts := strings.Split(str, ",")
		for _, part := range parts {
			conf.Add.Fixity = append(conf.Add.Fixity, part)
		}
	}
	if str := getFlagString(cmd, "user-name"); str != "" {
		conf.Add.User.Name = str
	}
	if str := getFlagString(cmd, "user-address"); str != "" {
		conf.Add.User.Address = str
	}
	if str := getFlagString(cmd, "message"); str != "" {
		conf.Add.Message = str
	}
	if str := getFlagString(cmd, "default-object-extensions"); str != "" {
		conf.Add.ObjectExtensionFolder = str
	}
	if b := getFlagBool(cmd, "deduplicate"); b {
		conf.Add.Deduplicate = b
	}
	if b := getFlagBool(cmd, "no-compress"); b {
		conf.Add.NoCompress = b
	}

	var addr string
	var localCache bool

	indexerActions, err := ironmaiden.InitActionDispatcher(map[string]fs.FS{}, *conf.Indexer, daLogger)
	if err != nil {
		daLogger.Panicf("cannot init indexer: %v", err)
	}

	t := startTimer()
	defer func() { daLogger.Infof("Duration: %s", t.String()) }()

	fmt.Printf("opening '%s'\n", ocflPath)
	daLogger.Infof("opening '%s'", ocflPath)

	var fixityAlgs = []checksum.DigestAlgorithm{}
	for _, alg := range conf.Add.Fixity {
		alg = strings.TrimSpace(strings.ToLower(alg))
		if alg == "" {
			continue
		}
		if _, err := checksum.GetHash(checksum.DigestAlgorithm(alg)); err != nil {
			daLogger.Errorf("unknown hash function '%s': %v", alg, err)
			return
		}
		fixityAlgs = append(fixityAlgs, checksum.DigestAlgorithm(alg))
	}

	if _, err := os.Stat(srcPath); err != nil {
		daLogger.Errorf("cannot stat '%s': %v", srcPath, err)
		return
	}

	if str := getFlagString(cmd, "digest"); str != "" {
		conf.Add.Digest = checksum.DigestAlgorithm(str)
	}
	if _, err := checksum.GetHash(conf.Add.Digest); err != nil {
		_ = cmd.Help()
		cobra.CheckErr(errors.Errorf("invalid digest '%s' for flag 'digest' or 'Init.DigestAlgorithm' config file entry", conf.Add.Digest))
	}

	fsFactory, err := initializeFSFactory([]checksum.DigestAlgorithm{conf.Add.Digest}, nil, nil, true, false, daLogger)
	if err != nil {
		daLogger.Errorf("cannot create filesystem factory: %v", err)
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	sourceFS, err := fsFactory.Get(srcPath)
	if err != nil {
		daLogger.Errorf("cannot get filesystem for '%s': %v", srcPath, err)
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
			daLogger.Errorf("error closing filesystem '%s': %v", destFS, err)
			daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		}
	}()

	area := viper.GetString("Add.DefaultArea")
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
			daLogger.Errorf("cannot get filesystem for '%s': %v", args[i], err)
			daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
			return
		}
	}

	mig, err := migration.GetMigrations()
	if err != nil {
		daLogger.Errorf("cannot get migrations: %v", err)
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	thumb, err := thumbnail.GetThumbnails()
	if err != nil {
		daLogger.Errorf("cannot get thumbnails: %v", err)
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}
	thumb.SetSourceFS(sourceFS)

	extensionParams := GetExtensionParamValues(cmd)
	extensionFactory, err := initExtensionFactory(extensionParams, addr, localCache, indexerActions, mig, thumb, sourceFS, daLogger)
	if err != nil {
		daLogger.Errorf("cannot initialize extension factory: %v", err)
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}
	_, objectExtensions, err := initDefaultExtensions(extensionFactory, "", conf.Add.ObjectExtensionFolder)
	if err != nil {
		daLogger.Errorf("cannot initialize default extensions: %v", err)
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	ctx := ocfl.NewContextValidation(context.TODO())
	storageRoot, err := ocfl.LoadStorageRoot(ctx, destFS, extensionFactory, daLogger)
	if err != nil {
		daLogger.Errorf("cannot open storage root: %v", err)
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}
	if storageRoot.GetDigest() == "" {
		storageRoot.SetDigest(checksum.DigestAlgorithm(conf.Add.Digest))
	} else {
		if storageRoot.GetDigest() != conf.Add.Digest {
			daLogger.Errorf("storageroot already uses digest '%s' not '%s'", storageRoot.GetDigest(), conf.Add.Digest)
			return
		}
	}

	exists, err := storageRoot.ObjectExists(flagObjectID)
	if err != nil {
		daLogger.Errorf("cannot check for object: %v", err)
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}
	if exists {
		fmt.Printf("Object '%s' already exist, exiting", flagObjectID)
		return
	}

	_, err = addObjectByPath(
		storageRoot,
		fixityAlgs,
		objectExtensions,
		conf.Add.Deduplicate,
		flagObjectID,
		conf.Add.User.Name,
		conf.Add.User.Address,
		conf.Add.Message,
		sourceFS,
		area,
		areaPaths,
		false)
	if err != nil {
		daLogger.Errorf("error adding content to storageroot filesystem '%s': %v", destFS, err)
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}
	_ = showStatus(ctx)

}

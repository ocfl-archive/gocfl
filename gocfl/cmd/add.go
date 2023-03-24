package cmd

import (
	"context"
	"emperror.dev/emperror"
	"emperror.dev/errors"
	"fmt"
	"github.com/je4/gocfl/v2/pkg/indexer"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	ironmaiden "github.com/je4/indexer/v2/pkg/indexer"
	"github.com/je4/utils/v2/pkg/checksum"
	lm "github.com/je4/utils/v2/pkg/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/exp/slices"
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
	emperror.Panic(addCmd.MarkFlagRequired("object-id"))

	addCmd.Flags().String("default-object-extensions", "", "folder with initial extension configurations for new OCFL objects")
	emperror.Panic(viper.BindPFlag("Add.ObjectExtensions", addCmd.Flags().Lookup("default-object-extensions")))

	addCmd.Flags().StringP("message", "m", "", "message for new object version (required)")
	//	addCmd.MarkFlagRequired("message")
	emperror.Panic(viper.BindPFlag("Add.Message", addCmd.Flags().Lookup("message")))

	addCmd.Flags().StringP("user-name", "u", "", "user name for new object version (required)")
	//	addCmd.MarkFlagRequired("user-name")
	emperror.Panic(viper.BindPFlag("Add.UserName", addCmd.Flags().Lookup("user-name")))

	addCmd.Flags().StringP("user-address", "a", "", "user address for new object version (required)")
	//	addCmd.MarkFlagRequired("user-address")
	emperror.Panic(viper.BindPFlag("Add.UserAddress", addCmd.Flags().Lookup("user-address")))

	addCmd.Flags().StringP("fixity", "f", "", "comma separated list of digest algorithms for fixity")
	emperror.Panic(viper.BindPFlag("Add.Fixity", addCmd.Flags().Lookup("fixity")))

	addCmd.Flags().StringP("digest", "d", "", "digest to use for ocfl checksum")
	emperror.Panic(viper.BindPFlag("Add.DigestAlgorithm", addCmd.Flags().Lookup("digest")))

	addCmd.Flags().Bool("deduplicate", false, "force deduplication (slower)")
	emperror.Panic(viper.BindPFlag("Add.Deduplicate", addCmd.Flags().Lookup("deduplicate")))

	addCmd.Flags().Bool("no-compress", false, "do not compress data in zip file")
	emperror.Panic(viper.BindPFlag("Add.NoCompression", initCmd.Flags().Lookup("no-compress")))

	addCmd.Flags().Bool("encrypt-aes", false, "create encrypted container (only for container target)")
	emperror.Panic(viper.BindPFlag("Add.AES", addCmd.Flags().Lookup("encrypt-aes")))

	addCmd.Flags().String("aes-key", "", "key to use for encrypted container in hex format (64 chars, empty: generate random key)")
	emperror.Panic(viper.BindPFlag("Add.AESKey", addCmd.Flags().Lookup("aes-key")))

	addCmd.Flags().String("aes-iv", "", "initialisation vector to use for encrypted container in hex format (32 chars empty: generate random vector)")
	emperror.Panic(viper.BindPFlag("Add.AESKey", addCmd.Flags().Lookup("aes-key")))
}

// initAdd executes the gocfl add command
func doAdd(cmd *cobra.Command, args []string) {
	var err error

	notSet := []string{}
	ocflPath := filepath.ToSlash(filepath.Clean(args[0]))
	srcPath := filepath.ToSlash(filepath.Clean(args[1]))
	persistentFlagLogfile := viper.GetString("LogFile")
	persistentFlagLoglevel := strings.ToUpper(viper.GetString("LogLevel"))
	if !slices.Contains([]string{"DEBUG", "ERROR", "WARNING", "INFO", "CRITICAL"}, persistentFlagLoglevel) {
		_ = cmd.Help()
		cobra.CheckErr(errors.Errorf("invalid log level '%s' for flag 'log-level' or 'LogLevel' config file entry", persistentFlagLoglevel))
	}

	flagFixity := viper.GetString("Add.Fixity")
	flagUserName := viper.GetString("Add.UserName")
	if flagUserName == "" {
		notSet = append(notSet, "user-name")
	}
	flagUserAddress := viper.GetString("Add.UserAddress")
	if flagUserAddress == "" {
		notSet = append(notSet, "user-address")
	}
	flagMessage := viper.GetString("Add.Message")
	if flagMessage == "" {
		notSet = append(notSet, "message")
	}
	flagObjectExtensionFolder := viper.GetString("Add.ObjectExtensions")
	flagDeduplicate := viper.GetBool("Add.Deduplicate")

	if len(notSet) > 0 {
		_ = cmd.Help()
		cobra.CheckErr(errors.Errorf("required flag(s) %s not set", strings.Join(notSet, ", ")))
	}

	daLogger, lf := lm.CreateLogger("ocfl", persistentFlagLogfile, nil, persistentFlagLoglevel, LOGFORMAT)
	defer lf.Close()

	var indexerActions *ironmaiden.ActionDispatcher
	var addr string
	if viper.GetBool("Indexer.Enable") {
		siegfried, err := indexer.GetSiegfried()
		if err != nil {
			daLogger.Errorf("cannot load indexer Siegfried: %v", err)
			return
		}
		mimeRelevance, err := indexer.GetMimeRelevance()
		if err != nil {
			daLogger.Errorf("cannot load indexer MimeRelevance: %v", err)
			return
		}

		ffmpeg, err := indexer.GetFFMPEG()
		if err != nil {
			daLogger.Warningf("cannot load indexer FFMPEG: %v", err)
			//			return
		}
		imageMagick, err := indexer.GetImageMagick()
		if err != nil {
			daLogger.Warningf("cannot load indexer ImageMagick: %v", err)
			//return
		}
		tika, err := indexer.GetTika()
		if err != nil {
			daLogger.Warningf("cannot load indexer Tika: %v", err)
			//return
		}

		indexerActions, err = indexer.InitActions(mimeRelevance, siegfried, ffmpeg, imageMagick, tika, daLogger)
	}

	t := startTimer()
	defer func() { daLogger.Infof("Duration: %s", t.String()) }()

	fmt.Printf("opening '%s'\n", ocflPath)
	daLogger.Infof("opening '%s'", ocflPath)

	var fixityAlgs = []checksum.DigestAlgorithm{}
	for _, alg := range strings.Split(flagFixity, ",") {
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

	flagDigest := strings.ToLower(viper.GetString("Add.DigestAlgorithm"))
	if flagDigest == "" {
		flagDigest = "sha512"
	}
	if _, err := checksum.GetHash(checksum.DigestAlgorithm(flagDigest)); err != nil {
		_ = cmd.Help()
		cobra.CheckErr(errors.Errorf("invalid digest '%s' for flag 'digest' or 'Init.DigestAlgorithm' config file entry", flagDigest))
	}

	fsFactory, err := initializeFSFactory("Add", cmd, []checksum.DigestAlgorithm{checksum.DigestAlgorithm(flagDigest)}, daLogger)
	if err != nil {
		daLogger.Errorf("cannot create filesystem factory: %v", err)
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	sourceFS, err := fsFactory.GetFS(srcPath)
	if err != nil {
		daLogger.Errorf("cannot get filesystem for '%s': %v", srcPath, err)
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}
	destFS, err := fsFactory.GetFSRW(ocflPath, false)
	if err != nil {
		daLogger.Errorf("cannot get filesystem for '%s': %v", ocflPath, err)
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	area := viper.GetString("Add.DefaultArea")
	if area == "" {
		area = "content"
	}
	var areaPaths = map[string]ocfl.OCFLFSRead{}
	for i := 2; i < len(args); i++ {
		matches := areaPathRegexp.FindStringSubmatch(args[i])
		if matches == nil {
			daLogger.Errorf("no area given in areapath '%s'", args[i])
			continue
		}
		areaPaths[matches[1]], err = fsFactory.GetFS(matches[2])
		if err != nil {
			daLogger.Errorf("cannot get filesystem for '%s': %v", args[i], err)
			daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
			return
		}
	}

	extensionParams := GetExtensionParamValues(cmd)
	extensionFactory, err := initExtensionFactory(extensionParams, addr, indexerActions, sourceFS, daLogger)
	if err != nil {
		daLogger.Errorf("cannot initialize extension factory: %v", err)
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}
	_, objectExtensions, err := initDefaultExtensions(extensionFactory, "", flagObjectExtensionFolder, daLogger)
	if err != nil {
		daLogger.Errorf("cannot initialize default extensions: %v", err)
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	ctx := ocfl.NewContextValidation(context.TODO())
	defer showStatus(ctx)
	if !destFS.HasContent() {

	}
	storageRoot, err := ocfl.LoadStorageRoot(ctx, destFS, extensionFactory, daLogger)
	if err != nil {
		daLogger.Errorf("cannot open storage root: %v", err)
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}
	if storageRoot.GetDigest() == "" {
		storageRoot.SetDigest(checksum.DigestAlgorithm(flagDigest))
	} else {
		if storageRoot.GetDigest() != checksum.DigestAlgorithm(flagDigest) {
			daLogger.Errorf("storageroot already uses digest '%s' not '%s'", storageRoot.GetDigest(), flagDigest)
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
		flagDeduplicate,
		flagObjectID,
		flagUserName,
		flagUserAddress,
		flagMessage,
		sourceFS,
		area,
		areaPaths,
		false)
	if err != nil {
		daLogger.Errorf("error adding content to storageroot filesystem '%s': %v", destFS, err)
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
	}

	if err := destFS.Close(); err != nil {
		daLogger.Errorf("error closing filesystem '%s': %v", destFS, err)
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
	}

}

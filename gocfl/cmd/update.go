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
	emperror.Panic(updateCmd.MarkFlagRequired("object-id"))

	updateCmd.Flags().StringP("message", "m", "", "message for new object version (required)")
	emperror.Panic(updateCmd.MarkFlagRequired("message"))
	//emperror.Panic(viper.BindPFlag("Update.Message", updateCmd.Flags().Lookup("message")))

	updateCmd.Flags().StringP("user-name", "u", "", "user name for new object version (required)")
	//	updateCmd.MarkFlagRequired("user-name")
	emperror.Panic(viper.BindPFlag("Update.UserName", updateCmd.Flags().Lookup("user-name")))

	updateCmd.Flags().StringP("user-address", "a", "", "user address for new object version (required)")
	//	updateCmd.MarkFlagRequired("user-address")
	emperror.Panic(viper.BindPFlag("Update.UserAddress", updateCmd.Flags().Lookup("user-address")))

	updateCmd.Flags().StringP("digest", "d", "", "digest to use for zip file checksum")
	emperror.Panic(viper.BindPFlag("Update.DigestAlgorithm", addCmd.Flags().Lookup("digest")))

	updateCmd.Flags().Bool("no-deduplicate", false, "disable deduplication (faster)")
	emperror.Panic(viper.BindPFlag("Update.NoDeduplicate", updateCmd.Flags().Lookup("no-deduplicate")))

	updateCmd.Flags().Bool("echo", false, "update strategy 'echo' (reflects deletions). if not set, update strategy is 'contribute'")
	emperror.Panic(viper.BindPFlag("Update.Echo", updateCmd.Flags().Lookup("echo")))

	updateCmd.Flags().Bool("no-compress", false, "do not compress data in zip file")
	emperror.Panic(viper.BindPFlag("Update.NoCompression", updateCmd.Flags().Lookup("no-compress")))

	updateCmd.Flags().Bool("encrypt-aes", false, "set flag to create encrypted container (only for container target)")
	emperror.Panic(viper.BindPFlag("Update.AES", updateCmd.Flags().Lookup("encrypt-aes")))

	updateCmd.Flags().String("aes-key", "", "key to use for encrypted container in hex format (64 chars, empty: generate random key")
	emperror.Panic(viper.BindPFlag("Update.AESKey", updateCmd.Flags().Lookup("aes-key")))

	updateCmd.Flags().String("aes-iv", "", "initialisation vector to use for encrypted container in hex format (32 charsempty: generate random vector")
	emperror.Panic(viper.BindPFlag("Update.AESKey", updateCmd.Flags().Lookup("aes-key")))

}

func doUpdate(cmd *cobra.Command, args []string) {
	var err error

	notSet := []string{}
	ocflPath := filepath.ToSlash(filepath.Clean(args[0]))
	srcPath := filepath.ToSlash(filepath.Clean(args[1]))
	persistentFlagLogfile := viper.GetString("LogFile")
	persistentFlagLoglevel := strings.ToUpper(viper.GetString("LogLevel"))
	if persistentFlagLoglevel == "" {
		persistentFlagLoglevel = "INFO"
	}
	if !slices.Contains([]string{"DEBUG", "ERROR", "WARNING", "INFO", "CRITICAL"}, persistentFlagLoglevel) {
		emperror.Panic(cmd.Help())
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
	flagMessage, err := cmd.Flags().GetString("message")
	if err != nil {
		emperror.Panic(cmd.Help())
		cobra.CheckErr(errors.Wrap(err, "error getting flag 'message'"))
	}
	if flagMessage == "" {
		notSet = append(notSet, "message")
	}
	flagNoDeduplicate := viper.GetBool("Update.NoDeduplicate")
	flagEcho := viper.GetBool("Update.Echo")

	area := viper.GetString("Add.DefaultArea")
	if area == "" {
		area = "content"
	}
	if matches := areaPathRegexp.FindStringSubmatch(srcPath); matches != nil {
		area = matches[1]
		srcPath = matches[2]
	}

	if len(notSet) > 0 {
		emperror.Panic(cmd.Help())
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
			daLogger.Errorf("cannot load indexer FFMPEG: %v", err)
			return
		}
		imageMagick, err := indexer.GetImageMagick()
		if err != nil {
			daLogger.Errorf("cannot load indexer ImageMagick: %v", err)
			return
		}
		tika, err := indexer.GetTika()
		if err != nil {
			daLogger.Errorf("cannot load indexer Tika: %v", err)
			return
		}
		indexerActions, err = indexer.InitActions(mimeRelevance, siegfried, ffmpeg, imageMagick, tika, daLogger)
	}

	t := startTimer()
	defer func() { daLogger.Infof("Duration: %s", t.String()) }()

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

	fsFactory, err := initializeFSFactory("Update", cmd, []checksum.DigestAlgorithm{checksum.DigestAlgorithm(flagDigest)}, daLogger)
	if err != nil {
		daLogger.Errorf("cannot create filesystem factory: %v", err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	sourceFS, err := fsFactory.GetFS(srcPath)
	if err != nil {
		daLogger.Errorf("cannot get filesystem for '%s': %v", srcPath, err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}
	destFS, err := fsFactory.GetFSRW(ocflPath, false)
	if err != nil {
		daLogger.Errorf("cannot get filesystem for '%s': %v", ocflPath, err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	extensionParams := GetExtensionParamValues(cmd)
	extensionFactory, err := initExtensionFactory(extensionParams, addr, indexerActions, sourceFS, daLogger)
	if err != nil {
		daLogger.Errorf("cannot initialize extension factory: %v", err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}
	_, objectExtensions, err := initDefaultExtensions(extensionFactory, "", "", daLogger)
	if err != nil {
		daLogger.Errorf("cannot initialize default extensions: %v", err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	var areaPaths = map[string]ocfl.OCFLFSRead{}
	for i := 2; i < len(args); i++ {
		matches := areaPathRegexp.FindStringSubmatch(args[i])
		if matches == nil {
			continue
		}
		areaPaths[matches[1]], err = fsFactory.GetFS(matches[2])
		if err != nil {
			daLogger.Errorf("cannot get filesystem for '%s': %v", args[i], err)
			daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
			return
		}
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

	exists, err := storageRoot.ObjectExists(flagObjectID)
	if err != nil {
		daLogger.Errorf("cannot check for object: %v", err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}
	if !exists {
		fmt.Printf("Object '%s' does not exists, exiting", flagObjectID)
		return
	}

	_, err = addObjectByPath(
		storageRoot,
		fixityAlgs,
		objectExtensions,
		!flagNoDeduplicate,
		flagObjectID,
		flagUserName,
		flagUserAddress,
		flagMessage,
		sourceFS,
		area,
		areaPaths,
		flagEcho)
	if err != nil {
		daLogger.Errorf("error adding content to storageroot filesystem '%s': %v", destFS, err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
	}

	if err := destFS.Close(); err != nil {
		daLogger.Errorf("error closing filesystem '%s': %v", destFS, err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
	}
}

package cmd

import (
	"context"
	"emperror.dev/errors"
	"encoding/hex"
	"fmt"
	"github.com/je4/gocfl/v2/pkg/checksum"
	"github.com/je4/gocfl/v2/pkg/ocfl"
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
	updateCmd.MarkFlagRequired("object-id")

	updateCmd.Flags().StringP("message", "m", "", "message for new object version (required)")
	//	updateCmd.MarkFlagRequired("message")
	viper.BindPFlag("Add.Message", updateCmd.Flags().Lookup("message"))

	updateCmd.Flags().StringP("user-name", "u", "", "user name for new object version (required)")
	//	updateCmd.MarkFlagRequired("user-name")
	viper.BindPFlag("Add.UserName", updateCmd.Flags().Lookup("user-name"))

	updateCmd.Flags().StringP("user-address", "a", "", "user address for new object version (required)")
	//	updateCmd.MarkFlagRequired("user-address")
	viper.BindPFlag("Add.UserAddress", updateCmd.Flags().Lookup("user-address"))

	updateCmd.Flags().StringP("digest", "d", "", "digest to use for zip file checksum")
	viper.BindPFlag("Add.DigestAlgorithm", addCmd.Flags().Lookup("digest"))

	updateCmd.Flags().Bool("no-deduplicate", false, "disable deduplication (faster)")
	viper.BindPFlag("Update.NoDeduplicate", updateCmd.Flags().Lookup("no-deduplicate"))

	updateCmd.Flags().Bool("echo", false, "update strategy 'echo' (reflects deletions). if not set, update strategy is 'contribute'")
	viper.BindPFlag("Update.Echo", updateCmd.Flags().Lookup("echo"))

	updateCmd.Flags().Bool("encrypt-aes", false, "set flag to create encrypted container (only for container target)")
	viper.BindPFlag("Init.AES", updateCmd.Flags().Lookup("encrypt-aes"))

	updateCmd.Flags().String("aes-key", "", "key to use for encrypted container in hex format (64 chars, empty: generate random key")
	viper.BindPFlag("Init.AESKey", updateCmd.Flags().Lookup("aes-key"))

	updateCmd.Flags().String("aes-iv", "", "initialisation vector to use for encrypted container in hex format (32 charsempty: generate random vector")
	viper.BindPFlag("Init.AESKey", updateCmd.Flags().Lookup("aes-key"))

}

func doUpdate(cmd *cobra.Command, args []string) {
	notSet := []string{}
	ocflPath := filepath.ToSlash(filepath.Clean(args[0]))
	srcPath := filepath.ToSlash(filepath.Clean(args[1]))
	persistentFlagLogfile := viper.GetString("LogFile")
	persistentFlagLoglevel := strings.ToUpper(viper.GetString("LogLevel"))
	if persistentFlagLoglevel == "" {
		persistentFlagLoglevel = "INFO"
	}
	if !slices.Contains([]string{"DEBUG", "ERROR", "WARNING", "INFO", "CRITICAL"}, persistentFlagLoglevel) {
		cmd.Help()
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
	flagNoDeduplicate := viper.GetBool("Update.NoDeduplicate")
	flagEcho := viper.GetBool("Update.Echo")

	flagAddDigest := viper.GetString("Update.DigestAlgorithm")
	if _, err := checksum.GetHash(checksum.DigestAlgorithm(flagAddDigest)); err != nil {
		cmd.Help()
		cobra.CheckErr(errors.Errorf("invalid digest '%s' for flag 'digest' or 'Update.DigestAlgorithm' config file entry", flagAddDigest))
	}
	var zipAlgs = []checksum.DigestAlgorithm{checksum.DigestAlgorithm(flagAddDigest)}

	flagAES := viper.GetBool("Init.AES")
	flagAESKey := viper.GetString("Init.AESKey")
	if flagAESKey != "" && len(flagAESKey) != 64 {
		cmd.Help()
		cobra.CheckErr(errors.Errorf("invalid format '%s' for flag 'aes-key' or 'Init.AESKey' config file entry. 64 character hex value needed", flagAESKey))
	}
	var aesKey []byte
	if flagAESKey != "" {
		aesKey = make([]byte, hex.DecodedLen(len(flagAESKey)))
		if _, err := hex.Decode(aesKey, []byte(flagAESKey)); err != nil {
			aesKey = nil
			cmd.Help()
			cobra.CheckErr(errors.Errorf("invalid format '%s' for flag 'aes-key' or 'Init.AESKey' config file entry. 64 character hex value needed: %v", flagAESKey, err))
		}
	}
	flagAESIV := viper.GetString("Init.AESIV")
	if flagAESIV != "" && len(flagAESIV) != 32 {
		cmd.Help()
		cobra.CheckErr(errors.Errorf("invalid format '%s' for flag 'aes-iv' or 'Init.AESIV' config file entry. 32 character hex value needed", flagAESIV))
	}
	var aesIV []byte
	if flagAESIV != "" {
		aesIV = make([]byte, hex.DecodedLen(len(flagAESIV)))
		if _, err := hex.Decode(aesIV, []byte(flagAESIV)); err != nil {
			aesIV = nil
			cmd.Help()
			cobra.CheckErr(errors.Errorf("invalid format '%s' for flag 'aes-iv' or 'Init.AESIV' config file entry. 64 character hex value needed: %v", flagAESIV, err))
		}
	}

	area := viper.GetString("Add.DefaultArea")
	if area == "" {
		area = "content"
	}
	if matches := areaPathRegexp.FindStringSubmatch(srcPath); matches != nil {
		area = matches[1]
		srcPath = matches[2]
	}

	if len(notSet) > 0 {
		cmd.Help()
		cobra.CheckErr(errors.Errorf("required flag(s) %s not set", strings.Join(notSet, ", ")))
	}

	daLogger, lf := lm.CreateLogger("ocfl", persistentFlagLogfile, nil, persistentFlagLoglevel, LOGFORMAT)
	defer lf.Close()
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

	extensionParams := GetExtensionParamValues(cmd)
	extensionFactory, err := initExtensionFactory(extensionParams, daLogger)
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

	fsFactory, err := initializeFSFactory(zipAlgs, flagAES, aesKey, aesIV, daLogger)
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
	destFS, err := fsFactory.GetFSRW(ocflPath)
	if err != nil {
		daLogger.Errorf("cannot get filesystem for '%s': %v", ocflPath, err)
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

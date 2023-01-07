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
	"path/filepath"
	"strings"
)

var createCmd = &cobra.Command{
	Use:     "create [path to ocfl structure] [path to content folder]",
	Aliases: []string{},
	Short:   "creates a new ocfl structure with initial content of one object",
	Long: "initializes an empty ocfl structure and adds contents of a directory subtree to it\n" +
		"This command is a combination of init and add",
	Example: "gocfl create ./archive.zip /tmp/testdata --digest sha512 -u 'Jane Doe' -a 'mailto:user@domain' -m 'initial add' -object-id 'id:abc123'",
	Args:    cobra.MinimumNArgs(2),
	Run:     doCreate,
}

func initCreate() {
	createCmd.Flags().String("default-storageroot-extensions", "", "folder with initial extension configurations for new OCFL Storage Root")
	viper.BindPFlag("Init.StorageRootExtensions", createCmd.Flags().Lookup("default-storageroot-extensions"))

	createCmd.Flags().String("ocfl-version", "1.1", "ocfl version for new storage root")
	viper.BindPFlag("Init.OCFLVersion", createCmd.Flags().Lookup("ocfl-version"))

	createCmd.Flags().StringVarP(&flagObjectID, "object-id", "i", "", "object id to update (required)")
	createCmd.MarkFlagRequired("object-id")

	createCmd.Flags().String("default-object-extensions", "", "folder with initial extension configurations for new OCFL objects")
	viper.BindPFlag("Init.ObjectExtensions", createCmd.Flags().Lookup("default-object-extensions"))

	createCmd.Flags().StringP("message", "m", "", "message for new object version (required)")
	//	createCmd.MarkFlagRequired("message")
	viper.BindPFlag("Add.Message", createCmd.Flags().Lookup("message"))

	createCmd.Flags().StringP("user-name", "u", "", "user name for new object version (required)")
	//	createCmd.MarkFlagRequired("user-name")
	viper.BindPFlag("Add.UserName", createCmd.Flags().Lookup("user-name"))

	createCmd.Flags().StringP("user-address", "a", "", "user address for new object version (required)")
	//	createCmd.MarkFlagRequired("user-address")
	viper.BindPFlag("Add.UserAddress", createCmd.Flags().Lookup("user-address"))

	createCmd.Flags().StringP("fixity", "f", "", fmt.Sprintf("comma separated list of digest algorithms for fixity %v", checksum.DigestsNames))
	viper.BindPFlag("Add.Fixity", createCmd.Flags().Lookup("fixity"))

	createCmd.Flags().StringP("digest", "d", "", "digest to use for ocfl checksum")
	viper.BindPFlag("Add.DigestAlgorithm", createCmd.Flags().Lookup("digest"))
	viper.BindPFlag("Init.DigestAlgorithm", createCmd.Flags().Lookup("digest"))

	createCmd.Flags().Bool("deduplicate", false, "set flag to force deduplication (slower)")
	viper.BindPFlag("Add.Deduplicate", createCmd.Flags().Lookup("deduplicate"))

	createCmd.Flags().Bool("encrypt-aes", false, "set flag to create encrypted container (only for container target)")
	viper.BindPFlag("Init.AES", createCmd.Flags().Lookup("encrypt-aes"))

	createCmd.Flags().String("aes-key", "", "key to use for encrypted container in hex format (64 chars, empty: generate random key)")
	viper.BindPFlag("Init.AESKey", createCmd.Flags().Lookup("aes-key"))

	createCmd.Flags().String("aes-iv", "", "initialisation vector to use for encrypted container in hex format (32 char, sempty: generate random vector)")
	viper.BindPFlag("Init.AESKey", createCmd.Flags().Lookup("aes-key"))
}

func doCreate(cmd *cobra.Command, args []string) {
	notSet := []string{}
	ocflPath := filepath.ToSlash(filepath.Clean(args[0]))
	srcPath := filepath.ToSlash(filepath.Clean(args[1]))
	area := "content"
	if matches := areaPathRegexp.FindStringSubmatch(srcPath); matches != nil {
		area = matches[1]
		srcPath = matches[2]
	}
	persistentFlagLogfile := viper.GetString("LogFile")
	persistentFlagLoglevel := strings.ToUpper(viper.GetString("LogLevel"))
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
	flagStorageRootExtensionFolder := viper.GetString("Init.StorageRootExtensions")
	flagObjectExtensionFolder := viper.GetString("Add.ObjectExtensions")
	flagDeduplicate := viper.GetBool("Add.Deduplicate")

	flagInitDigest := viper.GetString("Init.DigestAlgorithm")
	if _, err := checksum.GetHash(checksum.DigestAlgorithm(flagInitDigest)); err != nil {
		cmd.Help()
		cobra.CheckErr(errors.Errorf("invalid digest '%s' for flag 'digest' or 'Init.DigestAlgorithm' config file entry", flagInitDigest))
	}

	flagAddDigest := viper.GetString("Add.DigestAlgorithm")
	if _, err := checksum.GetHash(checksum.DigestAlgorithm(flagAddDigest)); err != nil {
		cmd.Help()
		cobra.CheckErr(errors.Errorf("invalid digest '%s' for flag 'digest' or 'Add.DigestAlgorithm' config file entry", flagAddDigest))
	}
	var zipAlgs = []checksum.DigestAlgorithm{checksum.DigestAlgorithm(flagAddDigest)}

	flagVersion := viper.GetString("Init.OCFLVersion")
	if !ocfl.ValidVersion(ocfl.OCFLVersion(flagVersion)) {
		cmd.Help()
		cobra.CheckErr(errors.Errorf("invalid version '%s' for flag 'ocfl-version' or 'Init.OCFLVersion' config file entry", flagVersion))
	}

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

	if len(notSet) > 0 {
		cmd.Help()
		cobra.CheckErr(errors.Errorf("required flag(s) %s not set", strings.Join(notSet, ", ")))
	}

	daLogger, lf := lm.CreateLogger("ocfl", persistentFlagLogfile, nil, persistentFlagLoglevel, LOGFORMAT)
	defer lf.Close()

	t := startTimer()
	defer func() { daLogger.Infof("Duration: %s", t.String()) }()

	daLogger.Infof("creating '%s'", ocflPath)

	//	extensionFlags := getExtensionFlags(cmd)

	fmt.Printf("creating '%s'\n", ocflPath)

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

	extensionParams := GetExtensionParamValues(cmd)
	extensionFactory, err := initExtensionFactory(extensionParams, daLogger)
	if err != nil {
		daLogger.Errorf("cannot initialize extension factory: %v", err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}
	storageRootExtensions, objectExtensions, err := initDefaultExtensions(extensionFactory, flagStorageRootExtensionFolder, flagObjectExtensionFolder, daLogger)
	if err != nil {
		daLogger.Errorf("cannot initialize default extensions: %v", err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	ctx := ocfl.NewContextValidation(context.TODO())
	defer showStatus(ctx)
	storageRoot, err := ocfl.CreateStorageRoot(ctx,
		destFS,
		ocfl.OCFLVersion(flagVersion),
		extensionFactory,
		storageRootExtensions,
		checksum.DigestAlgorithm(flagAddDigest),
		daLogger,
	)
	if err != nil {
		destFS.Discard()
		daLogger.Errorf("cannot create new storageroot: %v", err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
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
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
	}

	if err := destFS.Close(); err != nil {
		daLogger.Errorf("error closing filesystem '%s': %v", destFS, err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
	}
}

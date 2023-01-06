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

var initCmd = &cobra.Command{
	Use:     "init [path to ocfl structure]",
	Aliases: []string{"check"},
	Short:   "initializes an empty ocfl structure",
	Long:    "initializes an empty ocfl structure",
	Example: "gocfl init ./archive.zip",
	Args:    cobra.ExactArgs(1),
	Run:     doInit,
}

func initInit() {
	initCmd.Flags().String("default-storageroot-extensions", "", "folder with initial extension configurations for new OCFL Storage Root")
	viper.BindPFlag("Init.StorageRootExtensions", initCmd.Flags().Lookup("default-storageroot-extensions"))

	initCmd.Flags().String("ocfl-version", "v", "ocfl version for new storage root")
	viper.BindPFlag("Init.OCFLVersion", initCmd.Flags().Lookup("ocfl-version"))

	initCmd.Flags().StringP("digest", "d", "", "digest to use for ocfl checksum")
	viper.BindPFlag("Init.DigestAlgorithm", initCmd.Flags().Lookup("digest"))

	initCmd.Flags().Bool("encrypt-aes", false, "set flag to create encrypted container (only for container target)")
	viper.BindPFlag("Init.AES", initCmd.Flags().Lookup("encrypt-aes"))

	initCmd.Flags().String("aes-key", "", "key to use for encrypted container in hex format (64 chars, empty: generate random key")
	viper.BindPFlag("Init.AESKey", initCmd.Flags().Lookup("aes-key"))

	initCmd.Flags().String("aes-iv", "", "initialisation vector to use for encrypted container in hex format (32 charsempty: generate random vector")
	viper.BindPFlag("Init.AESKey", initCmd.Flags().Lookup("aes-key"))
}

func doInit(cmd *cobra.Command, args []string) {
	t := startTimer()
	defer fmt.Fprintf(os.Stdout, "Duration: %s\n", t.String())

	ocflPath := filepath.ToSlash(filepath.Clean(args[0]))
	persistentFlagLogfile := viper.GetString("LogFile")

	persistentFlagLoglevel := strings.ToUpper(viper.GetString("LogLevel"))
	if !slices.Contains([]string{"DEBUG", "ERROR", "WARNING", "INFO", "CRITICAL"}, persistentFlagLoglevel) {
		cmd.Help()
		cobra.CheckErr(errors.Errorf("invalid log level '%s' for flag 'log-level' or 'LogLevel' config file entry", persistentFlagLoglevel))
	}

	flagStorageRootExtensionFolder := viper.GetString("Init.StorageRootExtensions")

	flagVersion := viper.GetString("Init.OCFLVersion")
	if !ocfl.ValidVersion(ocfl.OCFLVersion(flagVersion)) {
		cmd.Help()
		cobra.CheckErr(errors.Errorf("invalid version '%s' for flag 'ocfl-version' or 'Init.OCFLVersion' config file entry", flagVersion))
	}

	flagInitDigest := viper.GetString("Init.DigestAlgorithm")
	if _, err := checksum.GetHash(checksum.DigestAlgorithm(flagInitDigest)); err != nil {
		cmd.Help()
		cobra.CheckErr(errors.Errorf("invalid digest '%s' for flag 'digest' or 'Init.DigestAlgorithm' config file entry", flagInitDigest))
	}
	var zipAlgs = []checksum.DigestAlgorithm{checksum.DigestAlgorithm(flagInitDigest)}

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

	fmt.Printf("creating '%s'\n", ocflPath)

	daLogger, lf := lm.CreateLogger("ocfl", persistentFlagLogfile, nil, persistentFlagLoglevel, LOGFORMAT)
	defer lf.Close()
	daLogger.Infof("creating '%s'", ocflPath)

	fmt.Printf("creating '%s'\n", ocflPath)

	fsFactory, err := initializeFSFactory(zipAlgs, flagAES, aesKey, aesIV, daLogger)
	if err != nil {
		daLogger.Errorf("cannot create filesystem factory: %v", err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	destFS, err := fsFactory.GetFSRW(ocflPath)
	if err != nil {
		daLogger.Errorf("cannot get filesystem for '%s': %v", ocflPath, err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	extensionParams := GetExtensionParamValues(cmd)
	extensionFactory, err := initExtensionFactory(extensionParams, daLogger)
	if err != nil {
		daLogger.Errorf("cannot initialize extension factory: %v", err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}
	storageRootExtensions, _, err := initDefaultExtensions(extensionFactory, flagStorageRootExtensionFolder, "", daLogger)
	if err != nil {
		daLogger.Errorf("cannot initialize default extensions: %v", err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	ctx := ocfl.NewContextValidation(context.TODO())
	defer showStatus(ctx)
	if _, err := ocfl.CreateStorageRoot(ctx, destFS, ocfl.OCFLVersion(flagVersion), extensionFactory, storageRootExtensions, checksum.DigestAlgorithm(flagInitDigest), daLogger); err != nil {
		destFS.Discard()
		daLogger.Errorf("cannot create new storageroot: %v", err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	if err := destFS.Close(); err != nil {
		daLogger.Errorf("error closing filesystem '%s': %v", destFS, err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}
}

package cmd

import (
	"context"
	"emperror.dev/errors"
	"fmt"
	lm "github.com/je4/utils/v2/pkg/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.ub.unibas.ch/gocfl/v2/pkg/checksum"
	"go.ub.unibas.ch/gocfl/v2/pkg/ocfl"
	"golang.org/x/exp/slices"
	"os"
	"path/filepath"
	"strings"
)

var createCmd = &cobra.Command{
	Use:     "create [path to ocfl structure]",
	Aliases: []string{},
	Short:   "creates a new ocfl structure with initial content of one object",
	Long: "initializes an empty ocfl structure and adds contents of a directory subtree to it\n" +
		"This command is a combination of init and add",
	Example: "gocfl create ./archive.zip /tmp/testdata --sha512 -u 'Jane Doe' -a 'mailto:user@domain' -m 'initial add' -object-id 'id:abc123'",
	Args:    cobra.ExactArgs(2),
	Run:     doCreate,
}

func initCreate() {
	createCmd.Flags().String("default-storageroot-extensions", "", "folder with initial extension configurations for new OCFL Storage Root")
	viper.BindPFlag("Init.StoragerootExtensions", createCmd.Flags().Lookup("default-storageroot-extensions"))

	createCmd.Flags().String("ocfl-version", "v", "ocfl version for new storage root")
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

	createCmd.Flags().StringP("fixity", "f", "", "comma separated list of digest algorithms for fixity")
	viper.BindPFlag("Add.Fixity", createCmd.Flags().Lookup("fixity"))

	createCmd.Flags().StringP("digest", "d", "", "digest to use for ocfl checksum")
	viper.BindPFlag("Add.DigestAlgorithm", createCmd.Flags().Lookup("digest"))
	viper.BindPFlag("Init.DigestAlgorithm", createCmd.Flags().Lookup("digest"))

	createCmd.Flags().Bool("deduplicate", false, "set flag to force deduplication (slower)")
	viper.BindPFlag("Add.Deduplicate", createCmd.Flags().Lookup("deduplicate"))
}

func doCreate(cmd *cobra.Command, args []string) {
	notSet := []string{}
	ocflPath := filepath.ToSlash(filepath.Clean(args[0]))
	srcPath := filepath.ToSlash(filepath.Clean(args[1]))
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
	flagStoragerootExtensionFolder := viper.GetString("Init.StoragerootExtensions")
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

	flagVersion := viper.GetString("Init.OCFLVersion")
	if !ocfl.ValidVersion(ocfl.OCFLVersion(flagVersion)) {
		cmd.Help()
		cobra.CheckErr(errors.Errorf("invalid version '%s' for flag 'ocfl-version' or 'Init.OCFLVersion' config file entry", flagVersion))
	}

	if len(notSet) > 0 {
		cmd.Help()
		cobra.CheckErr(errors.Errorf("required flag(s) %s not set", strings.Join(notSet, ", ")))
	}

	daLogger, lf := lm.CreateLogger("ocfl", persistentFlagLogfile, nil, persistentFlagLoglevel, LOGFORMAT)
	defer lf.Close()
	daLogger.Infof("creating '%s'", ocflPath)

	extensionFlags, err := getExtensionFlags(cmd)
	if err != nil {
		daLogger.Errorf("cannot get extension flags: %v", err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

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

	if _, err := os.Stat(srcPath); err != nil {
		daLogger.Errorf("cannot stat '%s': %v", srcPath, err)
		return
	}

	finfo, err := os.Stat(ocflPath)
	if err != nil {
		if !(os.IsNotExist(err) && strings.HasSuffix(strings.ToLower(ocflPath), ".zip")) {
			daLogger.Errorf("cannot stat '%s': %v", ocflPath, err)
			daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
			return
		}
	} else {
		if strings.HasSuffix(strings.ToLower(ocflPath), ".zip") {
			daLogger.Errorf("path '%s' already exists", ocflPath)
			fmt.Printf("path '%s' already exists\n", ocflPath)
			return
		}
		if !finfo.IsDir() {
			daLogger.Errorf("'%s' is not a directory", ocflPath)
			daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
			return
		}
	}

	extensionFactory, err := ocfl.NewExtensionFactory(daLogger)
	if err != nil {
		daLogger.Errorf("cannot instantiate extension factory: %v", err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}
	if err := initExtensionFactory(extensionFactory, extensionFlags); err != nil {
		daLogger.Errorf("cannot initialize extension factory: %v", err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}
	storageRootExtensions, objectExtensions, err := initDefaultExtensions(extensionFactory, flagStoragerootExtensionFolder, flagObjectExtensionFolder, daLogger)
	if err != nil {
		daLogger.Errorf("cannot initialize default extensions: %v", err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	tempFile := fmt.Sprintf("%s.tmp", ocflPath)
	reader, writer, ocfs, err := OpenRW(ocflPath, tempFile, daLogger)
	if err != nil {
		daLogger.Errorf("cannot create target filesystem: %v", err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	ctx := ocfl.NewContextValidation(context.TODO())
	defer showStatus(ctx)
	storageRoot, err := ocfl.CreateStorageRoot(ctx, ocfs, ocfl.OCFLVersion(flagVersion), extensionFactory, storageRootExtensions, checksum.DigestAlgorithm(flagAddDigest), daLogger)
	if err != nil {
		ocfs.Discard()
		daLogger.Errorf("cannot create new storageroot: %v", err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	_, err = addObjectByPath(storageRoot, fixityAlgs, objectExtensions, flagDeduplicate, flagObjectID, flagUserName, flagUserAddress, flagMessage, srcPath, false)
	if err != nil {
		daLogger.Errorf("error adding content to storageroot filesystem '%s': %v", ocfs, err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
	}

	if err := ocfs.Close(); err != nil {
		daLogger.Errorf("error closing filesystem '%s': %v", ocfs, err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
	} else {
		if reader != nil && reader != (*os.File)(nil) {
			if err := reader.Close(); err != nil {
				daLogger.Errorf("error closing reader: %v", err)
				daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
			}
		}
		if err := writer.Close(); err != nil {
			daLogger.Errorf("error closing writer: %v", err)
			daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		}
		if err := os.Rename(tempFile, ocflPath); err != nil {
			daLogger.Errorf("cannot rename '%s' -> '%s': %v", tempFile, ocflPath, err)
		}
	}

}

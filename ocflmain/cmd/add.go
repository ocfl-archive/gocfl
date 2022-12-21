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

var addCmd = &cobra.Command{
	Use:     "add [path to ocfl structure]",
	Aliases: []string{},
	Short:   "adds new object to existing ocfl structure",
	Long:    "opens an existing ocfl structure and adds a new object. if an object with the given id already exists, an error is produced",
	Example: "gocfl add ./archive.zip /tmp/testdata -u 'Jane Doe' -a 'mailto:user@domain' -m 'initial add' -object-id 'id:abc123'",
	Args:    cobra.ExactArgs(2),
	Run:     doAdd,
}

func initAdd() {
	addCmd.Flags().StringVarP(&flagObjectID, "object-id", "i", "", "object id to update (required)")
	addCmd.MarkFlagRequired("object-id")

	addCmd.Flags().String("default-object-extensions", "", "folder with initial extension configurations for new OCFL objects")
	viper.BindPFlag("Init.ObjectExtensions", addCmd.Flags().Lookup("default-object-extensions"))

	addCmd.Flags().StringP("message", "m", "", "message for new object version (required)")
	//	addCmd.MarkFlagRequired("message")
	viper.BindPFlag("Add.Message", addCmd.Flags().Lookup("message"))

	addCmd.Flags().StringP("user-name", "u", "", "user name for new object version (required)")
	//	addCmd.MarkFlagRequired("user-name")
	viper.BindPFlag("Add.UserName", addCmd.Flags().Lookup("user-name"))

	addCmd.Flags().StringP("user-address", "a", "", "user address for new object version (required)")
	//	addCmd.MarkFlagRequired("user-address")
	viper.BindPFlag("Add.UserAddress", addCmd.Flags().Lookup("user-address"))

	addCmd.Flags().StringP("fixity", "f", "", "comma separated list of digest algorithms for fixity")
	viper.BindPFlag("Add.Fixity", addCmd.Flags().Lookup("fixity"))

	addCmd.Flags().StringP("digest", "d", "", "digest to use for ocfl checksum")
	viper.BindPFlag("Add.DigestAlgorithm", addCmd.Flags().Lookup("digest"))

	addCmd.Flags().Bool("deduplicate", false, "set flag to force deduplication (slower)")
	viper.BindPFlag("Add.Deduplicate", addCmd.Flags().Lookup("deduplicate"))
}

func doAdd(cmd *cobra.Command, args []string) {
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
	flagObjectExtensionFolder := viper.GetString("Add.ObjectExtensions")
	flagDeduplicate := viper.GetBool("Add.Deduplicate")

	flagAddDigest := viper.GetString("Add.DigestAlgorithm")
	if flagAddDigest != "" {
		if _, err := checksum.GetHash(checksum.DigestAlgorithm(flagAddDigest)); err != nil {
			cmd.Help()
			cobra.CheckErr(errors.Errorf("invalid digest '%s' for flag 'digest' or 'Add.DigestAlgorithm' config file entry", flagAddDigest))
		}
	}

	if len(notSet) > 0 {
		cmd.Help()
		cobra.CheckErr(errors.Errorf("required flag(s) %s not set", strings.Join(notSet, ", ")))
	}

	daLogger, lf := lm.CreateLogger("ocfl", persistentFlagLogfile, nil, persistentFlagLoglevel, LOGFORMAT)
	defer lf.Close()

	extensionFlags, err := getExtensionFlags(cmd)
	if err != nil {
		daLogger.Errorf("cannot get extension flags: %v", err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

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

	_, err = os.Stat(ocflPath)
	if err != nil {
		if !(os.IsNotExist(err) && strings.HasSuffix(strings.ToLower(ocflPath), ".zip")) {
			daLogger.Errorf("cannot stat '%s': %v", ocflPath, err)
			daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
			return
		}
		daLogger.Errorf("path '%s' does not exist", ocflPath)
		fmt.Printf("path '%s' does not exists\n", ocflPath)
		return
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
	_, objectExtensions, err := initDefaultExtensions(extensionFactory, "", flagObjectExtensionFolder, daLogger)
	if err != nil {
		daLogger.Errorf("cannot initialize default extensions: %v", err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	tempFile := fmt.Sprintf("%s.tmp", ocflPath)
	reader, writer, ocfs, tempActive, err := OpenRW(ocflPath, tempFile, daLogger)
	if err != nil {
		daLogger.Errorf("cannot open target filesystem: %v", err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	ctx := ocfl.NewContextValidation(context.TODO())
	defer showStatus(ctx)
	if !ocfs.HasContent() {

	}
	storageRoot, err := ocfl.LoadStorageRoot(ctx, ocfs, extensionFactory, daLogger)
	if err != nil {
		daLogger.Errorf("cannot open storage root: %v", err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}
	if storageRoot.GetDigest() == "" {
		storageRoot.SetDigest(checksum.DigestAlgorithm(flagAddDigest))
	} else {
		if storageRoot.GetDigest() != checksum.DigestAlgorithm(flagAddDigest) {
			daLogger.Errorf("storageroot already uses digest '%s' not '%s'", storageRoot.GetDigest(), flagAddDigest)
			return
		}
	}

	exists, err := storageRoot.ObjectExists(flagObjectID)
	if err != nil {
		daLogger.Errorf("cannot check for object: %v", err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}
	if exists {
		fmt.Printf("Object '%s' already exists, exiting", flagObjectID)
		return
	}

	_, err = addObjectByPath(storageRoot, fixityAlgs, objectExtensions, flagDeduplicate, flagObjectID, flagUserName, flagUserAddress, flagMessage, srcPath, "", nil, false)
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
		if tempActive {
			if err := os.Rename(tempFile, ocflPath); err != nil {
				daLogger.Errorf("cannot rename '%s' -> '%s': %v", tempFile, ocflPath, err)
			}
		}
	}

}

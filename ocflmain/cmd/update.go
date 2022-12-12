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

	updateCmd.Flags().Bool("no-deduplicate", false, "set flag to disable deduplication (faster)")
	viper.BindPFlag("Update.NoDeduplicate", updateCmd.Flags().Lookup("no-deduplicate"))

	updateCmd.Flags().Bool("echo", false, "set flag to update strategy 'echo' (reflects deletions). if not set, update strategy is 'contribute'")
	viper.BindPFlag("Update.Echo", updateCmd.Flags().Lookup("echo"))
}

func doUpdate(cmd *cobra.Command, args []string) {
	notSet := []string{}
	ocflPath := filepath.ToSlash(filepath.Clean(args[0]))
	srcPath := filepath.ToSlash(filepath.Clean(args[1]))
	area := "content"
	if matches := areaPathRegexp.FindStringSubmatch(srcPath); matches != nil {
		area = matches[1]
		srcPath = matches[2]
	}
	var areaPaths = map[string]string{}
	for i := 2; i < len(args); i++ {
		matches := areaPathRegexp.FindStringSubmatch(args[i])
		if matches == nil {
			continue
		}
		areaPaths[matches[1]] = matches[2]
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
	flagNoDeduplicate := viper.GetBool("Update.NoDeduplicate")
	flagEcho := viper.GetBool("Update.Echo")

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
	_, objectExtensions, err := initDefaultExtensions(extensionFactory, "", "", daLogger)
	if err != nil {
		daLogger.Errorf("cannot initialize default extensions: %v", err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	tempFile := fmt.Sprintf("%s.tmp", ocflPath)
	reader, writer, ocfs, err := OpenRW(ocflPath, tempFile, daLogger)
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
		srcPath,
		area,
		areaPaths,
		flagEcho)
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

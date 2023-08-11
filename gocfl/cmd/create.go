package cmd

import (
	"context"
	"emperror.dev/emperror"
	"emperror.dev/errors"
	"fmt"
	"github.com/je4/filesystem/v2/pkg/writefs"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	indexer2 "github.com/je4/gocfl/v2/pkg/subsystem/indexer"
	"github.com/je4/gocfl/v2/pkg/subsystem/migration"
	"github.com/je4/gocfl/v2/pkg/subsystem/thumbnail"
	ironmaiden "github.com/je4/indexer/v2/pkg/indexer"
	"github.com/je4/utils/v2/pkg/checksum"
	lm "github.com/je4/utils/v2/pkg/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/exp/slices"
	"io/fs"
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

// initCreate initializes the gocfl create command
func initCreate() {
	createCmd.Flags().String("default-storageroot-extensions", "", "folder with initial extension configurations for new OCFL Storage Root")
	emperror.Panic(viper.BindPFlag("Create.StorageRootExtensions", createCmd.Flags().Lookup("default-storageroot-extensions")))

	createCmd.Flags().String("ocfl-version", "1.1", "ocfl version for new storage root")
	emperror.Panic(viper.BindPFlag("Create.OCFLVersion", createCmd.Flags().Lookup("ocfl-version")))

	createCmd.Flags().StringVarP(&flagObjectID, "object-id", "i", "", "object id to update (required)")
	emperror.Panic(createCmd.MarkFlagRequired("object-id"))

	createCmd.Flags().String("default-object-extensions", "", "folder with initial extension configurations for new OCFL objects")
	emperror.Panic(viper.BindPFlag("Create.ObjectExtensions", createCmd.Flags().Lookup("default-object-extensions")))

	createCmd.Flags().StringP("message", "m", "", "message for new object version (required)")
	//	createCmd.MarkFlagRequired("message")
	emperror.Panic(viper.BindPFlag("Create.Message", createCmd.Flags().Lookup("message")))

	createCmd.Flags().StringP("user-name", "u", "", "user name for new object version (required)")
	//	createCmd.MarkFlagRequired("user-name")
	emperror.Panic(viper.BindPFlag("Create.UserName", createCmd.Flags().Lookup("user-name")))

	createCmd.Flags().StringP("user-address", "a", "", "user address for new object version (required)")
	//	createCmd.MarkFlagRequired("user-address")
	emperror.Panic(viper.BindPFlag("Create.UserAddress", createCmd.Flags().Lookup("user-address")))

	createCmd.Flags().StringP("fixity", "f", "", fmt.Sprintf("comma separated list of digest algorithms for fixity %v", checksum.DigestNames))
	emperror.Panic(viper.BindPFlag("Create.Fixity", createCmd.Flags().Lookup("fixity")))

	createCmd.Flags().StringP("digest", "d", "", "digest to use for ocfl checksum")
	emperror.Panic(viper.BindPFlag("Create.DigestAlgorithm", createCmd.Flags().Lookup("digest")))

	createCmd.Flags().String("default-area", "", "default area for update or ingest (default: content)")
	emperror.Panic(viper.BindPFlag("Create.DefaultArea", createCmd.Flags().Lookup("default-area")))

	createCmd.Flags().Bool("deduplicate", false, "force deduplication (slower)")
	emperror.Panic(viper.BindPFlag("Create.Deduplicate", createCmd.Flags().Lookup("deduplicate")))

	createCmd.Flags().Bool("no-compress", false, "do not compress data in zip file")
	emperror.Panic(viper.BindPFlag("Create.NoCompression", createCmd.Flags().Lookup("no-compress")))

	createCmd.Flags().Bool("encrypt-aes", false, "create encrypted container (only for container target)")
	emperror.Panic(viper.BindPFlag("Create.AES", createCmd.Flags().Lookup("encrypt-aes")))

	createCmd.Flags().String("aes-key", "", "key to use for encrypted container in hex format (64 chars, empty: generate random key)")
	emperror.Panic(viper.BindPFlag("Create.AESKey", createCmd.Flags().Lookup("aes-key")))

	createCmd.Flags().String("aes-iv", "", "initialisation vector to use for encrypted container in hex format (32 char, sempty: generate random vector)")
	emperror.Panic(viper.BindPFlag("Create.AESKey", createCmd.Flags().Lookup("aes-key")))

	createCmd.Flags().String("keypass-file", "", "file with keypass2 database")
	emperror.Panic(viper.BindPFlag("Create.KeyPassFile", createCmd.Flags().Lookup("keypass-file")))

	createCmd.Flags().String("keypass-entry", "", "keypass2 entry to use for key encryption")
	emperror.Panic(viper.BindPFlag("Create.KeyPassEntry", createCmd.Flags().Lookup("keypass-entry")))

	createCmd.Flags().String("keypass-key", "", "key to use for keypass2 database decryption")
	emperror.Panic(viper.BindPFlag("Create.KeyPassKey", createCmd.Flags().Lookup("keypass-key")))

	//createCmd.Flags().Bool("force", false, "force overwrite of existing files")
}

// initCreate executes the gocfl create command
func doCreate(cmd *cobra.Command, args []string) {
	var err error
	notSet := []string{}
	ocflPath := filepath.ToSlash(args[0])
	srcPath := filepath.ToSlash(args[1])
	persistentFlagLogfile := viper.GetString("LogFile")
	persistentFlagLoglevel := strings.ToUpper(viper.GetString("LogLevel"))
	if !slices.Contains([]string{"DEBUG", "ERROR", "WARNING", "INFO", "CRITICAL"}, persistentFlagLoglevel) {
		_ = cmd.Help()
		cobra.CheckErr(errors.Errorf("invalid log level '%s' for flag 'log-level' or 'LogLevel' config file entry", persistentFlagLoglevel))
	}

	daLogger, lf := lm.CreateLogger("ocfl", persistentFlagLogfile, nil, persistentFlagLoglevel, LOGFORMAT)
	defer lf.Close()

	flagUserName := viper.GetString("Create.UserName")
	if flagUserName == "" {
		notSet = append(notSet, "user-name")
	}
	flagUserAddress := viper.GetString("Create.UserAddress")
	if flagUserAddress == "" {
		notSet = append(notSet, "user-address")
	}
	flagMessage := viper.GetString("Create.Message")
	if flagMessage == "" {
		notSet = append(notSet, "message")
	}
	flagStorageRootExtensionFolder := viper.GetString("Create.StorageRootExtensions")
	flagObjectExtensionFolder := viper.GetString("Create.ObjectExtensions")
	flagDeduplicate := viper.GetBool("Create.Deduplicate")

	flagInitDigest := viper.GetString("Create.DigestAlgorithm")
	if _, err := checksum.GetHash(checksum.DigestAlgorithm(flagInitDigest)); err != nil {
		_ = cmd.Help()
		cobra.CheckErr(errors.Errorf("invalid digest '%s' for flag 'digest' or 'Create.DigestAlgorithm' config file entry", flagInitDigest))
	}

	flagAddDigest := viper.GetString("Create.DigestAlgorithm")
	if _, err := checksum.GetHash(checksum.DigestAlgorithm(flagAddDigest)); err != nil {
		_ = cmd.Help()
		cobra.CheckErr(errors.Errorf("invalid digest '%s' for flag 'digest' or 'Create.DigestAlgorithm' config file entry", flagAddDigest))
	}

	area := viper.GetString("Create.DefaultArea")
	if area == "" {
		area = "content"
	}
	if matches := areaPathRegexp.FindStringSubmatch(srcPath); matches != nil {
		area = matches[1]
		srcPath = matches[2]
	}
	daLogger.Infof("source path '%s:%s'", area, srcPath)

	if len(notSet) > 0 {
		_ = cmd.Help()
		cobra.CheckErr(errors.Errorf("required flag(s) %s not set", strings.Join(notSet, ", ")))
	}

	var indexerActions *ironmaiden.ActionDispatcher
	var addr string
	var localCache bool
	if viper.GetBool("Indexer.Enable") {
		localCache = viper.GetBool("Indexer.LocalCache")
		siegfried, err := indexer2.GetSiegfried()
		if err != nil {
			daLogger.Warningf("cannot load indexer Siegfried: %v", err)
			//return
		}
		mimeRelevance, err := indexer2.GetMimeRelevance()
		if err != nil {
			daLogger.Warningf("cannot load indexer MimeRelevance: %v", err)
			// return
		}

		ffmpeg, err := indexer2.GetFFMPEG()
		if err != nil {
			daLogger.Warningf("cannot load indexer FFMPEG: %v", err)
			//			return
		}
		imageMagick, err := indexer2.GetImageMagick()
		if err != nil {
			daLogger.Warningf("cannot load indexer ImageMagick: %v", err)
			//return
		}
		tika, err := indexer2.GetTika()
		if err != nil {
			daLogger.Warningf("cannot load indexer Tika: %v", err)
			//return
		}

		indexerActions, err = indexer2.InitActions(mimeRelevance, siegfried, ffmpeg, imageMagick, tika, daLogger)

	}

	t := startTimer()
	defer func() { daLogger.Infof("Duration: %s", t.String()) }()

	daLogger.Infof("creating '%s'", ocflPath)

	//	extensionFlags := getExtensionFlags(cmd)

	fmt.Printf("creating '%s'\n", ocflPath)

	var fixityAlgs = []checksum.DigestAlgorithm{}
	flagFixity := viper.GetString("Create.Fixity")
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

	flagDigest := strings.ToLower(viper.GetString("Add.DigestAlgorithm"))
	if flagDigest == "" {
		flagDigest = "sha512"
	}
	if _, err := checksum.GetHash(checksum.DigestAlgorithm(flagDigest)); err != nil {
		_ = cmd.Help()
		cobra.CheckErr(errors.Errorf("invalid digest '%s' for flag 'digest' or 'Init.DigestAlgorithm' config file entry", flagDigest))
	}

	fsFactory, err := initializeFSFactory("Create", cmd, []checksum.DigestAlgorithm{checksum.DigestAlgorithm(flagDigest)}, false, daLogger)
	if err != nil {
		daLogger.Errorf("cannot create filesystem factory: %v", err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	sourceFS, err := fsFactory.Get(srcPath)
	if err != nil {
		daLogger.Errorf("cannot get filesystem for '%s': %v", srcPath, err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	destFS, err := fsFactory.Get(ocflPath)
	if err != nil {
		daLogger.Errorf("cannot get filesystem for '%s': %v", ocflPath, err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}
	defer func() {
		if err := writefs.Close(destFS); err != nil {
			daLogger.Errorf("error closing filesystem '%s': %v", destFS, err)
			daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		}
	}()

	var areaPaths = map[string]fs.FS{}
	for i := 2; i < len(args); i++ {
		matches := areaPathRegexp.FindStringSubmatch(args[i])
		if matches == nil {
			daLogger.Warningf("no area prefix for '%s'", args[i])
			continue
		}
		daLogger.Infof("additional path '%s:%s'", matches[1], matches[2])
		areaPaths[matches[1]], err = fsFactory.Get(matches[2])
		if err != nil {
			daLogger.Errorf("cannot get filesystem for '%s': %v", args[i], err)
			daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
			return
		}
	}

	mig, err := migration.GetMigrations()
	if err != nil {
		daLogger.Errorf("cannot get migrations: %v", err)
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}
	mig.SetSourceFS(sourceFS)

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

	storageRootExtensions, objectExtensions, err := initDefaultExtensions(extensionFactory, flagStorageRootExtensionFolder, flagObjectExtensionFolder)
	if err != nil {
		daLogger.Errorf("cannot initialize default extensions: %v", err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	flagVersion := viper.GetString("Init.OCFLVersion")
	if !ocfl.ValidVersion(ocfl.OCFLVersion(flagVersion)) {
		_ = cmd.Help()
		cobra.CheckErr(errors.Errorf("invalid version '%s' for flag 'ocfl-version' or 'Init.OCFLVersion' config file entry", flagVersion))
	}

	ctx := ocfl.NewContextValidation(context.TODO())
	storageRoot, err := ocfl.CreateStorageRoot(ctx, destFS, ocfl.OCFLVersion(flagVersion), extensionFactory, storageRootExtensions, checksum.DigestAlgorithm(flagAddDigest), daLogger)
	if err != nil {
		if err := writefs.Close(destFS); err != nil {
			daLogger.Errorf("cannot discard filesystem '%v': %v", destFS, err)
		}
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
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	_ = showStatus(ctx)
}

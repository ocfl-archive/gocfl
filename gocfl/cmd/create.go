package cmd

import (
	"context"
	"fmt"
	"github.com/je4/filesystem/v2/pkg/writefs"
	"github.com/je4/gocfl/v2/internal"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	"github.com/je4/gocfl/v2/pkg/subsystem/migration"
	"github.com/je4/gocfl/v2/pkg/subsystem/thumbnail"
	ironmaiden "github.com/je4/indexer/v2/pkg/indexer"
	"github.com/je4/utils/v2/pkg/checksum"
	lm "github.com/je4/utils/v2/pkg/logger"
	"github.com/spf13/cobra"
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
	createCmd.Flags().String("ocfl-version", "1.1", "ocfl version for new storage root")
	createCmd.Flags().StringVarP(&flagObjectID, "object-id", "i", "", "object id to update (required)")
	createCmd.MarkFlagRequired("object-id")
	createCmd.Flags().String("default-object-extensions", "", "folder with initial extension configurations for new OCFL objects")
	createCmd.Flags().StringP("message", "m", "", "message for new object version (required)")
	createCmd.Flags().StringP("user-name", "u", "", "user name for new object version (required)")
	createCmd.Flags().StringP("user-address", "a", "", "user address for new object version (required)")
	createCmd.Flags().StringP("fixity", "f", "", fmt.Sprintf("comma separated list of digest algorithms for fixity %v", checksum.DigestNames))
	createCmd.Flags().StringP("digest", "d", "", "digest to use for ocfl checksum")
	createCmd.Flags().String("default-area", "", "default area for update or ingest (default: content)")
	createCmd.Flags().Bool("deduplicate", false, "force deduplication (slower)")
	createCmd.Flags().Bool("no-compress", false, "do not compress data in zip file")
	createCmd.Flags().Bool("encrypt-aes", false, "create encrypted container (only for container target)")
	createCmd.Flags().String("aes-key", "", "key to use for encrypted container in hex format (64 chars, empty: generate random key)")
	createCmd.Flags().String("aes-iv", "", "initialisation vector to use for encrypted container in hex format (32 char, sempty: generate random vector)")
	createCmd.Flags().String("keypass-file", "", "file with keypass2 database")
	createCmd.Flags().String("keypass-entry", "", "keypass2 entry to use for key encryption")
	createCmd.Flags().String("keypass-key", "", "key to use for keypass2 database decryption")
}

// initCreate executes the gocfl create command
func doCreate(cmd *cobra.Command, args []string) {
	var err error

	if err := cmd.ValidateRequiredFlags(); err != nil {
		cobra.CheckErr(err)
		return
	}

	ocflPath := filepath.ToSlash(args[0])
	srcPath := filepath.ToSlash(args[1])

	daLogger, lf := lm.CreateLogger("ocfl", persistentFlagLogfile, nil, conf.LogLevel, conf.LogFormat)
	defer lf.Close()

	doInitConf(cmd)
	doAddConf(cmd)

	var addr string
	var localCache bool

	var fss = map[string]fs.FS{"internal": internal.InternalFS}

	indexerActions, err := ironmaiden.InitActionDispatcher(fss, *conf.Indexer, daLogger)
	if err != nil {
		daLogger.Panicf("cannot init indexer: %v", err)
	}

	t := startTimer()
	defer func() { daLogger.Infof("Duration: %s", t.String()) }()

	daLogger.Infof("creating '%s'", ocflPath)

	//	extensionFlags := getExtensionFlags(cmd)

	fmt.Printf("creating '%s'\n", ocflPath)

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

	fsFactory, err := initializeFSFactory([]checksum.DigestAlgorithm{conf.Init.Digest}, conf.AES, conf.S3, true, false, daLogger)
	if err != nil {
		daLogger.Errorf("cannot create filesystem factory: %v", err)
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	sourceFS, err := fsFactory.Get(srcPath)
	if err != nil {
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		daLogger.Panicf("cannot get filesystem for '%s': %v", srcPath, err)
	}
	destFS, err := fsFactory.Get(ocflPath)
	if err != nil {
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		daLogger.Panicf("cannot get filesystem for '%s': %v", ocflPath, err)
	}
	defer func() {
		if err := writefs.Close(destFS); err != nil {
			daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
			daLogger.Panicf("error closing filesystem '%s': %v", destFS, err)
		}
	}()

	area := conf.DefaultArea
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
			daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
			daLogger.Panicf("cannot get filesystem for '%s': %v", args[i], err)
		}
	}

	mig, err := migration.GetMigrations(conf)
	if err != nil {
		daLogger.Errorf("cannot get migrations: %v", err)
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}
	mig.SetSourceFS(sourceFS)

	thumb, err := thumbnail.GetThumbnails(conf)
	if err != nil {
		daLogger.Errorf("cannot get thumbnails: %v", err)
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}
	thumb.SetSourceFS(sourceFS)

	extensionParams := GetExtensionParamValues(cmd, conf)
	extensionFactory, err := initExtensionFactory(extensionParams, addr, localCache, indexerActions, mig, thumb, sourceFS, daLogger)
	if err != nil {
		daLogger.Errorf("cannot initialize extension factory: %v", err)
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	storageRootExtensions, objectExtensions, err := initDefaultExtensions(extensionFactory, conf.Init.StorageRootExtensionFolder, conf.Add.ObjectExtensionFolder)
	if err != nil {
		daLogger.Errorf("cannot initialize default extensions: %v", err)
		daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		return
	}

	ctx := ocfl.NewContextValidation(context.TODO())
	storageRoot, err := ocfl.CreateStorageRoot(
		ctx,
		destFS,
		ocfl.OCFLVersion(conf.Init.OCFLVersion),
		extensionFactory,
		storageRootExtensions,
		conf.Init.Digest,
		daLogger,
	)
	if err != nil {
		if err := writefs.Close(destFS); err != nil {
			daLogger.Errorf("cannot discard filesystem '%v': %v", destFS, err)
		}
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		daLogger.Panicf("cannot create new storageroot: %v", err)
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
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
		daLogger.Panicf("error adding content to storageroot filesystem '%s': %v", destFS, err)
	}
	_ = showStatus(ctx)

}

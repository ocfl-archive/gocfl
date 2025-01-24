package cmd

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"strings"

	"github.com/je4/filesystem/v3/pkg/writefs"
	ironmaiden "github.com/je4/indexer/v3/pkg/indexer"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v2/internal"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl"
	"github.com/ocfl-archive/gocfl/v2/pkg/subsystem/migration"
	"github.com/ocfl-archive/gocfl/v2/pkg/subsystem/thumbnail"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/cobra"
	ublogger "gitlab.switch.ch/ub-unibas/go-ublogger/v2"
	"go.ub.unibas.ch/cloud/certloader/v2/pkg/loader"
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

func isEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}

// initCreate executes the gocfl create command
func doCreate(cmd *cobra.Command, args []string) {
	var err error

	if err := cmd.ValidateRequiredFlags(); err != nil {
		cobra.CheckErr(err)
		return
	}

	ocflPath, err := ocfl.Fullpath(args[0])
	if err != nil {
		cobra.CheckErr(err)
		return
	}
	srcPath, err := ocfl.Fullpath(args[1])
	if err != nil {
		cobra.CheckErr(err)
		return
	}

	// create logger instance
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("cannot get hostname: %v", err)
	}

	var loggerTLSConfig *tls.Config
	var loggerLoader io.Closer
	if conf.Log.Stash.TLS != nil {
		loggerTLSConfig, loggerLoader, err = loader.CreateClientLoader(conf.Log.Stash.TLS, nil)
		if err != nil {
			log.Fatalf("cannot create client loader: %v", err)
		}
		defer loggerLoader.Close()
	}

	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	_logger, _logstash, _logfile, err := ublogger.CreateUbMultiLoggerTLS(conf.Log.Level, conf.Log.File,
		ublogger.SetDataset(conf.Log.Stash.Dataset),
		ublogger.SetLogStash(conf.Log.Stash.LogstashHost, conf.Log.Stash.LogstashPort, conf.Log.Stash.Namespace, conf.Log.Stash.LogstashTraceLevel),
		ublogger.SetTLS(conf.Log.Stash.TLS != nil),
		ublogger.SetTLSConfig(loggerTLSConfig),
	)
	if err != nil {
		log.Fatalf("cannot create logger: %v", err)
	}
	if _logstash != nil {
		defer _logstash.Close()
	}

	if _logfile != nil {
		defer _logfile.Close()
	}

	l2 := _logger.With().Timestamp().Str("host", hostname).Logger() //.Output(output)
	var logger zLogger.ZLogger = &l2

	doInitConf(cmd)
	doAddConf(cmd)

	var addr string
	var localCache bool

	var fss = map[string]fs.FS{"internal": internal.InternalFS}

	indexerActions, err := ironmaiden.InitActionDispatcher(fss, *conf.Indexer, logger)
	if err != nil {
		logger.Panic().Stack().Err(err).Msg("cannot init indexer")
	}

	t := startTimer()
	defer func() {
		logger.Info().Any(
			ErrorFactory.LogError(
				ErrorGOCFL,
				fmt.Sprintf("duration: %s", t.String()),
				err,
			),
		).Msg("")
	}()

	var fixityAlgs = []checksum.DigestAlgorithm{}
	for _, alg := range conf.Add.Fixity {
		alg = strings.TrimSpace(strings.ToLower(alg))
		if alg == "" {
			continue
		}
		if _, err := checksum.GetHash(checksum.DigestAlgorithm(alg)); err != nil {
			logger.Error().Any(
				ErrorFactory.LogError(
					ErrorExtensionInitErr,
					fmt.Sprintf("unknown hash function '%s'", alg),
					err,
				)).Msg("")
			return
		}
		fixityAlgs = append(fixityAlgs, checksum.DigestAlgorithm(alg))
	}

	fsFactory, err := initializeFSFactory([]checksum.DigestAlgorithm{conf.Init.Digest}, conf.AES, conf.S3, conf.Add.NoCompress, false, logger)
	if err != nil {
		logger.Error().Stack().Any(
			ErrorFactory.LogError(
				ErrorFS, "cannot create filesystem factory", err,
			)).Msg("")
		return
	}

	if fi, err := os.Stat(ocflPath); err == nil {
		if fi.IsDir() {
			if empty, err := isEmpty(ocflPath); err != nil {
				logger.Error().Stack().Any(
					ErrorFactory.LogError(
						ErrorOCFLCreation,
						fmt.Sprintf("cannot check if directory '%s' is empty", ocflPath),
						err,
					)).Msg("")
			} else if !empty {
				logger.Error().Any(
					ErrorFactory.LogError(
						ErrorOCFLCreation,
						fmt.Sprintf("directory '%s' is not empty", ocflPath),
						nil,
					)).Msg("")
				return
			}
		} else {
			logger.Error().Any(
				ErrorFactory.LogError(
					ErrorOCFLCreation,
					fmt.Sprintf("'%s' already exists and is not an empty directory", ocflPath),
					nil,
				)).Msg("")
			return
		}
	}

	sourceFS, err := fsFactory.Get(srcPath, true)
	if err != nil {
		logger.Panic().Stack().Err(err).Msgf("cannot get filesystem for '%s'", srcPath)
	}
	destFS, err := fsFactory.Get(ocflPath, false)
	if err != nil {
		logger.Panic().Stack().Msgf("cannot get filesystem for '%s'", ocflPath)
	}
	defer func() {
		if err := writefs.Close(destFS); err != nil {
			logger.Panic().Stack().Any(
				ErrorFactory.LogError(
					ErrorFS,
					fmt.Sprintf("error closing filesystem '%s'", destFS),
					err,
				),
			).Msg("")
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
			logger.Error().Any(
				ErrorFactory.LogError(
					ErrorFS,
					fmt.Sprintf("no area given in areapath '%s'", args[i]),
					nil,
				),
			).Msg("")
			continue
		}
		path, err := ocfl.Fullpath(matches[2])
		if err != nil {
			logger.Panic().Stack().Err(err).Msgf("cannot get fullpath for '%s'", matches[2])
		}
		areaPaths[matches[1]], err = fsFactory.Get(path, true)
		if err != nil {
			logger.Panic().Stack().Err(err).Msgf("cannot get filesystem for '%s'", args[i])
		}
	}

	mig, err := migration.GetMigrations(conf)
	if err != nil {
		logger.Error().Stack().Err(err).Any(
			ErrorFactory.LogError(
				ErrorExtensionInit, "cannot get migrations", err,
			),
		).Msg("")
		return
	}
	mig.SetSourceFS(sourceFS)

	thumb, err := thumbnail.GetThumbnails(conf)
	if err != nil {
		logger.Error().Stack().Any(
			ErrorFactory.LogError(
				ErrorExtensionInit, "cannot get thumbnails", err,
			),
		).Msg("")
		return
	}
	thumb.SetSourceFS(sourceFS)

	extensionParams := GetExtensionParamValues(cmd, conf)
	extensionFactory, err := InitExtensionFactory(extensionParams, addr, localCache, indexerActions, mig, thumb, sourceFS, logger)
	if err != nil {
		logger.Error().Stack().Any(
			ErrorFactory.LogError(
				ErrorExtensionInit, "cannot create extension factory", err,
			),
		).Msg("")
		return
	}

	storageRootExtensionManager, objectExtensionManager, err := initDefaultExtensions(extensionFactory, conf.Init.StorageRootExtensionFolder, conf.Add.ObjectExtensionFolder, logger)
	if err != nil {
		logger.Error().Stack().Any(
			ErrorFactory.LogError(
				ErrorExtensionInit,
				"cannot initialize default extensions",
				err,
			),
		).Msg("")
		return
	}
	defer func() {
		if err := objectExtensionManager.Terminate(); err != nil {
			logger.Error().Stack().Any(
				ErrorFactory.LogError(
					ErrorExtensionRunner,
					"cannot terminate object extension manager",
					err,
				),
			).Msg("")
		}
		if err := storageRootExtensionManager.Terminate(); err != nil {
			logger.Error().Stack().Any(
				ErrorFactory.LogError(
					ErrorExtensionRunner,
					"cannot terminate storage root extension manager",
					err,
				),
			).Msg("")
		}
	}()

	ctx := ocfl.NewContextValidation(context.TODO())
	storageRoot, err := ocfl.CreateStorageRoot(
		ctx,
		destFS,
		ocfl.OCFLVersion(conf.Init.OCFLVersion),
		extensionFactory,
		storageRootExtensionManager,
		conf.Init.Digest,
		logger,
		ErrorFactory,
	)
	if err != nil {
		if err := writefs.Close(destFS); err != nil {
			logger.Error().Stack().Any(
				ErrorFactory.LogError(
					ErrorFS,
					fmt.Sprintf("cannot close filesystem '%s'", destFS),
					err,
				),
			).Msg("")
		}
		logger.Panic().Stack().Err(err).Msg("cannot create new storage root")
	}

	_, err = addObjectByPath(
		storageRoot,
		fixityAlgs,
		objectExtensionManager,
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
		logger.Panic().Stack().Err(err).Msgf("error adding content to storageroot filesystem '%s'", destFS)
	}
	_ = showStatus(ctx, logger)

}

package cmd

import (
	"context"
	"crypto/tls"
	"emperror.dev/errors"
	"fmt"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v2/internal"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl"
	"github.com/ocfl-archive/gocfl/v2/pkg/subsystem/migration"
	"github.com/ocfl-archive/gocfl/v2/pkg/subsystem/thumbnail"
	ironmaiden "github.com/ocfl-archive/indexer/v3/pkg/indexer"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/cobra"
	ublogger "gitlab.switch.ch/ub-unibas/go-ublogger/v2"
	"go.ub.unibas.ch/cloud/certloader/v2/pkg/loader"
	"io"
	"io/fs"
	"log"
	"os"
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
	updateCmd.Flags().StringP("message", "m", "", "message for new object version (required)")
	updateCmd.Flags().StringP("user-name", "u", "", "user name for new object version (required)")
	updateCmd.Flags().StringP("user-address", "a", "", "user address for new object version (required)")
	updateCmd.Flags().StringP("digest", "d", "", "digest to use for zip file checksum")
	updateCmd.Flags().Bool("no-deduplicate", false, "disable deduplication (faster)")
	updateCmd.Flags().Bool("echo", false, "update strategy 'echo' (reflects deletions). if not set, update strategy is 'contribute'")
	updateCmd.Flags().Bool("no-compress", false, "do not compress data in zip file")
	updateCmd.Flags().Bool("encrypt-aes", false, "set flag to create encrypted container (only for container target)")
	updateCmd.Flags().String("aes-key", "", "key to use for encrypted container in hex format (64 chars, empty: generate random key")
	updateCmd.Flags().String("aes-iv", "", "initialisation vector to use for encrypted container in hex format (32 charsempty: generate random vector")
}

func doUpdateConf(cmd *cobra.Command) {
	if str := getFlagString(cmd, "user-name"); str != "" {
		conf.Update.User.Name = str
	}
	if str := getFlagString(cmd, "user-address"); str != "" {
		conf.Update.User.Address = str
	}
	if str := getFlagString(cmd, "message"); str != "" {
		conf.Update.Message = str
	}
	if str := getFlagString(cmd, "digest"); str != "" {
		conf.Update.Digest = checksum.DigestAlgorithm(str)
	}
	if conf.Update.Digest == "" {
		conf.Update.Digest = checksum.DigestSHA512
	}
	if _, err := checksum.GetHash(conf.Update.Digest); err != nil {
		_ = cmd.Help()
		cobra.CheckErr(errors.Errorf("invalid digest '%s' for flag 'digest' or 'Init.DigestAlgorithm' config file entry", conf.Add.Digest))
	}
	if b, ok := getFlagBool(cmd, "no-deduplicate"); ok {
		conf.Update.Deduplicate = !b
	}
	if b, ok := getFlagBool(cmd, "no-compress"); ok {
		conf.Update.NoCompress = b
	}
	if b, ok := getFlagBool(cmd, "echo"); ok {
		conf.Update.Echo = b
	}

}

func doUpdate(cmd *cobra.Command, args []string) {
	var err error

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

	doUpdateConf(cmd)

	var addr string
	var localCache bool
	var fss = map[string]fs.FS{"internal": internal.InternalFS}

	indexerActions, err := ironmaiden.InitActionDispatcher(fss, *conf.Indexer, logger)
	if err != nil {
		logger.Panic().Stack().Err(err).Msg("cannot init indexer")
	}

	t := startTimer()
	defer func() { logger.Info().Msgf("Duration: %s", t.String()) }()

	fmt.Printf("opening '%s'\n", ocflPath)
	logger.Info().Msgf("opening '%s'", ocflPath)

	if _, err := os.Stat(srcPath); err != nil {
		logger.Panic().Stack().Err(err).Msgf("cannot stat '%s'", srcPath)
	}

	fsFactory, err := initializeFSFactory([]checksum.DigestAlgorithm{conf.Update.Digest}, nil, nil, conf.Update.NoCompress, false, logger)
	if err != nil {
		logger.Panic().Stack().Err(err).Msg("cannot create filesystem factory")
	}

	sourceFS, err := fsFactory.Get(srcPath, true)
	if err != nil {
		logger.Panic().Stack().Err(err).Msgf("cannot get filesystem for '%s'", srcPath)
	}
	destFS, err := fsFactory.Get(ocflPath, false)
	if err != nil {
		logger.Panic().Stack().Err(err).Msgf("cannot get filesystem for '%s'", ocflPath)
	}
	var doNotClose = false
	defer func() {
		if doNotClose {
			logger.Panic().Msgf("filesystem '%s' not closed", destFS)
		} else {
			if err := writefs.Close(destFS); err != nil {
				logger.Panic().Stack().Err(err).Msgf("error closing filesystem '%s'", destFS)
			}
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
			logger.Error().Msgf("invalid areapath '%s'", args[i])
			continue
		}
		areaPaths[matches[1]], err = fsFactory.Get(matches[2], true)
		if err != nil {
			doNotClose = true
			logger.Panic().Stack().Err(err).Msgf("cannot get filesystem for '%s'", args[i])
		}
	}

	mig, err := migration.GetMigrations(conf)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("cannot get migrations")
		doNotClose = true
		return
	}
	mig.SetSourceFS(sourceFS)

	thumb, err := thumbnail.GetThumbnails(conf)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("cannot get thumbnails")
		doNotClose = true
		return
	}
	thumb.SetSourceFS(sourceFS)

	extensionParams := GetExtensionParamValues(cmd, conf)
	extensionFactory, err := InitExtensionFactory(
		extensionParams,
		addr,
		localCache,
		indexerActions,
		mig,
		thumb,
		sourceFS,
		logger,
		conf.TempDir,
	)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("cannot initialize extension factory")
		doNotClose = true
		return
	}
	_, objectExtensions, err := initDefaultExtensions(
		extensionFactory,
		"",
		"",
		logger,
	)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("cannot initialize default extensions")
		doNotClose = true
		return
	}

	ctx := ocfl.NewContextValidation(context.TODO())
	if !writefs.HasContent(destFS) {

	}
	storageRoot, err := ocfl.LoadStorageRoot(ctx, destFS, extensionFactory, logger, ErrorFactory, conf.Init.Documentation)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("cannot load storage root")
		doNotClose = true
		return
	}

	exists, err := storageRoot.ObjectExists(flagObjectID)
	if err != nil {
		logger.Error().Stack().Err(err).Msgf("cannot check for object '%s'", flagObjectID)
		doNotClose = true
		return
	}
	if !exists {
		fmt.Printf("Object '%s' does not exists, exiting", flagObjectID)
		doNotClose = true
		return
	}

	_, err = addObjectByPath(
		storageRoot,
		nil,
		objectExtensions,
		conf.Update.Deduplicate,
		flagObjectID,
		conf.Update.User.Name,
		conf.Update.User.Address,
		conf.Update.Message,
		sourceFS,
		area,
		areaPaths,
		conf.Update.Echo,
		logger,
	)
	if err != nil {
		logger.Error().Stack().Err(err).Msgf("cannot content to storageroot filesystem '%s'", destFS)
		doNotClose = true
	}
	_ = showStatus(ctx, logger)

}

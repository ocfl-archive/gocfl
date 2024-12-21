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

	"emperror.dev/errors"
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
	"golang.org/x/exp/slices"
)

var addCmd = &cobra.Command{
	Use:     "add [path to ocfl structure]",
	Aliases: []string{},
	Short:   "adds new object to existing ocfl structure",
	Long:    "opens an existing ocfl structure and adds a new object. if an object with the given id already exists, an error is produced",
	Example: "gocfl add ./archive.zip /tmp/testdata -u 'Jane Doe' -a 'mailto:user@domain' -m 'initial add' -object-id 'id:abc123'",
	Args:    cobra.MinimumNArgs(2),
	Run:     doAdd,
}

// initAdd initializes the gocfl add command
func initAdd() {
	addCmd.Flags().StringVarP(&flagObjectID, "object-id", "i", "", "object id to update (required)")
	addCmd.MarkFlagRequired("object-id")
	addCmd.Flags().String("default-object-extensions", "", "folder with initial extension configurations for new OCFL objects")
	addCmd.Flags().StringP("message", "m", "", "message for new object version (required)")
	addCmd.Flags().StringP("user-name", "u", "", "user name for new object version (required)")
	addCmd.Flags().StringP("user-address", "a", "", "user address for new object version (required)")
	addCmd.Flags().StringP("fixity", "f", "", "comma separated list of digest algorithms for fixity")
	addCmd.Flags().StringP("digest", "d", "", "digest to use for ocfl checksum")
	addCmd.Flags().Bool("deduplicate", false, "force deduplication (slower)")
	addCmd.Flags().Bool("no-compress", false, "do not compress data in zip file")
}

func doAddConf(cmd *cobra.Command) {
	if str := getFlagString(cmd, "fixity"); str != "" {
		parts := strings.Split(str, ",")
		for _, part := range parts {
			conf.Add.Fixity = append(conf.Add.Fixity, part)
		}
	}
	for _, alg := range conf.Add.Fixity {
		alg = strings.TrimSpace(strings.ToLower(alg))
		if alg == "" {
			continue
		}
		if _, err := checksum.GetHash(checksum.DigestAlgorithm(alg)); err != nil {
			_ = cmd.Help()
			cobra.CheckErr(errors.Errorf("invalid fixity '%s' for flag 'fixity' or 'Add.Fixity' config file entry", conf.Add.Fixity))
		}
	}

	if str := getFlagString(cmd, "user-name"); str != "" {
		conf.Add.User.Name = str
	}
	if str := getFlagString(cmd, "user-address"); str != "" {
		conf.Add.User.Address = str
	}
	if str := getFlagString(cmd, "message"); str != "" {
		conf.Add.Message = str
	}
	if str := getFlagString(cmd, "default-object-extensions"); str != "" {
		conf.Add.ObjectExtensionFolder = str
	}
	if b := getFlagBool(cmd, "deduplicate"); b {
		conf.Add.Deduplicate = b
	}
	if b := getFlagBool(cmd, "no-compress"); b {
		conf.Add.NoCompress = b
	}

	if str := getFlagString(cmd, "digest"); str != "" {
		conf.Add.Digest = checksum.DigestAlgorithm(str)
	}
	if conf.Add.Digest == "" {
		conf.Add.Digest = checksum.DigestSHA512
	}
	if _, err := checksum.GetHash(conf.Add.Digest); err != nil {
		_ = cmd.Help()
		cobra.CheckErr(errors.Errorf("invalid digest '%s' for flag 'digest' or 'Init.DigestAlgorithm' config file entry", conf.Add.Digest))
	}

}

// initAdd executes the gocfl add command
func doAdd(cmd *cobra.Command, args []string) {
	var err error

	if err := cmd.ValidateRequiredFlags(); err != nil {
		cobra.CheckErr(err)
		return
	}

	// todo: migration not working

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

	if !slices.Contains([]string{"DEBUG", "ERROR", "WARNING", "INFO", "CRITICAL"}, conf.Log.Level) {
		_ = cmd.Help()
		cobra.CheckErr(errors.Errorf("invalid log level '%s' for flag 'log-level' or 'LogLevel' config file entry", persistentFlagLoglevel))
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

	doAddConf(cmd)

	var addr string
	var localCache bool
	var fss = map[string]fs.FS{"internal": internal.InternalFS}

	indexerActions, err := ironmaiden.InitActionDispatcher(fss, *conf.Indexer, logger)
	if err != nil {
		logger.Panic().Stack().Err(err).Msg("cannot init indexer")
	}

	t := startTimer()
	defer func() { logger.Info().Msgf("duration: %s", t.String()) }()
	logger.Info().Msgf("opening '%s'", ocflPath)

	var fixityAlgs = []checksum.DigestAlgorithm{}
	for _, alg := range conf.Add.Fixity {
		alg = strings.TrimSpace(strings.ToLower(alg))
		if alg == "" {
			continue
		}
		fixityAlgs = append(fixityAlgs, checksum.DigestAlgorithm(alg))
	}

	if _, err := os.Stat(srcPath); err != nil {
		logger.Panic().Stack().Err(err).Msgf("cannot stat '%s'", srcPath)
	}

	fsFactory, err := initializeFSFactory([]checksum.DigestAlgorithm{conf.Add.Digest}, nil, nil, conf.Add.NoCompress, false, logger)
	if err != nil {
		logger.Debug().Stack().Any(
			ErrorFactory.LogError(
				ErrorReplaceMe,
				"cannot create filesystem factory",
				err,
			)).Msg("")
		logger.Panic().Err(err).Msg("cannot create filesystem factory")
	}

	sourceFS, err := fsFactory.Get(srcPath)
	if err != nil {
		logger.Panic().Stack().Err(err).Msgf("cannot get filesystem for '%s'", srcPath)
	}
	destFS, err := fsFactory.Get(ocflPath)
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
			logger.Error().Stack().Any(
				ErrorFactory.LogError(
					ErrorReplaceMe,
					fmt.Sprintf("no area given in areapath '%s'", args[i]),
					nil,
				)).Msg("")
			continue
		}
		areaPaths[matches[1]], err = fsFactory.Get(matches[2])
		if err != nil {
			doNotClose = true
			logger.Panic().Stack().Err(err).Msgf("cannot get filesystem for '%s'", args[i])
		}
	}

	mig, err := migration.GetMigrations(conf)
	if err != nil {
		doNotClose = true
		logger.Panic().Err(err).Msg("cannot get migrations")
	}
	mig.SetSourceFS(sourceFS)

	thumb, err := thumbnail.GetThumbnails(conf)
	if err != nil {
		doNotClose = true
		logger.Panic().Stack().Err(err).Msg("cannot get thumbnails")
	}
	thumb.SetSourceFS(sourceFS)

	extensionParams := GetExtensionParamValues(cmd, conf)
	extensionFactory, err := InitExtensionFactory(extensionParams, addr, localCache, indexerActions, mig, thumb, sourceFS, (logger))
	if err != nil {
		doNotClose = true
		logger.Panic().Stack().Err(err).Msg("cannot initialize extension factory")
	}
	_, objectExtensionManager, err := initDefaultExtensions(extensionFactory, "", conf.Add.ObjectExtensionFolder, logger)
	if err != nil {
		doNotClose = true
		logger.Panic().Stack().Err(err).Msg("cannot initialize default extensions")
	}

	ctx := ocfl.NewContextValidation(context.TODO())
	storageRoot, err := ocfl.LoadStorageRoot(ctx, destFS, extensionFactory, (logger), ErrorFactory)
	if err != nil {
		doNotClose = true
		logger.Panic().Stack().Err(err).Msg("cannot open storage root")
	}
	if storageRoot.GetDigest() == "" {
		storageRoot.SetDigest(checksum.DigestAlgorithm(conf.Add.Digest))
	} else {
		if storageRoot.GetDigest() != conf.Add.Digest {
			doNotClose = true
			logger.Panic().Msgf("storageroot already uses digest '%s' not '%s'", storageRoot.GetDigest(), conf.Add.Digest)
		}
	}

	exists, err := storageRoot.ObjectExists(flagObjectID)
	if err != nil {
		doNotClose = true
		logger.Panic().Stack().Err(err).Msgf("cannot check for object '%s'", flagObjectID)
	}
	if exists {
		logger.Warn().Any(
			ErrorFactory.LogError(
				ErrorReplaceMe,
				fmt.Sprintf("object '%s' already exist, exiting", flagObjectID),
				nil,
			)).Msg("")
		return
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
		doNotClose = true
		logger.Panic().Stack().Err(err).Msgf("error adding content to storageroot filesystem '%s'", destFS)
	}
	_ = showStatus(ctx, logger)

}

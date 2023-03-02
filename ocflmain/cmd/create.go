package cmd

import (
	"context"
	"emperror.dev/emperror"
	"emperror.dev/errors"
	"encoding/hex"
	"fmt"
	"github.com/je4/gocfl/v2/pkg/checksum"
	"github.com/je4/gocfl/v2/pkg/indexer"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	ironmaiden "github.com/je4/indexer/pkg/indexer"
	lm "github.com/je4/utils/v2/pkg/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/exp/slices"
	"net"
	"path/filepath"
	"strings"
	"time"
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

	createCmd.Flags().StringP("fixity", "f", "", fmt.Sprintf("comma separated list of digest algorithms for fixity %v", checksum.DigestsNames))
	emperror.Panic(viper.BindPFlag("Create.Fixity", createCmd.Flags().Lookup("fixity")))

	createCmd.Flags().StringP("digest", "d", "", "digest to use for ocfl checksum")
	emperror.Panic(viper.BindPFlag("Create.DigestAlgorithm", createCmd.Flags().Lookup("digest")))

	createCmd.Flags().String("default-area", "", "default area for update or ingest (default: content)")
	emperror.Panic(viper.BindPFlag("Create.DefaultArea", createCmd.Flags().Lookup("default-area")))

	createCmd.Flags().Bool("deduplicate", false, "force deduplication (slower)")
	emperror.Panic(viper.BindPFlag("Create.Deduplicate", createCmd.Flags().Lookup("deduplicate")))

	createCmd.Flags().Bool("encrypt-aes", false, "create encrypted container (only for container target)")
	emperror.Panic(viper.BindPFlag("Create.AES", createCmd.Flags().Lookup("encrypt-aes")))

	createCmd.Flags().String("aes-key", "", "key to use for encrypted container in hex format (64 chars, empty: generate random key)")
	emperror.Panic(viper.BindPFlag("Create.AESKey", createCmd.Flags().Lookup("aes-key")))

	createCmd.Flags().String("aes-iv", "", "initialisation vector to use for encrypted container in hex format (32 char, sempty: generate random vector)")
	emperror.Panic(viper.BindPFlag("Create.AESKey", createCmd.Flags().Lookup("aes-key")))
}

// initCreate executes the gocfl create command
func doCreate(cmd *cobra.Command, args []string) {
	var err error
	notSet := []string{}
	ocflPath := filepath.ToSlash(filepath.Clean(args[0]))
	srcPath := filepath.ToSlash(filepath.Clean(args[1]))
	persistentFlagLogfile := viper.GetString("LogFile")
	persistentFlagLoglevel := strings.ToUpper(viper.GetString("LogLevel"))
	if !slices.Contains([]string{"DEBUG", "ERROR", "WARNING", "INFO", "CRITICAL"}, persistentFlagLoglevel) {
		cmd.Help()
		cobra.CheckErr(errors.Errorf("invalid log level '%s' for flag 'log-level' or 'LogLevel' config file entry", persistentFlagLoglevel))
	}

	daLogger, lf := lm.CreateLogger("ocfl", persistentFlagLogfile, nil, persistentFlagLoglevel, LOGFORMAT)
	defer lf.Close()

	flagFixity := viper.GetString("Create.Fixity")
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
		cmd.Help()
		cobra.CheckErr(errors.Errorf("invalid digest '%s' for flag 'digest' or 'Create.DigestAlgorithm' config file entry", flagInitDigest))
	}

	flagAddDigest := viper.GetString("Create.DigestAlgorithm")
	if _, err := checksum.GetHash(checksum.DigestAlgorithm(flagAddDigest)); err != nil {
		cmd.Help()
		cobra.CheckErr(errors.Errorf("invalid digest '%s' for flag 'digest' or 'Create.DigestAlgorithm' config file entry", flagAddDigest))
	}
	var zipAlgs = []checksum.DigestAlgorithm{checksum.DigestAlgorithm(flagAddDigest)}

	flagVersion := viper.GetString("Create.OCFLVersion")
	if !ocfl.ValidVersion(ocfl.OCFLVersion(flagVersion)) {
		cmd.Help()
		cobra.CheckErr(errors.Errorf("invalid version '%s' for flag 'ocfl-version' or 'Create.OCFLVersion' config file entry", flagVersion))
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

	flagAES := viper.GetBool("Create.AES")
	flagAESKey := viper.GetString("Create.AESKey")
	if flagAESKey != "" && len(flagAESKey) != 64 {
		cmd.Help()
		cobra.CheckErr(errors.Errorf("invalid format '%s' for flag 'aes-key' or 'Create.AESKey' config file entry. 64 character hex value needed", flagAESKey))
	}
	var aesKey []byte
	if flagAESKey != "" {
		aesKey = make([]byte, hex.DecodedLen(len(flagAESKey)))
		if _, err := hex.Decode(aesKey, []byte(flagAESKey)); err != nil {
			aesKey = nil
			cmd.Help()
			cobra.CheckErr(errors.Errorf("invalid format '%s' for flag 'aes-key' or 'Create.AESKey' config file entry. 64 character hex value needed: %v", flagAESKey, err))
		}
	}
	flagAESIV := viper.GetString("Create.AESIV")
	if flagAESIV != "" && len(flagAESIV) != 32 {
		cmd.Help()
		cobra.CheckErr(errors.Errorf("invalid format '%s' for flag 'aes-iv' or 'Create.AESIV' config file entry. 32 character hex value needed", flagAESIV))
	}
	var aesIV []byte
	if flagAESIV != "" {
		aesIV = make([]byte, hex.DecodedLen(len(flagAESIV)))
		if _, err := hex.Decode(aesIV, []byte(flagAESIV)); err != nil {
			aesIV = nil
			cmd.Help()
			cobra.CheckErr(errors.Errorf("invalid format '%s' for flag 'aes-iv' or 'Create.AESIV' config file entry. 64 character hex value needed: %v", flagAESIV, err))
		}
	}

	if len(notSet) > 0 {
		cmd.Help()
		cobra.CheckErr(errors.Errorf("required flag(s) %s not set", strings.Join(notSet, ", ")))
	}

	var idx *ironmaiden.Server
	var addr string
	if withIndexer := viper.GetBool("Indexer.Local"); withIndexer {
		siegfried, err := indexer.GetSiegfried()
		if err != nil {
			daLogger.Errorf("cannot load indexer Siegfried: %v", err)
			return
		}
		mimeRelevance, err := indexer.GetMimeRelevance()
		if err != nil {
			daLogger.Errorf("cannot load indexer MimeRelevance: %v", err)
			return
		}
		ffmpeg, err := indexer.GetFFMPEG()
		if err != nil {
			daLogger.Errorf("cannot load indexer FFMPEG: %v", err)
			return
		}
		imageMagick, err := indexer.GetImageMagick()
		if err != nil {
			daLogger.Errorf("cannot load indexer ImageMagick: %v", err)
			return
		}
		tika, err := indexer.GetTika()
		if err != nil {
			daLogger.Errorf("cannot load indexer Tika: %v", err)
			return
		}
		var netAddr net.Addr
		idx, netAddr, err = indexer.StartIndexer(
			siegfried,
			ffmpeg,
			imageMagick,
			tika,
			mimeRelevance,
			daLogger)
		if err != nil {
			daLogger.Errorf("cannot start indexer: %v", err)
			return
		}
		addr = fmt.Sprintf("http://%s/v2", netAddr.String())
		defer func() {
			daLogger.Info("shutting down indexer")
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			idx.Shutdown(ctx)
		}()
	}

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
			daLogger.Warningf("no area prefix for '%s'", args[i])
			continue
		}
		daLogger.Infof("additional path '%s:%s'", matches[1], matches[2])
		areaPaths[matches[1]], err = fsFactory.GetFS(matches[2])
		if err != nil {
			daLogger.Errorf("cannot get filesystem for '%s': %v", args[i], err)
			daLogger.Errorf("%v%+v", err, ocfl.GetErrorStacktrace(err))
			return
		}
	}

	extensionParams := GetExtensionParamValues(cmd)
	extensionFactory, err := initExtensionFactory(extensionParams, addr, sourceFS, daLogger)
	if err != nil {
		daLogger.Errorf("cannot initialize extension factory: %v", err)
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
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
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
	}

	if err := destFS.Close(); err != nil {
		daLogger.Errorf("error closing filesystem '%s': %v", destFS, err)
		daLogger.Debugf("%v%+v", err, ocfl.GetErrorStacktrace(err))
	}
}

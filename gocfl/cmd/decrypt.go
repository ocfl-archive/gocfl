package cmd

import (
	"bytes"
	"crypto/tls"
	"emperror.dev/emperror"
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"github.com/google/tink/go/core/registry"
	"github.com/google/tink/go/keyset"
	"github.com/google/tink/go/streamingaead"
	statickms "github.com/je4/utils/v2/pkg/StaticKMS"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/encrypt"
	"github.com/je4/utils/v2/pkg/keepass2kms"
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/cobra"
	ublogger "gitlab.switch.ch/ub-unibas/go-ublogger/v2"
	"go.ub.unibas.ch/cloud/certloader/v2/pkg/loader"
	"io"
	"log"
	"os"
	"path/filepath"
)

var decryptCmd = &cobra.Command{
	Use:   "decrypt [path to ocfl structure]",
	Short: "decrypt an encrypted OCFL file",
	//Long:    "an utterly useless command for testing",
	Example: "gocfl decrypt ./archive.zip.aes",
	Args:    cobra.RangeArgs(1, 2),
	Run:     doDecrypt,
}

func initDecrypt() {
	decryptCmd.Flags().String("aes-key", "", "key to use for encrypted container in hex format (64 chars, empty: generate random key)")
	decryptCmd.Flags().String("aes-iv", "", "initialisation vector to use for encrypted container in hex format (32 char, sempty: generate random vector)")
	decryptCmd.Flags().String("keepass-file", "", "file with keepass2 database")
	decryptCmd.Flags().String("keepass-entry", "", "keepass2 entry to use for key encryption")
	decryptCmd.Flags().String("keepass-key", "", "key to use for keepass2 database decryption")
	decryptCmd.Flags().Bool("checksum-only", false, "do not output decrypted file, only checksum")
}

func doDecryptConf(cmd *cobra.Command) {
	if b, ok := getFlagBool(cmd, "encrypt-aes"); ok {
		conf.AES.Enable = b
	}
	if str := getFlagString(cmd, "aes-key"); str != "" {
		conf.AES.Key.UnmarshalText(([]byte)(str))
	}
	if str := getFlagString(cmd, "aes-iv"); str != "" {
		conf.AES.IV.UnmarshalText(([]byte)(str))
	}
	if str := getFlagString(cmd, "keepass-file"); str != "" {
		conf.AES.KeepassFile.UnmarshalText(([]byte)(str))
	}
	if str := getFlagString(cmd, "keepass-entry"); str != "" {
		conf.AES.KeepassEntry.UnmarshalText(([]byte)(str))
	}
	if str := getFlagString(cmd, "keepass-key"); str != "" {
		conf.AES.KeepassKey.UnmarshalText(([]byte)(str))
	}
}

func doDecrypt(cmd *cobra.Command, args []string) {
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

	doDecryptConf(cmd)

	onlyChecksum, _ := getFlagBool(cmd, "checksum-only")

	if len(args) < 2 && !onlyChecksum {
		cobra.CheckErr(errors.New("missing output path arguments"))
		return
	}
	aesConfig := conf.AES
	encFilePath, err := ocfl.Fullpath(args[0])
	if err != nil {
		logger.Error().Stack().Err(err).Msgf("cannot convert '%s' to absolute path", args[0])
		return
	}
	var targetFile string
	if !onlyChecksum {
		targetFile = args[1]
	}

	var client registry.KMSClient
	var keyPath string
	if aesConfig.Key != "" {
		logger.Info().Msgf("using static KMS client")
		client, err = statickms.NewClient(string(aesConfig.Key))
		if err != nil {
			err = ErrorFactory.NewError(
				ErrorFS,
				"cannot create static kms client",
				err,
			)
			logger.Error().Stack().Err(err).Msg("cannot create static kms client")
			return
		}
		keyPath = "static://"
	} else {
		logger.Info().Msgf("using keepass2kms client with file '%s'", aesConfig.KeepassFile)
		db, err := keepass2kms.LoadKeePassDBFromFile(string(aesConfig.KeepassFile), string(aesConfig.KeepassKey))
		if err != nil {
			err = ErrorFactory.NewError(
				ErrorFS,
				fmt.Sprintf("cannot load keepass file '%s'", aesConfig.KeepassFile),
				err,
			)
			logger.Error().Stack().Err(err).Msgf("cannot load keepass file '%s'", aesConfig.KeepassFile)
			return
		}
		client, err = keepass2kms.NewClient(db, filepath.Base(string(aesConfig.KeepassFile)))
		if err != nil {
			err = ErrorFactory.NewError(
				ErrorFS,
				"cannot create keepass2kms client",
				err,
			)
			logger.Error().Stack().Err(err).Msg("cannot create keepass2kms client")
			return
		}
		keyPath = string(aesConfig.KeepassEntry)
	}
	//registry.RegisterKMSClient(client)

	keyfile := encFilePath + ".key.json"
	keyData, err := os.ReadFile(keyfile)
	if err != nil {
		logger.Error().Stack().Err(err).Msgf("cannot read key file '%s'", keyfile)
		return
	}
	kStruct := &encrypt.KeyStruct{}
	if err := json.Unmarshal(keyData, kStruct); err != nil {
		logger.Error().Stack().Err(err).Msgf("cannot unmarshal key file '%s'", keyfile)
		return
	}

	aead, err := client.GetAEAD(keyPath)
	if err != nil {
		logger.Error().Stack().Err(err).Msgf("cannot get aead '%s'", keyPath)
		return
	}

	kh, err := keyset.Read(keyset.NewBinaryReader(bytes.NewBuffer(kStruct.EncryptedKey)), aead)
	if err != nil {

		logger.Error().Stack().Err(err).Msgf("cannot read keyset '%s'", keyfile)
		return
	}

	fp, err := os.Open(encFilePath)
	if err != nil {
		logger.Error().Stack().Err(err).Msgf("cannot open file '%s'", encFilePath)
		return
	}
	defer fp.Close()

	stream, err := streamingaead.New(kh)
	if err != nil {
		logger.Error().Stack().Err(err).Msgf("cannot create streamingaead '%s'", keyfile)
		return
	}
	dec, err := stream.NewDecryptingReader(fp, kStruct.Aad)
	if err != nil {
		logger.Error().Stack().Err(err).Msgf("cannot create decrypting reader '%s'", keyfile)
		return
	}

	var out io.Writer
	if !onlyChecksum && targetFile != "" {
		if targetFile == "-" {
			out = os.Stdout
		} else {
			outFP, err := os.Create(targetFile)
			if err != nil {
				emperror.Panic(errors.Errorf("cannot create output file '%s': %v", targetFile, err))
			}
			defer outFP.Close()
			out = outFP
		}
	} else {
		out = io.Discard
	}

	csWriter, err := checksum.NewChecksumWriter([]checksum.DigestAlgorithm{conf.Init.Digest}, out)
	if err != nil {
		logger.Error().Stack().Err(err).Msgf("cannot create checksum writer '%s'", targetFile)
		return
	}

	if _, err := io.Copy(csWriter, dec); err != nil {
		logger.Error().Stack().Err(err).Msgf("cannot copy decrypted data to '%s'", targetFile)
		return
	}
	if err := csWriter.Close(); err != nil {
		logger.Error().Stack().Err(err).Msgf("cannot close checksum writer '%s'", targetFile)
		return
	}
	css, err := csWriter.GetChecksums()
	if err != nil {
		logger.Error().Stack().Err(err).Msgf("cannot get checksums '%s'", targetFile)
		return
	}
	for csName, csValue := range css {
		if onlyChecksum {
			fmt.Printf("%s: %s\n", csName, csValue)
		} else {
			logger.Info().Msgf("checksum %s: %s", csName, csValue)
		}
	}
	if !onlyChecksum {
		logger.Info().Msgf("decrypted file '%s' to '%s'", encFilePath, targetFile)
	}
}

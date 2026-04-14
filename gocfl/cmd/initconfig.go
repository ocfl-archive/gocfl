package cmd

import (
	"context"
	"crypto/tls"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"emperror.dev/errors"
	"github.com/BurntSushi/toml"
	"github.com/je4/utils/v2/pkg/config"
	"github.com/ocfl-archive/gocfl/v2/internal"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/util"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfllogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/subsystem/thumbnail"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/cobra"
	ublogger "gitlab.switch.ch/ub-unibas/go-ublogger/v2"
	"go.ub.unibas.ch/cloud/certloader/v2/pkg/loader"
)

func quoteCmdArg(s string) string {
	// Einfacher Ansatz für cmd.exe: Anführungszeichen um das Argument,
	// innere Quotes verdoppeln.
	s = strings.ReplaceAll(s, `"`, `""`)
	return `"` + s + `"`
}

func quoteShellArg(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'"'"'`) + "'"
}

var initConfigCmd = &cobra.Command{
	Use:     "initconfig [path to config file]",
	Aliases: []string{},
	Short:   "store configuration of gocfl in toml format",
	//Long:    "an utterly useless command for testing",
	Example: "gocfl initconfig",
	Args:    cobra.MaximumNArgs(1),
	Run:     doInitConfig,
}

func initInitConfig() {
	initConfigCmd.Flags().String("toml", "", "name of toml config file")
	initConfigCmd.Flags().String("extension-folder", "", "folder for extension templates")
	initConfigCmd.Flags().String("script-folder", "", "folder for extension scripts")
	initConfigCmd.Flags().Bool("fullconfig", false, "store all configuration options instead of minimal configuration")
	initConfigCmd.Flags().Bool("extensions", false, "extract extension templates")
	initConfigCmd.Flags().Bool("scripts", false, "extract extension scripts")
}

func doInitConfigConf(cmd *cobra.Command) {
	if str := getFlagString(cmd, "toml"); str != "" {
		conf.InitConfig.TOMLFile = str
	}
	if str := getFlagString(cmd, "extension-folder"); str != "" {
		conf.InitConfig.ExtensionFolder = str
	}
	if str := getFlagString(cmd, "script-folder"); str != "" {
		conf.InitConfig.ScriptFolder = str
	}
	if b, ok := getFlagBool(cmd, "fullconfig"); ok {
		conf.InitConfig.FullConfig = b
	}
	if b, ok := getFlagBool(cmd, "extensions"); ok {
		conf.InitConfig.Extensions = b
	}
	if b, ok := getFlagBool(cmd, "scripts"); ok {
		conf.InitConfig.Scripts = b
	}

}

func doInitConfig(cmd *cobra.Command, args []string) {
	var configFolder string
	var err error
	if len(args) == 0 {
		configFolder = conf.InitConfig.ConfigFolder
	} else {
		configFolder = args[0]
	}
	configFolder, err = util.Fullpath(configFolder)
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
	ctx := context.TODO()
	var logger = ocfllogger.NewOCFLLogger(ctx, &l2, nil, version.Default, nil)

	doInitConfigConf(cmd)

	var scriptFolder = conf.InitConfig.ScriptFolder
	var extensionFolder = conf.InitConfig.ExtensionFolder
	var tomlPath = conf.InitConfig.TOMLFile
	if !filepath.IsAbs(scriptFolder) {
		scriptFolder = filepath.ToSlash(filepath.Join(configFolder, scriptFolder))
	}
	logger.Info().Msgf("Script Folder: %s", scriptFolder)
	if !filepath.IsAbs(extensionFolder) {
		extensionFolder = filepath.ToSlash(filepath.Join(configFolder, extensionFolder))
	}
	if !filepath.IsAbs(scriptFolder) {
		scriptFolder = filepath.ToSlash(filepath.Join(configFolder, scriptFolder))
	}
	logger.Info().Msgf("Extension Folder: %s", extensionFolder)
	if !filepath.IsAbs(tomlPath) {
		tomlPath = filepath.ToSlash(filepath.Join(configFolder, tomlPath))
	}
	logger.Info().Msgf("TOML File: %s", tomlPath)

	//	scripts := []string{}
	newMiniConfig := config.MiniConfig{
		"log.level":      conf.Log.Level,
		"add.user":       conf.Add.User,
		"add.message":    conf.Add.Message,
		"update.user":    conf.Update.User,
		"update.message": conf.Update.Message,
	}

	if conf.InitConfig.Extensions {
		if err := os.MkdirAll(extensionFolder, 0755); err != nil {
			logger.Fatal().Err(err).Msgf("cannot create extension folder: %s", extensionFolder)
		}
		extFS, err := fs.Sub(internal.InternalFS, "extensions")
		if err != nil {
			logger.Fatal().Err(err).Msg("cannot create subfs for internal:extensions")
		}
		if err := fs.WalkDir(extFS, ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return errors.WithStack(err)
			}
			if d.IsDir() {
				return nil
			}
			target := filepath.Join(extensionFolder, path)
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return errors.Wrapf(err, "cannot create directory: %s", filepath.Dir(target))
			}
			src, err := extFS.Open(path)
			if err != nil {
				return errors.Wrapf(err, "cannot open file: internal:%s", path)
			}
			defer func(src fs.File) {
				err := src.Close()
				if err != nil {
					logger.Error().Err(err).Msgf("cannot close file: internal:%s", path)
				}
			}(src)
			out, err := os.Create(target)
			if err != nil {
				return errors.Wrapf(err, "cannot create file: %s", target)
			}
			defer func(out *os.File) {
				err := out.Close()
				if err != nil {
					logger.Error().Err(err).Msgf("cannot close file: %s", target)
				}
			}(out)
			logger.Info().Msgf("copying extension config: internal:%s -> %s", path, target)
			if _, err := io.Copy(out, src); err != nil {
				return errors.Wrapf(err, "cannot copy file: internal:%s -> %s", path, target)
			}
			return nil
		}); err != nil {
			logger.Fatal().Err(err).Msg("cannot walk internal:extensions")
		}

		conf.Init.StorageRootExtensionFolder = filepath.ToSlash(filepath.Join(extensionFolder, "storageroot"))
		newMiniConfig["init.storagerootextensions"] = conf.Init.StorageRootExtensionFolder

		conf.Add.ObjectExtensionFolder = filepath.ToSlash(filepath.Join(extensionFolder, "object"))
		newMiniConfig["add.objectextensions"] = conf.Init.StorageRootExtensionFolder
	}
	thumbConf, thumbMiniconfig, err := thumbnail.InitConfig(conf.Thumbnail, scriptFolder, logger.Logger())
	if err != nil {
		logger.Fatal().Err(err).Msg("cannot init thumbnail")
	}
	for k, v := range thumbMiniconfig {
		newMiniConfig["thumbnail."+k] = v
	}

	conf.Thumbnail = thumbConf
	if err := os.MkdirAll(filepath.Dir(tomlPath), 0755); err != nil {
		logger.Fatal().Err(err).Msgf("cannot create thumbnail directory: %s", filepath.Dir(tomlPath))
	}
	fp, err := os.Create(tomlPath)
	if err != nil {
		log.Fatalf("cannot create config file: %v", err)
	}
	defer func(fp *os.File) {
		err := fp.Close()
		if err != nil {
			logger.Error().Msgf("cannot close config file: %v", err)
		}
	}(fp)
	for k, v := range newMiniConfig {
		miniConfig[k] = v
	}
	var buf []byte
	if conf.InitConfig.FullConfig {
		buf, err = toml.Marshal(conf)
		if err != nil {
			logger.Fatal().Msgf("cannot encode config: %v", err)
		}
	} else {
		buf, err = toml.Marshal(miniConfig)
		if err != nil {
			logger.Fatal().Msgf("cannot encode config: %v", err)
		}
	}
	if err := os.WriteFile(tomlPath, buf, 0644); err != nil {
		logger.Fatal().Msgf("cannot write config file: %v", err)
	}
}

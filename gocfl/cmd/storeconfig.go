package cmd

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"emperror.dev/errors"
	"github.com/BurntSushi/toml"
	"github.com/ocfl-archive/gocfl/v2/config"
	"github.com/ocfl-archive/gocfl/v2/internal"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/util"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfllogger"
	indexerutil "github.com/ocfl-archive/indexer/v3/pkg/util"
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

var storeConfigCmd = &cobra.Command{
	Use:     "storeconfig [path to config file]",
	Aliases: []string{},
	Short:   "store configuration of gocfl in toml format",
	//Long:    "an utterly useless command for testing",
	Example: "gocfl storeconfig ~/",
	Args:    cobra.MaximumNArgs(1),
	Run:     doStoreConfig,
}

func initStoreConfig() {
	storeConfigCmd.Flags().String("toml", "", "name of toml config file")
	storeConfigCmd.Flags().String("extension-folder", "", "folder for extension templates")
	storeConfigCmd.Flags().String("script-folder", "", "folder for extension scripts")
	storeConfigCmd.Flags().Bool("fullconfig", false, "store all configuration options instead of minimal configuration")
	storeConfigCmd.Flags().Bool("extensions", false, "extract extension templates")
	storeConfigCmd.Flags().Bool("scripts", false, "extract extension scripts")
}

func doStoreConfigConf(cmd *cobra.Command) {
	if str := getFlagString(cmd, "toml"); str != "" {
		conf.StoreConfig.TOMLFile = str
	}
	if str := getFlagString(cmd, "extension-folder"); str != "" {
		conf.StoreConfig.ExtensionFolder = str
	}
	if str := getFlagString(cmd, "script-folder"); str != "" {
		conf.StoreConfig.ScriptFolder = str
	}
	if b, ok := getFlagBool(cmd, "fullconfig"); ok {
		conf.StoreConfig.FullConfig = b
	}
	if b, ok := getFlagBool(cmd, "extensions"); ok {
		conf.StoreConfig.Extensions = b
	}
	if b, ok := getFlagBool(cmd, "scripts"); ok {
		conf.StoreConfig.Scripts = b
	}

}

func doStoreConfig(cmd *cobra.Command, args []string) {
	var configFolder string
	var err error
	if len(args) == 0 {
		configFolder = conf.StoreConfig.ConfigFolder
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

	doStoreConfigConf(cmd)

	var scriptFolder = conf.StoreConfig.ScriptFolder
	var extensionFolder = conf.StoreConfig.ExtensionFolder
	var tomlPath = conf.StoreConfig.TOMLFile
	if !filepath.IsAbs(scriptFolder) {
		scriptFolder = filepath.ToSlash(filepath.Join(configFolder, scriptFolder))
	}
	logger.Info().Msgf("Script Folder: %s", scriptFolder)
	if !filepath.IsAbs(extensionFolder) {
		extensionFolder = filepath.ToSlash(filepath.Join(configFolder, extensionFolder))
	}
	logger.Info().Msgf("Extension Folder: %s", extensionFolder)
	if !filepath.IsAbs(tomlPath) {
		tomlPath = filepath.ToSlash(filepath.Join(configFolder, tomlPath))
	}
	logger.Info().Msgf("TOML File: %s", tomlPath)

	scripts := []string{}
	miniConfig := map[string]interface{}{
		"loglevel": "info",
	}

	// check ghostscript and image magick convert
	gsPath, gsOK := indexerutil.CheckProgram(indexerutil.CheckProgramGhostscript, "")
	if gsOK {
		logger.Info().Msgf("ghostscript found as '%s'", gsPath)
	} else {
		logger.Error().Msgf("ghostscript not found")
	}
	convertPath, convertOK := indexerutil.CheckProgram(indexerutil.CheckProgramMagickConvert, "")
	if convertOK {
		logger.Info().Msgf("magick convert found as '%s'", convertPath)
	} else {
		logger.Error().Msgf("magick convert not found")
	}
	ffmpegPath, ffmpegOK := indexerutil.CheckProgram(indexerutil.CheckProgramFFMpeg, "")
	if ffmpegOK {
		logger.Info().Msgf("ffmpeg found as '%s'", ffmpegPath)
	} else {
		logger.Error().Msgf("ffmpeg not found")
	}
	convertParts := strings.Split(convertPath, " ")
	ffmpegParts := strings.Split(ffmpegPath, " ")
	gsParts := strings.Split(gsPath, " ")
	doReplace := func(str string) string {
		var params string
		str = strings.Replace(str, "%%CONVERT%%", convertParts[0], -1)
		params = ""
		if len(convertParts) > 1 {
			params = convertParts[1]
		}
		str = strings.Replace(str, "%%CONVERT_PARAMS%%", params, -1)

		str = strings.Replace(str, "%%FFMPEG%%", ffmpegParts[0], -1)
		params = ""
		if len(ffmpegParts) > 1 {
			params = ffmpegParts[1]
		}
		str = strings.Replace(str, "%%FFMPEG_PARAMS%%", params, -1)

		str = strings.Replace(str, "%%GHOSTSCRIPT%%", gsParts[0], -1)
		params = ""
		if len(gsParts) > 1 {
			params = gsParts[1]
		}
		str = strings.Replace(str, "%%GHOSTSCRIPT_PARAMS%%", params, -1)
		return str
	}
	swAvailable := func(str string) bool {
		if strings.Contains(str, "%%FFMPEG%%") && !ffmpegOK {
			logger.Error().Msg("ffmpeg needed, but not found on machine")
			return false
		}
		if strings.Contains(str, "%%CONVERT%%") && !convertOK {
			logger.Error().Msg("image magick convert needed, but not found on machine")
			return false
		}
		if strings.Contains(str, "%%GHOSTSCRIPT%%") && !gsOK {
			logger.Error().Msg("ghostscript needed, but not found on machine")
			return false
		}
		return true
	}

	if conf.StoreConfig.Scripts {
		files, err := fs.ReadDir(internal.InternalFS, "thumbnail/scripts")
		if err != nil {
			logger.Fatal().Err(err).Msg("cannot read internal:thumbnail/scripts")
		}
		if len(files) > 0 {
			if err := os.MkdirAll(scriptFolder, 0755); err != nil {
				logger.Fatal().Err(err).Msgf("cannot create script folder: %s", conf.StoreConfig.ScriptFolder)
			}
		}
		for _, f := range files {
			script := f.Name()
			contentBytes, err := fs.ReadFile(internal.InternalFS, path.Join("thumbnail/scripts", script))
			if err != nil {
				logger.Fatal().Err(err).Msgf("cannot read script %s", script)
			}
			contentStr := string(contentBytes)
			if !swAvailable(contentStr) {
				logger.Info().Msgf("ignoring script %s", script)
				continue
			}
			contentStr = doReplace(contentStr)
			logger.Info().Msgf("storing script %s", path.Join(scriptFolder, script))
			if err := os.WriteFile(path.Join(scriptFolder, script), []byte(contentStr), 0755); err != nil {
				logger.Error().Msgf("cannot write script file: %v", err)
				continue
			}
			scripts = append(scripts, f.Name())
		}
	}
	if conf.StoreConfig.Extensions {
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
		miniConfig["init"] = map[string]string{
			"storagerootextensions": filepath.ToSlash(filepath.Join(extensionFolder, "storageroot")),
		}
		conf.Init.StorageRootExtensionFolder = filepath.ToSlash(filepath.Join(extensionFolder, "storageroot"))

		miniConfig["add"] = map[string]string{
			"objectextensions": filepath.ToSlash(filepath.Join(extensionFolder, "object")),
		}
		conf.Add.ObjectExtensionFolder = filepath.ToSlash(filepath.Join(extensionFolder, "object"))
	}
	thumbConf := config.Thumbnail{}
	if _, err := toml.DecodeFS(internal.InternalFS, "thumbnail/thumbnail.toml", &thumbConf); err != nil {
		logger.Fatal().Err(err).Msg("cannot decode internal:thumbnail/thumbnail.toml")
	}
	for key, fn := range thumbConf.Function {
		if !swAvailable(fn.Command) {
			logger.Info().Msgf("removing function %s", key)
			delete(thumbConf.Function, key)
			continue
		}
		var isScript bool
		for _, script := range scripts {
			if strings.HasPrefix(fn.Command, script) {
				isScript = true
				if runtime.GOOS == "windows" {
					thumbConf.Function[key].Command = fmt.Sprintf("powershell -File \"%s\" %s", filepath.ToSlash(filepath.Join(scriptFolder, script)), fn.Command[len(script)+1:])
				} else {
					thumbConf.Function[key].Command = "bash -c " + quoteShellArg(fn.Command)
				}
			}
		}
		if !isScript && !strings.HasPrefix(fn.Command, "%%") {
			logger.Info().Msgf("removing function %s - no script and no prefix %%", key)
			delete(thumbConf.Function, key)
			continue
		}
		thumbConf.Function[key].Command = doReplace(fn.Command)
	}
	miniConfig["thumbnail"] = thumbConf
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
	tenc := toml.NewEncoder(fp)
	tenc.Indent = "  "
	var cfg any = miniConfig
	if conf.StoreConfig.FullConfig {
		cfg = conf
	}
	if err := tenc.Encode(cfg); err != nil {
		logger.Error().Msgf("cannot encode config: %v", err)
	}
}

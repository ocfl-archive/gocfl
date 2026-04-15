package thumbnail

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/BurntSushi/toml"
	configutil "github.com/je4/utils/v2/pkg/config"
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v2/config"
	"github.com/ocfl-archive/gocfl/v2/internal"
	"github.com/ocfl-archive/gocfl/v2/pkg/util"
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

func Autoconfig(conf *config.Thumbnail, scripts map[string]string, logger zLogger.ZLogger) (configutil.MiniConfig, error) {
	if len(scripts) == 0 {
		logger.Info().Msg("thumbnail: no scripts to autoconfig - do nothing")
		return configutil.MiniConfig{}, nil
	}
	prgs := util.DetectPrg(logger)
	if conf == nil {
		conf = &config.Thumbnail{}
	}
	miniConfig := configutil.MiniConfig{}

	if _, err := toml.DecodeFS(internal.InternalFS, "thumbnail/thumbnail.toml", conf); err != nil {
		logger.Fatal().Err(err).Msg("cannot decode internal:thumbnail/thumbnail.toml")
	}
	for key, fn := range conf.Function {
		if !util.SWAvailable(fn.Command, prgs, logger) {
			logger.Info().Msgf("removing function %s", key)
			delete(conf.Function, key)
			continue
		}
		var isScript bool
		for script, _ := range scripts {
			if strings.HasPrefix(fn.Command, script) {
				isScript = true
			}
		}
		if !isScript {
			if !strings.HasPrefix(fn.Command, "%%") {
				logger.Info().Msgf("removing function %s - no script and no prefix %%", key)
				delete(conf.Function, key)
				continue
			} else {
				conf.Function[key].Command = util.SWDoReplace(fn.Command, prgs)
			}
			miniConfig[fmt.Sprintf("function.%s.command", strings.ToLower(key))] = fn.Command
		}
	}
	return miniConfig, nil
}

func InitConfig(conf *config.Thumbnail, scriptFolder string, logger zLogger.ZLogger) (*config.Thumbnail, configutil.MiniConfig, error) {
	prgs := util.DetectPrg(logger)
	var scripts = map[string]string{}
	miniConfig := configutil.MiniConfig{}

	newConf := &config.Thumbnail{}
	if _, err := toml.DecodeFS(internal.InternalFS, "thumbnail/thumbnail.toml", newConf); err != nil {
		logger.Error().Err(err).Msg("cannot decode thumbnail config")
	}
	for key, fn := range newConf.Function {
		conf.Function[key] = fn
	}

	// get all script files
	files, err := fs.ReadDir(internal.InternalFS, "thumbnail/scripts")
	if err != nil {
		logger.Fatal().Err(err).Msg("cannot read internal:thumbnail/scripts")
	}
	if len(files) > 0 {
		if err := os.MkdirAll(scriptFolder, 0755); err != nil {
			logger.Fatal().Err(err).Msgf("cannot create script folder: %s", scriptFolder)
		}
	}
	// copy the script files while replacing software variables
	for _, f := range files {
		script := f.Name()
		contentBytes, err := fs.ReadFile(internal.InternalFS, path.Join("thumbnail/scripts", script))
		if err != nil {
			logger.Fatal().Err(err).Msgf("cannot read script %s", script)
		}
		contentStr := string(contentBytes)
		// ignore script, if we do not have the software for it
		if !util.SWAvailable(contentStr, prgs, logger) {
			logger.Info().Msgf("ignoring script %s", script)
			continue
		}
		contentStr = util.SWDoReplace(contentStr, prgs)
		logger.Info().Msgf("storing script %s", path.Join(scriptFolder, script))
		if err := os.WriteFile(path.Join(scriptFolder, script), []byte(contentStr), 0755); err != nil {
			logger.Error().Msgf("cannot write script file: %v", err)
			continue
		}
		scripts[script] = path.Join(scriptFolder, script)
	}
	for key, fn := range conf.Function {
		if !util.SWAvailable(fn.Command, prgs, logger) {
			logger.Info().Msgf("removing function %s", key)
			delete(conf.Function, key)
			continue
		}
		//		var isScript bool
		for script, p := range scripts {
			if strings.HasPrefix(fn.Command, script) {
				//				isScript = true
				if runtime.GOOS == "windows" {
					conf.Function[key].Command = fmt.Sprintf("powershell -File \"%s\" %s", filepath.ToSlash(p), fn.Command[len(script)+1:])
				} else {
					conf.Function[key].Command = "bash -c " + quoteShellArg(fn.Command)
				}
				miniConfig[fmt.Sprintf("function.%s.command", strings.ToLower(key))] = conf.Function[key].Command
				break
			}
		}
		/*
			if !isScript && !strings.HasPrefix(fn.Command, "%%") {
				logger.Info().Msgf("removing function %s - no script and no prefix %%", key)
				delete(conf.Function, key)
				continue
			}
		*/
		conf.Function[key].Command = util.SWDoReplace(fn.Command, prgs)
		miniConfig[fmt.Sprintf("function.%s.command", strings.ToLower(key))] = conf.Function[key].Command
		miniConfig[fmt.Sprintf("function.%s.id", strings.ToLower(key))] = conf.Function[key].ID
		miniConfig[fmt.Sprintf("function.%s.title", strings.ToLower(key))] = conf.Function[key].Title
		miniConfig[fmt.Sprintf("function.%s.pronoms", strings.ToLower(key))] = conf.Function[key].Pronoms
		miniConfig[fmt.Sprintf("function.%s.mime", strings.ToLower(key))] = conf.Function[key].Mime
		miniConfig[fmt.Sprintf("function.%s.timeout", strings.ToLower(key))] = conf.Function[key].Timeout.String()

	}
	return conf, miniConfig, nil
}

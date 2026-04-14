package util

import (
	"maps"
	"slices"
	"strings"

	"github.com/je4/utils/v2/pkg/zLogger"
	indexerutil "github.com/ocfl-archive/indexer/v3/pkg/util"
)

type Program struct {
	path   string
	Prg    string
	Params string
}

var _prgs = map[string]Program{}

func DetectPrg(logger zLogger.ZLogger) map[string]Program {
	if _, ok := _prgs[indexerutil.CheckProgramGhostscript]; !ok {
		// check ghostscript and image magick convert
		gsPath, gsOK := indexerutil.CheckProgram(indexerutil.CheckProgramGhostscript, "")
		if gsOK {
			logger.Info().Msgf("ghostscript found as '%s'", gsPath)
			parts := strings.SplitN(gsPath, " ", 2)
			params := ""
			if len(parts) > 1 {
				params = parts[1]
			}
			_prgs[indexerutil.CheckProgramGhostscript] = Program{
				path:   gsPath,
				Prg:    parts[0],
				Params: params,
			}
		} else {
			logger.Error().Msgf("ghostscript not found")
		}
	}
	if _, ok := _prgs[indexerutil.CheckProgramMagickConvert]; !ok {
		convertPath, convertOK := indexerutil.CheckProgram(indexerutil.CheckProgramMagickConvert, "")
		if convertOK {
			logger.Info().Msgf("magick convert found as '%s'", convertPath)
			parts := strings.SplitN(convertPath, " ", 2)
			params := ""
			if len(parts) > 1 {
				params = parts[1]
			}
			_prgs[indexerutil.CheckProgramMagickConvert] = Program{
				path:   convertPath,
				Prg:    parts[0],
				Params: params,
			}
		} else {
			logger.Error().Msgf("magick convert not found")
		}
	}
	if _, ok := _prgs[indexerutil.CheckProgramFFMpeg]; !ok {
		ffmpegPath, ffmpegOK := indexerutil.CheckProgram(indexerutil.CheckProgramFFMpeg, "")
		if ffmpegOK {
			logger.Info().Msgf("ffmpeg found as '%s'", ffmpegPath)
			parts := strings.SplitN(ffmpegPath, " ", 2)
			params := ""
			if len(parts) > 1 {
				params = parts[1]
			}
			_prgs[indexerutil.CheckProgramFFMpeg] = Program{
				path:   ffmpegPath,
				Prg:    parts[0],
				Params: params,
			}
		} else {
			logger.Error().Msgf("ffmpeg not found")
		}
	}
	return _prgs
}

func SWAvailable(str string, prgs map[string]Program, logger zLogger.ZLogger) bool {
	var prgNames = SeqToSlice(maps.Keys(prgs))
	if strings.Contains(str, "%%FFMPEG%%") && !slices.Contains(prgNames, indexerutil.CheckProgramFFMpeg) {
		logger.Error().Msg("ffmpeg needed, but not found on machine")
		return false
	}
	if strings.Contains(str, "%%CONVERT%%") && !slices.Contains(prgNames, indexerutil.CheckProgramMagickConvert) {
		logger.Error().Msg("image magick convert needed, but not found on machine")
		return false
	}
	if strings.Contains(str, "%%GHOSTSCRIPT%%") && !slices.Contains(prgNames, indexerutil.CheckProgramGhostscript) {
		logger.Error().Msg("ghostscript needed, but not found on machine")
		return false
	}
	return true
}

func SWDoReplace(str string, prgs map[string]Program) string {
	if _, ok := prgs[indexerutil.CheckProgramMagickConvert]; ok {
		str = strings.Replace(str, "%%CONVERT%%", prgs[indexerutil.CheckProgramMagickConvert].Prg, -1)
		str = strings.Replace(str, "%%CONVERT_PARAMS%%", prgs[indexerutil.CheckProgramMagickConvert].Params, -1)
	}
	if _, ok := prgs[indexerutil.CheckProgramFFMpeg]; ok {
		str = strings.Replace(str, "%%FFMPEG%%", prgs[indexerutil.CheckProgramFFMpeg].Prg, -1)
		str = strings.Replace(str, "%%FFMPEG_PARAMS%%", prgs[indexerutil.CheckProgramFFMpeg].Params, -1)
	}
	if _, ok := prgs[indexerutil.CheckProgramGhostscript]; ok {
		str = strings.Replace(str, "%%GHOSTSCRIPT%%", prgs[indexerutil.CheckProgramGhostscript].Prg, -1)
		str = strings.Replace(str, "%%GHOSTSCRIPT_PARAMS%%", prgs[indexerutil.CheckProgramGhostscript].Params, -1)
	}
	return str
}

package indexer

import (
	"emperror.dev/errors"
	ironmaiden "github.com/je4/indexer/v3/pkg/indexer"
	"github.com/je4/utils/v2/pkg/zLogger"
	"os"
	"time"
)

func InitActions(relevance map[int]ironmaiden.MimeWeightString, siegfried *Siegfried, ffmpeg *FFMPEG, magick *ImageMagick, tika *Tika, logger zLogger.ZLogger) (*ironmaiden.ActionDispatcher, error) {
	ad := ironmaiden.NewActionDispatcher(relevance)
	if siegfried != nil && siegfried.Signature != "" {
		signatureData, err := os.ReadFile(siegfried.Signature)
		if err != nil {
			logger.Warningf("no siegfried signature file provided. using default signature file. please provide a recent signature file.")
		}
		logger.Info("indexer action siegfried added")
		_ = ironmaiden.NewActionSiegfried("siegfried", signatureData, siegfried.MimeMap, nil, ad)
	}
	if ffmpeg != nil && ffmpeg.Enabled {
		timeout, err := time.ParseDuration(ffmpeg.Timeout)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot parse ffmpeg timeout '%s'", ffmpeg.Timeout)
		}
		_ = ironmaiden.NewActionFFProbe(
			"ffprobe",
			ffmpeg.FFProbe,
			ffmpeg.WSL,
			timeout,
			ffmpeg.Online,
			ffmpeg.Mime,
			nil,
			ad)
		logger.Info("indexer action ffprobe added")
	}
	if magick != nil && magick.Enabled {
		timeout, err := time.ParseDuration(magick.Timeout)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot parse magick timeout '%s'", magick.Timeout)
		}
		_ = ironmaiden.NewActionIdentifyV2("identify", magick.Identify, magick.Convert, magick.WSL, timeout, magick.Online, nil, ad)
		logger.Info("indexer action identify added")
	}
	if tika != nil && tika.Enabled {
		timeout, err := time.ParseDuration(tika.Timeout)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot parse magick timeout '%s'", magick.Timeout)
		}
		_ = ironmaiden.NewActionTika("tika", tika.AddressMeta, timeout, tika.RegexpMimeMeta, tika.RegexpMimeMetaNot, "", tika.Online, nil, ad)
		logger.Info("indexer action tika added")

		_ = ironmaiden.NewActionTika("fulltext", tika.AddressFulltext, timeout, tika.RegexpMimeFulltext, tika.RegexpMimeFulltextNot, "X-TIKA:content", tika.Online, nil, ad)
		logger.Info("indexer action fulltext added")

	}

	return ad, nil
}

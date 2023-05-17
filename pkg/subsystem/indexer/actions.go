package indexer

import (
	"emperror.dev/errors"
	datasiegfried "github.com/je4/gocfl/v2/data/siegfried"
	ironmaiden "github.com/je4/indexer/v2/pkg/indexer"
	"github.com/op/go-logging"
	"os"
	"time"
)

func InitActions(relevance map[int]ironmaiden.MimeWeightString, siegfried *Siegfried, ffmpeg *FFMPEG, magick *ImageMagick, tika *Tika, logger *logging.Logger) (*ironmaiden.ActionDispatcher, error) {
	ad := ironmaiden.NewActionDispatcher(relevance)
	signatureData, err := os.ReadFile(siegfried.Signature)
	if err != nil {
		logger.Warningf("no siegfried signature file provided. using default signature file. please provide a recent signature file.")
		signatureData = datasiegfried.DefaultSig
	}
	_ = ironmaiden.NewActionSiegfried("siegfried", signatureData, siegfried.MimeMap, nil, ad)
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
		logger.Info("indexer action siegfried added")
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

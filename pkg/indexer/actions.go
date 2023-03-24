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
		logger.Warningf("no signature file provided. using default signature file. please provide a recent signature file.")
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
	}
	if magick != nil && magick.Enabled {
		timeout, err := time.ParseDuration(magick.Timeout)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot parse magick timeout '%s'", magick.Timeout)
		}
		_ = ironmaiden.NewActionIdentifyV2("identify", magick.Identify, magick.Convert, magick.WSL, timeout, magick.Online, nil, ad)
	}
	if tika != nil && tika.Enabled {
		timeout, err := time.ParseDuration(tika.Timeout)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot parse magick timeout '%s'", magick.Timeout)
		}
		_ = ironmaiden.NewActionTika("tika", tika.AddressMeta, timeout, tika.RegexpMimeMeta, tika.RegexpMimeMetaNot, "", tika.Online, nil, ad)
		_ = ironmaiden.NewActionTika("fulltext", tika.AddressFulltext, timeout, tika.RegexpMimeFulltext, tika.RegexpMimeFulltextNot, "X-TIKA:content", tika.Online, nil, ad)
	}

	return ad, nil
}

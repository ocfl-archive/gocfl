package indexer

import (
	"emperror.dev/errors"
	ironmaiden "github.com/je4/indexer/pkg/indexer"
	"github.com/op/go-logging"
	"time"
)

func InitActions(relevance map[int]ironmaiden.MimeWeightString, siegfried *Siegfried, ffmpeg *FFMPEG, magick *ImageMagick, tika *Tika, logger *logging.Logger) (*ironmaiden.ActionDispatcher, error) {
	ad := ironmaiden.NewActionDispatcher(relevance)
	_ = ironmaiden.NewActionSiegfried("siegfried", siegfried.Signature, siegfried.MimeMap, nil, ad)
	if ffmpeg.Enabled {
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
	if magick.Enabled {
		timeout, err := time.ParseDuration(magick.Timeout)
		/*
			if err != nil {
				return nil, errors.Wrapf(err, "cannot parse magick timeout '%s'", magick.Timeout)
			}
			_ = ironmaiden.NewActionIdentify("identify", magick.Identify, magick.Convert, magick.WSL, timeout, magick.Online, nil)
		*/
		if err != nil {
			return nil, errors.Wrapf(err, "cannot parse magick timeout '%s'", magick.Timeout)
		}
		_ = ironmaiden.NewActionIdentifyV2("identify", magick.Identify, magick.Convert, magick.WSL, timeout, magick.Online, nil, ad)
	}
	if tika.Enabled {
		timeout, err := time.ParseDuration(tika.Timeout)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot parse magick timeout '%s'", magick.Timeout)
		}
		_ = ironmaiden.NewActionTika("tika", tika.AddressMeta, timeout, tika.RegexpMimeMeta, tika.RegexpMimeMetaNot, "", tika.Online, nil, ad)
		_ = ironmaiden.NewActionTika("fulltext", tika.AddressFulltext, timeout, tika.RegexpMimeFulltext, tika.RegexpMimeFulltextNot, "X-TIKA:content", tika.Online, nil, ad)
	}

	return ad, nil
}

//go:build exclude

package indexer

import (
	"emperror.dev/errors"
	datasiegfried "github.com/je4/gocfl/v2/data/siegfried"
	ironmaiden "github.com/je4/indexer/v2/pkg/indexer"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/op/go-logging"
	"html/template"
	"net"
	"os"
	"time"
)

func StartIndexer(
	siegfried *Siegfried,
	ffmpeg *FFMPEG,
	magick *ImageMagick,
	tika *Tika,
	mimeRelevance map[int]ironmaiden.MimeWeightString,
	ad *ironmaiden.ActionDispatcher,
	logger *logging.Logger,
) (*ironmaiden.Server, net.Addr, error) {
	errorTemplate, err := template.New("foo").Parse(
		`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>{{.Status}} - {{.StatusText}}</title>
</head>
<body>
<h1>{{.Status}} - {{.StatusText}}</h1>
<h3>{{.Message}}</h3>
</body>
</html>
`)
	if err != nil {
		return nil, nil, errors.Wrap(err, "cannot parse error template")
	}
	srv, err := ironmaiden.NewServer(
		10*time.Second,
		2048,
		"",
		0,
		mimeRelevance,
		"",
		[]string{},
		false,
		logger,
		&checksum.NullWriter{},
		errorTemplate,
		"",
		nil,
		nil,
	)
	if err != nil {
		return nil, nil, errors.Wrap(err, "cannot create new server")
	}
	signatureData, err := os.ReadFile(siegfried.Signature)
	if err != nil {
		logger.Warningf("no signature file provided. using default signature file. please provide a recent signature file.")
		signatureData = datasiegfried.DefaultSig
	}
	_ = ironmaiden.NewActionSiegfried("siegfried", signatureData, siegfried.MimeMap, srv, ad)
	if ffmpeg.Enabled {
		timeout, err := time.ParseDuration(ffmpeg.Timeout)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "cannot parse ffmpeg timeout '%s'", ffmpeg.Timeout)
		}
		_ = ironmaiden.NewActionFFProbe(
			"ffprobe",
			ffmpeg.FFProbe,
			ffmpeg.WSL,
			timeout,
			ffmpeg.Online,
			ffmpeg.Mime,
			srv,
			ad)
	}
	if magick.Enabled {
		timeout, err := time.ParseDuration(magick.Timeout)
		/*
			if err != nil {
				return nil, nil, errors.Wrapf(err, "cannot parse magick timeout '%s'", magick.Timeout)
			}
			_ = ironmaiden.NewActionIdentify("identify", magick.Identify, magick.Convert, magick.WSL, timeout, magick.Online, srv)
		*/
		if err != nil {
			return nil, nil, errors.Wrapf(err, "cannot parse magick timeout '%s'", magick.Timeout)
		}
		_ = ironmaiden.NewActionIdentifyV2("identify", magick.Identify, magick.Convert, magick.WSL, timeout, magick.Online, srv, ad)
	}
	if tika.Enabled {
		timeout, err := time.ParseDuration(tika.Timeout)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "cannot parse magick timeout '%s'", magick.Timeout)
		}
		_ = ironmaiden.NewActionTika("tika", tika.AddressMeta, timeout, tika.RegexpMimeMeta, tika.RegexpMimeMetaNot, "", tika.Online, srv, ad)
		_ = ironmaiden.NewActionTika("fulltext", tika.AddressFulltext, timeout, tika.RegexpMimeFulltext, tika.RegexpMimeFulltextNot, "X-TIKA:content", tika.Online, srv, ad)
	}

	var addr net.Addr
	if srv != nil {
		for _, action := range ad.GetActions() {
			srv.AddActions(action)
		}

		// get a random free port
		l, err := net.Listen("tcp", "localhost:0")
		if err != nil {
			return nil, nil, errors.Wrap(err, "cannot listen to port :0")
		}
		addr = l.Addr()
		if err := l.Close(); err != nil {
			return nil, nil, errors.Wrap(err, "error closing test listener")
		}

		go func() {
			logger.Infof("starting indexer http server at http://%s", addr.String())
			if err := srv.ListenAndServe(addr.String(), "", ""); err != nil {
				logger.Errorf("http server stopped: %v", err)
			} else {
				logger.Infof("http server stopped normally")
			}
		}()
		time.Sleep(1 * time.Second)
	}
	return srv, addr, err
}

package indexer

import (
	"emperror.dev/errors"
	"github.com/je4/gocfl/v2/pkg/checksum"
	ironmaiden "github.com/je4/indexer/pkg/indexer"
	"github.com/op/go-logging"
	"html/template"
	"net"
	"time"
)

func StartIndexer(
	siegfried *Siegfried,
	ffmpeg *FFMPEG,
	magick *ImageMagick,
	tika *Tika,
	mimeRelevance map[int]ironmaiden.MimeWeightString,
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
	_ = ironmaiden.NewActionSiegfried(siegfried.Signature, siegfried.MimeMap, srv)
	if ffmpeg.Enabled {
		timeout, err := time.ParseDuration(ffmpeg.Timeout)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "cannot parse ffmpeg timeout '%s'", ffmpeg.Timeout)
		}
		_ = ironmaiden.NewActionFFProbe(
			ffmpeg.FFProbe,
			ffmpeg.WSL,
			timeout,
			ffmpeg.Online,
			ffmpeg.Mime,
			srv)
	}
	if magick.Enabled {
		timeout, err := time.ParseDuration(magick.Timeout)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "cannot parse magick timeout '%s'", magick.Timeout)
		}
		_ = ironmaiden.NewActionIdentify(magick.Identify, magick.Convert, magick.WSL, timeout, magick.Online, srv)
	}
	if tika.Enabled {
		timeout, err := time.ParseDuration(tika.Timeout)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "cannot parse magick timeout '%s'", magick.Timeout)
		}
		_ = ironmaiden.NewActionTika(tika.Address, timeout, tika.RegexpMime, tika.Online, srv)
	}

	// get a random free port
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return nil, nil, errors.Wrap(err, "cannot listen to port :0")
	}
	addr := l.Addr()
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
	return srv, addr, err
}

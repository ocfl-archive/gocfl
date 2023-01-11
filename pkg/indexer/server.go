package indexer

import (
	"emperror.dev/errors"
	"github.com/je4/gocfl/v2/pkg/checksum"
	iron "github.com/je4/indexer/pkg/indexer"
	"github.com/op/go-logging"
	"net"
	"time"
)

func StartIndexer(signatureFile string, mimeMap map[string]string, mimeRelevance map[int]iron.MimeWeightString, logger *logging.Logger) (*iron.Server, error) {
	srv, err := iron.NewServer(
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
		"",
		"",
		nil,
		nil,
	)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create new server")
	}
	_ = iron.NewActionSiegfried(signatureFile, mimeMap, srv)

	// get a random free port
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return nil, errors.Wrap(err, "cannot listen to port :0")
	}
	addr := l.Addr()
	l.Close()

	go func() {
		if err := srv.ListenAndServe(addr.String(), "", ""); err != nil {
			logger.Errorf("http server stopped: %v", err)
		} else {
			logger.Infof("http server stopped normally")
		}
	}()

	return srv, err

}

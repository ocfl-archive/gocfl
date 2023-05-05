package display

import (
	"context"
	"crypto/tls"
	"emperror.dev/emperror"
	"emperror.dev/errors"
	"github.com/gin-gonic/gin"
	"github.com/je4/gocfl/v2/data/displaydata"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	dcert "github.com/je4/utils/v2/pkg/cert"
	"github.com/op/go-logging"
	"html/template"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"time"
)

type Server struct {
	service        string
	host, port     string
	name, password string
	srv            *http.Server
	linkTokenExp   time.Duration
	jwtKey         string
	jwtAlg         []string
	log            *logging.Logger
	urlExt         *url.URL
	accessLog      io.Writer
	dataFS         fs.FS
	storageRoot    ocfl.StorageRoot
	object         ocfl.Object
	metadata       *ocfl.ObjectMetadata
}

func NewServer(storageRoot ocfl.StorageRoot, service, addr string, urlExt *url.URL, dataFS fs.FS, log *logging.Logger, accessLog io.Writer) (*Server, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, emperror.Wrapf(err, "cannot split address %s", addr)
	}

	srv := &Server{
		service:     service,
		host:        host,
		port:        port,
		urlExt:      urlExt,
		dataFS:      dataFS,
		log:         log,
		accessLog:   accessLog,
		storageRoot: storageRoot,
	}

	return srv, nil
}

func (s *Server) ListenAndServe(cert, key string) (err error) {
	route := gin.Default()

	route.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	tpl, err := template.New("dashboard.tmpl").Funcs(template.FuncMap{
		"formatTime": func(t time.Time) string {
			return t.Format(time.RFC3339)
		},
	}).Parse(displaydata.DashboardTemplate)
	if err != nil {
		return errors.Wrap(err, "cannot parse dashboard template")
	}
	route.SetHTMLTemplate(tpl)
	route.GET("/", s.dashboard)
	route.GET("/load/id/:id", s.loadObjectID)
	route.GET("/load/path/:path", s.loadObjectPath)

	route.StaticFS("/static", http.FS(s.dataFS))

	s.srv = &http.Server{
		Addr:    net.JoinHostPort(s.host, s.port),
		Handler: route.Handler(),
	}

	if cert == "auto" || key == "auto" {
		s.log.Info("generating new certificate")
		cert, err := dcert.DefaultCertificate()
		if err != nil {
			return errors.Wrap(err, "cannot generate default certificate")
		}
		s.srv.TLSConfig = &tls.Config{Certificates: []tls.Certificate{*cert}}
		s.log.Infof("starting gocfl viewer at %v - https://%s:%v/", s.urlExt.String(), s.host, s.port)
		return errors.WithStack(s.srv.ListenAndServeTLS("", ""))
	} else if cert != "" && key != "" {
		s.log.Infof("starting gocfl viewer at %v - https://%s:%v/", s.urlExt.String(), s.host, s.port)
		return errors.WithStack(s.srv.ListenAndServeTLS(cert, key))
	} else {
		s.log.Infof("starting gocfl viewer at %v - http://%s:%v/", s.urlExt.String(), s.host, s.port)
		return errors.WithStack(s.srv.ListenAndServe())
	}
}

func (s *Server) dashboard(c *gin.Context) {

	c.HTML(http.StatusOK, "dashboard.tmpl", gin.H{
		"title": "Dashboard",
	})
}

func (s *Server) loadObjectID(c *gin.Context) {
	var err error
	type idParam struct {
		ID string `uri:"id" binding:"required"`
	}
	var iop idParam
	if err = c.ShouldBindUri(&iop); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	s.object, err = s.storageRoot.LoadObjectByID(iop.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	s.metadata, err = s.object.GetMetadata()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrapf(err, "cannot get metadata for object %s", s.object.GetID()).Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": s.object.GetID()})
}

func (s *Server) loadObjectPath(c *gin.Context) {
	var err error
	type pathParam struct {
		Path string `uri:"path" binding:"required"`
	}
	var iop pathParam
	if err = c.ShouldBindUri(&iop); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	s.object, err = s.storageRoot.LoadObjectByFolder(iop.Path)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	s.metadata, err = s.object.GetMetadata()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrapf(err, "cannot get metadata for object %s", s.object.GetID()).Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": s.object.GetID()})
}

func (s *Server) Shutdown(ctx context.Context) error {
	return errors.WithStack(s.srv.Shutdown(ctx))
}

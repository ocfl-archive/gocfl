package display

import (
	"context"
	"crypto/tls"
	"emperror.dev/emperror"
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"github.com/Masterminds/sprig/v3"
	"github.com/dustin/go-humanize"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	dcert "github.com/je4/utils/v2/pkg/cert"
	"github.com/je4/utils/v2/pkg/checksum"
	iou "github.com/je4/utils/v2/pkg/io"
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/extension"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl"
	"github.com/ocfl-archive/indexer/v3/pkg/indexer"
	"golang.org/x/exp/maps"
	"html/template"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"sort"
	"strings"
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
	log            zLogger.ZLogger
	urlExt         *url.URL
	accessLog      io.Writer
	dataFS         fs.FS
	storageRoot    ocfl.StorageRoot
	object         ocfl.Object
	metadata       *ocfl.ObjectMetadata
	templateFS     fs.FS
	obfuscate      bool
	objectFS       http.FileSystem
}

func NewServer(storageRoot ocfl.StorageRoot, service, addr string, urlExt *url.URL, dataFS, templateFS fs.FS, obfuscate bool, log zLogger.ZLogger, accessLog io.Writer) (*Server, error) {
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
		templateFS:  templateFS,
		log:         log,
		accessLog:   accessLog,
		storageRoot: storageRoot,
		obfuscate:   obfuscate,
	}

	return srv, nil
}

func (s *Server) ListenAndServe(cert, key string) (err error) {
	gin.SetMode(gin.ReleaseMode)
	route := gin.Default()
	route.UseRawPath = true
	route.UnescapePathValues = false

	route.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	mt := multitemplate.New()

	var tplfiles []string = []string{
		"storageroot.gohtml",
		"object.gohtml",
		"manifest.gohtml",
		"version.gohtml",
		"detail.gohtml",
		"report.gohtml",
	}

	for _, tplfile := range tplfiles {
		funcMap := sprig.FuncMap()
		funcMap["basename"] = func(str string) string {
			return filepath.Base(str)
		}
		funcMap["PathEscape"] = func(str string) string {
			return url.PathEscape(str)
		}
		funcMap["humanizeBytes"] = func(size uint64) string {
			return humanize.Bytes(size)
		}
		funcMap["humanizeTime"] = func(t time.Time) string {
			return t.Format("2006-01-02 15:04:05")
		}

		tpl, err := template.New(tplfile).Funcs(funcMap).ParseFS(s.templateFS, tplfile)
		if err != nil {
			return errors.Wrapf(err, "cannot parse template %s", tplfile)
		}
		mt.Add(tplfile, tpl)
	}

	route.HTMLRender = mt
	route.GET("/", s.storageroot)
	//	route.GET("/:id", s.dashboard)

	route.GET("/object/id/:id", s.loadObjectID)
	route.GET("/object/id/:id/manifest", s.manifest)
	route.GET("/object/id/:id/version/:version", s.version)
	route.GET("/object/id/:id/detail/:checksum", s.detail)
	route.GET("/object/id/:id/report", s.report)
	route.GET("/object/id/:id/download/:checksum/:filename", s.download)
	route.GET("/object/id/:id/extension/:extension/download/*path", s.downloadExtFile)
	route.GET("/object/folder/*path", s.loadObjectPath)
	route.GET("/object/id/:id/browse/*path", s.loadObjectBrowser)

	route.StaticFS("/static", http.FS(s.dataFS))

	s.srv = &http.Server{
		Addr:    net.JoinHostPort(s.host, s.port),
		Handler: route.Handler(),
	}

	if cert == "auto" || key == "auto" {
		s.log.Info().Msg("generating new certificate")
		cert, err := dcert.DefaultCertificate()
		if err != nil {
			return errors.Wrap(err, "cannot generate default certificate")
		}
		s.srv.TLSConfig = &tls.Config{Certificates: []tls.Certificate{*cert}}
		fmt.Printf("starting gocfl viewer at %v - https://%s:%v/", s.urlExt.String(), s.host, s.port)
		return errors.WithStack(s.srv.ListenAndServeTLS("", ""))
	} else if cert != "" && key != "" {
		fmt.Printf("starting gocfl viewer at %v - https://%s:%v/", s.urlExt.String(), s.host, s.port)
		return errors.WithStack(s.srv.ListenAndServeTLS(cert, key))
	} else {
		fmt.Printf("starting gocfl viewer at %v - http://%s:%v/", s.urlExt.String(), s.host, s.port)
		return errors.WithStack(s.srv.ListenAndServe())
	}
}

func (s *Server) downloadExtFile(c *gin.Context) {
	var err error
	type idParam struct {
		ID        string `uri:"id" binding:"required"`
		Path      string `uri:"path" binding:"required"`
		Extension string `uri:"extension" binding:"required"`
	}
	var iop idParam
	if err = c.ShouldBindUri(&iop); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iop.ID, err = url.PathUnescape(iop.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrapf(err, "cannot unescape '%s'", iop.ID).Error()})
		return
	}
	iop.Path, err = url.PathUnescape(iop.Path)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrapf(err, "cannot unescape '%s'", iop.Path).Error()})
		return
	}
	iop.Extension, err = url.PathUnescape(iop.Extension)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrapf(err, "cannot unescape '%s'", iop.Extension).Error()})
		return
	}

	if s.object != nil && s.object.GetID() == iop.ID {
		if s.metadata == nil {
			s.metadata, err = s.object.GetMetadata()
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrapf(err, "cannot get metadata for object %s", s.object.GetID()).Error()})
				return
			}
			if s.obfuscate {
				if err := s.metadata.Obfuscate(); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrapf(err, "cannot obfuscate metadata").Error()})
					return
				}
			}
		}
	} else {
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
		if s.obfuscate {
			if err := s.metadata.Obfuscate(); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrapf(err, "cannot obfuscate metadata").Error()})
				return
			}
		}
	}
	pathStr := filepath.ToSlash(filepath.Join("extensions", iop.Extension, iop.Path))
	fp, err := s.object.GetFS().Open(pathStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer fp.Close()
	fi, err := fp.Stat()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mimeReader, err := iou.NewMimeReader(fp)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrapf(err, "cannot instantiate mimereader for object %s - %s", s.object.GetID(), pathStr).Error()})
		return
	}
	contentType, err := mimeReader.DetectContentType()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrapf(err, "cannot detect content-type for object %s - %s", s.object.GetID(), pathStr).Error()})
		return
	}
	c.DataFromReader(http.StatusOK, fi.Size(), contentType, mimeReader, map[string]string{})
}

func (s *Server) download(c *gin.Context) {
	var err error
	type idParam struct {
		ID       string `uri:"id" binding:"required"`
		Checksum string `uri:"checksum" binding:"required"`
	}
	var iop idParam
	if err = c.ShouldBindUri(&iop); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iop.ID, err = url.PathUnescape(iop.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrapf(err, "cannot unescape '%s'", iop.ID).Error()})
		return
	}

	if s.object != nil && s.object.GetID() == iop.ID {
		if s.metadata == nil {
			s.metadata, err = s.object.GetMetadata()
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrapf(err, "cannot get metadata for object %s", s.object.GetID()).Error()})
				return
			}
			if s.obfuscate {
				if err := s.metadata.Obfuscate(); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrapf(err, "cannot obfuscate metadata").Error()})
					return
				}
			}
		}
	} else {
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
		if s.obfuscate {
			if err := s.metadata.Obfuscate(); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrapf(err, "cannot obfuscate metadata").Error()})
				return
			}
		}
	}

	file, ok := s.metadata.Files[iop.Checksum]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Errorf("no file with checksum %s found", iop.Checksum).Error()})
		return
	}

	fp, err := s.object.GetFS().Open(file.InternalName[0])
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer fp.Close()
	fi, err := fp.Stat()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.DataFromReader(http.StatusOK, fi.Size(), "application/octet-stream", fp, map[string]string{})

}

func (s *Server) detail(c *gin.Context) {
	var err error
	type idParam struct {
		ID       string `uri:"id" binding:"required"`
		Checksum string `uri:"checksum" binding:"required"`
	}
	var iop idParam
	if err = c.ShouldBindUri(&iop); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iop.ID, err = url.PathUnescape(iop.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrapf(err, "cannot unescape '%s'", iop.ID).Error()})
		return
	}

	if s.object != nil && s.object.GetID() == iop.ID {
		if s.metadata == nil {
			s.metadata, err = s.object.GetMetadata()
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrapf(err, "cannot get metadata for object %s", s.object.GetID()).Error()})
				return
			}
			if s.obfuscate {
				if err := s.metadata.Obfuscate(); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrapf(err, "cannot obfuscate metadata").Error()})
					return
				}
			}
		}
	} else {
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
		if s.obfuscate {
			if err := s.metadata.Obfuscate(); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrapf(err, "cannot obfuscate metadata").Error()})
				return
			}
		}
	}

	file, ok := s.metadata.Files[iop.Checksum]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Errorf("no file with checksum %s found", iop.Checksum).Error()})
		return
	}
	type extFEntry struct {
		Name  string
		ATime string
		CTime string
		MTime string
		Size  string
		Attr  string
		OS    string
		Sys   string
	}
	type detailStatus struct {
		Checksum        string                              `json:"checksum"`
		DigestAlgorithm checksum.DigestAlgorithm            `json:"digestAlgorithm"`
		InternalNames   []string                            `json:"internalNames"`
		ExternalNames   map[string][]*extFEntry             `json:"externalNames"`
		Fixity          map[checksum.DigestAlgorithm]string `json:"fixity"`
		Indexer         *indexer.ResultV2                   `json:"indexer"`
		IndexerJSON     string
		Migration       *extension.MigrationResult
		Thumbnail       *extension.ThumbnailResult
	}

	status := &detailStatus{
		Checksum:        iop.Checksum,
		DigestAlgorithm: s.metadata.DigestAlgorithm,
		InternalNames:   file.InternalName,
		ExternalNames:   map[string][]*extFEntry{},
		Fixity:          file.Checksums,
	}

	extFilesystemAny, _ := file.Extension[extension.FilesystemName]
	var extFilesystem map[string][]*extension.FileSystemLine
	if extFilesystemAny != nil {
		extFilesystem, _ = extFilesystemAny.(map[string][]*extension.FileSystemLine)
	}

	for ver, names := range file.VersionName {
		extFilesystemVersion, _ := extFilesystem[ver]
		if status.ExternalNames[ver] == nil {
			status.ExternalNames[ver] = []*extFEntry{}
		}
		for _, name := range names {
			efe := &extFEntry{
				Name: name,
			}
			if extFilesystemVersion != nil {
				for _, fs := range extFilesystemVersion {
					if fs.Path == name {
						efe.ATime = fs.Meta.ATime.Format(time.DateTime)
						efe.CTime = fs.Meta.CTime.Format(time.DateTime)
						efe.MTime = fs.Meta.MTime.Format(time.DateTime)
						efe.Size = humanize.Bytes(uint64(fs.Meta.Size))
						efe.Attr = fs.Meta.Attr
						efe.OS = fs.Meta.OS
						sys, _ := json.MarshalIndent(fs.Meta.SystemStat, "", "  ")
						efe.Sys = string(sys)
						break
					}
				}
			}
			status.ExternalNames[ver] = append(status.ExternalNames[ver], efe)
		}
	}

	extIndexerAny, _ := file.Extension[extension.IndexerName]
	var extIndexer *indexer.ResultV2
	if extIndexerAny != nil {
		extIndexer, _ = extIndexerAny.(*indexer.ResultV2)
	}

	if extIndexer != nil {
		iData, err := json.MarshalIndent(extIndexer.Metadata, "", "  ")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrapf(err, "cannot marshal indexer metadata for object %s", s.object.GetID()).Error()})
			return
		}
		//	extIndexer.Metadata = nil
		status.Indexer = extIndexer
		status.IndexerJSON = string(iData)
	}

	extMigrationAny, _ := file.Extension[extension.MigrationName]
	var extMigration *extension.MigrationResult
	if extMigrationAny != nil {
		extMigration, _ = extMigrationAny.(*extension.MigrationResult)
	}
	if extMigration != nil {
		status.Migration = extMigration
	}

	extThumbnailAny, _ := file.Extension[extension.ThumbnailName]
	if extThumbnailAny != nil {
		if extThumbnail, ok := extThumbnailAny.(extension.ThumbnailResult); ok {
			status.Thumbnail = &extThumbnail
		} else {
			if extThumbnail, ok := extThumbnailAny.(*extension.ThumbnailResult); ok {
				status.Thumbnail = extThumbnail
			}
		}
	}

	var params = map[string]any{
		"title":  "Detail",
		"id":     s.object.GetID(),
		"status": status,
		//		"metadata": s.metadata,
		"file": file,
	}

	c.HTML(http.StatusOK, "detail.gohtml", gin.H(params))

}

func (s *Server) dashboard(c *gin.Context) {

	var id string
	if s.object != nil {
		id = s.object.GetID()
	}
	c.HTML(http.StatusOK, "object.gohtml", gin.H{
		"title": "gocfl",
		"id":    id,
	})
}

func (s *Server) storageroot(c *gin.Context) {

	var id string
	if s.object != nil {
		id = s.object.GetID()
	}

	if s.storageRoot == nil {
		c.JSON(http.StatusInternalServerError, "no storage root loaded")
		return
	}

	folders, err := s.storageRoot.GetObjectFolders()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.HTML(http.StatusOK, "storageroot.gohtml", gin.H{
		"title":       "gocfl",
		"id":          id,
		"folders":     folders,
		"storageroot": s.storageRoot.String(),
	})
}

func (s *Server) manifest(c *gin.Context) {
	var err error
	type idParam struct {
		ID string `uri:"id" binding:"required"`
	}
	var iop idParam
	if err = c.ShouldBindUri(&iop); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iop.ID, err = url.PathUnescape(iop.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrapf(err, "cannot unescape '%s'", iop.ID).Error()})
		return
	}

	if s.object != nil && s.object.GetID() == iop.ID {
		if s.metadata == nil {
			s.metadata, err = s.object.GetMetadata()
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrapf(err, "cannot get metadata for object %s", s.object.GetID()).Error()})
				return
			}
			if s.obfuscate {
				if err := s.metadata.Obfuscate(); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrapf(err, "cannot obfuscate metadata").Error()})
					return
				}
			}

		}
	} else {
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
		if s.obfuscate {
			if err := s.metadata.Obfuscate(); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrapf(err, "cannot obfuscate metadata").Error()})
				return
			}
		}

	}

	type fEntry struct {
		Checksum  string
		Pronom    string
		Mimetype  string
		IdxSize   string
		Migration map[string]string
	}
	var files = map[string]*fEntry{}
	var filenames = []string{}

	for checksum, file := range s.metadata.Files {
		extMigrationAny, _ := file.Extension[extension.MigrationName]
		var extMigration *extension.MigrationResult
		if extMigrationAny != nil {
			extMigration = extMigrationAny.(*extension.MigrationResult)
		}
		extIndexerAny, _ := file.Extension[extension.IndexerName]
		var extIndexer *indexer.ResultV2
		if extIndexerAny != nil {
			extIndexer, _ = extIndexerAny.(*indexer.ResultV2)
		}

		for _, name := range file.InternalName {
			fe := &fEntry{
				Checksum:  checksum,
				Migration: map[string]string{},
			}
			if extIndexer != nil {
				fe.Pronom = extIndexer.Pronom
				fe.Mimetype = extIndexer.Mimetype
				fe.IdxSize = humanize.Bytes(extIndexer.Size)
			}
			if extMigration != nil {
				fe.Migration[extMigration.ID] = extMigration.Source
			}
			files[name] = fe
			filenames = append(filenames, name)
		}
	}

	var params = map[string]any{
		"title":     "Manifest",
		"id":        s.object.GetID(),
		"versions":  s.metadata.Versions,
		"files":     files,
		"filenames": filenames,
	}

	c.HTML(http.StatusOK, "manifest.gohtml", gin.H(params))

}

func (s *Server) version(c *gin.Context) {
	var err error
	type idParam struct {
		ID      string `uri:"id" binding:"required"`
		Version string `uri:"version" binding:"required"`
	}
	var iop idParam
	if err = c.ShouldBindUri(&iop); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iop.ID, err = url.PathUnescape(iop.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrapf(err, "cannot unescape '%s'", iop.ID).Error()})
		return
	}

	if s.object != nil && s.object.GetID() == iop.ID {
		if s.metadata == nil {
			s.metadata, err = s.object.GetMetadata()
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrapf(err, "cannot get metadata for object %s", s.object.GetID()).Error()})
				return
			}
			if s.obfuscate {
				if err := s.metadata.Obfuscate(); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrapf(err, "cannot obfuscate metadata").Error()})
					return
				}
			}

		}
	} else {
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
		if s.obfuscate {
			if err := s.metadata.Obfuscate(); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrapf(err, "cannot obfuscate metadata").Error()})
				return
			}
		}

	}

	type fEntry struct {
		CTime     string
		Size      string
		Checksum  string
		Pronom    string
		Mimetype  string
		IdxSize   string
		Attr      string
		OS        string
		Migration map[string]string
	}
	var files = map[string]*fEntry{}
	var filenames = []string{}

	for checksum, file := range s.metadata.Files {
		extMigrationAny, _ := file.Extension[extension.MigrationName]
		var extMigration *extension.MigrationResult
		if extMigrationAny != nil {
			extMigration = extMigrationAny.(*extension.MigrationResult)
		}
		extIndexerAny, _ := file.Extension[extension.IndexerName]
		var extIndexer *indexer.ResultV2
		if extIndexerAny != nil {
			extIndexer, _ = extIndexerAny.(*indexer.ResultV2)
		}
		extFilesystemAny, _ := file.Extension[extension.FilesystemName]
		var extFilesystem map[string][]*extension.FileSystemLine
		if extFilesystemAny != nil {
			extFilesystem, _ = extFilesystemAny.(map[string][]*extension.FileSystemLine)
		}
		extFilesystemVersion, _ := extFilesystem[iop.Version]
		if vNames, ok := file.VersionName[iop.Version]; ok {
			for _, name := range vNames {
				fe := &fEntry{
					Checksum:  checksum,
					Migration: map[string]string{},
				}
				if extMigration != nil {
					fe.Migration[extMigration.ID] = extMigration.Source
				}
				if extFilesystemVersion != nil {
					for _, fsLine := range extFilesystemVersion {
						if fsLine.Path == name {
							fe.Size = humanize.Bytes(fsLine.Meta.Size)
							fe.CTime = fsLine.Meta.CTime.Format(time.RFC3339)
							fe.OS = fsLine.Meta.OS
							fe.Attr = fsLine.Meta.Attr
							break
						}
					}
				}
				if extIndexer != nil {
					fe.Pronom = extIndexer.Pronom
					fe.Mimetype = extIndexer.Mimetype
					fe.IdxSize = humanize.Bytes(extIndexer.Size)
				}
				files[name] = fe
				filenames = append(filenames, name)
			}
		}
	}

	var params = map[string]any{
		"title":     "Version",
		"id":        s.object.GetID(),
		"versions":  s.metadata.Versions,
		"files":     files,
		"filenames": filenames,
		"version":   iop.Version,
	}

	c.HTML(http.StatusOK, "version.gohtml", gin.H(params))

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
	iop.ID, err = url.PathUnescape(iop.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrapf(err, "cannot unescape '%s'", iop.ID).Error()})
		return
	}
	if s.object != nil && s.object.GetID() == iop.ID {
		// already loaded
		if s.metadata == nil {
			s.metadata, err = s.object.GetMetadata()
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrapf(err, "cannot get metadata for object %s", s.object.GetID()).Error()})
				return
			}
			if s.obfuscate {
				if err := s.metadata.Obfuscate(); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrapf(err, "cannot obfuscate metadata").Error()})
					return
				}
			}
		}
	} else {
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
		if s.obfuscate {
			if err := s.metadata.Obfuscate(); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrapf(err, "cannot obfuscate metadata").Error()})
				return
			}
		}
	}
	s.displayObject(c)
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
	s.object, err = s.storageRoot.LoadObjectByFolder(strings.Trim(iop.Path, "/"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	s.metadata, err = s.object.GetMetadata()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrapf(err, "cannot get metadata for object %s", s.object.GetID()).Error()})
		return
	}
	if s.obfuscate {
		if err := s.metadata.Obfuscate(); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrapf(err, "cannot obfuscate metadata").Error()})
			return
		}
	}

	c.Redirect(http.StatusPermanentRedirect, s.urlExt.String()+fmt.Sprintf("/object/id/%s", url.PathEscape(s.object.GetID())))
	//	s.displayObject(c)
}

func (s *Server) displayObject(c *gin.Context) {

	if s.metadata == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no metadata loaded"})
		return
	}
	var numFiles int
	var size uint64
	var noSizeFiles int
	var mimeTypes = make(map[string]int)
	var pronoms = make(map[string]int)
	for _, v := range s.metadata.Files {
		numFiles += len(v.InternalName)
		_fs, _ := v.Extension[extension.FilesystemName]
		_idx, _ := v.Extension[extension.IndexerName]
		var fs map[string]any
		var idx *indexer.ResultV2
		var ok bool
		var sizeDone bool
		if _fs != nil {
			if fs, ok = _fs.(map[string]any); ok {
				if fs["size"] != nil {
					size += fs["size"].(uint64)
					sizeDone = true
				}
			}
		}
		if _idx != nil {
			if idx, ok = _idx.(*indexer.ResultV2); ok {
				size += idx.Size
				if idx.Size > 0 {
					sizeDone = true
				}
				if idx.Mimetype != "" {
					if _, ok := mimeTypes[idx.Mimetype]; !ok {
						mimeTypes[idx.Mimetype] = 0
					}
					mimeTypes[idx.Mimetype]++
				}
				if idx.Pronom != "" {
					if _, ok := pronoms[idx.Pronom]; !ok {
						pronoms[idx.Pronom] = 0
					}
					pronoms[idx.Pronom]++
				}
			}
		}
		if !sizeDone {
			noSizeFiles++
		}
	}
	var params = map[string]any{
		"title":          "gocfl",
		"id":             s.object.GetID(),
		"versions":       s.metadata.Versions,
		"differentFiles": len(s.metadata.Files),
		"numFiles":       numFiles,
		"size":           humanize.Bytes(size),
		"noSizeFiles":    noSizeFiles,
		"mimeTypes":      mimeTypes,
		"pronoms":        pronoms,
	}

	c.HTML(http.StatusOK, "object.gohtml", gin.H(params))
}

func (s *Server) loadObjectBrowser(c *gin.Context) {
	var err error
	type idParam struct {
		ID string `uri:"id" binding:"required"`
	}
	var iop idParam
	if err = c.ShouldBindUri(&iop); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iop.ID, err = url.PathUnescape(iop.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrapf(err, "cannot unescape '%s'", iop.ID).Error()})
		return
	}
	if s.object != nil && s.object.GetID() == iop.ID {
		// already loaded
		if s.metadata == nil {
			s.metadata, err = s.object.GetMetadata()
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrapf(err, "cannot get metadata for object %s", s.object.GetID()).Error()})
				return
			}
			if s.obfuscate {
				if err := s.metadata.Obfuscate(); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrapf(err, "cannot obfuscate metadata").Error()})
					return
				}
			}
		}
	} else {
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
		if s.obfuscate {
			if err := s.metadata.Obfuscate(); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrapf(err, "cannot obfuscate metadata").Error()})
				return
			}
		}
	}
	if s.objectFS == nil {
		objectFS, err := NewObjectFS(s.object)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrapf(err, "cannot get filesystem for object %s", s.object.GetID()).Error()})
			return
		}
		s.objectFS = http.FS(objectFS)
	}
	s.displayObjectBrowse(c)
}
func (s *Server) displayObjectBrowse(c *gin.Context) {
	path := c.Param("path")
	c.FileFromFS(path, s.objectFS)

}

func (s *Server) loadObjectContentRoot(c *gin.Context) {
	var err error
	type idParam struct {
		ID string `uri:"id" binding:"required"`
	}
	var iop idParam
	if err = c.ShouldBindUri(&iop); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iop.ID, err = url.PathUnescape(iop.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrapf(err, "cannot unescape '%s'", iop.ID).Error()})
		return
	}
	if s.object != nil && s.object.GetID() == iop.ID {
		// already loaded
		if s.metadata == nil {
			s.metadata, err = s.object.GetMetadata()
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrapf(err, "cannot get metadata for object %s", s.object.GetID()).Error()})
				return
			}
			if s.obfuscate {
				if err := s.metadata.Obfuscate(); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrapf(err, "cannot obfuscate metadata").Error()})
					return
				}
			}
		}
	} else {
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
		if s.obfuscate {
			if err := s.metadata.Obfuscate(); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrapf(err, "cannot obfuscate metadata").Error()})
				return
			}
		}
	}
	s.displayObjectContentRoot(c)
}

func (s *Server) displayObjectContentRoot(c *gin.Context) {
	if s.metadata == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no metadata loaded"})
		return
	}
	// currURL := strings.TrimSuffix(c.Request.URL.String(), "/") + "/"
	currURL := c.Request.URL.String()
	versions := maps.Keys(s.metadata.Versions)
	sort.Strings(versions)
	c.Header("Content-Type", "text/html")
	head := fmt.Sprintf(
		`<html>
<head><title>%s</title></head>
<body>
   <ul>
`,
		s.metadata.ID)
	io.WriteString(c.Writer, head)
	latestStr, _ := url.JoinPath(currURL, "latest")
	for _, version := range versions {
		var str string
		versionStr, _ := url.JoinPath(currURL, version)
		if version == s.metadata.Head {
			str = fmt.Sprintf(`<li><a href="%s">%s</a> (<a href="%s">latest</a>)</li>`, versionStr, version, latestStr)
		} else {
			str = fmt.Sprintf(`<li><a href="%s">%s</a></li>`, versionStr, version)
		}
		io.WriteString(c.Writer, str)
	}
	footer := `</ul>
</body>
</html>`
	io.WriteString(c.Writer, footer)
}

func (s *Server) report(c *gin.Context) {

	var err error
	type idParam struct {
		ID string `uri:"id" binding:"required"`
	}
	var iop idParam
	if err = c.ShouldBindUri(&iop); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iop.ID, err = url.PathUnescape(iop.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrapf(err, "cannot unescape '%s'", iop.ID).Error()})
		return
	}
	full := c.DefaultQuery("full", "none") != "none"

	if s.object != nil && s.object.GetID() == iop.ID {
		if s.metadata == nil {
			s.metadata, err = s.object.GetMetadata()
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrapf(err, "cannot get metadata for object %s", s.object.GetID()).Error()})
				return
			}
			if s.obfuscate {
				if err := s.metadata.Obfuscate(); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrapf(err, "cannot obfuscate metadata").Error()})
					return
				}
			}

		}
	} else {
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
		if s.obfuscate {
			if err := s.metadata.Obfuscate(); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrapf(err, "cannot obfuscate metadata").Error()})
				return
			}
		}

	}

	if s.metadata == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no metadata loaded"})
		return
	}

	extManager := s.object.GetExtensionManager()
	inventory := s.object.GetInventory()

	type mimeCount struct {
		SizeStr string
		Size    uint64
		Count   int
	}
	var numFiles int
	var size uint64
	var noSizeFiles int
	var mimeTypes = make(map[string]*mimeCount)
	var pronoms = make(map[string]*mimeCount)
	var videoSecs uint
	for _, v := range s.metadata.Files {
		numFiles += len(v.InternalName)
		_fs, _ := v.Extension[extension.FilesystemName]
		_idx, _ := v.Extension[extension.IndexerName]
		if _thumb, ok := v.Extension[extension.ThumbnailName]; ok {
			if thumb, ok := _thumb.(extension.ThumbnailResult); ok {
				if len(thumb.SourceDigest) > 0 {
					continue
				}
			}
		}
		var fs map[string]any
		var idx *indexer.ResultV2
		var ok bool
		var sizeDone bool
		if _fs != nil {
			if fs, ok = _fs.(map[string]any); ok {
				if fs["size"] != nil {
					size += fs["size"].(uint64)
					sizeDone = true
				}
			}
		}
		if _idx != nil {
			if idx, ok = _idx.(*indexer.ResultV2); ok {
				size += idx.Size
				videoSecs += idx.Duration
				if idx.Size > 0 {
					sizeDone = true
				}
				if idx.Mimetype != "" {
					if _, ok := mimeTypes[idx.Mimetype]; !ok {
						mimeTypes[idx.Mimetype] = &mimeCount{
							SizeStr: "",
							Size:    0,
							Count:   0,
						}
					}
					mimeTypes[idx.Mimetype].Count++
					mimeTypes[idx.Mimetype].Size += idx.Size
				}
				if idx.Pronom != "" {
					if _, ok := pronoms[idx.Pronom]; !ok {
						pronoms[idx.Pronom] = &mimeCount{
							Size:  0,
							Count: 0,
						}
					}
					pronoms[idx.Pronom].Count++
					pronoms[idx.Pronom].Size += idx.Size
				}
			}
		}
		if !sizeDone {
			noSizeFiles++
		}
	}

	for _, pronomSize := range pronoms {
		pronomSize.SizeStr = humanize.Bytes(pronomSize.Size)
	}
	for _, mimeSize := range mimeTypes {
		mimeSize.SizeStr = humanize.Bytes(mimeSize.Size)
	}

	var objectpath string
	if fsStringer, ok := s.object.GetFS().(fmt.Stringer); ok {
		objectpath = fsStringer.String()
	}

	cfg, err := extManager.GetConfigName(extension.MetaFileName)
	if err != nil {
		cfg = &extension.MetaFileConfig{
			ExtensionConfig: &ocfl.ExtensionConfig{ExtensionName: extension.MetaFileName},
			StorageType:     "area",
			StorageName:     "metadata",
			MetaName:        "info.json",
			MetaSchema:      "none",
		}
	}

	metafileCfg, ok := cfg.(*extension.MetaFileConfig)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errors.Errorf("invalid config format %v", cfg)})
		return
	}

	var infoBytes []byte
	if metafileCfg.StorageType == "extension" {
		fsys, err := extManager.GetFSName(extension.MetaFileName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		infoname := strings.TrimLeft(filepath.ToSlash(filepath.Join(metafileCfg.StorageName, metafileCfg.MetaName)), "/")
		infoBytes, err = fs.ReadFile(fsys, infoname)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": errors.Wrapf(err, "cannot open %v/%s", fsys, infoname).Error()})
			return
		}
	} else {
		area := "content"
		path := metafileCfg.StorageName
		if metafileCfg.StorageType == "area" {
			area = metafileCfg.StorageName
			path = ""
		}
		fname := filepath.ToSlash(filepath.Join(path, metafileCfg.MetaName))
		mPath, err := extManager.BuildObjectManifestPath(s.object, fname, area)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": errors.Wrapf(err, "cannot map %s:%s", area, fname).Error()})
			return
		}

		// search for info file
		for ver, _ := range s.metadata.Versions {
			fullpath := filepath.ToSlash(filepath.Join(ver, "content", mPath))
			jsonData, err := fs.ReadFile(s.object.GetFS(), fullpath)
			if err == nil && len(jsonData) > 0 {
				infoBytes = jsonData
			}
		}
	}
	var info = map[string]any{}
	if len(infoBytes) > 0 {
		if err := json.Unmarshal(infoBytes, &info); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": errors.Wrapf(err, "cannot unmarshal %s", metafileCfg.MetaName).Error()})
			return
		}
	}

	var filenames = []string{}

	type edge struct {
		indent   uint
		children []*edge
		parent   *edge
		name     string
	}

	var tree = &edge{
		indent:   0,
		children: []*edge{},
		parent:   nil,
		name:     "",
	}

	var maxDepth uint
	var addToTree func(parts []string, e *edge)
	addToTree = func(parts []string, e *edge) {
		if len(parts) == 0 {
			return
		}
		for _, child := range e.children {
			if child.name == parts[0] {
				addToTree(parts[1:], child)
				return
			}
		}
		newEdge := &edge{
			indent:   e.indent + 1,
			children: []*edge{},
			parent:   e,
			name:     parts[0],
		}
		if maxDepth <= e.indent {
			maxDepth = e.indent + 1
		}
		e.children = append(e.children, newEdge)
		addToTree(parts[1:], newEdge)
	}

	for _, file := range s.metadata.Files {
		if _thumb, ok := file.Extension[extension.ThumbnailName]; ok {
			if thumb, ok := _thumb.(extension.ThumbnailResult); ok {
				if len(thumb.SourceDigest) > 0 {
					continue
				}
			}
		}
		for _, files := range file.VersionName {
			for _, filename := range files {
				filenames = append(filenames, filename)
				parts := strings.Split(filename, "/")
				if !full {
					parts = parts[0 : len(parts)-1]
				}
				addToTree(parts, tree)
			}
		}
	}

	type flatEdge struct {
		Left  int
		Right int
		Name  string
	}
	var flatTree = []*flatEdge{}
	var flattenTree func(e *edge)
	flattenTree = func(e *edge) {
		if strings.TrimSpace(e.name) != "" {
			flatTree = append(flatTree, &flatEdge{
				Left:  int(e.indent),
				Right: int(maxDepth - e.indent),
				Name:  e.name,
			})
		}
		for _, child := range e.children {
			flattenTree(child)
		}
	}
	flattenTree(tree)

	var files = map[string]*ocfl.FileMetadata{}
	var filesNoData int64
	for key, file := range s.metadata.Files {
		if _thumb, ok := file.Extension[extension.ThumbnailName]; ok {
			if thumb, ok := _thumb.(extension.ThumbnailResult); ok {
				if len(thumb.SourceDigest) > 0 {
					continue
				}
			}
		}
		if file.Extension[extension.IndexerName] == nil && file.Extension[extension.FilesystemName] == nil {
			filesNoData++
		}
		if full {
			files[key] = file
		}
	}

	var params = map[string]any{
		"objectpath":     objectpath,
		"gocfl":          "gocfl",
		"head":           inventory.GetHead(),
		"id":             s.object.GetID(),
		"versions":       s.metadata.Versions,
		"differentFiles": len(s.metadata.Files),
		"numFiles":       numFiles,
		"filesNoData":    filesNoData,
		"size":           size,
		"noSizeFiles":    noSizeFiles,
		"mimeTypes":      mimeTypes,
		"pronoms":        pronoms,
		"files":          files,
		"info":           info,
		"avLength":       fmtDuration(time.Duration(int64(videoSecs) * int64(time.Second))),
		"tree":           flatTree,
		"full":           full,
	}

	c.HTML(http.StatusOK, "report.gohtml", gin.H(params))
}

func fmtDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	return fmt.Sprintf("%d:%02d:%02d", h, m, s)
}

func (s *Server) Shutdown(ctx context.Context) error {
	return errors.WithStack(s.srv.Shutdown(ctx))
}

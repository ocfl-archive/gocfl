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
	"github.com/je4/gocfl/v2/pkg/extension"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	"github.com/je4/indexer/v2/pkg/indexer"
	dcert "github.com/je4/utils/v2/pkg/cert"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/op/go-logging"
	"html/template"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
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
	log            *logging.Logger
	urlExt         *url.URL
	accessLog      io.Writer
	dataFS         fs.FS
	storageRoot    ocfl.StorageRoot
	object         ocfl.Object
	metadata       *ocfl.ObjectMetadata
	templateFS     fs.FS
}

func NewServer(storageRoot ocfl.StorageRoot, service, addr string, urlExt *url.URL, dataFS fs.FS, templateFS fs.FS, log *logging.Logger, accessLog io.Writer) (*Server, error) {
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
	}

	for _, tplfile := range tplfiles {
		funcMap := sprig.FuncMap()
		funcMap["basename"] = func(str string) string {
			return filepath.Base(str)
		}
		funcMap["PathEscape"] = func(str string) string {
			return url.PathEscape(str)
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
	route.GET("/object/id/:id/download/:checksum/:filename", s.download)
	route.GET("/object/folder/*path", s.loadObjectPath)

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

func (s *Server) Shutdown(ctx context.Context) error {
	return errors.WithStack(s.srv.Shutdown(ctx))
}

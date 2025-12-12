package extension

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"strings"

	"emperror.dev/errors"
	"github.com/andybalholm/brotli"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/object"
)

func ReadFile(object object.Object, name, version, storageType, storageName string, fsys fs.FS) ([]byte, error) {
	var targetname string
	switch storageType {
	case "area":
		path, err := object.GetAreaPath(storageName)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot get area path for '%s'", storageName)
		}
		targetname = object.GetInventory().BuildManifestNameVersion(fmt.Sprintf("%s/%s", path, name), version)
		//targetname = fmt.Sprintf("%s/content/%s/indexer_%s.jsonl%s", version, path, version, ext)
		fsys = object.GetFS()
	case "path":
		path, err := object.GetAreaPath("content")
		if err != nil {
			return nil, errors.Wrapf(err, "cannot get area path for '%s'", "content")
		}
		targetname = object.GetInventory().BuildManifestNameVersion(fmt.Sprintf("%s/%s/%s", path, storageName, name), version)
		//targetname = fmt.Sprintf("%s/content/%s/indexer_%s.jsonl%s", v, sl.IndexerConfig.StorageName, v, ext)
		fsys = object.GetFS()
	case "extension":
		targetname = strings.TrimLeft(fmt.Sprintf("%s/%s", storageName, name), "/")
	default:
		return nil, errors.Errorf("unsupported storage type '%s'", storageType)
	}

	return fs.ReadFile(fsys, targetname)
}

func ReadJsonL(object object.Object, name, version, compress, storageType, storageName string, fsys fs.FS) ([]byte, error) {
	if fsys == nil {
		return nil, errors.Errorf("[%s/%s] %s: fsys is nil", object.GetID(), version, name)
	}
	var ext string
	switch compress {
	case "brotli":
		ext = ".br"
	case "gzip":
		ext = ".gz"
	case "none":
	default:
		return nil, errors.Errorf("invalid compression '%s'", compress)
	}
	var targetname string
	switch storageType {
	case "area":
		path, err := object.GetAreaPath(storageName)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot get area path for '%s'", storageName)
		}
		targetname = object.GetInventory().BuildManifestNameVersion(fmt.Sprintf("%s/%s_%s.jsonl%s", path, name, version, ext), version)
		//targetname = fmt.Sprintf("%s/content/%s/indexer_%s.jsonl%s", version, path, version, ext)
		fsys = object.GetFS()
	case "path":
		path, err := object.GetAreaPath("content")
		if err != nil {
			return nil, errors.Wrapf(err, "cannot get area path for '%s'", "content")
		}
		targetname = object.GetInventory().BuildManifestNameVersion(fmt.Sprintf("%s/%s/%s_%s.jsonl%s", path, storageName, name, version, ext), version)
		//targetname = fmt.Sprintf("%s/content/%s/indexer_%s.jsonl%s", v, sl.IndexerConfig.StorageName, v, ext)
		fsys = object.GetFS()
	case "extension":
		targetname = strings.TrimLeft(fmt.Sprintf("%s/%s_%s.jsonl%s", storageName, name, version, ext), "/")
	default:
		return nil, errors.Errorf("unsupported storage type '%s'", storageType)
	}

	var reader io.Reader
	f, err := fsys.Open(targetname)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot open '%v/%s'", fsys, targetname)
	}
	switch compress {
	case "brotli":
		reader = brotli.NewReader(f)
	case "gzip":
		reader, err = gzip.NewReader(f)
		if err != nil {
			f.Close()
			return nil, errors.Wrapf(err, "cannot open gzip reader on '%s'", targetname)
		}
	case "none":
		reader = f
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		if f != nil {
			f.Close()
		}
		return nil, errors.Wrapf(err, "cannot read '%s'", targetname)
	}
	if f != nil {
		if err := f.Close(); err != nil {
			return nil, errors.Wrapf(err, "cannot close '%s'", targetname)
		}
	}
	return data, nil
}

func WriteJsonL(object object.Object, name string, brotliData []byte, compress, storageType, storageName string, fsys fs.FS) error {
	var bufReader = bytes.NewBuffer(brotliData)
	var ext string
	var reader io.Reader
	switch compress {
	case "brotli":
		ext = ".br"
		reader = bufReader
	case "gzip":
		ext = ".gz"
		brotliReader := brotli.NewReader(bufReader)
		pr, pw := io.Pipe()
		go func() {
			defer pw.Close()
			gzipWriter := gzip.NewWriter(pw)
			defer gzipWriter.Close()
			if _, err := io.Copy(gzipWriter, brotliReader); err != nil {
				pw.CloseWithError(errors.Wrapf(err, "error on gzip compressor"))
			}
		}()
		reader = pr
	case "none":
		reader = brotli.NewReader(bufReader)
	default:
		return errors.Errorf("invalid compression '%s'", compress)
	}

	head := object.GetInventory().GetHead()
	switch strings.ToLower(storageType) {
	case "area":
		targetname := fmt.Sprintf("%s_%s.jsonl%s", name, head, ext)
		if _, err := object.AddReader(io.NopCloser(reader), []string{targetname}, storageName, true, false); err != nil {
			return errors.Wrapf(err, "cannot write '%s'", targetname)
		}
	case "path":
		path, err := object.GetAreaPath("content")
		if err != nil {
			return errors.Wrapf(err, "cannot get area path for '%s'", "content")
		}
		targetname := fmt.Sprintf("%s/%s/%s_%s.jsonl%s", path, storageName, name, head, ext)

		//targetname := fmt.Sprintf("%s/%s_%s.jsonl%s", name, storageName, head, ext)
		if _, err := object.AddReader(io.NopCloser(reader), []string{targetname}, "", true, false); err != nil {
			return errors.Wrapf(err, "cannot write '%s'", targetname)
		}
	case "extension":
		targetname := strings.TrimLeft(fmt.Sprintf("%s/%s_%s.jsonl%s", storageName, name, head, ext), "/")
		fp, err := writefs.Create(fsys, targetname)
		if err != nil {
			return errors.Wrapf(err, "cannot create '%v/%s'", fsys, targetname)
		}
		defer fp.Close()
		if _, err := io.Copy(fp, reader); err != nil {
			return errors.Wrapf(err, "cannot write '%v/%s'", fsys, targetname)
		}
	default:
		return errors.Errorf("unsupported storage type '%s'", storageType)
	}

	return nil
}

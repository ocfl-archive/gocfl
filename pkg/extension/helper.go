package extension

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"path"
	"strings"

	"emperror.dev/errors"
	"github.com/andybalholm/brotli"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object"
)

func ReadFile(extensionName string, fsys fs.FS, obj object.Object, name string, version *inventory.VersionNumber, storageType, storageName string) ([]byte, error) {
	var targetname string
	switch storageType {
	case "area":
		path, err := obj.GetExtensionManager().GetAreaPath(storageName)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot get area path for '%s'", storageName)
		}
		targetname = obj.GetInventory().BuildManifestNameVersion(fmt.Sprintf("%s/%s", path, name), version)
		//targetname = fmt.Sprintf("%s/content/%s/indexer_%s.jsonl%s", version, path, version, ext)
	case "path":
		path, err := obj.GetExtensionManager().GetAreaPath("content")
		if err != nil {
			return nil, errors.Wrapf(err, "cannot get area path for '%s'", "content")
		}
		targetname = obj.GetInventory().BuildManifestNameVersion(fmt.Sprintf("%s/%s/%s", path, storageName, name), version)
		//targetname = fmt.Sprintf("%s/content/%s/indexer_%s.jsonl%s", v, sl.IndexerConfig.StorageName, v, ext)
	case "extension":
		targetname = strings.TrimLeft(fmt.Sprintf("%s/%s", path.Join("extensions", extensionName, storageName), name), "/")
	default:
		return nil, errors.Errorf("unsupported storage type '%s'", storageType)
	}

	return fs.ReadFile(fsys, targetname)
}

func ReadJsonL(extensionName string, fsys fs.FS, obj object.Object, version *inventory.VersionNumber, name, compress, storageType, storageName string) ([]byte, error) {
	if fsys == nil {
		return nil, errors.Errorf("[%s/%s] %s: fsys is nil", obj.GetID(), version, name)
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
		path, err := obj.GetExtensionManager().GetAreaPath(storageName)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot get area path for '%s'", storageName)
		}
		targetname = obj.GetInventory().BuildManifestNameVersion(fmt.Sprintf("%s/%s_%s.jsonl%s", path, name, version, ext), version)
		//targetname = fmt.Sprintf("%s/content/%s/indexer_%s.jsonl%s", version, path, version, ext)
	case "path":
		path, err := obj.GetExtensionManager().GetAreaPath("content")
		if err != nil {
			return nil, errors.Wrapf(err, "cannot get area path for '%s'", "content")
		}
		targetname = obj.GetInventory().BuildManifestNameVersion(fmt.Sprintf("%s/%s/%s_%s.jsonl%s", path, storageName, name, version, ext), version)
		//targetname = fmt.Sprintf("%s/content/%s/indexer_%s.jsonl%s", v, sl.IndexerConfig.StorageName, v, ext)
	case "extension":
		targetname = strings.TrimLeft(fmt.Sprintf("%s/%s_%s.jsonl%s", path.Join("extensions", extensionName, storageName), name, version, ext), "/")
	default:
		return nil, errors.Errorf("unsupported storage type '%s'", storageType)
	}

	var reader io.Reader
	f, err := fsys.Open(targetname)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot open '%v/%s'", fsys, targetname)
	}
	defer f.Close()
	switch compress {
	case "brotli":
		reader = brotli.NewReader(f)
	case "gzip":
		r, err := gzip.NewReader(f)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot open gzip reader on '%s'", targetname)
		}
		reader = r
		defer r.Close()
	case "none":
		reader = f
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot read '%s'", targetname)
	}
	return data, nil
}

func WriteJsonL(extensionName string, obj object.VersionWriter, name string, brotliData []byte, compress, storageType, storageName string) error {
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

	head := obj.GetInventory().GetHead()
	switch strings.ToLower(storageType) {
	case "area":
		targetname := fmt.Sprintf("%s_%s.jsonl%s", name, head, ext)
		if _, err := obj.AddReader(io.NopCloser(reader), []string{targetname}, storageName, true, false); err != nil {
			return errors.Wrapf(err, "cannot write '%s'", targetname)
		}
	case "path":
		path, err := obj.GetExtensionManager().GetAreaPath("content")
		if err != nil {
			return errors.Wrapf(err, "cannot get area path for '%s'", "content")
		}
		targetname := fmt.Sprintf("%s/%s/%s_%s.jsonl%s", path, storageName, name, head, ext)

		//targetname := fmt.Sprintf("%s/%s_%s.jsonl%s", name, storageName, head, ext)
		if _, err := obj.AddReader(io.NopCloser(reader), []string{targetname}, "", true, false); err != nil {
			return errors.Wrapf(err, "cannot write '%s'", targetname)
		}
	case "extension":
		targetname := strings.TrimLeft(fmt.Sprintf("extensions/%s/%s/%s_%s.jsonl%s", extensionName, storageName, name, head, ext), "/")
		fsys := obj.GetFS()
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

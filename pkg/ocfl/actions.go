// Package ocfl provides high-level functions for interacting with OCFL objects.
// It orchestrates lower-level modules to perform common tasks like validating,
// extracting, and retrieving metadata from OCFL objects.
package ocfl

import (
	"context"
	"fmt"
	"io/fs"

	"emperror.dev/errors"
	"github.com/ocfl-archive/filesystem/pkg/appendfs"
	"github.com/ocfl-archive/filesystem/pkg/writefs"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/validation"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

// ValidateObject validates an OCFL object located at objectFS.
// It loads the object and performs structural and fixity checks using the object's validator.
func ValidateObject(ctx context.Context, objectFS fs.FS, logger ocfllogger.OCFLLogger) error {
	fmt.Printf("object folder '%v'\n", objectFS)
	obj, err := LoadObject(ctx, objectFS, nil, logger)
	if err != nil {
		logger.ValidationError(validation.E001, "invalid fsys '%v': %v", objectFS, err)
		return errors.Wrapf(err, "cannot load object from folder '%v'", objectFS)
	}
	defer func() { _ = obj.Close() }()
	validator := obj.GetValidator()
	defer func() { _ = validator.Close() }()
	if err := validator.Validate(); err != nil {
		return errors.Wrapf(err, "cannot validate object from folder '%v'", objectFS)
	}
	return nil
}

// Extract retrieves the content of an OCFL object at the specified path and version,
// and writes it to the destination filesystem (destFS).
// If version is invalid or empty, the latest version is used.
// area specifies the storage area (e.g. "content" or "metadata") if applicable.
func Extract(ctx context.Context, objectFS fs.FS, destFS appendfs.FS, path string, version *inventory.VersionNumber, withManifest bool, area string, logger ocfllogger.OCFLLogger) error {
	if !version.IsValid() {
		version = inventory.NewVersionNumber().WithLatest()
	}

	logger.Debug().Msgf("Extracting object '%s' with version '%s'", path, version)
	var o object.Object
	var err error
	objFsys, err := writefs.Sub(objectFS, path)
	if err != nil {
		return errors.Wrapf(err, "cannot create subfs  '%v' / %s", objectFS, path)
	}
	o, err = LoadObject(ctx, objFsys, nil, logger)
	if err != nil {
		return errors.Wrapf(err, "cannot load object '%s'", path)
	}
	defer func() {
		_ = o.Close()
	}()
	extractor := o.GetExtractor().WithDestFS(destFS)
	defer func() { _ = extractor.Close() }()
	if err := extractor.Extract(version, withManifest, area); err != nil {
		return errors.Wrapf(err, "cannot extract object '%s'", path)
	}

	logger.Debug().Msgf("extraction done")
	return nil
}

// ExtractMeta retrieves metadata information from an OCFL object at the given path.
// It returns a pointer to inventory.Metadata containing object details.
func ExtractMeta(ctx context.Context, fsys fs.FS, path string, logger ocfllogger.OCFLLogger) (*inventory.Metadata, error) {
	logger.Debug().Msgf("Extracting object '%s'", path)
	objFsys, err := writefs.Sub(fsys, path)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot sub fsys '%s'", path)
	}
	obj, err := LoadObject(ctx, objFsys, nil, logger)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot load object '%s'", path)
	}
	defer func() { _ = obj.Close() }()
	defer logger.Debug().Msgf("extraction done")
	extractor := obj.GetExtractor().WithDestFS(nil)
	defer func() { _ = extractor.Close() }()
	return extractor.GetMetadata()
}

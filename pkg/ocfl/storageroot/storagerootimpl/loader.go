package storagerootimpl

import (
	"context"
	"io/fs"

	"emperror.dev/errors"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/storageroot"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/validation"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

func NewLoader(ctx context.Context, logger ocfllogger.OCFLLogger) *Loader {
	return &Loader{
		ctx:    ctx,
		logger: logger.With("task", "storage root loader"),
	}
}

type Loader struct {
	storageroot.StorageRoot
	ctx              context.Context
	extensionFactory extension.Factory
	sourceFS         fs.FS
	logger           ocfllogger.OCFLLogger
}

func (loader *Loader) Close() error {
	loader.sourceFS = nil
	return nil
}

func (loader *Loader) Load() error {
	if err := loader.loadExtensionManager(); err != nil {
		return errors.Wrap(err, "loading extension manager")
	}
	return nil
}

func (loader *Loader) WithExtensionFactory(factory extension.Factory) storageroot.Loader {
	loader.extensionFactory = factory
	return loader
}

func (loader *Loader) WithStorageRoot(sr storageroot.StorageRoot) storageroot.Loader {
	loader.StorageRoot = sr
	return loader
}

func (loader *Loader) WithFS(sourceFS fs.FS) storageroot.Loader {
	loader.sourceFS = sourceFS
	return loader
}

func (loader *Loader) loadExtensionManager() error {
	extensionFS, err := fs.Sub(loader.sourceFS, "extensions")
	if err != nil {
		return errors.Wrapf(err, "cannot create subfs of %v for folder '%s'", loader.sourceFS, "extensions")
	}
	manager, err := loader.extensionFactory.LoadExtensionManager(extensionFS)
	if err != nil {
		loader.logger.ValidationError(validation.W000, "cannot initialize all extensions in folder '%s': %v", extensionFS, err)
		if manager == nil {
			return errors.Wrap(err, "cannot create extension manager")
		}
	}
	extensionManager, ok := manager.(storageroot.ExtensionManager)
	if !ok {
		return errors.Errorf("extension manager does not implement storageroot.ExtensionManager")
	}
	loader.StorageRoot.WithExtensionManager(extensionManager)
	return nil
}

var _ storageroot.Loader = (*Loader)(nil)

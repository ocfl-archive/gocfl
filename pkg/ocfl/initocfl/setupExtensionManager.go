package initocfl

import (
	"io/fs"

	"emperror.dev/errors"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension/extensionimpl"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

func NewExtensionFactory[T extension.ManagerCore[T]](
	params map[string]string,
	logger ocfllogger.OCFLLogger,
) (extension.Factory[T], error) {
	factory, err := extensionimpl.NewFactory[T](params, logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create extension factory")
	}
	return factory, nil
}

// SetupExtensionManager initializes an extension factory and loads the extension manager from the provided filesystem.
// The function uses the type parameter T (usually storageroot.ExtensionManager or object.ExtensionManager)
// to instantiate the appropriate manager type.
// It takes configuration parameters, a filesystem containing extension configuration files, and a logger.
// It is intended to be called once for the storage root and once for objects to manage their respective extensions.
// Returns the loaded manager, the used factory, and any error encountered.
func SetupExtensionManager[T extension.ManagerCore[T]](
	params map[string]string,
	fsys fs.FS,
	logger ocfllogger.OCFLLogger,
) (T, extension.Factory[T], error) {
	factory, err := NewExtensionFactory[T](params, logger)
	if err != nil {
		var result T
		return result, nil, errors.Wrap(err, "cannot create extension factory")
	}
	m, err := factory.LoadExtensionManager(fsys)
	if err != nil {
		var result T
		return result, nil, errors.Wrapf(err, "loading extension manager [%T]", result)
	}
	return m, factory, nil
}

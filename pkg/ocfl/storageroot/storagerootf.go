package storageroot

/*


func ValidVersion(ver version.OCFLVersion) bool {
	switch ver {
	case version.Version1_0:
		return true
	case version.Version1_1:
		return true
	case version.Version2_0:
		return true
	default:
		return false
	}
}
func CreateStorageRoot(
	ctx context.Context,
	fsys fs.FS,
	fact factory.Factory,
	digest checksum.DigestAlgorithm,
	logger zLogger.ZLogger,
) (StorageRoot, error) {
	storageRoot := fact.NewStorageRoot(ctx).WithFS(fsys)

	if err := storageRoot.Init(digest); err != nil {
		return nil, errors.Wrap(err, "cannot initialize storage root")
	}

	return storageRoot, nil
}

func LoadStorageRoot(ctx context.Context, fsys fs.FS, extensionFactory *extensionimpl.ExtensionFactory, logger zLogger.ZLogger) (StorageRoot, error) {
	ver, err := util.GetVersion(ctx, fsys, ".", "ocfl_")
	if err != nil && !errors.Is(err, ocflerrors.ErrVersionNone) {
		return nil, errors.WithStack(err)
	}
	if ver == "" {
		dirs, err := fs.ReadDir(fsys, ".")
		if err != nil {
			return nil, errors.WithStack(err)
		}
		if len(dirs) > 0 {
			err := validation.GetValidationError(version.Version1_1, validation.E069).AppendDescription("storage root not empty without version information").AppendContext("storage root '%s'", fsys)
			validation.AddValidationErrors(ctx, err)
			//			return nil, err
		}
		ver = version.Version1_1
	}
	extFSys, err := writefs.Sub(fsys, "extensions")
	if err != nil {
		return nil, errors.Wrap(err, "cannot create sub filesystem 'extensions'")
	}
	extensionManager, err := extensionFactory.CreateExtensions(extFSys, nil)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create extension manager")
	}
	storageRoot, err := newStorageRoot(ctx, fsys, ver, extensionFactory, extensionManager.(ExtensionManager), logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot instantiate storage root")
	}

	if err := storageRoot.Load(); err != nil {
		return nil, errors.Wrap(err, "cannot load storage root")
	}
	return storageRoot, nil
}

func LoadStorageRootRO(ctx context.Context, fsys fs.FS, extensionFactory *extensionimpl.ExtensionFactory, logger zLogger.ZLogger) (StorageRoot, error) {
	ver, err := util.GetVersion(ctx, fsys, ".", "ocfl_")
	if err != nil && !errors.Is(err, ocflerrors.ErrVersionNone) {
		return nil, errors.WithStack(err)
	}
	if ver == "" {
		dirs, err := fs.ReadDir(fsys, ".")
		if err != nil {
			return nil, errors.WithStack(err)
		}
		if len(dirs) > 0 {
			err := validation.GetValidationError(version.Version1_1, validation.E069).AppendDescription("storage root not empty without version information").AppendContext("storage root '%s'", fsys)
			validation.
				AddValidationErrors(ctx, err)
			//			return nil, err
		}
		ver = version.Version1_1
	}
	extFSys, err := writefs.Sub(fsys, "extensions")
	if err != nil {
		return nil, errors.Wrap(err, "cannot create sub filesystem 'extensions'")
	}
	extensionManager, err := extensionFactory.CreateExtensions(extFSys, nil)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create extension manager")
	}
	storageRoot, err := newStorageRoot(ctx, fsys, ver, extensionFactory, extensionManager.(ExtensionManager), logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot instantiate storage root")
	}

	if err := storageRoot.Load(); err != nil {
		return nil, errors.Wrap(err, "cannot load storage root")
	}
	return storageRoot, nil
}

*/

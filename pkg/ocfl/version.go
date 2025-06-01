package ocfl

import (
	"context"
	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/zLogger"
	archiveerror "github.com/ocfl-archive/error/pkg/error"
	"io/fs"
)

func NewVersion(
	objectID string,
	version string,
	ocflVersion OCFLVersion,
	ctx context.Context,
	fsys fs.FS,
	inventory Inventory,
	packages VersionPackages,
	manager ExtensionManager,
	digestAlgorithm checksum.DigestAlgorithm,
	logger zLogger.ZLogger,
	factory *archiveerror.Factory,
) (Version, error) {
	switch ocflVersion {
	case Version1_0:
		return newVersionV1_0(objectID, version, ctx, fsys, inventory, manager, digestAlgorithm, logger, factory)
	case Version1_1:
		return newVersionV1_1(objectID, version, ctx, fsys, inventory, manager, digestAlgorithm, logger, factory)
	case Version2_0:
		return newVersionV2_0(objectID, version, ctx, fsys, inventory, packages, manager, digestAlgorithm, logger, factory)
	default:
		return nil, errors.Errorf("Unsupported version: %s", version)
	}
}

type Version interface {
	Check() error
}

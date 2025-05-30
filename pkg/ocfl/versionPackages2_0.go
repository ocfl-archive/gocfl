package ocfl

import (
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/zLogger"
)

func newVersionPackageV2_0(digestAlgorithm checksum.DigestAlgorithm, logger zLogger.ZLogger) (VersionPackages, error) {
	vpb := newVersionPackagesBase(digestAlgorithm, logger)
	return &VersionPackageV2_0{
		VersionPackagesBase: vpb,
	}, nil
}

type VersionPackageV2_0 struct {
	*VersionPackagesBase
}

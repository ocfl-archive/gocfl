package initocfl

import (
	"context"

	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/validation"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger/ocflloggerimpl"
)

func NewOCFLLogger(ctx context.Context, logger zLogger.ZLogger, data map[string]string, ver version.OCFLVersion, validationStatus *validation.Status) ocfllogger.OCFLLogger {
	return ocflloggerimpl.NewOCFLLogger(ctx, logger, data, ver, validationStatus)
}

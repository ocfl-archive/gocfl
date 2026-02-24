package object

import (
	"context"
)

type Factory interface {
	NewObject(ctx context.Context) Object
	NewLoader(ctx context.Context) Loader
	NewInitializer(ctx context.Context) Initializer
	NewChecker(ctx context.Context) Checker
	NewExtractor(ctx context.Context) Extractor
}

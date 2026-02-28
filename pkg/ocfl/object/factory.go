package object

import (
	"context"
)

// Factory is the interface for creating object components.
//
// This interface is also implemented by the Unified Factory in [pkg/ocfl/factory/factory.go].
// It's recommended to use the Unified Factory instead of calling this interface directly.
type Factory interface {
	NewObject(ctx context.Context) Object
	NewLoader(ctx context.Context) Loader
	NewInitializer(ctx context.Context) Initializer
	NewChecker(ctx context.Context) Checker
	NewExtractor(ctx context.Context) Extractor
}

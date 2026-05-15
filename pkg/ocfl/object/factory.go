package object

import (
	"context"
)

// Factory is the interface for creating object components.
//
// This interface is also implemented by the Unified Factory in [pkg/ocfl/factory/factory.go].
// It's recommended to use the Unified Factory instead of calling this interface directly.
type Factory interface {
	// NewObject creates a new Object instance.
	NewObject(ctx context.Context) Object
	// NewLoader creates a new Loader instance.
	NewLoader(ctx context.Context) Loader
	// NewInitializer creates a new Initializer instance.
	NewInitializer(ctx context.Context) Initializer
	// NewChecker creates a new Checker instance.
	NewChecker(ctx context.Context) Checker
	// NewExtractor creates a new Extractor instance.
	NewExtractor(ctx context.Context) Extractor
	// NewVersionWriter creates a new VersionWriter instance.
	NewVersionWriter(ctx context.Context) VersionWriter
}

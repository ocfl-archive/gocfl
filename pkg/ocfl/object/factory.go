package object

import (
	"context"
)

// ConfigName represents the name of a configuration parameter for the object factory.
type ConfigName string

const (
	// ObjectName is the configuration name for the object instance.
	ObjectName ConfigName = "object"
	// CheckerName is the configuration name for the checker instance.
	CheckerName ConfigName = "checker"
	// LoaderName is the configuration name for the loader instance.
	LoaderName ConfigName = "loader"
	// InitializerName is the configuration name for the initializer instance.
	InitializerName ConfigName = "initializer"
	// ExtractorName is the configuration name for the extractor instance.
	ExtractorName ConfigName = "extractor"
	// VersionWriterName is the configuration name for the version writer instance.
	VersionWriterName ConfigName = "versionwriter"
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

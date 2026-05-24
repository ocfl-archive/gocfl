// Package factory provides unified factory interfaces for creating OCFL objects and storage roots.
// It combines sub-factories from various OCFL components into a single, cohesive interface.
package factory

import (
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/storageroot"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
)

// FactoryObject is the unified factory interface for OCFL objects.
// It combines object and inventory factories and provides methods for configuration and cloning.
type FactoryObject interface {
	// Factory provides methods for creating object-related components.
	object.Factory
	// Factory provides methods for creating inventory-related components.
	inventory.Factory
	// GetVersion returns the OCFL version used by the factory.
	GetVersion() version.OCFLVersion
	// WithConfig sets the configuration for the factory.
	WithConfig(config map[object.ConfigName]any) FactoryObject
	// GetConfig returns the configuration for the factory.
	GetConfig() map[object.ConfigName]any
	// Copy returns a deep copy of the factory.
	Copy() FactoryObject
}

// FactoryStorageRoot is the unified factory interface for OCFL storage roots.
// It provides methods for configuration and cloning of storage root factories.
type FactoryStorageRoot interface {
	// Factory provides methods for creating storage root-related components.
	storageroot.Factory
	// GetVersion returns the OCFL version used by the factory.
	GetVersion() version.OCFLVersion
	// WithConfig sets the configuration for the factory.
	WithConfig(config map[storageroot.ConfigName]any) FactoryStorageRoot
	// Copy returns a deep copy of the factory.
	Copy() FactoryStorageRoot
}

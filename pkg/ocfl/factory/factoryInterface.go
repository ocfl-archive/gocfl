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
// It combines object and inventory factories and provides methods for version management and cloning.
type FactoryObject interface {
	// Factory provides methods for creating object-related components.
	object.Factory
	// Factory provides methods for creating inventory-related components.
	inventory.Factory
	// GetVersion returns the OCFL version used by the factory.
	GetVersion() version.OCFLVersion
	// WithNewVersion returns a new factory instance with the specified OCFL version.
	WithNewVersion(version.OCFLVersion) FactoryObject
	// Copy returns a deep copy of the factory.
	Copy() FactoryObject
}

// FactoryStorageRoot is the unified factory interface for OCFL storage roots.
// It provides methods for version management and cloning of storage root factories.
type FactoryStorageRoot interface {
	// Factory provides methods for creating storage root-related components.
	storageroot.Factory
	// GetVersion returns the OCFL version used by the factory.
	GetVersion() version.OCFLVersion
	// WithNewVersion returns a new factory instance with the specified OCFL version.
	WithNewVersion(version.OCFLVersion) FactoryStorageRoot
	// Copy returns a deep copy of the factory.
	Copy() FactoryStorageRoot
}

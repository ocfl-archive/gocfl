// Package inventory defines the structures and interfaces for the OCFL (Oxford Common File Layout) Inventory.
//
// The inventory is the central management unit of an OCFL object and contains all metadata
// about the object's versions, files, and fixity information.
//
// The inventory structure is divided into several modules:
//   - Inventory: The central management unit.
//   - Versions: Management of the version history and the state of each version.
//   - Manifest: Mapping of content digests to physical file paths.
//   - Fixity: Optional additional fixity information for physical files.
//
// For a visual overview of the architecture, see docs/architecture.md.
//
// It is recommended to use the factory in pkg/ocfl/factory to create inventory components,
// as it ensures correct creation according to the desired OCFL version.
package inventory

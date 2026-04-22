# DILCIS Go Packages

This directory contains Go data structures and helper functions for working with DILCIS (Digital Information LifeCycle Interoperability Standard) XML schemas, specifically METS, PREMIS, and EAD3.

Further information and tools can be found at: [https://earkaip.dilcis.eu/](https://earkaip.dilcis.eu/)

## Overview

The packages provided here are used to marshal and unmarshal XML documents that follow the DILCIS specifications. They are essential for generating standardized metadata in the OCFL (Oxford Common File Layout) environment.

For detailed documentation on the METS extension implementation, see:
- **[NNNN-mets.md](../../../docs/NNNN-mets.md)**: Documentation of the METS extension using these packages.

### Subpackages

- **`ead3`**: Go structures for the Encoded Archival Description (EAD3) standard.
- **`mets`**: Go structures for the Metadata Encoding and Transmission Standard (METS), including DILCIS extensions like CSIP (Common Specification for Information Packages).
- **`premis`**: Go structures for the PREMIS (Preservation Metadata: Implementation Strategies) Data Dictionary for Preservation Metadata.

## Schema Definitions

The root of this directory contains the original XML Schema Definition (`.xsd`) files which serve as the source of truth for the structure of the Go types:

- `mets.xsd`, `DILCISExtensionMETS.xsd`, `DILCISExtensionSIPMETS.xsd`: Definitions for METS and its DILCIS extensions.
- `premis-v3-0.xsd`: Definitions for PREMIS version 3.0.
- `ead3.xsd`, `ead3_undeprecated.xsd`: Definitions for EAD3.
- `xlink.xsd`: Standard XLink definitions used by the other schemas.

The Go code in the subdirectories (e.g., `pkg/dilcis/mets/mets.go`) was largely generated from these schemas using tools like `xgen`.

## Usage in GOCFL

These packages are primarily used by the OCFL extensions to generate standards-compliant metadata. 

Documentation of the implementation can be found in:
- **[docs/NNNN-mets.md](../../../docs/NNNN-mets.md)**: Detailed information about the METS extension which utilizes the `pkg/dilcis/mets` package to build and update METS documents during the OCFL object lifecycle. It maps internal OCFL state and file information to the METS structure defined in this package.

## Maintenance

When schemas are updated:
1. Replace the `.xsd` files in this directory.
2. Re-generate the Go code in the respective subpackages.
3. Update manual helper functions (like those in `premisFunc.go`) if the underlying structures have changed.

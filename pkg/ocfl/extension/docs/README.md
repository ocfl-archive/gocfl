# OCFL Extension Documentation

This directory contains detailed documentation for the components of the `extension` package.

## Components

0. [**System Architecture**](ARCHITECTURE.md): Overview of how components work together.
1. [**Extension Interface**](EXTENSION.md): Functional units for specific OCFL extensions.
2. [**Custom Extensions Guide**](../../../docs/custom_extensions.md): Step-by-step guide to creating your own extensions with code examples.
3. [**Manager Interface**](MANAGER.md): Coordination of multiple extensions.
4. [**Factory Interface**](FACTORY.md): Registry-based creation of extensions and managers.

## Core Concepts: Manager vs. Factory

While both components deal with extensions, they have distinct roles in the system:

- **Extension Factory**: The "Knowledge Base" and "Manufacturer". It knows which extension names belong to which Go implementations and handles the **instantiation** (creation) of extensions from configuration data.
- **Extension Manager**: The "Orchestrator" at runtime. It holds the active extension instances and **coordinates** their execution during OCFL operations (e.g. path mapping, fixity checks).

In short: The **Factory** *creates* the extensions, and the **Manager** *uses* them. For a deeper dive, see the [Architecture Documentation](ARCHITECTURE.md).

## Related Documentation

- [**Initial Extension Spec**](../../../pkg/extensions/ext_initial/initial.md): Identification of the primary manager.
- [**OCFL Specification**](../../version/ocfl_spec_1.1.md#5-extensions): OCFL 1.1 Section 5 (Extensions).
- [**External OCFL Extensions**](https://ocfl.io/extensions/): Official OCFL extension registry.

---
- [Back to Extension Overview](../README.md)

# OCFL Community Extension NNNN: Gocfl Extension Manager

* **Extension Name:** NNNN-gocfl-extension-manager
* **Authors:** Jürgen Enge (Basel)
* **Minimum OCFL Version:** 1.0
* **OCFL Community Extensions Version:** 1.0
* **Obsoletes:** n/a
* **Obsoleted by:** n/a

## Overview

This extension is used as an [initial extension](https://github.com/OCFL/extensions#optional-initial-extension)
to manage the OCFL extensions used in an OCFL repository.

There are two functionalities provided by this extension:

### Sorting of extensions

The extensions are sorted in the order they are defined in the `config.json` file. This enforces the correct execution order of the extensions. This is especially important for the extensions manipulating 
the object content path. 

### Exclusion of extension

Several extensions must not be executed together for the same object. For example extensions which map 
Object Ids to file paths must not be executed together. This extension provides a mechanism to exclude 
extensions from execution if conflicting extensions are enabled.

## Extension Types

Since extensions can be used in different contexts, there are different types of extensions which are
separated by the hooks, they are using. The following types are defined till now:

### Storage Root Hooks
#### `StorageRootPath`
For change of the Storage Root Object Path.
This hook is used by "Storage Root Layout Extensions" i.e. extension [0002-flat-direct-storage-layout](https://github.com/OCFL/extensions/blob/00bd9dcec83d9b27a2e4faae854a8c1e66997e0c/docs/0002-flat-direct-storage-layout.md), [0003-hash-and-id-n-tuple-storage-layout.md](https://github.com/OCFL/extensions/blob/00bd9dcec83d9b27a2e4faae854a8c1e66997e0c/docs/0003-hash-and-id-n-tuple-storage-layout.md), [0004-hashed-n-tuple-storage-layout.md](https://github.com/OCFL/extensions/blob/00bd9dcec83d9b27a2e4faae854a8c1e66997e0c/docs/0004-hashed-n-tuple-storage-layout.md), [0006-flat-omit-prefix-storage-layout.md](https://github.com/OCFL/extensions/blob/00bd9dcec83d9b27a2e4faae854a8c1e66997e0c/docs/0006-flat-omit-prefix-storage-layout.md), [0006-flat-omit-prefix-storage-layout.md](https://github.com/OCFL/extensions/blob/00bd9dcec83d9b27a2e4faae854a8c1e66997e0c/docs/0006-flat-omit-prefix-storage-layout.md) or [0010-differential-n-tuple-omit-prefix-storage-layout.md](https://github.com/OCFL/extensions/blob/00bd9dcec83d9b27a2e4faae854a8c1e66997e0c/docs/0010-differential-n-tuple-omit-prefix-storage-layout.md).

There are two hooks
* `WriteLayout` for writing the correct `ocfl_layout.json`
* `BuildStorageRootPath` Executed after the storage root path is known to the OCFL tool. 

### Object Hooks
#### `ObjectContentPath`
For change of the filepath within the Object versions content folder.
Used by extensions, which manipulate the path like [0011-direct-clean-path-layout](https://github.com/OCFL/extensions/blob/00bd9dcec83d9b27a2e4faae854a8c1e66997e0c/docs/0011-direct-clean-path-layout.md) or move content do subdirectories like [NNNN-content-subpath](https://github.com/ocfl-archive/gocfl/blob/3fa65107121024aaa3cfc17bbfa02ba2d89e679f/docs/NNNN-content-subpath.md).

There is one hook
* `BuildObjectManifestPath` Executed after the path within the version content folder is known to the OCFL tool. 

#### `ObjectExtractPath`
For change of the content file extraction path. Used by extensions like [NNNN-content-subpath](https://github.com/ocfl-archive/gocfl/blob/3fa65107121024aaa3cfc17bbfa02ba2d89e679f/docs/NNNN-content-subpath.md) which have inserted additional folders and want to remove them before extractions.

There is one hook
* `BuildObjectExtractPath` This extension hook is executed after the relative path of the external file is known to the OCFL tool. 

#### `ObjectStatePath`
For change of the version path within the inventory. This hook can be used by migration extensions, which have moved files from the manifest area to different folders and want to reflect a correct external path for extraction utilities.

There is one hook
* `BuildObjectStatePath` This extension hook is executed before the state filepath of the current version is written into the inventory. 

#### `ContentChange`

Whenever a file within the version content is written, a hook is called. Since there are no "real" update or delete operations within OCFL they are emulated for the hookd.

There are six hooks available.
* `AddFileBefore` Before a new file is added to the content
* `UpdateFileBefore` Before a file, which is already in the prior version, is added to the content
* `DeleteFileBefore` Before a file is not written to the state, which is already in the prior version
* `AddFileAfter` After a new file was added to the content
* `UpdateFileAfter` After a file, which was already in a prior version, was added to the content
* `DeleteFileAfter` After a file was not written to the state, which is already in the prior version

#### `ObjectChange`

#### `FixityDigest`

#### `Metadata`

#### `Area`


## Parameters

### Summary

* **Name:** `sort`
    * **Description:** Map of extension types to a list of extension names
    * **Type:** map[string][]string
    * **Constraints:** 
        * The keys must be one of the extension types
        * The values must be a list of extension names
    * **Default:** empty map
* **Name:** `exclude`
    * **Description:** Map of extension types to a list of extension names
    * **Type:** map[string][]string
    * **Constraints:**
        * The keys must be one of the extension types
        * The values must be a list of extension names
    * **Default:** empty map

## Procedure

### Sort

Create a list of extensions which fit to the same extension type. The order of the extensions is defined
by the list of the extension names in the `sort` parameter corresponding to the type. If the extension is
not defined, it is added to the end of the list.

### Exclude

Create a list of extensions which fit to the same extension type. If more than one extension is listed in 
the `exclude` parameter corresponding to the type, only the extension with the highest priority is will be
used. Extensions which are not listed are kept in the list.

## Example

```json
{
  "extensionName": "NNNN-gocfl-extension-manager",
  "sort": {
    "ObjectContentPath": [
      "NNNN-direct-clean-path-layout",
      "NNNN-content-subpath"
    ]
  },
  "exclusion": {
    "ObjectContentPath": [
      [
        "NNNN-direct-clean-path-layout",
        "NNNN-direct-path-layout"
     ]
    ]
  }
}
```

This example makes sure, that the `NNNN-direct-clean-path-layout` extension is executed before the
`NNNN-content-subpath` extension.
The `NNNN-direct-path-layout` extension is excluded from execution, if the `NNNN-direct-clean-path-layout`
is used.

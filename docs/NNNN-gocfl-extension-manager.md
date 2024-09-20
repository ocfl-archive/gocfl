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

## Hook Groups

Since extensions can be used in different contexts, there are different types of extensions which are
separated by the hooks, they are using. The following types are defined till now:

### Storage Root Hook Groups
#### `StorageRootPath`
For change of the Storage Root Object Path.
This hook is used by "Storage Root Layout Extensions" i.e. extension [0002-flat-direct-storage-layout](https://github.com/OCFL/extensions/blob/00bd9dcec83d9b27a2e4faae854a8c1e66997e0c/docs/0002-flat-direct-storage-layout.md), [0003-hash-and-id-n-tuple-storage-layout.md](https://github.com/OCFL/extensions/blob/00bd9dcec83d9b27a2e4faae854a8c1e66997e0c/docs/0003-hash-and-id-n-tuple-storage-layout.md), [0004-hashed-n-tuple-storage-layout.md](https://github.com/OCFL/extensions/blob/00bd9dcec83d9b27a2e4faae854a8c1e66997e0c/docs/0004-hashed-n-tuple-storage-layout.md), [0006-flat-omit-prefix-storage-layout.md](https://github.com/OCFL/extensions/blob/00bd9dcec83d9b27a2e4faae854a8c1e66997e0c/docs/0006-flat-omit-prefix-storage-layout.md), [0006-flat-omit-prefix-storage-layout.md](https://github.com/OCFL/extensions/blob/00bd9dcec83d9b27a2e4faae854a8c1e66997e0c/docs/0006-flat-omit-prefix-storage-layout.md) or [0010-differential-n-tuple-omit-prefix-storage-layout.md](https://github.com/OCFL/extensions/blob/00bd9dcec83d9b27a2e4faae854a8c1e66997e0c/docs/0010-differential-n-tuple-omit-prefix-storage-layout.md).

There are two hooks
* `WriteLayout` for writing the correct `ocfl_layout.json`
* `BuildStorageRootPath` Executed after the storage root path is known to the OCFL tool. 

### Object Hook Groups
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

This group of hooks is needed by extensions, which deal with the change of an object. 

There are two hooks.
* `UpdateObjectBefore` Before a new version of an OCFL object is generate
* `UpdateObjectAfter` After all content to the new version is written

#### `Stream`

To provide performant operation on binary content, there is the need to hijack a copy of the content data stream. This allows the creation of extensions, which can create checksums on the fly or determine technical metadata directly from the data stream.
An example is [NNNN-indexer](https://github.com/ocfl-archive/gocfl/blob/38c300010fccc8e5562719c04c902243b664d27c/docs/NNNN-indexer.md).

There is one hook availabe.
* `StreamObject` is called with a data stream as parameter. It should be executed on a parallel process to make sure, that speed of creating the OCFL object is not affected to heavily.

#### `Version`

After a version is done, there can be extension, which want to create another version. One example is a migration extension, which does preservation management like [NNNN-migration](https://github.com/ocfl-archive/gocfl/blob/38c300010fccc8e5562719c04c902243b664d27c/docs/NNNN-migration.md). These new versions are created before the inventory file is finally written to the object root.

There are two hooks available.
* `NeedNewVersion` determines whether a new version has to be created
* `DoNewVersion` gives control to the extension to create and fill up a new object version.

#### `FixityDigest`

Since fixity algorithms are done by extensions, there's need to get a list of all available fixity checksum algorithms.

There's one hook available.
* `GetFixityDigests` used by ocfl tools to determine which fixity algorithms have to be used for checksum calculations. There can be multiple algorithms in addition to the main one.

#### `Metadata

For looking inside of an OCFL object or for reporting purposes, there's often need for more metadata than the content of `inventory.json`. Any extension, which knows more about the content must be able to provide additional metadata. This could be metadata for specific content files or the whole object. Metadata for content-files should be mapped with the checksum as key.

There is one hook available.
* `GetMetadata` is called by any tool, which wants to report about content.

#### `Area`

Extensions, which have defined areas within the version content folder must provide a way of getting the internal path of the area. 

There is one hook available
* `GetAreaPath` this hook provides the internal (relative) path to a specific content area.

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

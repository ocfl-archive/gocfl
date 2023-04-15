# OCFL Community Extension NNNN: Migration

* **Extension Name:** NNNN-migration
* **Authors:** Jürgen Enge (Basel)
* **Minimum OCFL Version:** 1.0
* **OCFL Community Extensions Version:** 1.0
* **Obsoletes:** n/a
* **Obsoleted by:** n/a

## Overview

Preservation management requires a migration strategy. This extension provides a way to 
migrate old file formats to new ones. It needs the [NNNN-indexer](NNNN-indexer.md) extension to 
get Pronom IDs for the files. The migration is done by an external migration service. 
The migrated files are stored in a new version of the OCFL object. 

### Usage Scenario

If you have for example old PDF files, you can migrate them to PDF/A. This is a standard for
archival PDF files. It is a good idea to migrate the files to PDF/A, because it is a more
future-proof format.

## Parameters

### Summary

* **Name:** `storageType`
    * **Description:** Location Type where the technical metadata is stored. Possible values are
      `area`, `path` or `extension`.
        * **area:** within an `area` defined by [NNNN-content-subpath](NNNN-content-subpath.md)
          extension
        * **path:** directly within content folder
        * **extension:** within the extension subfolder
    * **Type:** string
    * **Default:**

* **Name:** `storageName`
    * **Description:** Location within the specified Type
        * **area:** area name
        * **path:** subfolder within content folder
        * **extension:** subfolder within extension folder
    * **Type:** string
    * **Default:**

* **Name:** `compress`
    * **Description:** Compression type for JSONL file
        * **none:** no compression
        * **gzip:** [gzip compression](https://en.wikipedia.org/wiki/Gzip)
        * **brotli:** [brotli compression](https://en.wikipedia.org/wiki/Brotli)
    * **Type:** string
    * **Default:**


## Caveat

The migration rules itself are not part of this extension. They are defined by the migration service.

## Procedure (tbd.)

Every migrated entry whithin the [OCFL Object Manifest](https://ocfl.io/1.1/spec/#manifest)
is represented by a JSON line in a file called  `migration_<version>.jsonl[.gz|.br]`.
Since this file is immutable, every version of the ocfl object gets its own indexer file.

## Examples

JSON Entry for a migrated pdf
```json
{
	"path": "v2/content/data/=u007Eblä=u0020blubb=u005Bin=u005D/Modulhandbuch_MA_Gestaltung.pdf",
	"migration": {
		"source": "v1/content/data/=u007Eblä=u0020blubb=u005Bin=u005D/Modulhandbuch_MA_Gestaltung.pdf",
		"id": "PDFA#01"
	}
}
```

### Result


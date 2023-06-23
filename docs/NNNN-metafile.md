# OCFL Community Extension NNNN: Metafile

* **Extension Name:** NNNN-metafile
* **Authors:** Jürgen Enge (Basel)
* **Minimum OCFL Version:** 1.0
* **OCFL Community Extensions Version:** 1.0
* **Obsoletes:** n/a
* **Obsoleted by:** n/a

## Overview

This object extension allows the import of one metadata file, which is 
validated against a json schema.

### Usage Scenario

To allow OCFL Viewers the display of easy to find semantic metadata, this
extension gives a schema and a position of the metadata file.

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
*
* **Name:** `storageName`
    * **Description:** Location within the specified Type
        * **area:** area name
        * **path:** subfolder within content folder
        * **extension:** subfolder within extension folder
    * **Type:** string
    * **Default:**

* **Name:** `format`
    * **Description:** format 
    * **Type:** string
    * **Default:**

* **Name:** `compress`
    * **Description:** Compression type for JSONL file
        * **none:** no compression
        * **gzip:** [gzip compression](https://en.wikipedia.org/wiki/Gzip)
        * **brotli:** [brotli compression](https://en.wikipedia.org/wiki/Brotli)
    * **Type:** string
    * **Default:**


## Procedure (tbd.)

Every entry whithin the [OCFL Object Manifest](https://ocfl.io/1.1/spec/#manifest)
is represented by a JSON line in a file called  `filesystem_<version>.jsonl[.gz|.br]`.
Since this file is immutable, every version of the ocfl object gets its own indexer file.

## Examples

JSON Entry for a file from Linux OS
```json
{
  "path": "data/test.odt",
  "meta": {
    "aTime": "2023-05-03T11:52:02.6948384+02:00",
    "mTime": "2023-01-15T13:49:09.7643455+01:00",
    "cTime": "2023-01-15T13:52:22.9096886+01:00",
    "attr": "-rwxrwxrwx",
    "os": "linux",
    "sysStat": {
      "Dev": 72,
      "Ino": 23643898043722929,
      "Nlink": 1,
      "Mode": 33279,
      "Uid": 1000,
      "Gid": 1000,
      "X__pad0": 0,
      "Rdev": 0,
      "Size": 4456,
      "Blksize": 4096,
      "Blocks": 16,
      "Atim": {
        "Sec": 1683107522,
        "Nsec": 694838400
      },
      "Mtim": {
        "Sec": 1673786949,
        "Nsec": 764345500
      },
      "Ctim": {
        "Sec": 1673787142,
        "Nsec": 909688600
      },
      "X__unused": [
        0,
        0,
        0
      ]
    },
    "stateVersion": "v1"
  }
}
```

JSON Entry for a file from Windows OS
```json
{
  "path": "data/test.odt",
  "meta": {
    "aTime": "2023-05-05T10:16:16.634688+02:00",
    "mTime": "2023-01-15T13:49:09.7643455+01:00",
    "cTime": "2023-01-15T13:50:27.6733047+01:00",
    "attr": "Archive",
    "os": "windows",
    "sysStat": {
      "FileAttributes": 32,
      "CreationTime": {
        "LowDateTime": 4049774711,
        "HighDateTime": 31008991
      },
      "LastAccessTime": {
        "LowDateTime": 3711819904,
        "HighDateTime": 31031081
      },
      "LastWriteTime": {
        "LowDateTime": 3270685119,
        "HighDateTime": 31008991
      },
      "FileSizeHigh": 0,
      "FileSizeLow": 4456
    },
    "stateVersion": "v1"
  }
}
```


### Result


# OCFL Community Extension NNNN: Metafile

* __Extension Name:__ NNNN-ro-crate
* **Authors:** Jürgen Enge (Basel)
* **Minimum OCFL Version:** 1.0
* **OCFL Community Extensions Version:** 1.0
* **Obsoletes:** n/a
* **Obsoleted by:** n/a

## Overview

This object extension enables the use of an existing ro-crate-metadata.json 
file to create an info.json metafile ([NNNN-metafile](NNNN-metafile.md)) 
and integrates the ro-crate metadata into metadata-export and -viewer. 

### Usage Scenario

To allow the use of ro-crate for various purposes, the ro-crate metadata extension
makes sure, that ro-crate-metadata.json is available and can be used for further processing.

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

* **Name:** `metafilename`
    * **Description:** the name of the metadata file. Extension MUST be `.json`. If empty, no metafile will be created. This name MUST be the same as in [NNNN-metafile](NNNN-metafile.md). 
    * **Type:** string
    * **Default:** `info.json`


## Procedure (tbd.)


## Examples

### Parameters

```json
{
  "extensionName": "NNNN-ro-crate",
  "storageType": "extension",
  "storageName": "metadata",
  "metafilename": "info.json"
}
```
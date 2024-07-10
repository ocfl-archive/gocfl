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
extension gives a schema, position and name of the metadata file. This can be used for external 
archive Managers to store semantic metadata.

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

* **Name:** `schemaUrl`
    * **Description:** url of the json metadata schema, to check the metafile content 
    * **Type:** string
    * **Default:**

* **Name:** `schema`
    * **Description:** local filename of schema, which contains the content of `schemaUrl`
    * **Type:** string
    * **Default:**

* **Name:** `name`
    * **Description:** the name of the metadata file. Extension MUST be `.json`
    * **Type:** string
    * **Default:** `info.json`


## Procedure (tbd.)

While adding or updating an OCFL object, a metadata file is added at the specified storage location with 
the specified name. Within this process, the file is validated against the given json schema.
The schema file is stored next to the config.json file within the extension folder.

## Examples

### Parameters

```json
{
  "extensionName": "NNNN-metafile",
  "storageType": "extension",
  "storageName": "metadata",
  "name": "info.json",
  "schema": "gocfl-info-1.0.json",
  "schemaUrl": "https://raw.githubusercontent.com/ocfl-archive/gocfl/main/gocfl-info-1.0.json"
}
```
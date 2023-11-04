# OCFL Community Extension NNNN: METS

* **Extension Name:** NNNN-mets
* **Authors:** Jürgen Enge (Basel)
* **Minimum OCFL Version:** 1.0
* **OCFL Community Extensions Version:** 1.0
* **Uses:** [NNNN-indexer](NNNN-indexer.md)(optional), [NNNN-metafile](NNNN-metafile.md)(optional), [NNNN-migration](NNNN-migration.md)(optional), [NNNN-content-subpath](NNNN-content-subpath.md)(optional)
* **Obsoletes:** n/a
* **Obsoleted by:** n/a

## Overview

For enhancing compatibility with classic archive information package 
formats (i.e. https://dilcis.eu), this extension provides a way to 
integrate a METS and a Premis file for every version of 
the OCFL object based on the inventory.

Technical metadata is provided by the [NNNN-indexer](NNNN-indexer.md) extension.
The [NNNN-metafile](NNNN-metafile.md) extension is for the mandatory
entry of the descriptive metadata section within the mets file.
Migrations provided by [NNNN-migration](NNNN-migration.md) are referenced
within the premis file. Using the [NNNN-content-subpath](NNNN-content-subpath.md) 
extension the METS and Premis files can be stored within the metadata folder.

### Usage Scenario

In order to be compatible with classic archive information package formats,
the OCFL object should contain a METS file for every version.
This extension avoids the need to generate the METS file manually or to embed 
another aip format within the OCFL object.

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

* **Name:** `primaryDescriptiveMetadata`
    * **Description:** File with primary descriptive metadata (mets:dmdSec)
    * **Format:** <type>:<area>:<filename>
    * **Type:** string
    * **Default:** `info:metadata:info.json`

* **Name:** `metsFile`
    * **Description:** Name of the mets file
    * **Type:** string
    * **Default:** `mets.xml`

* **Name:** `premisFile`
    * **Description:** Name of the premis file
    * **Type:** string
    * **Default:** `premis.xml`


## Caveat

Make sure, that the extensions are used in the correct order.
(`NNNN-mets` is normally at the end of the `UpdateObjectAfter` chain)


## Procedure (tbd.)

This extension is used to place a METS and Premis file within the object at the given location.
The source of these files depends on the implementation of this extension.

### Generation of METS and Premis files
Both files are generated based on information available in the inventory and the metadata files.

### Insertion
Both files already exists und are inserted into the object.


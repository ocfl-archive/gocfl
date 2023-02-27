# OCFL Community Extension NNNN: Indexer

* **Extension Name:** NNNN-indexer
* **Authors:** Jürgen Enge (Basel)
* **Minimum OCFL Version:** 1.0
* **OCFL Community Extensions Version:** 1.0
* **Obsoletes:** n/a
* **Obsoleted by:** n/a

## Overview

This object extension integrates technical metadata extraction during the ingest process.
It uses an external indexer service that provides the required functionality. The technical 
metadata is stored as newline delimited JSON format. 

### Usage Scenario

In order to have a better overview of the content contained in an OCFL object structure, technical metadata is very 
helpful. It is an aid to reporting as well as to preservation management.

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

* **Name:** `actions`
    * **Description:** actions to extract technical metadata. Capabilities depend on the indexer service.
        * **siegfried:** mimetype and pronom recognition using [Siegfried](https://www.itforarchivists.com/siegfried/) 
          library.
        * **ffprobe:** [ffmpeg](https://ffmpeg.org/) for audio/video metadata extraction
        * **identify:** [Image Magick](https://imagemagick.org/) for image metadata extraction
        * **tika:** [Tika](https://tika.apache.org/) for metadata extraction of office files, pdf etc.
        * **...**
    * **Type:** array of strings
    * **Default:**



## Caveat

Make sure, that there's an indexer service available to use this extension for ingest.
[gocfl](https://github.com/je4/gocfl) comes with a build in service.

## Procedure

For every version within the OCFL object, there's a file called `indexer_<version>.jsonl[.gz|.br]`. 
Every added or updated file has a line within this indexer json. 

## Examples

### Parameters


```json
{
  "extensionName": "NNNN-indexer",
  "storageType": "area",
  "storageName": "index",
  "actions": ["siegfried", "ffprobe", "identify", "tika"],
  "compress": "gzip"
}
```

### Result

#### File Structure

```
\---content
    |   README.md
    |
    +---data
    |   |   [...]
    |   |
    |   \---[...]
    |
    \---index
            indexer_v1.jsonl
```

#### readme.md
```markdown
### Description of folders


##### data
Payload of archival object

##### metadata
additional semantic metadata

##### index
additional technical metadata
```
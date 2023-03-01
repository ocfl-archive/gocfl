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

Every entry whithin the [OCFL Object Manifest](https://ocfl.io/1.1/spec/#manifest) 
is represented by a JSON line in a file called  `indexer_<version>.jsonl[.gz|.br]`.
Since this file is immutable, every version of the ocfl object gets its own indexer file.

## Examples

JSON Entry for an Image
```json
{
	"Digest": "5cb8c60eb3c7641561df988493acdd0fbc6b6325ec396a6eaf6a9cbc329e1790b006d61b4465371c21a105b0fb5a77dff9a219ed57ead6cd074d6b8a6e2be896",
	"Metadata": {
        "errors": {},
        "mimetype": "image/jpeg",
        "mimetypes": [ "image/jpeg" ],
        "height": 1512,
        "size": 668629,
        "width": 2016,
        "identify": { <complete result from identify command> },
        "siegfried": [
          {
            "Basis": [
              "extension match jpeg",
              "byte match at [[0 16] [256 12] [668627 2]] (signature 1/2)"
            ],
            "ID": "fmt/645",
            "MIME": "image/jpeg",
            "Name": "Exchangeable Image File Format (Compressed)",
            "Namespace": "pronom",
            "Version": "2.2.1",
            "Warning": ""
          }
       ]
    }
}
```

### Result


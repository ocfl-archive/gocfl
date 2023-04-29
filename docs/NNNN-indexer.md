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
        * **fulltext:** [Tika](https://tika.apache.org/) for fulltext extraction of pdf files
        * **...**
    * **Type:** array of strings
    * **Default:**



## Caveat

Make sure, that there's an indexer service available to use this extension for ingest.
[gocfl](https://github.com/je4/gocfl) comes with a build in service.

## Procedure (tbd.)

Every entry whithin the [OCFL Object Manifest](https://ocfl.io/1.1/spec/#manifest) 
is represented by a JSON line in a file called  `indexer_<version>.jsonl[.gz|.br]`.
Since this file is immutable, every version of the ocfl object gets its own indexer file.

## Examples

JSON Entry for an Image
```json
{
  "Path": "v1/content/data/together.png",
  "Indexer": {
    "mimetype": "image/png",
    "mimetypes": [
      "image/png"
    ],
    "pronom": "fmt/12",
    "pronoms": [
      "fmt/12"
    ],
    "width": 1920,
    "height": 1080,
    "size": 645553,
    "metadata": {
      "identify": {
        "magick": {
          "version": "1.0",
          "image": {
            "name": "together.png",
            "permissions": 666,
            "format": "PNG",
            "formatDescription": "Portable Network Graphics",
            "mimeType": "image/png",
            "class": "DirectClass",
            "geometry": {
              "width": 1920,
              "height": 1080
            },
            "resolution": {
              "x": 37.79,
              "y": 37.79
            },
            "printSize": {
              "x": 50.8071,
              "y": 28.579
            },
            "units": "PixelsPerCentimeter",
            "type": "TrueColor",
            "baseType": "Undefined",
            "endianness": "Undefined",
            "colorspace": "sRGB",
            "depth": 8,
            "baseDepth": 8,
            "channelDepth": {
              "blue": 1,
              "green": 8,
              "red": 8
            },
            "pixels": 6220800,
            "imageStatistics": {
              "Overall": {
                "max": 255,
                "mean": 205.498,
                "median": 219,
                "standardDeviation": 69.5403,
                "kurtosis": 3.43043,
                "skewness": -2.16984,
                "entropy": 0.369045
              }
            },
            "channelStatistics": {
            [...]
```

### Result


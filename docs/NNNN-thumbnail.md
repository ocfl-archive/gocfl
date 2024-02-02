# OCFL Community Extension NNNN: Thumbnail

* **Extension Name:** NNNN-thumbnail
* **Authors:** Jürgen Enge (Basel)
* **Minimum OCFL Version:** 1.0
* **OCFL Community Extensions Version:** 1.0
* **Obsoletes:** n/a
* **Obsoleted by:** n/a

## Overview

This object extension generates thumbnails of Media-Files

### Usage Scenario

To generate reports or content sites for OCFL Objects, it is helpful to display thumbnails to get an overview of 
the content.  

## Parameters

### Summary

* **Name:** `compress`
    * **Description:** Compression type for JSONL file
        * **none:** no compression
        * **gzip:** [gzip compression](https://en.wikipedia.org/wiki/Gzip)
        * **brotli:** [brotli compression](https://en.wikipedia.org/wiki/Brotli)
    * **Type:** string
    * **Default:** none
* **Name:** `ext`
    * **Description:** Image Format (Extension)
    * **Type:** string
    * **Default:** png
* **Name:** `width`
    * **Description:** Thumbnail width
    * **Type:** integer
    * **Default:** 256
* **Name:** `height`
    * **Description:** Thumbnail height
    * **Type:** integer
    * **Default:** 256
* **Name:** `singleDirectory`
    * **Description:** Write all thumbnails to a single directory. This is useful, if the number of thumbnails do not create problems with directory size.
    * **Type:** boolean
    * **Default:** false


## Procedure (tbd.)

Every entry whithin the [OCFL Object Manifest](https://ocfl.io/1.1/spec/#manifest)
is represented by a JSON line in a file called  `thumbnail_<version>.jsonl[.gz|.br]`.
Since this file is immutable, every version of the ocfl object gets its own indexer file.
Since thumbnails only add visual value to the OCFL Object, data and metadata is stored within the extension folder.

## Examples

It is not necessary to specify any parameters to use the default configuration.
However, if you were to do so, it would look like the following:

```json
{
  "extensionName": "NNNN-thumbnail",
  "compress": "gzip",
  "ext": "png",
  "width": 256,
  "height": 256
}
```

### Result

JSON Entry for a file with a "png" thumbnail based on process "Image#01". The file is identified by its digest/checksum.
```json
{
  "ext": "png",
  "id": "Image#01",
  "checksum": "5cb8c60eb3c7641561df988493acdd0fbc6b6325ec396a6eaf6a9cbc329e1790b006d61b4465371c21a105b0fb5a77dff9a219ed57ead6cd074d6b8a6e2be896"
}
```
The thumbnail itself resides in the file `extensions/NNNN-thumbnail/data/5/c/5cb8c60eb3c7641561df988493acdd0fbc6b6325ec396a6eaf6a9cbc329e1790b006d61b4465371c21a105b0fb5a77dff9a219ed57ead6cd074d6b8a6e2be896.png`


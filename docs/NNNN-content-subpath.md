# OCFL Community Extension NNNN: Content Subpath

* **Extension Name:** NNNN-content-subpath
* **Authors:** Jürgen Enge (Basel)
* **Minimum OCFL Version:** 1.0
* **OCFL Community Extensions Version:** 1.0
* **Obsoletes:** n/a
* **Obsoleted by:** n/a

## Overview

This object extension allows the creation of an additional path hierarchy within the content folder
of an object version.  

Internally, the paths are called `area` and OCFL software which supports this extension is able to
add or extract data from these specific areas.

### Usage Scenario

With this additional path layer, there can be for example a metadata, a data and a log subfolder
whereas the data folder contains the payload of the archived object.

## Parameters

### Summary

* **Name:** `subPath`
    * **Description:** map of named `PathDescription`. The entry name is the `area`.
    * **Type:** map
    * **Default:** 

#### `PathDescription`

* **Name:** `path`
    * **Description:** subpath in object content 
    * **Type:** string
    * **Default:**

* **Name:** `description`
    * **Description:** description of content belonging to this subfolder
    * **Type:** string
    * **Default:**

## Caveat

There MUST exist an `area` called `content` since this is the default area for adding payload 
files.

## Procedure

When adding a content file the subfolder will be automatically inserted  into the content path of the
manifest. Within the version `content` folder write a `readme.md` file containing the description of
the folders.

## Examples

### Parameters

It is not necessary to specify any parameters to use the default configuration.
However, if you were to do so, it would look like the following:

```json
{
  "extensionName": "NNNN-content-subpath",
  "subPath": {
    "content": {
      "path": "data",
      "description": "Payload of archival object"
    },
    "metadata": {
      "path": "metadata",
      "description": "additional semantic metadata"
    },
    "index": {
      "path": "index",
      "description": "additional technical metadata"
    }
  }
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
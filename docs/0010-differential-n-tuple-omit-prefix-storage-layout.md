# OCFL Community Extension 0010: Differential N-Tuple Omit Prefix Storage Layout

  * **Extension Name:** `0010-differential-n-tuple-omit-prefix-storage-layout`
  * **Authors:** Mike Giarlo
  * **Minimum OCFL Version:** 1.1
  * **OCFL Community Extensions Version:** 1.0
  * **Related to:** n/a

## Overview

This OCFL storage root extension describes a storage layout combining a pairtree-inspired root directory structure intended to support identifiers with differential tuple sizes. For identifiers such as the [DRUID ("Digital Resource Unique IDentifier")](https://sdr.library.stanford.edu/documentation/purls-dois-and-orcid-ids), n-tuples in the identifier are of different sizes. For example, for the identifier `bc123df5678`, one would expect four tuples: `bc`, `123`, `df`, and `5678`.

The OCFL object identifiers may contain prefixes which may be removed in the mapping to directory names. The OCFL object identifier prefix is defined as all characters before and including a configurable delimiter. The extension includes an option to include the full identifier as a leaf node.

This layout is based on the [0007-n-tuple-omit-prefix-storage-layout extension](https://ocfl.github.io/extensions/0007-n-tuple-omit-prefix-storage-layout), adding the ability to handle tuples of different sizes.

The limitations of this layout are filesystem dependent (with one exception), and are generally as follows:

* The size of object identifiers, minus the length of the prefix, cannot exceed the maximum allowed directory name size (eg. 255 characters)
* Object identifiers cannot include characters that are illegal in directory names
* This extension is defined over the ASCII subset of UTF-8 (code points 0x20 to 0x7F). Any character outside of this range in either an identifier or a path is an error.

## Parameters

### Summary

* **Name:** `delimiter`
  * **Description:** The case-insensitive delimiter marking the end of the OCFL object identifier prefix; MUST consist
    of a character string of length one or greater. If the delimiter is found multiple times in the OCFL object
    identifier, its last occurence (right-most) will be used to select the termination of the prefix.
  * **Type:** string
  * **Constraints:** Must not be empty
  * **Default:** :
* **Name**: `tupleSegmentSizes`
  * **Description:** Indicates the tuple segment sizes (in characters) to split the identifier into
  * **Type:** array
  * **Constraints:** An array of integers, each of which corresponds to a tuple segment size
  * **Default:** [2, 3, 2, 4]
* **Name:** `fullIdentifierAsObjectRoot`
  * **Description:** When true, indicates that the prefix-omitted, object identifier will be included as the leaf node of the layout, holding the object root.
  * **Type:** boolean
  * **Default:** false

## Procedure

The following is an outline of the steps to map an OCFL object identifier to an OCFL object root path:

1. Remove the prefix, which is everything to the left of the right-most instance of the delimiter, as well as the delimiter. If there is no delimiter, the whole id is used; if the delimiter is found at the end, an error is thrown.
1. Starting at the leftmost character of the resulting id and working right, divide the id into segments, where the number of segments is equal to the number of elements in the `tupleSegmentSizes` parameter array and the character size of each segment from left to right equals the corresponding integer value in the `tupleSegmentSizes` array. If the length of the identifier does not equal the sum of the `tupleSegmentSizes`, an error is thrown.
1. Create the start of the object root path by joining the tuples, in order, using the filesystem path separator.
1. Optionally, if `fullIdentifierAsObjectRoot` is true, complete the object root path by joining the prefix-omitted id (from step 1) onto the end after another filesystem path separator.

## Examples

### Example 1

This example demonstrates mappings where the single-character delimiter is found one or more times in the OCFL object
identifier, with default `tupleSegmentSizes` and `fullIdentifierAsObjectRoot` false.

#### Parameters

```json
{
    "extensionName": "0010-differential-n-tuple-omit-prefix-storage-layout",
    "delimiter": ":",
    "tupleSegmentSizes": [2, 3, 2, 4],
    "fullIdentifierAsObjectRoot": false
}
```

#### Mappings

| Object ID | Object Root Path |
| --- | --- |
| druid:gh875jh5489 | `gh/875/jh/5489` |
| namespace:11887296672 | `11/887/29/6672` |
| urn:nbn:fi:111-0023815  `11/1-0/02/3815` |
| abc123xyz89 | `ab/c12/3x/yz89`

#### Storage Hierarchy

```
[storage_root]/
├── 0=ocfl_1.1
├── ocfl_layout.json
├── extensions/
│   └── 0010-differential-n-tuple-omit-prefix-storage-layout/
│       └── config.json
├── 11
│   ├── 1-0
│   │   └── 02
│   │       └── 3815
│   │           ├── 0=ocfl_object_1.1
│   │           ├── inventory.json
│   │           ├── inventory.json.sha512
│   │           └── v1 [...]
│   └── 887
│       └── 29
│           └── 6672
│               ├── 0=ocfl_object_1.1
│               ├── inventory.json
│               ├── inventory.json.sha512
│               └── v1 [...]
├── ab
│   └── c12
│       └── 3x
│           └── yz89
│               ├── 0=ocfl_object_1.1
│               ├── inventory.json
│               ├── inventory.json.sha512
│               └── v1 [...]
└── gh/
    └── 875/
        └── jh/
            └── 5489/
                ├── 0=ocfl_object_1.1
                ├── inventory.json
                ├── inventory.json.sha512
                └── v1 [...]
```

### Example 2

This example demonstrates mappings where the multi-character delimiter is found one or more times in the OCFL object identifier, with custom `tupleSegmentSizes`, a custom `delimiter`, and `fullIdentifierAsObjectRoot` turned on.

#### Parameters

```json
{
    "extensionName": "0010-differential-n-tuple-omit-prefix-storage-layout",
    "delimiter": "edu/",
    "tupleSegmentSizes": [3, 4],
    "fullIdentifierAsObjectRoot": true
}
```

#### Mappings

| Object ID | Object Root Path |
| --- | --- |
| https://institution.edu/3448793 | `344/8793/3448793` |
| https://institution.edu/abc/edu/f8a905v | `f8a/905v/f8a905v` |

#### Storage Hierarchy

```
[storage_root]/
├── 0=ocfl_1.1
├── ocfl_layout.json
├── extensions/
│   └── 0010-differential-n-tuple-omit-prefix-storage-layout/
│       └── config.json
├── 344/
│   └── 8793/
│       └── 3448793/
│           ├── 0=ocfl_object_1.1
│           ├── inventory.json
│           ├── inventory.json.sha512
│           └── v1 [...]
└── f8a/
    └── 905v/
        └── f8a905v/
            ├── 0=ocfl_object_1.1
            ├── inventory.json
            ├── inventory.json.sha512
            └── v1 [...]
```

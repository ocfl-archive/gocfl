# OCFL Community Extension 0012: Hashed Truncated N-tuple Trees with Non-prefixed Object ID Encapsulating Directory for OCFL Storage Hierarchies

  * **Extension Name:** `0012-hash-and-no-prefix-id-n-tuple-storage-layout`
  * **Authors:** Elie Roux
  * **Minimum OCFL Version:** 1.0
  * **OCFL Community Extensions Version:** 1.0
  * **Related to:** [0003: Hashed Truncated N-tuple Trees with Encapsulating Directory for OCFL Storage Hierarchies](0003-hash-and-id-n-tuple-storage-layout.md) (extends)

## Overview

This storage root extension describes how to safely map OCFL object identifiers
of any length, containing any characters, to OCFL object root directories.

Except for the addition of stripping a prefix, this extension is otherwise the same
as [`0003-hash-and-id-n-tuple-storage-layout`](0003-hash-and-id-n-tuple-storage-layout.md).

Using this extension, OCFL object identifiers are stripped of a prefix, hashed and encoded as hex
strings (all letters lower-case). These digests are then divided into _N_
n-tuple segments, which are used to create nested paths under the OCFL storage
root. Finally, the OCFL object identifier is percent-encoded to create a
directory name for the OCFL object root (see ["Encapsulation
Directory"](#encapsulation-directory) section below).

The n-tuple segments approach allows OCFL object identifiers to be evenly
distributed across the storage hierarchy. The maximum number of files under any
given directory is controlled by the number of characters in each n-tuple, and
the tree depth is controlled by the number of n-tuple segments each digest is
divided into. The encoded encapsulation directory name provides visibility into
the object identifier from the file path (see ["Encapsulation
Directory"](#encapsulation-directory) section below for details).

## Encapsulation Directory

For basic OCFL object identifiers, the object identifier with prefix removed is used as the name of
the encapsulation directory (ie. the object root directory).

Some object identifiers could contain characters that are not safe for directory
names on all filesystems. Safe characters are defined as A-Z, a-z, 0-9, '-' and
'\_'. When an unsafe character is encountered in an object identifier, it is
percent-encoded using the lower-case hex characters of its UTF-8 encoding.

Some object identifiers with prefix removed could also result in an encoded string that is longer
than can be supported as a directory name. To handle that scenario, if the
percent-encoded object identifier with prefix removed is longer than 100 characters, it is truncated
to 100 characters, and then the digest of the original object identifier with prefix removed is
appended to the encoded object identifier like this:
\<encoded-object-identifier-with-prefix-removed-first-100-chars>-\<digest>. Note: this means that it
is no longer possible to determine the full object identifier with prefix removed from the
encapsulation directory name - some characters have been removed, and even the
first 100 characters of the encoded object identifier with prefix removed cannot be fully, reliably
decoded, because the truncation may leave a partial encoding at the end of the
100 characters.

| Object ID | Delimiter | Encapsulation Directory Name |
| --- | --- | --- |
| prefix:object-01 | [":"] | object-01 |
| Bad$$..Hor/rib:lè-$id | ["$$"] | %2e%2eHor%2frib%3al%c3%a8-%24id |
| abcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghija | [":"] |abcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghij-5cc73e648fbcff136510e330871180922ddacf193b68fdeff855683a01464220 |

## Parameters

### Summary

* **Name:** `digestAlgorithm`
  * **Description:** The digest algorithm to apply to the OCFL object identifier; MUST be an algorithm that is allowed in the OCFL fixity block
  * **Type:** string
  * **Constraints:** Must not be empty
  * **Default:** sha256
* **Name**: `tupleSize`
  * **Description:** Indicates the size of the segments (in characters) that the digest is split into
  * **Type:** number
  * **Constraints:** An integer between 0 and 32 inclusive
  * **Default:** 3
* **Name:** `numberOfTuples`
  * **Description:** Indicates how many segments are used for path generation
  * **Type:** number
  * **Constraints:** An integer between 0 and 32 inclusive
  * **Default:** 3
* **Name:** `delimiters`
  * **Description:** A list of possible delimiters marking the end of the OCFL object identifier prefix
  * **Type:** array
  * **Constraints:** MUST consist of an array of character strings of length one or greater. The array can have a length of 0.
  * **Default:** []

### Details

#### digestAlgorithm

`digestAlgorithm` is defaulted to `sha256`, and it MUST either contain a digest
algorithm that's [officially supported by the OCFL
specification](https://ocfl.io/1.0/spec/#digest-algorithms) or defined in a community
extension. The specified algorithm is applied to OCFL object identifiers to
produce hex encoded digest values that are then mapped to OCFL object root
paths.

#### tupleSize

`tupleSize` determines the number of digest characters to include in each tuple.
The tuples are used as directory names. The default value is `3`, which means
that each intermediate directory in the OCFL storage hierarchy could contain up
to 4096 directories. Increasing this value increases the maximum number of
sub-directories per directory.

If `tupleSize` is set to `0`, then no tuples are created and `numberOfTuples`
MUST also equal `0`.

The product of `tupleSize` and `numberOfTuples` MUST be less than or equal to
the number of characters in the hex encoded digest.

#### numberOfTuples

`numberOfTuples` determines how many tuples to create from the digest. The
tuples are used as directory names, and each successive directory is nested
within the previous. The default value is `3`, which means that every OCFL
object root will be 4 directories removed from the OCFL storage root, 3 tuple
directories plus 1 encapsulation directory. Increasing this value increases the
depth of the OCFL storage hierarchy.

If `numberOfTuples` is set to `0`, then no tuples are created and `tupleSize`
MUST also equal `0`.

The product of `numberOfTuples` and `tupleSize` MUST be less than or equal to
the number of characters in the hex encoded digest.

#### delimiters

`delimiters` is a list of case-sensitive delimiters that will be used to strip the prefix from the
OCFL object identifier.

If it is empty, this extension is equivalent to `0003-hash-and-id-n-tuple-storage-layout` as no prefix will be removed.

If a delimiter occurs at the end of the OCFL object identifier, that occurrence of a delimiter is ignored and the last previous one, if present, is used.

If an array of delimiters is used, there is no precedence of the delimiters; however, the final occurrence of any delimiter in the array identifies the end of the prefix.

| Object ID | Delimiter | Object id with prefix removed |
| --- | --- | --- |
| abcd | ["d"] | abcd |
| abcd | ["c", "d"] | d |
| abcdd | ["d"] | d |

## Procedure

The following is an outline of the steps to follow to map an OCFL object
identifier to an OCFL object root path using this extension (also see the
["Python Code"](#python-code) section):

1. Remove the prefix, which is everything to the left of the right-most instance of any of the delimiters, as well as the corresponding delimiter. If there is no delimiter, the whole id is used; if a delimiter is found at the end, it is ignored.
2. The OCFL object identifier with prefix removed is encoded as UTF-8 and hashed using the specified
   `digestAlgorithm`.
3. The digest is encoded as a lower-case hex string.
4. Starting at the beginning of the digest and working forwards, the digest is
   divided into `numberOfTuples` tuples each containing `tupleSize` characters.
5. The tuples are joined, in order, using the filesystem path separator.
6. The OCFL object identifier with prefix removed is percent-encoded to create the encapsulation
   directory name (see ["Encapsulation Directory"](#encapsulation-directory)
   section above for details).
7. The encapsulation directory name is joined to the end of the path.

## Examples

### Example 1

This example demonstrates what the OCFL storage hierarchy looks like when using
this extension's default configuration.

#### Parameters

It is not necessary to specify any parameters to use the default configuration.
However, if you were to do so, it would look like the following:

```json
{
    "extensionName": "0012-hash-and-no-prefix-id-n-tuple-storage-layout",
    "digestAlgorithm": "sha256",
    "tupleSize": 3,
    "numberOfTuples": 3,
    "delimiters": []
}
```

#### Mappings

| Object ID | Object ID with prefix removed | Digest | Object Root Path |
| --- | --- | --- | --- |
| `object-01` | `object-01` | `3c0ff4240c1e116dba14c7627f2319b58aa3d77606d0d90dfc6161608ac987d4` | `3c0/ff4/240/object-01` |
| `..hor/rib:le-$id` | `..hor/rib:le-$id` | `487326d8c2a3c0b885e23da1469b4d6671fd4e76978924b4443e9e3c316cda6d` | `487/326/d8c/%2e%2ehor%2frib%3ale-%24id` |

#### Storage Hierarchy

```
[storage_root]/
├── 0=ocfl_1.0
├── ocfl_layout.json
├── extensions/0012-hash-and-no-prefix-id-n-tuple-storage-layout/config.json
├── 3c0/
│   └── ff4/
│       └── 240/
│           └── object-01/
│               ├── 0=ocfl_object_1.0
│               ├── inventory.json
│               ├── inventory.json.sha512
│               └── v1 [...]
└── a6b/
    └── 979/
        └── 238/
            └── rib%3ale-%24id/
                ├── 0=ocfl_object_1.0
                ├── inventory.json
                ├── inventory.json.sha512
                └── v1 [...]
```

### Example 2

This example demonstrates the effects of modifying the default parameters to use
a different `digestAlgorithm`, smaller `tupleSize`, larger `numberOfTuples` and
non-empty `delimiters` array.

#### Parameters

```json
{
    "extensionName": "0012-hash-and-no-prefix-id-n-tuple-storage-layout",
    "digestAlgorithm": "md5",
    "tupleSize": 2,
    "numberOfTuples": 15,
    "delimiters": ["/"]
}
```

#### Mappings

| Object ID | Object ID with prefix removed | Digest | Object Root Path |
| --- | --- | --- | --- |
| `object-01` | `object-01` | `ff75534492485eabb39f86356728884e` | `ff/75/53/44/92/48/5e/ab/b3/9f/86/35/67/28/88/object-01` |
| `..hor/rib:le-$id` | `rib:le-$id` | `5d6e4e8cb5cd0c7a8fbf65c1295127e3` | `5d/6e/4e/8c/b5/cd/0c/7a/8f/bf/65/c1/29/51/27/rib%3ale-%24id` |

#### Storage Hierarchy

```
[storage_root]/
├── 0=ocfl_1.0
├── ocfl_layout.json
├── extensions/0012-hash-and-no-prefix-id-n-tuple-storage-layout/config.json
├── 5d/
│   └── 6e/
│       └── 4e/
│           └── 8c/
│               └── b5/
│                   └── cd/
│                       └── 0c/
│                           └── 7a/
│                               └── 8f/
│                                   └── bf/
│                                       └── 65/
│                                           └── c1/
│                                               └── 29/
│                                                   └── 51/
│                                                       └── 27/
│                                                           └── rib%3ale-%24id/
│                                                               ├── 0=ocfl_object_1.0
│                                                               ├── inventory.json
│                                                               ├── inventory.json.sha512
│                                                               └── v1 [...]
└── ff/
    └── 75/
        └── 53/
            └── 44/
                └── 92/
                    └── 48/
                        └── 5e/
                            └── ab/
                                └── b3/
                                    └── 9f/
                                        └── 86/
                                            └── 35/
                                                └── 67/
                                                    └── 28/
                                                        └── 88/
                                                            └── object-01/
                                                                ├── 0=ocfl_object_1.0
                                                                ├── inventory.json
                                                                ├── inventory.json.sha512
                                                                └── v1 [...]
```

### Example 3

This example demonstrates what happens when `tupleSize` and `numberOfTuples` are
set to `0`. This is an edge case and not a recommended configuration.

#### Parameters

```json
{
    "extensionName": "0012-hash-and-no-prefix-id-n-tuple-storage-layout",
    "digestAlgorithm": "sha256",
    "tupleSize": 0,
    "numberOfTuples": 0,
    "delimiters": ["/"]
}
```

#### Mappings

| Object ID | Object ID with prefix removed | Digest | Object Root Path |
| --- | --- | --- | --- |
| `object-01` | `object-01` | `3c0ff4240c1e116dba14c7627f2319b58aa3d77606d0d90dfc6161608ac987d4` | `object-01` |
| `..hor/rib:le-$id` | `rib:le-$id` | `a6b979238e7131d89e45b913942c33374ce5c09d348727b50c5f995d0ce4f7f8` | `rib%3ale-%24id` |

#### Storage Hierarchy

```
[storage_root]/
├── 0=ocfl_1.0
├── ocfl_layout.json
├── extensions/0012-hash-and-no-prefix-id-n-tuple-storage-layout/config.json
├── object-id/
│   ├── 0=ocfl_object_1.0
│   ├── inventory.json
│   ├── inventory.json.sha512
│   └── v1 [...]
└── rib%3ale-%24id/
    ├── 0=ocfl_object_1.0
    ├── inventory.json
    ├── inventory.json.sha512
    └── v1 [...]
```

## Python Code

Here is some python code that implements the algorithm:

```
import codecs
import hashlib
import os
import re

def _remove_prefixes(object_id, delimiters):
    rightmost_idx = -1
    for delimiter in delimiters:
        # ignore empty string
        if len(delimiter) > 0:
            if (delimiter_idx := object_id.rfind(delimiter, 0, len(object_id)-1)) >= 0:
                rightmost_idx = max(rightmost_idx, delimiter_idx+len(delimiter))
    if rightmost_idx > 0:
        return object_id[rightmost_idx:]
    return object_id


def _percent_encode(c):
    c_bytes = c.encode('utf8')
    s = ''
    for b in c_bytes:
        s += '%' + codecs.encode(bytes([b]), encoding='hex_codec').decode('utf8')
    return s


def _get_encapsulation_directory(object_id, digest):
    d = ''
    for c in object_id:
        if re.match(r'[A-Za-z0-9-_]{1}', c):
            d += c
        else:
            d += _percent_encode(c)
    if len(d) > 100:
        return f'{d[:100]}-{digest}'
    return d


def ocfl_path(object_id, algorithm='sha256', tuple_size=3, number_of_tuples=3, delimiters=[]):
    object_id = _remove_prefixes(object_id, delimiters)
    object_id_utf8 = object_id.encode('utf8')
    if algorithm == 'md5':
        digest = hashlib.md5(object_id_utf8).hexdigest()
    elif algorithm == 'sha256':
        digest = hashlib.sha256(object_id_utf8).hexdigest()
    elif algorithm == 'sha512':
        digest = hashlib.sha512(object_id_utf8).hexdigest()
    digest = digest.lower()
    path = ''
    for i in range(number_of_tuples):
        part = digest[i*tuple_size:i*tuple_size+tuple_size]
        path = os.path.join(path, part)
    encapsulation_directory = _get_encapsulation_directory(object_id, digest=digest)
    path = os.path.join(path, encapsulation_directory)
    return path


def _check_path(object_id, correct_path, algorithm='sha256', tuple_size=3, number_of_tuples=3, delimiters=[]):
    p = ocfl_path(object_id, algorithm=algorithm, tuple_size=tuple_size, number_of_tuples=number_of_tuples, delimiters=delimiters)
    assert p == correct_path, f'{p} != {correct_path}'
    print(f'  "{object_id}" {algorithm} => {p}')


def run_tests():
    print('running tests...')
    assert _remove_prefixes('ab/cd', ['/']) == 'cd'
    assert _remove_prefixes('ab/cd', []) == 'ab/cd'
    assert _remove_prefixes('ab/cd:ef', ['/', ':']) == 'ef'
    assert _remove_prefixes('ab/cd:', ['/', ':']) == 'cd:'
    assert _remove_prefixes('abcd', ['d']) == 'abcd'
    assert _remove_prefixes('abcd', ['c', 'd']) == 'd'
    assert _remove_prefixes('abcdd', ['c', 'd']) == 'd'
    assert _remove_prefixes('abcde', ['abc']) == 'de'
    assert _remove_prefixes('abcde', ['bcd']) == 'e'
    assert _remove_prefixes('abcde', ['cde']) == 'abcde'
    assert _percent_encode('.') == '%2e'
    assert _percent_encode('ç') == '%c3%a7'
    _check_path(object_id='object-01', correct_path='3c0/ff4/240/object-01')
    _check_path(object_id='object-01', correct_path='938/db8/c9f/01', delimiters=['-'])
    _check_path(object_id='object-01', correct_path='ff7/553/449/object-01', algorithm='md5')
    _check_path(object_id='object-01', correct_path='ff755/34492/object-01', algorithm='md5', tuple_size=5, number_of_tuples=2)
    _check_path(object_id='object-01', correct_path='object-01', algorithm='md5', tuple_size=0, number_of_tuples=0)
    _check_path(object_id='object-01', correct_path='ff/75/53/44/92/48/5e/ab/b3/9f/86/35/67/28/88/object-01', algorithm='md5', tuple_size=2, number_of_tuples=15)
    _check_path(object_id='..hor/rib:le-$id', correct_path='487/326/d8c/%2e%2ehor%2frib%3ale-%24id')
    _check_path(object_id='..hor/rib:le-$id', correct_path='083/197/66f/%2e%2ehor%2frib%3ale-%24id', algorithm='md5') #08319766fb6c2935dd175b94267717e0
    _check_path(object_id='..Hor/rib:lè-$id', correct_path='373/529/21a/%2e%2eHor%2frib%3al%c3%a8-%24id')
    long_object_id = 'abcdefghij' * 26
    long_object_id_digest = '55b432806f4e270da0cf23815ed338742179002153cd8d896f23b3e2d8a14359'
    _check_path(object_id=long_object_id, correct_path=f'55b/432/806/abcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghij-{long_object_id_digest}')
    long_object_id_101 = 'abcdefghij' * 10 + 'a'
    long_object_id_101_digest = '5cc73e648fbcff136510e330871180922ddacf193b68fdeff855683a01464220'
    _check_path(object_id=long_object_id_101, correct_path=f'5cc/73e/648/abcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghij-{long_object_id_101_digest}')


if __name__ == '__main__':
    run_tests()
```

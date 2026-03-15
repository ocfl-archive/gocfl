# OCFL Community Extension 0005: Mutable HEAD

  * **Extension Name:** `0005-mutable-head`
  * **Authors:** Peter Winckles
  * **Minimum OCFL Version:** 1.0
  * **OCFL Community Extensions Version:** 1.0
  * **Related to:** n/a

## Overview

In the core OCFL specification, every mutation of an object is recorded as a new
immutable version. This extension enables an OCFL client to create a HEAD
version of an object that is able to be mutated in-place without generating
additional versions. This allows changes to an object to be accumulated over a
period of time within an object's OCFL object root before they are committed to
an immutable version, reducing version inflation and providing a client-agnostic
mechanism for interpreting content that has not yet been officially versioned in
OCFL.

## Important Note

Changes should be left in the mutable HEAD for as short of time as possible
before committing them into an immutable OCFL version. OCFL clients that do not
implement this extension CANNOT interact with content in the mutable HEAD, and
the creation of new versions before the mutable HEAD is committed invalidates
the contents of the mutable HEAD.

## Specification

### Terminology

* **Mutable HEAD:** A mutable OCFL version that contains the HEAD state of an
  object and is outside of the core OCFL specification.
* **Mutable HEAD version:** The version identifier of the mutable HEAD, eg `v3`
* **Root HEAD version:** The version identifier of the most recent immutable
  OCFL version under the object root.
* **Extension directory:** This extension’s root directory within objects that
  use it, `[object-root]/extensions/0005-mutable-head`.
* **Mutable HEAD version directory:** The directory within the extension
  directory that contains the OCFL version for the mutable HEAD,
  `[object-root]/extensions/0005-mutable-head/head`.
* **Mutable HEAD inventory:** The inventory located in the mutable HEAD version
  directory.
* **Root inventory:** The inventory located in the OCFL object’s root.
* **Mutable HEAD content directory:** The version content directory within the
  mutable HEAD version directory,
  `[object-root]/extensions/0005-mutable-head/head/content`.
* **Commit:** The action of moving the mutable HEAD out of the extension
  directory and into the OCFL object root as a new immutable version.
* **Revision:** An update that is applied to a mutable HEAD.
* **Revision marker:** A file that notes the existence of a revision.

### Mutable HEAD

A mutable HEAD is a OCFL version that may be mutated in-place until it is
committed to an immutable version. The OCFL specification does not support this
behavior; therefore the mutable HEAD is not kept alongside the rest of the
versions within the OCFL object root. Objects do not have mutable HEADs by
default. When a mutable HEAD is added to an object, the latest immutable version
of the object is NOT mutated. Instead, a new, mutable version is created within
the extension directory. This new version is treated as the HEAD of the object
even though it is not in the object root and not referenced in the root
inventory. When OCFL clients that implement this extension access an object they
MUST first check to see if the object contains a mutable HEAD, and, if so, load
the mutable HEAD inventory instead of the root inventory.

#### Structure

All files related to the mutable HEAD extension MUST be contained within
`[object-root]/extensions/0005-mutable-head`. This extension directory MUST
contain three children:

1. `root-inventory.json.sha512`: A copy of the root inventory's sidecar at the
   time the mutable HEAD was created.
2. `revisions`: A directory that contains revision markers.
3. `head`: The mutable HEAD version directory, the contents of which are
   identical to a standard OCFL version directory.

Here is an example:

```
[object-root]/
    ├── 0=ocfl_object_1.0
    ├── inventory.json
    ├── inventory.json.sha512
    ├── extensions/
    │   └── 0005-mutable-head/
    │       ├── root-inventory.json.sha512
    │       ├── revisions
    │       │   ├── r1/
    │       │   └── ... more revision markers ...
    │       └── head/
    │           ├── inventory.json
    │           ├── inventory.json.sha512
    │           └── content/
    │               ├── r1/
    │               │   └── ... files ...
    |               └── ... other revisions ...
    └── v1/
        ├── inventory.json
        ├── inventory.json.sha512
        └── content/
            └── ... files ...
```

The extension directory MUST contain a copy of the root inventory’s sidecar file
at the time that the mutable HEAD was created, named `root-inventory.json.ALGORITHM`
(eg. `root-inventory.json.sha512`). This file is used to ensure that the root
object has not been modified between the time the mutable HEAD was created and
it was committed.

The mutable HEAD MUST be a valid OCFL version. Unlike normal versions, it MUST
be stored at `[object-root]/extensions/0005-mutable-head/head`.

The mutable HEAD inventory file MUST NOT be written to the object root, and the
root inventory file MUST NOT reference files within the extension directory.

If there is not an active mutable HEAD, then the extension directory MUST NOT
exist. An object has an active mutable HEAD when a mutable HEAD inventory exists
at `[object-root]/extensions/0005-mutable-head/head/inventory.json`. When there
is an active mutable HEAD, new root object versions MUST NOT be created as this
results in version conflicts between the root object and the mutable HEAD.

The `revisions` directory MUST contain a revision marker for every mutation that
has been made to the mutable HEAD. Revisions are discussed in more detail in the
[revisions section](#revisions).

#### Inventory

When a mutable HEAD is created, if the HEAD version defined in the root
inventory file is `N`, then the mutable HEAD version MUST be `N+1`. Regardless
of how many times a mutable HEAD is mutated, its version MUST NEVER change. The
mutable HEAD’s inventory file contains everything in the root inventory file,
plus all content (manifest entries, version state, etc) pertinent to version
`N+1`.

Files that are added in the mutable HEAD are added to the inventory manifest as
normal. Content paths MUST be relative to the object root. The manifest MUST NOT
reference files in the mutable HEAD content directory that are no longer
referenced in the mutable HEAD version’s state.

The version fields `created`, `message`, and `user` SHOULD be overwritten on
every update of the mutable HEAD.

The following is an example mutable HEAD inventory file:

```json
{
  "digestAlgorithm": "sha512",
  "head": "v2",
  "id": "ark:/12345/bcd987",
  "manifest": {
    "4d27c8...b53": [ "extensions/0005-mutable-head/head/content/r1/foo/bar.xml" ],
    "9bb43j...n3a": [ "extensions/0005-mutable-head/head/content/r2/file1.txt" ],
    "u8b99v...7b2": [ "extensions/0005-mutable-head/head/content/r3/file1.txt" ],
    "7dcc35...c31": [ "v1/content/foo/bar.xml" ],
    "cf83e1...a3e": [ "v1/content/empty.txt" ],
    "ffccf6...62e": [ "v1/content/image.tiff" ]
  },
  "type": "https://ocfl.io/1.0/spec/#inventory",
  "versions": {
    "v1": {
      "created": "2018-01-01T01:01:01Z",
      "message": "Initial import",
      "state": {
        "7dcc35...c31": [ "foo/bar.xml" ],
        "cf83e1...a3e": [ "empty.txt" ],
        "ffccf6...62e": [ "image.tiff" ]
      },
      "user": {
        "address": "mailto:alice@example.com",
        "name": "Alice"
      }
    },
    "v2": {
      "created": "2018-02-02T02:02:02Z",
      "message": "Fix bar.xml, remove image.tiff, add empty2.txt, rename file1.txt to file2.txt, add updated file1.txt",
      "state": {
        "9bb43j...n3a": [ "file2.txt" ],
        "u8b99v...7b2": [ "file1.txt" ],
        "4d27c8...b53": [ "foo/bar.xml" ],
        "cf83e1...a3e": [ "empty.txt", "empty2.txt" ]
      },
      "user": {
        "address": "mailto:bob@example.com",
        "name": "Bob"
      }
    }
  }
}
```

#### Revisions

Any change that is made to a mutable HEAD is known as a revision. Each revision
is assigned a revision number in the form of `rN`, where `N` is a positive
integer greater than 0, similar to OCFL version numbers. When a mutable HEAD is
first created, its initial revision number is `r1`. Every subsequent change
that's made to the mutable HEAD MUST use the next available revision number. For
example, the second change would use revision `r2` and so forth.

Revisions are tracked by writing revision marker files to the revisions
directory located at `[object-root]/extensions/0005-mutable-head/revisions`. A
revision marker file MUST be named using the revision number (eg. `r1`).
Revision marker files MUST contain only the marker's revision number and no
whitespace.

A new revision marker MUST be created before applying an update to the mutable
HEAD. If the revision marker already exists, this indicates that a concurrent
update occurred, and the pending update MUST be aborted. When using storage
implementations that do not support atomic file creation, this check alone is
not sufficient to guard against concurrent modifications.

#### Content Directory

The mutable HEAD content directory MUST NOT contain any files that are not
referenced in the mutable HEAD inventory manifest.

The mutable HEAD content directory is further subdivided by directories named
for revision numbers. Every file that's added to the mutable HEAD MUST be placed
within the revision sub-directory that corresponds with the target revision
number.

For example, if `foo.txt` is added in the second revision of an object's mutable
HEAD, then the files content path is
`extensions/0005-mutable-head/head/content/r2/foo.txt`.

If a revision does not add any files with digests not in the mutable HEAD
inventory's manifest, then a new content revision directory MUST NOT be created.

### Mutable HEAD on New Objects

If a new object is created with a mutable HEAD, an empty OCFL `v1` version MUST
be created first. This is required because an OCFL object cannot exist without
at least one version, and the mutable HEAD is not recognized as a version by the
core specification. At the minimum, an inventory that contains a single version
with no files in its manifest and state MUST be created, along with the
corresponding version directory and inventory files in the object root, as
defined in the core specification. Then, the mutable HEAD can be created as
version `v2`.

### Version Conflicts

A version conflict occurs if a new OCFL object version is created after a
mutable HEAD is created but before it is committed. This results in two
different versions with the same version number but different contents. How
version conflicts are resolved is up to the client implementation.

### Operations

#### Mutate

Unlike the OCFL versions under the object root, the mutable HEAD MAY be mutated
in-place. Updates to the mutable HEAD MUST NOT change any files in the object
outside of the extension directory. The [implementation
notes](#mutable-head-mutation) contain suggestions on how to mutate it safely.

#### Commit

At some point, the mutable HEAD SHOULD be committed to an immutable version in
the object root. The end result of a successful commit operation MUST be that
the extension directory no longer exists, and the mutable HEAD is installed in
the object root as the root HEAD version. The manifest and fixity blocks in the
mutable HEAD inventory might reference files within the extensions directory,
and these paths MUST be rewritten when the mutable HEAD is committed. It is left
up to the client implementation to handle any version conflicts that are
encountered during a commit. Commit implementation suggestions are in the
[implementation notes](#committing-a-mutable-head).

#### Access

When a mutable HEAD exists, the mutable HEAD inventory MUST be used as the
current HEAD inventory rather than the inventory in the object root. If there is
a version conflict, then it is up to the client implementation to decide what to
do.

## Implementation Notes

Note: The following notes reference content paths as `[version]/content` and the
inventory sidecar as `inventory.json.sha512`. These are the default OCFL values,
but they can differ based on inventory file settings.

### Mutable HEAD Creation

When an object does not have an active mutable HEAD, then a new mutable version
is created following similar steps as outlined in the [OCFL implementation
notes](https://ocfl.io/1.0/implementation-notes/#an-example-approach-to-updating-ocfl-object-versions)
with two notable differences.

1. The content paths of newly added files must be relative to
   `extensions/0005-mutable-head/head/content/r1`.
2. Instead of moving the new version to `[object-root]/vN`, it must be moved to
   `[object-root]/extensions/0005-mutable-head/head`.

The procedure for creating a new mutable HEAD is as follows:

1. Stage a new OCFL version as you would normally, noting the above exceptions.
2. Write the file `[object-root]/extensions/0005-mutable-head/revisions/r1`
   containing the text `r1`. If this fails, abort because another process
   already created a mutable HEAD.
3. Move the staged OCFL version to `[object-root]/extensions/0005-mutable-head/head`.
4. Copy `[object-root]/inventory.json.sha512` to `[object-root]/extensions/0005-mutable-head/root-inventory.json.sha512`

### Mutable HEAD Mutation

Mutating an existing mutable HEAD requires creating a new revision and is more
involved than the standard process for creating a new OCFL version. The
following process should provide close to the same level of safety as is
expected when creating a new OCFL version.

1. Identify the next available revision number, `rN`, by inspecting the revision
   markers in `[object-root]/extensions/0005-mutable-head/revisions`.
2. Create a new revision directory somewhere in temporary space named `rN`.
3. Create a copy of the existing mutable HEAD inventory, hence forth referred to
   as the "new inventory".
4. Write any files that are added with digests not in the new inventory's
   manifest to the revision directory, and add corresponding manifest entries to
   the inventory.
5. Update the version state in the new inventory to reflect any changes that
   were made to the object (deletions, renames, additions, modifications).
6. Update the new inventory’s manifest and fixity blocks, removing any entries
   that were introduced in prior revisions of the mutable HEAD but are no longer
   referenced in the current version state.
7. Write the new inventory and its sidecar to temporary space.
8. Write the file `[object-root]/extensions/0005-mutable-head/revisions/rN`
   containing the text `rN`. If this fails, abort because another process
   already created a revision using the same revision number.
9. If the revision directory is not empty, move it to
   `[object-root]/extensions/0005-mutable-head/head/content/rN`.
10. Move the new inventory file and sidecar into `[object-root]/extensions/0005-mutable-head/head`.
11. Finally, iterate over all of the files under
    `[object-root]/extensions/0005-mutable-head/head/content`, deleting any that
    are no longer referenced in the inventory manifest.

### Committing a Mutable HEAD

A commit is simply the process of moving an object’s mutable HEAD into the
object’s root as an immutable OCFL version. Because the mutable HEAD is already
a valid OCFL version, this is relatively straightforward.

1. Compare the contents of
   `[object-root]/extensions/0005-mutable-head/root-inventory.json.sha512` and
   `[object-root]/inventory.json.sha512` to ensure that they are the same and
   that the root object has not been modified since the creation of the mutable
   HEAD. If their contents are different, there is a version conflict that is up
   to the implementation to resolve.
2. Rewrite the inventory manifest and fixity entries that reference
   `extensions/0005-mutable-head/head` to reference `vN` instead, where `vN` is
   the mutable HEAD version number.
3. Move `[object-root]/extensions/0005-mutable-head/head` to `[object-root]/vN`.
4. If the directory cannot be moved because there is already a version directory
   at that location, then there is a version conflict that is up to the
   implementation to resolve.
5. Write the updated inventory and sidecar to the object root.
6. If this fails, move `[object-root]/vN` back to
   `[object-root]/extensions/0005-mutable-head/head` and abort.
7. Otherwise, write the updated inventory and sidecar to `[object-root]/vN`.
8. Remove the extension directory.

The following is what the [mutable HEAD inventory example](#inventory) would
look like after it is committed:

```json
{
  "digestAlgorithm": "sha512",
  "head": "v2",
  "id": "ark:/12345/bcd987",
  "manifest": {
    "4d27c8...b53": [ "v2/content/r1/foo/bar.xml" ],
    "9bb43j...n3a": [ "v2/content/r2/file1.txt" ],
    "u8b99v...7b2": [ "v2/content/r3/file1.txt" ],
    "7dcc35...c31": [ "v1/content/foo/bar.xml" ],
    "cf83e1...a3e": [ "v1/content/empty.txt" ],
    "ffccf6...62e": [ "v1/content/image.tiff" ]
  },
  "type": "https://ocfl.io/1.0/spec/#inventory",
  "versions": {
    "v1": {
      "created": "2018-01-01T01:01:01Z",
      "message": "Initial import",
      "state": {
        "7dcc35...c31": [ "foo/bar.xml" ],
        "cf83e1...a3e": [ "empty.txt" ],
        "ffccf6...62e": [ "image.tiff" ]
      },
      "user": {
        "address": "mailto:alice@example.com",
        "name": "Alice"
      }
    },
    "v2": {
      "created": "2018-02-02T02:02:02Z",
      "message": "Fix bar.xml, remove image.tiff, add empty2.txt",
      "state": {
        "9bb43j...n3a": [ "file2.txt" ],
        "u8b99v...7b2": [ "file1.txt" ],
        "4d27c8...b53": [ "foo/bar.xml" ],
        "cf83e1...a3e": [ "empty.txt", "empty2.txt" ]
      },
      "user": {
        "address": "mailto:bob@example.com",
        "name": "Bob"
      }
    }
  }
}
```

### Accessing a Mutable HEAD

An object with a mutable HEAD is accessed in much the same way as an object that
does not have a mutable HEAD. The primary difference is that the inventory file
at `[object-root]/extensions/0005-mutable-head/head/inventory.json` is used
instead of the inventory file in the object root. Same as an immutable version,
the mutable HEAD version is a valid version with a version number and all of its
content paths relative the object root.

### Purging a Mutable HEAD

A mutable HEAD can be purged by simply deleting the extension directory. Purging
may be desirable if the changes the mutable HEAD contains are no longer wanted,
or as a means of resolving a version conflict between the mutable HEAD and the
root object.

### OCFL Version Creation

Implementations should not allow the creation of new OCFL versions while there
is an active mutable HEAD, as doing so causes version conflicts. Before a new
version can be created, the mutable HEAD must either be committed or purged.

### Resolving Conflicts

When a version conflict occurs, it is up to implementations to decide how to
resolve them. One simple approach is to fail whatever operation detected the
conflict until either the mutable HEAD is purged or the conflict is manually
resolved. Regardless, it is a good idea to check for conflicts every time an
object is accessed so that they are detected as early as possible rather than
waiting until whenever the mutable HEAD is committed.

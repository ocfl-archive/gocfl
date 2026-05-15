<img src="https://avatars0.githubusercontent.com/u/35607965" alt="OCFL Hand-drive logo" style="float:right;width:307px;height:307px;"/>

# Oxford Common File Layout Specification

## Recommendation 07 July 2020

This version:
* [https://ocfl.io/1.0/spec/](https://ocfl.io/1.0/spec/)

Latest published version:
* <https://ocfl.io/latest/spec/>

Editors:
* [Andrew Hankinson](https://orcid.org/0000-0003-2663-0003) ([Bodleian Libraries, University of
 Oxford](http://www.bodleian.ox.ac.uk/))
* [Neil Jefferies](https://orcid.org/0000-0003-3311-3741) ([Bodleian Libraries, University of
 Oxford](http://www.bodleian.ox.ac.uk/))
* [Rosalyn Metz](https://orcid.org/0000-0003-3526-2230) ([Emory
 University](https://web.library.emory.edu/))
* [Julian Morley](https://orcid.org/0000-0003-4176-1933) ([Stanford
 University](https://library.stanford.edu/))
* [Simeon Warner](https://orcid.org/0000-0002-7970-7855) ([Cornell
 University](https://www.library.cornell.edu/))
* [Andrew Woods](https://orcid.org/0000-0002-8318-4225) ([LYRASIS](https://lyrasis.org/))

Additional Documents:
* [Implementation Notes](https://ocfl.io/1.0/implementation-notes/)
* [Validation Codes](https://ocfl.io/1.0/spec/validation-codes.html)
* [Extensions](https://github.com/OCFL/extensions/)

Previous version:
* <https://ocfl.io/0.9/spec/>

Repository:
* [Github](https://github.com/ocfl/spec)
* [Issues](https://github.com/ocfl/spec/issues)
* [Commits](https://github.com/ocfl/spec/commits)
* [Use Cases](https://github.com/ocfl/Use-Cases)

This document is licensed under a [Creative Commons Attribution 4.0
License](https://creativecommons.org/licenses/by/4.0/).\
[OCFL logo:
hand-drive](https://avatars0.githubusercontent.com/u/35607965)
by [Patrick Hochstenbach](http://orcid.org/0000-0001-8390-6171) is
licensed under [CC BY
2.0](https://creativecommons.org/licenses/by/2.0/).

## Introduction

*This section is non-normative.*

This Oxford Common File Layout (OCFL) specification describes an
application-independent approach to the storage of digital objects in a
structured, transparent, and predictable manner. It is designed to
promote long-term access and management of digital objects within
digital repositories.

### Need

The OCFL initiative began as a discussion amongst digital repository
practitioners to identify well-defined, common, and
application-independent file management for a digital repository's
persisted objects and represents a specification of the community's
collective recommendations addressing five primary requirements:
completeness, parsability, versioning, robustness, and storage
diversity.

#### Completeness

The OCFL recommends storing metadata and the content it describes
together so the OCFL object can be fully understood in the absence of
original software. The OCFL does not make recommendations about what
constitutes an object, nor does it assume what type of metadata is
needed to fully understand the object, recognizing those decisions may
differ from one repository to another. However, it is recommended that
when making this decision, implementers consider what is necessary to
rebuild the objects from the files stored.

#### Parsability

One goal of the OCFL is to ensure objects remain fixed over time. This
can be difficult as software and infrastructure change, and content is
migrated. To combat this challenge, the OCFL ensures that both humans
and machines can understand the layout and corresponding inventory
regardless of the software or infrastructure used. This allows for
humans to read the layout and corresponding inventory, and understand it
without the use of machines. Additionally, if existing software were to
become obsolete, the OCFL could easily be understood by a light weight
application, even without the full feature repository that might have
been used in the past.

#### Versioning

Another need expressed by the community was the need to update and
change objects, either the content itself or the metadata associated
with the object. The OCFL relies heavily on the prior art in the
[[Moab](https://ocfl.io/1.0/spec/#bib-moab "The Moab Design for Digital Object Versioning")] Design for Digital Object Versioning which
utilizes forward deltas to track the history of the object. Utilizing
this schema allows implementers of the OCFL to easily recreate past
versions of an OCFL object. Like with objects, the OCFL remains silent
on when versioning should occur recognizing this may differ from
implementation to implementation.

#### Robustness

The OCFL also fills the need for robustness against errors, corruption,
and migration. The versioning schema ensures an OCFL object is robust
enough to allow for the discovery of human errors. The fixity checking
built into the OCFL via content addressable storage allows implementers
to identify file corruption that might happen outside of normal human
interactions. The OCFL eases content migrations by providing a
technology agnostic method for verifying OCFL objects have remained
fixed.

#### Storage diversity

Finally, the community expressed a need to store content on a wide
variety of storage technologies. With that in mind, the OCFL was written
with an eye toward various storage infrastructures including cloud
object stores.

### Note

This normative specification describes the nature of an OCFL Object (the
"object-at-rest") and the arrangement of OCFL Objects under an OCFL
Storage Root. A set of recommendations for how OCFL Objects should be
acted upon (the "object-in-motion") can be found in the
[[OCFL-Implementation-Notes](https://ocfl.io/1.0/spec/#bib-ocfl-implementation-notes "OCFL Implementation Notes")]. The OCFL editorial group recommends reading both
the specification and the implementation notes in order to understand
the full scope of the OCFL.

This specification is designed to operate on storage systems that employ
a hierarchical metaphor for presenting data to users. On traditional
disk-based storage this may take the form of files and directories, and
this is the terminology we use in this specification since it is widely
known. However, it may equally apply to object stores, where namespaces,
containers, and objects present a similar organization hierarchy to
users.

## Table of Contents

1. [Conformance](#1-conformance)
2. [Terminology](#2-terminology)
3. [OCFL Object](#3-ocfl-object)
    1. [Object Structure](#31-object-structure)
    2. [Object Conformance Declaration](#32-object-conformance-declaration)
    3. [Version Directories](#33-version-directories)
        1. [Content Directory](#331-content-directory)
    4. [Digest Algorithms](#34-digest-algorithms)
    5. [Inventory](#35-inventory)
        1. [Basic Structure](#351-basic-structure)
        2. [Manifest](#352-manifest)
        3. [Versions](#353-versions)
            1. [Version State](#3531-version-state)
        4. [Fixity](#354-fixity)
    6. [Inventory Digest](#36-inventory-digest)
    7. [Version Inventory and Inventory Digest](#37-version-inventory-and-inventory-digest)
    8. [Logs Directory](#38-logs-directory)
    9. [Object Extensions](#39-object-extensions)
4. [OCFL Storage Root](#4-ocfl-storage-root)
    1. [Root Structure](#41-root-structure)
    2. [Root Conformance Declaration](#42-root-conformance-declaration)
    3. [Storage Hierarchies](#43-storage-hierarchies)
    4. [Storage Root Extensions](#44-storage-root-extensions)
    5. [Filesystem Features](#45-filesystem-features)
5. [Examples](#5-examples)
    1. [Minimal OCFL Object](#51-minimal-ocfl-object)
    2. [Versioned OCFL Object](#52-versioned-ocfl-object)
    3. [Different Logical and Content Paths in an OCFL Object](#53-different-logical-and-content-paths-in-an-ocfl-object)
    4. [BagIt in an OCFL Object](#54-bagit-in-an-ocfl-object)
    5. [Moab in an OCFL Object](#55-moab-in-an-ocfl-object)
    6. [Example Extended OCFL Storage Root](#56-example-extended-ocfl-storage-root)
    7. [Example Extended OCFL Object](#57-example-extended-ocfl-object)
6. [References](#a-references)
    1. [Normative References](#a1-normative-references)
    2. [Informative References](#a2-informative-references)

## 1. Conformance

As well as sections marked as non-normative, all authoring guidelines,
diagrams, examples, and notes in this specification are non-normative.
Everything else in this specification is normative.

The key words *MAY*, *MUST*, *MUST NOT*, *SHOULD*, and *SHOULD NOT* in
this document are to be interpreted as described in [BCP
14](https://tools.ietf.org/html/bcp14)
[[RFC2119](https://ocfl.io/1.0/spec/#bib-rfc2119 "Key words for use in RFCs to Indicate Requirement Levels")]
[[RFC8174](https://ocfl.io/1.0/spec/#bib-rfc8174 "Ambiguity of Uppercase vs Lowercase in RFC 2119 Key Words")] when, and only when, they appear in all capitals,
as shown here.

## 2. Terminology

[Content Path]:
* The file path of a file on disk or in an object store, relative to
 the [OCFL Object
 Root](https://ocfl.io/1.0/spec/#dfn-ocfl-object-root). Content paths are used in the
 [Manifest](https://ocfl.io/1.0/spec/#dfn-manifest) within an
 [Inventory](https://ocfl.io/1.0/spec/#dfn-inventory).

[Digest]:
* An algorithmic characterization of the contents of a file conforming
 to a standard digest algorithm.

[Extension]:
* Extensions are used to collaborate, review, and publish additional
 non-normative functions related to OCFL. Extensions are intended to
 be informational and cite-able, but outside the scope of the normal
 specification process. Existing extensions may be found in the [OCFL
 Extensions repository.](https://ocfl.github.io/extensions/)

[Inventory]:
* A file, expressed in JSON, that tracks the history and current state
 of an OCFL Object.

[Logical Path]:
* A path that represents a file's location in the [logical
 state](https://ocfl.io/1.0/spec/#dfn-logical-state) of an object. Logical paths are used
 in conjunction with a digest to represent the file name and path for
 a given bitstream at a given version.

[Logical State]:
* A grouping of logical paths tied to their corresponding bitstreams
 that reflect the state of the object content for a given version.

[Logs Directory]:
* A directory for storing information about the content (e.g., actions
 performed) that is not part of the content itself.

[Manifest]
* A section of the
 [Inventory](https://ocfl.io/1.0/spec/#dfn-inventory) listing all files and their digests
 within an OCFL Object.

[OCFL Object]:
* A group of one or more content files and administrative information,
 that together have a unique identifier. The object may contain a
 sequence of versions of the files that represent the evolution of
 the object's contents.

[OCFL Object Root]:
* The base directory of an [OCFL
 Object](https://ocfl.io/1.0/spec/#dfn-ocfl-object), identified by a
 [[NAMASTE](https://ocfl.io/1.0/spec/#bib-namaste "Directory Description with Namaste Tags")] file "0=ocfl_object_1.0".

[OCFL Storage Root]:
* A base directory used to store OCFL Objects, identified by a
 [[NAMASTE](https://ocfl.io/1.0/spec/#bib-namaste "Directory Description with Namaste Tags")] file "0=ocfl_1.0".

[OCFL Version]:
* The state of an [OCFL
 Object](https://ocfl.io/1.0/spec/#dfn-ocfl-object)'s content which is constructed using
 the incremental changes recorded in the sequence of corresponding
 and prior version directories.

[Registered Extension Name]:
* The registered name of an extension is the name provided in the
 *Extension Name* property of the extension's definition.

## 3. OCFL Object

An OCFL Object is a group of one or more content files and
administrative information, that are together identified by a URI. The
object may contain a sequence of versions of the files that represent
the evolution of the object's contents.

A file is defined as a content bitstream that can be stored and
transmitted. Directories (also called "folders") allow for the
organization of files into tree-like hierarchies. The content of an OCFL
Object is the files and the directories they are organized in that are
stored *within* the hierarchy layout described in this specification.

An OCFL Object includes administrative information that identifies a
directory as an OCFL Object, and also provides a means of tracking
changes to the contents of the object over time.

An OCFL Object is therefore:

1. A conceptual gathering of all files (data and metadata), the
 directories they are organized in, and their changes over time which
 together form the digital representation of an entity that need to
 be managed, in preservation terms, as a single coherent whole (i.e.,
 content); and
2. A file and directory layout and administrative information on a
 storage medium that provides a defined structure for the storage of
 this content, and through which these files and their changes may be
 understood (i.e., structure).

A key goal of the OCFL is the rebuildability of a repository from an
OCFL Storage Root without additional information resources.
Consequently, a key implementation consideration should be to ensure
that OCFL Objects contain all the data and metadata required to achieve
this. With reference to the
[[OAIS](https://ocfl.io/1.0/spec/#bib-oais " Reference Model for an Open Archival Information System (OAIS), Issue 2")] model, this would include all the descriptive,
administrative, structural, representation and preservation metadata
relevant to the object.

A central feature of the OCFL specification is support for versioning.
This recognizes that digital objects will change over time, through new
requirements, fixes, updates, or format shifts. The specification takes
no position on what constitutes a version or a versionable action, but
it is recommended that implementers have a clear position on this within
their local storage policies.

### 3.1 Object Structure

The OCFL Object structure organizes content files and administrative
information in order to support content storage and object validation.
The structure for an object with one version is shown in the following
figure:

```
[object_root]
 ├── 0=ocfl_object_1.0
 ├── inventory.json
 ├── inventory.json.sha512
 └── v1
 ├── inventory.json
 ├── inventory.json.sha512
 └── content
 └── ... content files ...
```

The [OCFL Object
Root](https://ocfl.io/1.0/spec/#dfn-ocfl-object-root) [*MUST NOT*] contain files or
directories other than those specified in the following sections.

### 3.2 Object Conformance Declaration

The version declaration [*MUST*] be formatted according to the
[[NAMASTE](https://ocfl.io/1.0/spec/#bib-namaste "Directory Description with Namaste Tags")] specification. It [*MUST*] be a file in the
base directory of the OCFL Object Root giving the OCFL version in the
filename. The filename [*MUST*] conform to the pattern
`T=dvalue`, where `T` [*MUST*] be 0, and `dvalue` [*MUST*]
be `ocfl_object_`, followed by the OCFL specification version number.
The text contents of the file [*MUST*] be the same as `dvalue`,
followed by a newline (`\n`).

### 3.3 Version Directories

OCFL Object content [*MUST*] be stored as a sequence of one or
more versions. Each object version is stored in a version directory
under the object root. The sequence of version numbers is the sequence
of positive, base-ten integers: 1, 2, 3, etc., and the version directory
name is constructed by adding the prefix `v`. The version number
sequence [*MUST*] start at 1 and [*MUST*] be continuous
without missing integers.

Implementations [*SHOULD*] use version directory names
constructed without zero-padding the version number, ie. `v1`, `v2`,
`v3`, etc..

For compatibility with existing filesystem conventions, implementations
*MAY* use zero-padded version directory numbers, with the following
restriction: If zero-padded version directory numbers are used then they
[*MUST*] start with the prefix `v` and then a zero. For example,
in an implementation that uses five digits for version directory names
then `v00001` to `v09999` are allowed, `v10000` is not allowed.

The first version of an object defines the naming convention for all
version directories for the object. All version directories of an object
[*MUST*] use the same naming convention: either a non-padded
version directory number, or a zero-padded version directory number of
consistent length. Operations that add a new version to an object
[*MUST*] follow the version directory naming convention
established by earlier versions. In all cases, references to files
inside version directories from inventory files [*MUST*] use the
actual version directory names.

There [*MUST*] be no other files as children of a version
directory, other than an [inventory
file](https://ocfl.io/1.0/spec/#inventory) and a [inventory
digest](https://ocfl.io/1.0/spec/#inventory-digest). The version
directory [*SHOULD NOT*] contain any directories other than the
designated content sub-directory. Once created, the contents of a
version directory are expected to be immutable.

#### 3.3.1 Content Directory

Version directories [*MUST*] contain a designated content
sub-directory if the version contains files to be preserved, and
[*SHOULD NOT*] contain this sub-directory otherwise. The name of
this designated sub-directory *MAY* be defined in the [inventory
file](https://ocfl.io/1.0/spec/#inventory) using the key
`contentDirectory` with the value being the chosen sub-directory name as
a string, relative to the version directory. The `contentDirectory`
value [*MUST NOT*] contain the forward slash (`/`) path separator
and [*MUST NOT*] be either one or two periods (`.` or `..`). If
the key `contentDirectory` is set, it [*MUST*] be set in the
first version of the object and [*MUST NOT*] change between
versions of the same object.

If the key `contentDirectory` is not present in the [inventory
file](https://ocfl.io/1.0/spec/#inventory) then the name of the
designated content sub-directory [*MUST*] be `content`.
OCFL-compliant tools (including any validators) [*MUST*] ignore
all directories in the object version directory except for the
designated content directory.

Every file within a version's content directory [*MUST*] be
referenced in the [manifest](https://ocfl.io/1.0/spec/#manifest) section
of the inventory. There [*MUST NOT*] be empty directories within
a version's content directory. A directory that would otherwise be
empty *MAY* be maintained by creating a file within it named according
to local conventions, for example by making an empty `.keep` file.

### 3.4 Digest Algorithms

Digests play two roles in an OCFL Object. The first is that digests
allow for content-addressable reference to files within the OCFL Object.
That is, the connection between a file's [content
path](https://ocfl.io/1.0/spec/#dfn-content-path) on physical storage and its [logical
path](https://ocfl.io/1.0/spec/#dfn-logical-path) in a version of the object's content is
made with a digest of its contents, rather than its filename. This use
of the content digest facilitates de-duplication of files with the same
content within an object, such as files that are unchanged from one
version to the next. The second role that digests play is provide for
fixity checks to determine whether a file has become corrupt, through
hardware degradation or accident for example.

For content-addressing, OCFL Objects [*MUST*] use either `sha512`
or `sha256`, and [*SHOULD*] use `sha512`. The choice of the
`sha512` digest algorithm as default recognizes that it has no known
collision vulnerabilities and multiple implementations are available.

For storage of additional fixity values, or to support legacy content
migration, implementers [*MUST*] choose from the following
controlled vocabulary of digest algorithms, or from a list of additional
algorithms given in the
[[Digest-Algorithms-Extension](https://ocfl.io/1.0/spec/#bib-digest-algorithms-extension "OCFL Community Extension 0001: Digest Algorithms")]. OCFL clients [*MUST*] support all fixity
algorithms given in the table below, and *MAY* support additional
algorithms from the extensions. Optional fixity algorithms that are not
supported by a client [*MUST*] be ignored by that client.

 Digest Algorithm Name Note
 ----------------------- ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
 `md5` Insecure. Use only for legacy fixity values. MD5 algorithm and hex encoding defined by [[RFC1321](https://ocfl.io/1.0/spec/#bib-rfc1321 "The MD5 Message-Digest Algorithm")]. For example, the `md5` digest of a zero-length bitstream is `d41d8cd98f00b204e9800998ecf8427e`.
 `sha1` Insecure. Use only for legacy fixity values. SHA-1 algorithm defined by [[FIPS-180-4](https://ocfl.io/1.0/spec/#bib-fips-180-4 "FIPS PUB 180-4: Secure Hash Standard (SHS)")] and [*MUST*] be encoded using hex (base16) encoding [[RFC4648](https://ocfl.io/1.0/spec/#bib-rfc4648 "The Base16, Base32, and Base64 Data Encodings")]. For example, the `sha1` digest of a zero-length bitstream is `da39a3ee5e6b4b0d3255bfef95601890afd80709`.
 `sha256` Non-truncated form only; note performance implications. SHA-256 algorithm defined by [[FIPS-180-4](https://ocfl.io/1.0/spec/#bib-fips-180-4 "FIPS PUB 180-4: Secure Hash Standard (SHS)")] and [*MUST*] be encoded using hex (base16) encoding [[RFC4648](https://ocfl.io/1.0/spec/#bib-rfc4648 "The Base16, Base32, and Base64 Data Encodings")]. For example, the `sha256` digest of a zero-length bitstream starts `e3b0c44298fc1c149afbf4c8996fb92427ae41e4...` (64 hex digits long).
 `sha512` Default choice. Non-truncated form only. SHA-512 algorithm defined by [[FIPS-180-4](https://ocfl.io/1.0/spec/#bib-fips-180-4 "FIPS PUB 180-4: Secure Hash Standard (SHS)")] and [*MUST*] be encoded using hex (base16) encoding [[RFC4648](https://ocfl.io/1.0/spec/#bib-rfc4648 "The Base16, Base32, and Base64 Data Encodings")]. For example, the `sha512` digest of a zero-length bitstream starts `cf83e1357eefb8bdf1542850d66d8007d620e405...` (128 hex digits long).
 `blake2b-512` Full-length form only, using the 2B variant (64 bit) as defined by [[RFC7693](https://ocfl.io/1.0/spec/#bib-rfc7693 "The BLAKE2 Cryptographic Hash and Message Authentication Code (MAC)")]. [*MUST*] be encoded using hex (base16) encoding [[RFC4648](https://ocfl.io/1.0/spec/#bib-rfc4648 "The Base16, Base32, and Base64 Data Encodings")]. For example, the `blake2b-512` digest of a zero-length bitstream starts `786a02f742015903c6c6fd852552d272912f4740...` (128 hex digits long).

An OCFL Inventory *MAY* contain a fixity section that can store one or
more blocks containing fixity values using multiple digest algorithms.
See the [section on fixity](https://ocfl.io/1.0/spec/#fixity) below for
further details.

> Non-normative note: Implementers may also store copies of their file
> digests in a system external to their OCFL Object stores at the point
> of ingest, to further safeguard against the possibility of malicious
> manipulation of file contents and digests.
>
> Implementers should be aware that base16 digests are case insensitive.
> Different tools will generate digests in uppercase or lowercase, and
> this may lead to case differences between references to a digest and
> the digest itself within the inventory. If string-based methods are
> used to work with digests and inventories (as is the case in most
> common JSON libraries) then extra care must be taken to ensure
> case-insensitive comparisons are being made.

### 3.5 Inventory

An OCFL Object Inventory [*MUST*] follow the
[[JSON](https://ocfl.io/1.0/spec/#bib-json "The JavaScript Object Notation (JSON) Data Interchange Format")] structure described in this section and
[*MUST*] be named `inventory.json`. The order of entries in both
the
[[JSON](https://ocfl.io/1.0/spec/#bib-json "The JavaScript Object Notation (JSON) Data Interchange Format")] objects and arrays used in inventory files has no
significance. An OCFL Object Inventory [*MUST NOT*] contain any
keys not described in this specification.

The forward slash (/) path separator [*MUST*] be used in content
paths in the [manifest](https://ocfl.io/1.0/spec/#manifest) and
[fixity](https://ocfl.io/1.0/spec/#fixity) blocks within the inventory.
Implementations that target systems using other separators will need to
translate paths appropriately.

> Non-normative note: A
> [[JSON-Schema](https://ocfl.io/1.0/spec/#bib-json-schema "JSON Schema Validation: A Vocabulary for Structural Validation of JSON")] for validating OCFL Object Inventory files is
> provided at
> [inventory_schema.json](https://ocfl.io/1.0/spec/inventory_schema.json).

#### 3.5.1 Basic Structure

Every OCFL inventory [*MUST*] include the following keys:

`id`
* A unique identifier for the OCFL Object. This [*MUST*] be
 unique in the local context, and [*SHOULD*] be a URI
 [[RFC3986](https://ocfl.io/1.0/spec/#bib-rfc3986 "Uniform Resource Identifier (URI): Generic Syntax")]. There is no expectation that a URI used is
 resolvable. For example, URNs
 [[RFC8141](https://ocfl.io/1.0/spec/#bib-rfc8141 "Uniform Resource Names (URNs)")] *MAY* be used.

`type`
* A type for the inventory JSON object that also serves to document
 the OCFL specification version that the inventory complies with. In
 the object root inventory this [*MUST*] be the URI of the
 inventory section of the specification version matching the object
 conformance declaration. For the current specification version the
 value is `https://ocfl.io/1.0/spec/#inventory`.

`digestAlgorithm`
* The digest algorithm used for calculating digests for
 content-addressing within the OCFL Object and for the [Inventory
 Digest](https://ocfl.io/1.0/spec/#inventory-digest). This
 [*MUST*] be the algorithm used in the `manifest` and `state`
 blocks, see the [section on
 Digests](https://ocfl.io/1.0/spec/#digests) for more information
 about algorithms.

`head`
* The version directory name of the most recent version of the object.
 This [*MUST*] be the version directory name with the highest
 version number.

There *MAY* be the following key:

`contentDirectory`
* The name of the designated content directory within the version
 directories. If not specified then the content directory name is
 `content`.

In addition to these keys, there [*MUST*] be two other blocks
present, `manifest` and `versions`, which are discussed in the next two
sections.

#### 3.5.2 Manifest

The value of the `manifest` key is a JSON object, with keys
corresponding to the digests of every content file in all versions of
the [OCFL
Object](https://ocfl.io/1.0/spec/#dfn-ocfl-object). The value for each key [*MUST*] be
an array containing the [content
path](https://ocfl.io/1.0/spec/#dfn-content-path)s of files in the OCFL Object that have
content with the given digest. As JSON keys are case sensitive, while
digests may not be, there is an additional requirement that each digest
value [*MUST*] occur only once in the manifest regardless of
case. Content paths within a manifest block [*MUST*] be relative
to the [OCFL Object
Root](https://ocfl.io/1.0/spec/#dfn-ocfl-object-root). The following restrictions avoid
ambiguity and provide path safety for clients processing the `manifest`.

* The content path [*MUST*] be interpreted as a set of one or
 more path elements joined by a `/` path separator.
* Path elements [*MUST NOT*] be `.`, `..`, or empty (`//`).
* A content path [*MUST NOT*] begin or end with a forward slash
 (`/`).
* Within an inventory, content paths [*MUST*] be unique and
 non-conflicting, so the content path for a file cannot appear as the
 initial part of another content path.

> Non-normative note: If only one file is stored in the OCFL Object for
> each digest, fully de-duplicating the content, then there will be only
> one [content
> path](https://ocfl.io/1.0/spec/#dfn-content-path) for each digest. There may, however, be
> multiple logical paths for a given digest if the content was not
> entirely de-duplicated when constructing the OCFL Object.
>
> An example manifest object for three content paths, all in version 1,
> is shown below:
>
> ```
> "manifest": 
> ```

#### 3.5.3 Versions

An OCFL Object Inventory [*MUST*] include a block for storing
versions. This block [*MUST*] have the key of `versions` within
the inventory, and it [*MUST*] be a JSON object. The keys of this
object [*MUST*] correspond to the names of the [version
directories](https://ocfl.io/1.0/spec/#version-directories) used. Each
value [*MUST*] be another JSON object that characterizes the
version, as described in the [§ 3.5.3.1
Version](https://ocfl.io/1.0/spec/#version) section.

##### 3.5.3.1 Version State

A JSON object to describe one [OCFL
Version](https://ocfl.io/1.0/spec/#dfn-ocfl-version), which [*MUST*] include the
following keys:

`created`
* The value of this key is the datetime of creation of this version.
 It [*MUST*] be expressed in the Internet Date/Time Format
 defined by
 [[RFC3339](https://ocfl.io/1.0/spec/#bib-rfc3339 "Date and Time on the Internet: Timestamps")]. This format requires the inclusion of a
 timezone value or `Z` for UTC, and that the time component be
 granular to the second level (with optional fractional seconds).

`state`

* The value of this key is a JSON object, containing a list of keys
 and values corresponding to the [logical
 state](https://ocfl.io/1.0/spec/#dfn-logical-state) of the object at that version. The
 keys of this JSON object are digest values, each of which
 [*MUST*] exactly match a digest value key in the [manifest of
 the inventory](https://ocfl.io/1.0/spec/#manifest). The value for
 each key is an array containing [logical
 path](https://ocfl.io/1.0/spec/#dfn-logical-path) names of files in the OCFL Object
 state that have content with the given digest.

 [Logical
 path](https://ocfl.io/1.0/spec/#dfn-logical-path)s present the structure of an OCFL
 Object at a given version. This is given as an array of values, with
 the following restrictions to provide for path safety in the common
 case of the logical path value representing a file path.

 * The logical path [*MUST*] be interpreted as a set of one or
 more path elements joined by a `/` path separator.
 * Path elements [*MUST NOT*] be `.`, `..`, or empty (`//`).
 * A logical path [*MUST NOT*] begin or end with a forward
 slash (`/`).
 * Within a version, logical paths [*MUST*] be unique and
 non-conflicting, so the logical path for a file cannot appear as
 the initial part of another logical path.

 > Non-normative note: The [logical
 > state](https://ocfl.io/1.0/spec/#dfn-logical-state) of the object uses
 > content-addressing to map logical paths to their bitstreams, as
 > expressed in the manifest section of the inventory. Notably, the
 > version state provides de-duplication of content within the OCFL
 > Object by mapping multiple logical paths with the same content to
 > the same digest in the manifest. See
 > [[OCFL-Implementation-Notes](https://ocfl.io/1.0/spec/#bib-ocfl-implementation-notes "OCFL Implementation Notes")].
 >
 > An example state block is shown below:
 >
 > ```
 > "state": 
 > ```
 >
 > This state block describes an object with 3 files, two of which
 > have the same content (`empty.txt` and `empty2.txt`), and one of
 > which is in a sub-directory (`bar.xml`). The [logical
 > state](https://ocfl.io/1.0/spec/#dfn-logical-state) shown as a tree is thus:
 >
 > ```
 > ├── empty.txt
 > ├── empty2.txt
 > └── foo
 > └── bar.xml
 > ```

The JSON object describing an [OCFL
Version](https://ocfl.io/1.0/spec/#dfn-ocfl-version), [*SHOULD*] include the following
keys:

`message`
* The value of this key is freeform text, used to record the rationale
 for creating this version. It [*MUST*] be a JSON string.

`user`
* The value of this key is a JSON object intended to identify the user
 or agent that created the current [OCFL
 Version](https://ocfl.io/1.0/spec/#dfn-ocfl-version). The value of the `user` key
 [*MUST*] contain a user name key, `name` and
 [*SHOULD*] contain an address key, `address`. The `name`
 value is any readable name of the user, e.g., a proper name, user
 ID, agent ID. The `address` value [*SHOULD*] be a URI: either
 a mailto URI
 [[RFC6068](https://ocfl.io/1.0/spec/#bib-rfc6068 "The 'mailto' URI Scheme")] with the e-mail address of the user or a URL
 to a personal identifier, e.g., an ORCID iD.

#### 3.5.4 Fixity

An OCFL Object inventory *MAY* include a block for storing additional
fixity information to supplement the complete set of digests in the
[Manifest](https://ocfl.io/1.0/spec/#manifest), for example to support
legacy digests from a content migration. This block [*MUST*] have
the key of `fixity` within the inventory.

The `fixity` block [*MUST*] contain keys corresponding to the
controlled vocabulary given in the [digest
algorithms](https://ocfl.io/1.0/spec/#digest-algorithms) listed in the
[Digests](https://ocfl.io/1.0/spec/#digests) section, or in a table
given in an
[Extension](https://ocfl.io/1.0/spec/#dfn-extension). The value of the fixity block for a
particular digest algorithm [*MUST*] follow the structure of the
[`manifest`](https://ocfl.io/1.0/spec/#manifest) block; that is, a key
corresponding to the digest value, and an array of [content
path](https://ocfl.io/1.0/spec/#dfn-content-path)s. The fixity block for any digest
algorithm *MAY* include digest values for any subset of content paths in
the object. Where included, the digest values given [*MUST*]
match the digests of the files at the corresponding content paths. As
JSON keys are case sensitive, while digests may not be, there is an
additional requirement that each digest value [*MUST*] occur only
once in the fixity block for any digest algorithm, regardless of case.
There is no requirement that all content files have a value in the
fixity block, or that fixity values provided in one version are carried
forward to later versions.

> An example fixity block with `md5` and `sha1` digests is shown below.
> In this case the `md5` digest values are provided only for version 1
> content paths.
>
> ```
> "fixity": ,
> "sha1": 
> }
> ```

### 3.6 Inventory Digest

Every occurrence of an inventory file [*MUST*] have an
accompanying sidecar file named `inventory.json.ALGORITHM` stating its
digest, where `ALGORITHM` is the chosen digest algorithm for the object.
The ALGORITHM [*MUST*] match the value given for the
`digestAlgorithm` key in the inventory. An example might be
`inventory.json.sha512`.

The digest sidecar file [*MUST*] contain the digest of the
inventory file. This [*MUST*] follow the format:

```
DIGEST inventory.json
```

One or more whitespace characters (spaces or tabs) must separate DIGEST
from the string `inventory.json`; that is, the name of the inventory
file in the same directory.

The digest of the inventory [*MUST*] be computed only after all
changes to the inventory have been made, and thus writing the digest
sidecar file is the last step in the versioning process.

### 3.7 Version Inventory and Inventory Digest

Every OCFL Object [*MUST*] have an inventory file within the OCFL
Object Root, corresponding to the state of the OCFL Object at the
current version. Additionally, every version directory [*SHOULD*]
include an inventory file that is an
[Inventory](https://ocfl.io/1.0/spec/#inventory) of all content for
versions up to and including that particular version. Where an OCFL
Object contains `inventory.json` in version directories, the inventory
file in the OCFL Object Root [*MUST*] be the same as the file in
the most recent version. See also requirements for the corresponding
[Inventory Digest](https://ocfl.io/1.0/spec/#inventory-digest).

In the case that prior version directories include an inventory file
there will be multiple inventory files describing prior versions within
the OCFL Object. Each version block in each prior inventory file
[*MUST*] represent the same object state as the corresponding
version block in the current inventory file. Additionally, the values of
the `created`, `message` and `user` keys in each version block in each
prior inventory file [*SHOULD*] have the same values as the
corresponding keys in the corresponding version block in the current
inventory file.

> Non-normative note: Storing an inventory for every version provides
> redundancy for this critical information in a way that is compatible
> with storage strategies that have immutable version directories.

### 3.8 Logs Directory

The base directory of an OCFL Object *MAY* contain a directory named
`logs`, which *MAY* be empty. Implementers [*SHOULD*] use this
for storing files that contain a record of actions taken on the object.
Since these logs may be subject to local standards requirements, the
format of these logs is considered out-of-scope for the OCFL Object.
Clients operating on the object *MAY* log actions here that are not
otherwise captured.

> Non-normative note: The purpose of the logs directory is to provide
> implementers with a location for storing local information about
> actions to the OCFL Object's content that is not part of the content
> itself.
>
> As an example, implementers may have different local requirements to
> store audit information for their content. Some may wish to store a
> log entry indicating that an audit was conducted, and nothing was
> wrong, while others may wish to only store a log entry if an
> intervention was required.

### 3.9 Object Extensions

The base directory of an OCFL Object *MAY* contain a directory named
`extensions` for the purposes of extending the functionality of an OCFL
Object. The `extensions` directory [*MUST NOT*] contain any
files, and no sub-directories other than extension sub-directories.
Extension sub-directories [*SHOULD*] be named according to a
[registered extension
name](https://ocfl.io/1.0/spec/#dfn-registered-extension-name). The specific structure and function of
the extension, as well as a declaration of the registered extension name
[*MUST*] be defined in one of the following locations:

* The [OCFL Extensions repository](https://ocfl.github.io/extensions/)
* The Storage Root, as a plain text document directly in the Storage
 Root

> Non-normative note: Extension sub-directories should use the same name
> as a registered extension in order to both avoid the possiblity of an
> extension sub-directory colliding with the name of another registered
> extension as well as to facilitate the recognition of extensions by
> OCFL clients.

## 4. OCFL Storage Root

An [OCFL Storage
Root](https://ocfl.io/1.0/spec/#dfn-ocfl-storage-root) is the base directory of an OCFL storage
layout.

### 4.1 Root Structure

An OCFL Storage Root [*MUST*] contain a [Root Conformance
Declaration](https://ocfl.io/1.0/spec/#root-conformance-declaration)
identifying it as such.

An OCFL Storage Root *MAY* contain other files as direct children. These
might include a human-readable copy of the OCFL specification to make
the storage root self-documenting, or files used by [storage root
extensions](https://ocfl.io/1.0/spec/#storage-root-extensions). An OCFL
validator [*MUST*] ignore any files in the storage root it does
not understand.

An OCFL Storage Root [*MUST NOT*] contain directories or
sub-directories other than as a directory hierarchy used to store OCFL
Objects or for [storage root
extensions](https://ocfl.io/1.0/spec/#storage-root-extensions). The
directory hierarchy used to store OCFL Objects [*MUST NOT*]
contain files that are not part of an OCFL Object. Empty directories
[*MUST NOT*] appear under a storage root.

An OCFL Storage Root *MAY* contain a JSON file named `ocfl_layout.json`
to describe the arrangement of directories and OCFL objects under the
storage root. If present, this JSON document [*MUST*] include the
following two keys in the root JSON object:

* `extension` - An extension name that identifies an arrangement of
 directories and OCFL objects under the storage root, i.e. how OCFL
 object identifiers are mapped to directory hierarchies. The value of
 the `extension` key [*MUST*] be the [registered extension
 name](https://ocfl.io/1.0/spec/#dfn-registered-extension-name) for the extension defining the
 arrangement under the storage root.
* `description` - A human readable description of the arrangement of
 directories and OCFL objects under the storage root.

Although implementations may require multiple OCFL Storage Roots---that
is, several logical or physical volumes, or multiple "buckets" in an
object store---each OCFL Storage Root [*MUST*] be independent.

The following example OCFL Storage Root represents the minimal set of
files and folders:

```
[storage_root]
 ├── 0=ocfl_1.0
 ├── ocfl_1.0.txt (human-readable text of the OCFL specification; optional)
 └── ocfl_layout.json (description of storage hierarchy layout; optional)
```

### 4.2 Root Conformance Declaration

The OCFL version declaration [*MUST*] be formatted according to
the
[[NAMASTE](https://ocfl.io/1.0/spec/#bib-namaste "Directory Description with Namaste Tags")] specification. It [*MUST*] be a file in the
base directory of the [OCFL Storage
Root](https://ocfl.io/1.0/spec/#dfn-ocfl-storage-root) giving the OCFL version in the filename.
The filename [*MUST*] conform to the pattern `T=dvalue`, where
`T` [*MUST*] be 0, and `dvalue` [*MUST*] be `ocfl_`,
followed by the OCFL specification version number. The text contents of
the file [*MUST*] be the same as `dvalue`, followed by a newline
(`\n`).

Root conformance indicates that the OCFL Storage Root conforms to this
section (i.e. the OCFL Storage Root section) of the specification. OCFL
Objects within the OCFL Storage Root also include a conformance
declaration which [*MUST*] indicate OCFL Object conformance to
the same or earlier version of the specification.

### 4.3 Storage Hierarchies

[OCFL Object
Root](https://ocfl.io/1.0/spec/#dfn-ocfl-object-root)s [*MUST*] be stored either as the
terminal resource at the end of a directory storage hierarchy or as
direct children of a containing [OCFL Storage
Root](https://ocfl.io/1.0/spec/#dfn-ocfl-storage-root).

A common practice is to use a unique identifier scheme to compose this
storage hierarchy, typically arranged according to some form of the
[[PairTree](https://ocfl.io/1.0/spec/#bib-pairtree "Pairtrees for Object Storage")] specification. Irrespective of the pattern chosen
for the storage hierarchies, the following restrictions apply:

1. There [*MUST*] be a deterministic mapping from an object
 identifier to a unique storage path
2. Storage hierarchies [*MUST NOT*] include files within
 intermediate directories
3. Storage hierarchies [*MUST*] be terminated by OCFL Object
 Roots
4. Storage hierarchies within the same OCFL Storage Root
 [*SHOULD*] use just one layout pattern
5. Storage hierarchies within the same OCFL Storage Root
 [*SHOULD*] consistently use either a directory hierarchy of
 OCFL Objects or top-level OCFL Objects

### 4.4 Storage Root Extensions

The behavior of the storage root may be extended to support features
from other specifications.

The base directory of an OCFL Storage Root *MAY* contain a directory
named `extensions` for the purposes of extending the functionality of an
OCFL Storage Root. The storage root `extensions` directory
[*MUST*] conform to the same guidelines and limitations as those
defined for [object
extensions](https://ocfl.io/1.0/spec/#object-extensions).

> Non-normative note: Storage extensions can be used to support
> additional features, such as providing the storage hierarchy
> disposition when pairtree is in use, or additional human-readable text
> about the nature of the storage root.

### 4.5 Filesystem Features

In order to maximize the compatibility of the OCFL with different
filesystems, and thus improve the portability of OCFL Objects between
different systems, some restrictions on the use of certain filesystem
features are necessary. If the preservation of non-OCFL-compliant
features is required then the content [*MUST*] be wrapped in a
suitable disk or filesystem image format which OCFL can treat as a
regular file.

1. Filesystem metadata (e.g. permissions, access, and creation times)
 are not considered portable between filesystems or preservable
 through file transfer operations. These attributes also cannot be
 validated in terms of fixity in a consistent manner. As such, the
 OCFL does not support the portability of these attributes.
2. Hard and soft (symbolic) links are not portable and [*MUST
 NOT*] be used within OCFL Storage hierachies. A common use
 case for links is storage deduplication. OCFL inventories provide a
 portable method of achieving the same effect by using digests to
 address content.
3. File paths and filenames in the OCFL are case sensitive. Filesystems
 [*MUST*] preserve the case of OCFL filepaths and filenames.
4. Transparent filesystem features such as compression and encryption
 should be effectively invisible to OCFL operations. Consequently,
 they should not be expected to be portable.

## 5. Examples

*This section is non-normative.*

### 5.1 Minimal OCFL Object

The following example OCFL Object has content that is a single file
(`file.txt`), and just one version (`v1`):

```
[object root]
 ├── 0=ocfl_object_1.0
 ├── inventory.json
 ├── inventory.json.sha512
 └── v1
 ├── inventory.json
 ├── inventory.json.sha512
 └── content
 └── file.txt
```

The inventory for this OCFL Object, the same both at the top-level and
in the `v1` directory, might be:

``` 
,
 "type": "https://ocfl.io/1.0/spec/#inventory",
 "versions": ,
 "user": 
 }
 }
}
```

### 5.2 Versioned OCFL Object

The following example OCFL Object has three versions:

```
[object root]
 ├── 0=ocfl_object_1.0
 ├── inventory.json
 ├── inventory.json.sha512
 ├── v1
 │   ├── inventory.json
 │   ├── inventory.json.sha512
 │   └── content
 │ ├── empty.txt
 │ ├── foo
 │ │   └── bar.xml
 │ └── image.tiff
 ├── v2
 │   ├── inventory.json
 │   ├── inventory.json.sha512
 │   └── content
 │ └── foo
 │     └── bar.xml
 └── v3
 ├── inventory.json
 └── inventory.json.sha512
```

In `v1` there are three files, `empty.txt`, `foo/bar.xml`, and
`image.tiff`. In `v2` the content of `foo/bar.xml` is changed,
`empty2.txt` is added with the same content as `empty.txt`, and
`image.tiff` is removed. In `v3` the file `empty.txt` is removed, and
`image.tiff` is reinstated. As a result of forward-delta versioning, the
object tree above shows only new content added in each version. The
inventory shown below details the other changes, includes additional
fixity information using `md5` and `sha1` digest algorithms, and minimal
metadata for each version.

``` 
,
 "sha1": 
 },
 "head": "v3",
 "id": "ark:/12345/bcd987",
 "manifest": ,
 "type": "https://ocfl.io/1.0/spec/#inventory",
 "versions": ,
 "user": 
 },
 "v2": ,
 "user": 
 },
 "v3": ,
 "user": 
 }
 }
}
```

### 5.3 Different Logical and Content Paths in an OCFL Object

The following example OCFL Object inventory shows how content paths may
differ from logical paths. The example object has just one version,
`v1`, which has two files with logical paths `a file.wxy` and
`another file.xyz` as shown in the `state` block. The corresponding
content paths are `v1/content/3bacb119a98a15c5` and
`v1/content/9f2bab8ef869947d` respectively, as shown in the `manifest`.
Except for location within the appropriate version directory,
`v1/content` in this example, the OCFL specification does not constrain
the choice of content paths used when creating or updating an OCFL
object. The choice might depend on particular limitations of, or
optimizations for, the target storage system, or on portability
considerations. Any compliant implementation will be able to recover
version state with the original logical paths.

``` 
,
 "type": "https://ocfl.io/1.0/spec/#inventory",
 "versions": 
 }
 }
}
```

### 5.4 BagIt in an OCFL Object

[[BagIt](https://ocfl.io/1.0/spec/#bib-bagit "The BagIt File Packaging Format (V1.0)")] is a common file packaging specification, but
unlike the OCFL it does not provide a mechanism for content versioning.
Using the OCFL it is possible to store a BagIt structure with content
versioning, such that when the object state is resolved, it creates a
valid BagIt 'bag'. This example will illustrate one way this can be
accomplished, using the [example of a basic
bag](https://tools.ietf.org/html/draft-kunze-bagit-17#section-4.1) given
in the BagIt specification.

```
[object root]
 ├── 0=ocfl_object_1.0
 ├── inventory.json
 ├── inventory.json.sha512
 └── v1
 ├── inventory.json
 ├── inventory.json.sha512
 └── content
 └── myfirstbag
 ├── bagit.txt
 ├── data
 │   └── 27613-h
 │   └── images
 │   ├── q172.png
 │   └── q172.txt
 └── manifest-md5.txt
```

If, for example, a new directory were added in a subsequent version, the
OCFL Object would look like this:

```
[object root]
 ├── 0=ocfl_object_1.0
 ├── inventory.json
 ├── inventory.json.sha512
 ├── v1
 │ ├── inventory.json
 │ ├── inventory.json.sha512
 │ └── content
 │   └── myfirstbag
 │   ├── bagit.txt
 │   ├── data
 │   │   └── 27613-h
 │   │   └── images
 │   │   ├── q172.png
 │   │   └── q172.txt
 │   └── manifest-md5.txt
 └── v2
 ├── inventory.json
 ├── inventory.json.sha512
 └── content
 └── myfirstbag
 ├── data
 │   └── 27614-h
 │   └── images
 │   ├── q173.png
 │   └── q173.txt
 └── manifest-md5.txt
```

The state of the object at version 2 would be the following BagIt
object:

```
myfirstbag
 ├── bagit.txt
 ├── data
 │   ├── 27613-h
 │   │   └── images
 │   │   ├── q172.png
 │   │   └── q172.txt
 │   └── 27614-h
 │   └── images
 │   ├── q173.png
 │   └── q173.txt
 └── manifest-md5.txt
```

The OCFL Inventory for this object would be as follows:

``` 
,
 "type": "https://ocfl.io/1.0/spec/#inventory",
 "versions": ,
 "user": 
 },
 "v2": ,
 "user": 
 }
 }
}
```

### 5.5 Moab in an OCFL Object

[[Moab](https://ocfl.io/1.0/spec/#bib-moab "The Moab Design for Digital Object Versioning")] is an archive information package format developed
and used by Stanford University. Many of the ideas in Moab have been
refined by the OCFL, and the OCFL is designed to give institutions
currently using Moab an easy path to adoption.

Converting content preserved in a Moab object in a way that does not
compromise existing Moab access patterns whilst allowing for the
eventual use of OCFL-native workflows requires a Moab to OCFL conversion
tool. This tool uses the Moab-versioning gem to extract deltas and
digests of the Moab data directory for each Moab version and translate
those into version state blocks in an OCFL inventory file, which would
be placed in the root directory of the Moab object. The content of the
`data` directory in the Moab version directories (and thus, the
bitstreams that Moab is preserving) is tracked by OCFL, via the
`contentDirectory` value. The contents of the Moab `manifests`
directories are not tracked, as the intention is not to encapsulate a
Moab object inside an OCFL object, but rather to migrate Moab's
preserved bitstreams into an OCFL object without compromising legacy
access patterns.

During the transitionary period the OCFL inventory file exists only in
the root of the Moab object. Once OCFL-native object creation workflows
have been completed, future versions of that object will be fully OCFL
compliant - new versions will no longer have a manifests directory and
will contain an OCFL inventory file. At this stage OCFL tools will be
able to access all versions of the content originally preserved by Moab.

Consider the following sample Moab object:

```
[object root]
 └── bj102hs9687
 ├── v0001
 │   ├── data
 │   │   ├── content
 │   │   │   ├── eric-smith-dissertation-augmented.pdf
 │   │   │   └── eric-smith-dissertation.pdf
 │   │   └── metadata
 │   │   ├── contentMetadata.xml
 │   │   ├── descMetadata.xml
 │   │   ├── identityMetadata.xml
 │   │   ├── provenanceMetadata.xml
 │   │   ├── relationshipMetadata.xml
 │   │   ├── rightsMetadata.xml
 │   │   ├── technicalMetadata.xml
 │   │   └── versionMetadata.xml
 │   └── manifests
 │   ├── fileInventoryDifference.xml
 │   ├── manifestInventory.xml
 │   ├── signatureCatalog.xml
 │   ├── versionAdditions.xml
 │   └── versionInventory.xml
 ├── v0002
 │   ├── data
 │   │   └── metadata
 │  │   ├── contentMetadata.xml
 │  │   ├── embargoMetadata.xml
 │  │   ├── events.xml
 │  │   ├── identityMetadata.xml
 │  │   ├── provenanceMetadata.xml
 │  │   ├── relationshipMetadata.xml
 │  │   ├── rightsMetadata.xml
 │  │   ├── versionMetadata.xml
 │  │   └── workflows.xml
 │  └── manifests
 │  ├── fileInventoryDifference.xml
 │  ├── manifestInventory.xml
 │  ├── signatureCatalog.xml
 │  ├── versionAdditions.xml
 │  └── versionInventory.xml
 └── v0003
 ├── data
 │   └── metadata
 │   ├── contentMetadata.xml
 │   ├── descMetadata.xml
 │   ├── embargoMetadata.xml
 │   ├── events.xml
 │   ├── identityMetadata.xml
 │   ├── provenanceMetadata.xml
 │   ├── rightsMetadata.xml
 │   ├── technicalMetadata.xml
 │   ├── versionMetadata.xml
 │   └── workflows.xml
 └── manifests
 ├── fileInventoryDifference.xml
 ├── manifestInventory.xml
 ├── signatureCatalog.xml
 ├── versionAdditions.xml
 └── versionInventory.xml
```

An OCFL inventory that tracks the `data` directory would include a
manifest comprised as follows. Note the absence of the `manifests`
directory, as we are not encapsulating the Moab object in an OCFL
object, and the presence of `contentDirectory` to specify `data` as the
preserved content directory:

``` 
,
 "type": "https://ocfl.io/1.0/spec/#inventory",
 "versions": 
 },
 "v2": 
 },
 "v3": 
 }
 }
}
```

### 5.6 Example Extended OCFL Storage Root

The following example OCFL Storage Root has an extension containing
custom content. The OCFL Storage Root itself remains valid.

```
[storage root]
 ├── 0=ocfl_1.0
 ├── extensions
 │   └── 0000-example-extension
 │   └── file-example.txt
 ├── ocfl_1.0.txt
 └── ocfl_layout.json
```

### 5.7 Example Extended OCFL Object

The following example OCFL Object has an extension containing custom
content. The OCFL Object itself remains valid.

```
[object root]
 ├── 0=ocfl_object_1.0
 ├── inventory.json
 ├── inventory.json.sha512
 ├── extensions
 │   └── 0000-example-extension
 │   └── file1-draft.txt
 └── v1
 ├── inventory.json
 ├── inventory.json.sha512
 └── content
 └── file.txt
```

`<style>#gridmanContainer blockquote,#gridmanContainer dl,#gridmanContainer dd,#gridmanContainer h1,#gridmanContainer h2,#gridmanContainer h3,#gridmanContainer h4,#gridmanContainer h5,#gridmanContainer h6,#gridmanContainer hr,#gridmanContainer figure,#gridmanContainer p,#gridmanContainer pre#gridmanContainer h1,#gridmanContainer h2,#gridmanContainer h3,#gridmanContainer h4,#gridmanContainer h5,#gridmanContainer h6#gridmanContainer ol,#gridmanContainer ul#gridmanContainer img,#gridmanContainer svg,#gridmanContainer video,#gridmanContainer canvas,#gridmanContainer audio,#gridmanContainer iframe,#gridmanContainer embed,#gridmanContainer object#gridmanContainer img,#gridmanContainer video#gridmanContainer a#gridmanContainer a:hover#gridmanContainer *,#gridmanContainer :before,#gridmanContainer :after#gridmanContainer *#gridmanContainer *:focus#gridmanContainer *#gridmanContainer *::-webkit-scrollbar*,:before,:after::backdrop.container@media (min-width: 640px)}@media (min-width: 768px)}@media (min-width: 1024px)}@media (min-width: 1280px)}@media (min-width: 1536px)}.pointer-events-none.pointer-events-auto.fixed.absolute.relative.-bottom-1.-bottom-1\.5.-bottom-[1px].-bottom-full.-left-1.-left-1\.5.-left-5.-left-[1px].-left-full.-right-1.-right-1\.5.-right-full.-top-1.-top-1\.5.-top-5.-top-[1px],.-top-px.bottom-0.bottom-1\/3.left-0.left-1\/2.left-4.left-px.right-0.right-2.right-4.top-0.top-1\/2.top-4.top-px.z-[999999].z-[99999].m-auto.mr-1.mt-1.mt-3.mt-px.block.inline.flex.grid.h-0.h-3.h-4.h-5.h-6.h-8.h-9.h-[200px].h-fit.h-full.h-px.max-h-0.max-h-full.min-h-2.w-0.w-3.w-4.w-40.w-5.w-6.w-[200px].w-fit.w-full.w-min.min-w-2.max-w-0.max-w-full.max-w-screen-sm.border-collapse.-translate-x-1\/2.-translate-y-1\/2.translate-x-1\/2.translate-y-1\/2.-rotate-90.rotate-180.rotate-90.transform@keyframes spin}.animate-spin.cursor-not-allowed.cursor-pointer.resize-none.resize.flex-row.flex-col.flex-wrap.items-start.items-end.items-center.justify-center.gap-0.gap-0\.5.gap-1.gap-2.gap-4.gap-6.gap-[2px].self-start.overflow-hidden.whitespace-nowrap.rounded.rounded-full.rounded-lg.rounded-sm.border.border-0.border-4.border-b-2.border-l-2.border-r-2.border-t-2.border-solid.border-rose-200.border-rose-300.border-rose-500.border-rose-700\/30.bg-red-500.bg-rose-400.bg-rose-500.bg-rose-500\/10.bg-rose-500\/5.bg-rose-500\/50.bg-rose-500\/70.bg-rose-600.bg-rose-700.bg-rose-700\/50.bg-slate-200.bg-slate-700.bg-slate-800.bg-transparent.p-0.p-0\.5.p-1.p-1\.5.p-10.p-2.p-4.px-1.px-2.px-3.px-4.py-0.py-1.py-2.py-px.pb-1.pb-2.pl-1.pl-2.pr-1.pr-1\.5.pr-2.pt-0.pt-1.pt-2.text-center.text-right.font-mono.font-sans.text-2xs.text-lg.text-sm.text-xs.font-bold.uppercase.capitalize.italic.leading-normal.leading-tight.tracking-wide.tracking-wider.text-rose-200.text-rose-50.text-rose-700.text-slate-400.text-slate-400\/70.text-slate-500.text-white.opacity-20.opacity-50.shadow.shadow-2xl.shadow-inner.shadow-lg.shadow-md.filter.transition.transition-all.transition-transform.duration-200.ease-in-out#gridmanContainer .flex-c#gridmanContainer .absolute-center-t#gridmanContainer .absolute-center-m@keyframes rotate}#gridmanContainer .bg-stripes-pink.hover\:scale-103:hover.hover\:bg-rose-500\/10:hover.hover\:bg-rose-500\/20:hover.hover\:opacity-0:hover.hover\:opacity-90:hover.[\&\:\:placeholder]\:text-slate-300::-moz-placeholder.[\&\:\:placeholder]\:text-slate-300::placeholder.[\&\>div\:not\(\.Corner\)]\:h-fit>div:not(.Corner).[\&\>div]\:px-1>div.[\&\>div]\:text-xs>div.[\&\>div]\:leading-tight>div.[\&\>div]\:text-white>div.[\&\>div]\:hover\:bg-slate-800:hover>div.[\&\>svg]\:hover\:-translate-x-0\.5:hover>svg.[\&\>svg]\:hover\:translate-x-0\.5:hover>svg</style>`

## A. References

### A.1 Normative References

[Digest-Algorithms-Extension]
* [OCFL Community Extension 0001: Digest
 Algorithms](https://ocfl.github.io/extensions/0001-digest-algorithms.html).
 OCFL Editors. URL:
 <https://ocfl.github.io/extensions/0001-digest-algorithms.html>

[FIPS-180-4]
* [FIPS PUB 180-4: Secure Hash Standard
 (SHS)](https://nvlpubs.nist.gov/nistpubs/FIPS/NIST.FIPS.180-4.pdf).
 U.S. Department of Commerce/National Institute of Standards and
 Technology. August 2015. National Standard. URL:
 <https://nvlpubs.nist.gov/nistpubs/FIPS/NIST.FIPS.180-4.pdf>

[JSON]
* [The JavaScript Object Notation (JSON) Data Interchange
 Format](https://www.rfc-editor.org/rfc/rfc8259). T. Bray, Ed.. IETF.
 December 2017. Internet Standard. URL:
 <https://www.rfc-editor.org/rfc/rfc8259>

[NAMASTE]
* [Directory Description with Namaste
 Tags](https://confluence.ucop.edu/download/attachments/14254149/NamasteSpec.pdf). J.
 Kunze. 9 November 2009. URL:
 <https://confluence.ucop.edu/download/attachments/14254149/NamasteSpec.pdf>

[OAIS]
* [Reference Model for an Open Archival Information System (OAIS),
 Issue 2](https://public.ccsds.org/pubs/650x0m2.pdf). June 2012. URL:
 <https://public.ccsds.org/pubs/650x0m2.pdf>

[PairTree]
* [Pairtrees for Object
 Storage](https://confluence.ucop.edu/display/Curation/PairTree). J.
 Kunze; M. Haye; E. Hetzner; M. Reyes; C. Snavely. 12 August 2008.
 URL: <https://confluence.ucop.edu/display/Curation/PairTree>

[RFC1321]
* [The MD5 Message-Digest
 Algorithm](https://www.rfc-editor.org/rfc/rfc1321). R. Rivest. IETF.
 April 1992. Informational. URL:
 <https://www.rfc-editor.org/rfc/rfc1321>

[RFC2119]
* [Key words for use in RFCs to Indicate Requirement
 Levels](https://www.rfc-editor.org/rfc/rfc2119). S. Bradner. IETF.
 March 1997. Best Current Practice. URL:
 <https://www.rfc-editor.org/rfc/rfc2119>

[RFC3339]
* [Date and Time on the Internet:
 Timestamps](https://www.rfc-editor.org/rfc/rfc3339). G. Klyne; C.
 Newman. IETF. July 2002. Proposed Standard. URL:
 <https://www.rfc-editor.org/rfc/rfc3339>

[RFC3986]
* [Uniform Resource Identifier (URI): Generic
 Syntax](https://www.rfc-editor.org/rfc/rfc3986). T. Berners-Lee; R.
 Fielding; L. Masinter. IETF. January 2005. Internet Standard. URL:
 <https://www.rfc-editor.org/rfc/rfc3986>

[RFC4648]
* [The Base16, Base32, and Base64 Data
 Encodings](https://www.rfc-editor.org/rfc/rfc4648). S. Josefsson.
 IETF. October 2006. Proposed Standard. URL:
 <https://www.rfc-editor.org/rfc/rfc4648>

[RFC6068]
* [The 'mailto' URI
 Scheme](https://www.rfc-editor.org/rfc/rfc6068). M. Duerst; L.
 Masinter; J. Zawinski. IETF. October 2010. Proposed Standard. URL:
 <https://www.rfc-editor.org/rfc/rfc6068>

[RFC7693]
* [The BLAKE2 Cryptographic Hash and Message Authentication Code
 (MAC)](https://www.rfc-editor.org/rfc/rfc7693). M-J. Saarinen, Ed.;
 J-P. Aumasson. IETF. November 2015. Informational. URL:
 <https://www.rfc-editor.org/rfc/rfc7693>

[RFC8141]
* [Uniform Resource Names
 (URNs)](https://www.rfc-editor.org/rfc/rfc8141). P. Saint-Andre; J.
 Klensin. IETF. April 2017. Proposed Standard. URL:
 <https://www.rfc-editor.org/rfc/rfc8141>

[RFC8174]
* [Ambiguity of Uppercase vs Lowercase in RFC 2119 Key
 Words](https://www.rfc-editor.org/rfc/rfc8174). B. Leiba. IETF.
 May 2017. Best Current Practice. URL:
 <https://www.rfc-editor.org/rfc/rfc8174>

### A.2 Informative References

[BagIt]
* [The BagIt File Packaging Format
 (V1.0)](https://tools.ietf.org/html/draft-kunze-bagit-17). J.
 Kunze; J. Littman; E. Madden; J. Scancella; C. Adams. 17
 September 2018. URL:
 <https://tools.ietf.org/html/draft-kunze-bagit-17>

[JSON-Schema]
* [JSON Schema Validation: A Vocabulary for Structural Validation of
 JSON](https://json-schema.org/latest/json-schema-validation.html). A.
 Wright; H Andrews. 20 September 2018. URL:
 <https://json-schema.org/latest/json-schema-validation.html>

[Moab]
* [The Moab Design for Digital Object
 Versioning](https://journal.code4lib.org/articles/8482). Richard
 Anderson. 15 July 2013. URL:
 <https://journal.code4lib.org/articles/8482>

[OCFL-Implementation-Notes]
* [OCFL Implementation
 Notes](https://ocfl.io/1.0/implementation-notes). URL:
 [../implementation-notes](https://ocfl.io/1.0/implementation-notes)

[[↑]](https://ocfl.io/1.0/spec/#title)

[]

[Permalink](https://ocfl.io/1.0/spec/#dfn-content-path)

**Referenced in:**

* [3.4 Digests](https://ocfl.io/1.0/spec/#ref-for-dfn-content-path-1)
* [3.5.2 Manifest](https://ocfl.io/1.0/spec/#ref-for-dfn-content-path-2)
 [(2)](https://ocfl.io/1.0/spec/#ref-for-dfn-content-path-3)
* [3.5.4 Fixity](https://ocfl.io/1.0/spec/#ref-for-dfn-content-path-4)

[]

[Permalink](https://ocfl.io/1.0/spec/#dfn-digest)

**Referenced in:**

* Not referenced in this document.

[]

[Permalink](https://ocfl.io/1.0/spec/#dfn-extension)

**Referenced in:**

* [3.5.4 Fixity](https://ocfl.io/1.0/spec/#ref-for-dfn-extension-1)

[]

[Permalink](https://ocfl.io/1.0/spec/#dfn-inventory)

**Referenced in:**

* [2. Terminology](https://ocfl.io/1.0/spec/#ref-for-dfn-inventory-1)
 [(2)](https://ocfl.io/1.0/spec/#ref-for-dfn-inventory-2)

[]

[Permalink](https://ocfl.io/1.0/spec/#dfn-logical-path)

**Referenced in:**

* [3.4 Digests](https://ocfl.io/1.0/spec/#ref-for-dfn-logical-path-1)
* [3.5.3.1
 Version](https://ocfl.io/1.0/spec/#ref-for-dfn-logical-path-2)
 [(2)](https://ocfl.io/1.0/spec/#ref-for-dfn-logical-path-3)

[]

[Permalink](https://ocfl.io/1.0/spec/#dfn-logical-state)

**Referenced in:**

* [2.
 Terminology](https://ocfl.io/1.0/spec/#ref-for-dfn-logical-state-1)
* [3.5.3.1
 Version](https://ocfl.io/1.0/spec/#ref-for-dfn-logical-state-2)
 [(2)](https://ocfl.io/1.0/spec/#ref-for-dfn-logical-state-3)
 [(3)](https://ocfl.io/1.0/spec/#ref-for-dfn-logical-state-4)

[]

[Permalink](https://ocfl.io/1.0/spec/#dfn-logs-directory)

**Referenced in:**

* Not referenced in this document.

[]

[Permalink](https://ocfl.io/1.0/spec/#dfn-manifest)

**Referenced in:**

* [2. Terminology](https://ocfl.io/1.0/spec/#ref-for-dfn-manifest-1)

[]

[Permalink](https://ocfl.io/1.0/spec/#dfn-ocfl-object)

**Referenced in:**

* [2. Terminology](https://ocfl.io/1.0/spec/#ref-for-dfn-ocfl-object-1)
 [(2)](https://ocfl.io/1.0/spec/#ref-for-dfn-ocfl-object-2)
* [3.5.2 Manifest](https://ocfl.io/1.0/spec/#ref-for-dfn-ocfl-object-3)

[]

[Permalink](https://ocfl.io/1.0/spec/#dfn-ocfl-object-root)

**Referenced in:**

* [2.
 Terminology](https://ocfl.io/1.0/spec/#ref-for-dfn-ocfl-object-root-1)
* [3.1 Object
 Structure](https://ocfl.io/1.0/spec/#ref-for-dfn-ocfl-object-root-2)
* [3.5.2
 Manifest](https://ocfl.io/1.0/spec/#ref-for-dfn-ocfl-object-root-3)
* [4.3 Storage
 Hierarchies](https://ocfl.io/1.0/spec/#ref-for-dfn-ocfl-object-root-4)

[]

[Permalink](https://ocfl.io/1.0/spec/#dfn-ocfl-storage-root)

**Referenced in:**

* [4. OCFL Storage
 Root](https://ocfl.io/1.0/spec/#ref-for-dfn-ocfl-storage-root-1)
* [4.2 Root Conformance
 Declaration](https://ocfl.io/1.0/spec/#ref-for-dfn-ocfl-storage-root-2)
* [4.3 Storage
 Hierarchies](https://ocfl.io/1.0/spec/#ref-for-dfn-ocfl-storage-root-3)

[]

[Permalink](https://ocfl.io/1.0/spec/#dfn-ocfl-version)

**Referenced in:**

* [3.5.3.1
 Version](https://ocfl.io/1.0/spec/#ref-for-dfn-ocfl-version-1)
 [(2)](https://ocfl.io/1.0/spec/#ref-for-dfn-ocfl-version-2)
 [(3)](https://ocfl.io/1.0/spec/#ref-for-dfn-ocfl-version-3)

[]

[Permalink](https://ocfl.io/1.0/spec/#dfn-registered-extension-name)

**Referenced in:**

* [3.9 Object
 Extensions](https://ocfl.io/1.0/spec/#ref-for-dfn-registered-extension-name-1)
* [4.1 Root
 Structure](https://ocfl.io/1.0/spec/#ref-for-dfn-registered-extension-name-2)

`<style> #view_info_box#view_info_box>div#view_info_box .attr_name#view_info_box .attr_value#view_info_box #address_link a:link#view_info_box #address_link a:visited#view_info_box #close_button#view_info_box .info_header#view_info_box .input_link.spinner@keyframes spinnerto}</style>`


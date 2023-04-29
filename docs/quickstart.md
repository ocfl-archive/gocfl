
# Quickstart

This guide will walk you through the basics of
using the `gocfl` applications.

## Installation

The `gocfl` applications are available as a single
binary.  You can download the latest release from
the [releases page](https://github.com/je4/gocfl/releases/latest).

Alternatively, you can install the latest version with `go install`:

```bash
go install github.com/je4/gocfl/v2/gocfl@latest
```

## Initializing a new storage root

To initialize a new storage on an existing empty folder root, run the `init` command:

```bash
gocfl init ./storage_root
```

```bash
gocfl init ./ocfl.zip
```

## Add a new object to an existing storage root

To add a new object to an existing storage root, run the `add` command:

```bash
gocfl add ./storage_root C:/temp/ocfltest1 -u 'Jane Doe' -a 'mailto:user@domain' -m 'initial add' --object-id 'id:abc123'
```

```bash
gocfl add ./ocfl.zip C:/temp/ocfltest1 -u 'Jane Doe' -a 'mailto:user@domain' -m 'initial add' --object-id 'id:abc123'
```

## Adding a new version to an existing OCFL Repository

To add a new version to an existing OCFL Object, run the `update` command:

```bash
gocfl update ./storage_root /temp/ocfltest2 -u 'Jane Doe' -a 'mailto:user@domain' -m 'some new data' --object-id 'id:abc123'
```

```bash
gocfl update ./ocfl.zip /temp/ocfltest2 -u 'Jane Doe' -a 'mailto:user@domain' -m 'some new data' --object-id 'id:abc123'
```

This command adds a new version to the object `id:abc123` located in the storage root `archive.zip`.
The object contains the new or changed files from the directory `/temp/ocfltest1`.
The digest is derived from the existing manifest.

By default, deduplication is performed. If you want to disable deduplication, use
the `--no-deduplicate` flag (less I/O, faster).


## Creating an OCFL Repository with one object

To create a new OCFL repository, run the `create` command:

```bash
gocfl create ./storage_root_create /temp/ocfltest1 --digest sha512 -u 'Jane Doe' -a 'mailto:user@domain' -m 'initial add' --object-id 'id:abc123'
```

```bash
gocfl create ./ocfl_create.zip /temp/ocfltest1 --digest sha512 -u 'Jane Doe' -a 'mailto:user@domain' -m 'initial add' --object-id 'id:abc123'
```
This command creates a zip file containing an ocfl storageroot with one object.
The object-id is `id:abc`. The object contains the files from the directory `/temp/ocfltest1`. 
The digest algorithm used for the [OCFL Manifest](https://ocfl.io/1.1/spec/#manifest) is sha512.
As Metadata for the version `v1`, the user is Jane Doe, the email address is user@domain and the message is 'initial add'.

Instead of a zip file, you can also use a directory as target.

By default, no deduplication is performed. If you want to deduplicate the files, 
you can use the `--deduplicate` flag (more I/O, takes longer).

Using the --no-compress flag, you can disable compression of the files in the zip file.


## Validating an OCFL Storage Root or Object

To validate an OCFL Storage Root with all objects, run the `validate` 
command on the target folder or zip file

```bash
gocfl validate ./ocfl.zip
```

```bash
gocfl validate ./storage_root
```

```bash
gocfl validate ./ocfl_create.zip
```

```bash
gocfl validate ./storage_root_create
```



To validate a single OCFL Object, run the `validate` command on the
target folder or zip file and specify the object-id:

```bash
gocfl validate ./ocfl.zip --object-id 'id:abc123'
```

## Extracting an Object from an OCFL Repository

To extract an object from an OCFL Repository, run the `extract` command:

```bash
gocfl extract ./ocfl.zip ./extract --object-id 'id:abc123' --with-manifest
```

```bash
gocfl extract ./storage_root ./extract --object-id 'id:abc123' --with-manifest
```

This command extracts the object `id:abc123` from the storage root `archive.zip` to the directory `/temp/abc123`.
The `--with-manifest` flag adds the manifest file `manifest.XXX` to the extracted object 
where `XXX`is the digest algorithm used for the manifest. The target directory must be empty.

With sha512sum, you can check the integrity of the extracted object
```bash
cd extract; sha512sum -c manifest.sha512; cd ..
```


## Extracting Metadata from an OCFL Object

To extract metadata from an OCFL Object, run the `extractmeta` command:

```bash
gocfl extractmeta ./ocfl.zip --object-id 'id:abc123' --format json --output ./meta.json
```

```bash
gocfl extractmeta ./storage_root --object-id 'id:abc123' --format json --output ./meta.json
```

## Extracting Information from an OCFL Storage Root or Object

To extract information from an OCFL Object, run the `info` command:

```bash
gocfl info ./ocfl.zip --object-id 'id:abc123' 
```

```bash
gocfl info ./storage_root --object-id 'id:abc123' 
```

This command prints information about the object `id:abc123` from the storage root `archive.zip` or 
`storage_root`to the console.
It contains the Object ID, Digest, Head Version, Number of Files

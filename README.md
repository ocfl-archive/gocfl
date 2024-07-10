# Attention
New simplified config file format
Look at [gocfl2.toml](./config/gocfl2.toml)

# Installation

`go install github.com/ocfl-archive/gocfl/v2/gocfl@latest`

# Go OCFL Implementation

This library supports the Oxford Common Filesystem Layout ([OCFL](https://ocfl.io/)) 
and focuses on creation, update, validation and extraction of ocfl StorageRoots and Objects.

GOCFL command line tool supports the following subcommands
* [init](docs/init.md)
* [add](docs/add.md)
* [create](docs/create.md)
* [update](docs/update.md)
* [validate](docs/validate.md)
* [info](docs/stat.md)
* [extract](docs/extract.md)
* [extractmeta](docs/extractmeta.md)
* [display](docs/display.md)

There's a [quickstart guide](docs/quickstart.md) available.

## Why
There are several [OCFL tools & libraries](https://github.com/OCFL/spec/wiki/Implementations#code-libraries-validators-and-other-tools) 
which already exists. This software is build with the following motivation.

### I/O Performance
Regarding performance, Storage I/O generates the main performance issues. Therefor, every file 
should be read and written only once. Only in case of deduplication, the checksum of a file is
calculated before ingest and a second time while ingesting. 

### Container 
Serialization of an OCFL Storage Root into a container format like ZIP must not generate 
overhead on disk I/O. Therefor generation of an OCFL Container is possible without an intermediary
OCFL Storage Root on a filesystem.  

#### Encryption 
For storing OCFL containers in low-security locations (cloud storage, etc.), it's possible to 
create an AES-256 encrypted container on ingest.

### Extensions
The extensions described in the OCFL standard are quite open in their functionality and may 
belong to the [Storage Root](https://ocfl.io/1.1/spec/#storage-root-extensions) or
[Object](https://ocfl.io/1.1/spec/#object-extensions). Since there's no specification of a 
generic extension api, it's difficult to integrate specific extension hooks into other libraries. 
This library identifies 7 different extension hooks so far.

#### Indexer
When content is ingested into OCFL objects, technical metadata should be extracted and stored alongside the manifest data. This allows technical metadata to be extracted alongside the content.
Since the OCFL structure is quite rigid, there's a need for a special extension to support this.

## Functionality

- [x] Supports local filesystems
- [x] Supports S3 Cloud Storage (via [MinIO Client SDK](https://github.com/minio/minio-go))
- [ ] SFTP Storage
- [ ] Google Cloud Storage
- [x] Serialization into ZIP Container
- [x] AES Encryption of Container
- [x] Supports mixing of source and target storage systems
- [x] Non blocking validation (does not stop on validation errors)
- [x] Support for OCFL v1.0 and v1.1
- [ ] Documentation for API
- [x] Digest Algorithms for Manifest: SHA512, SHA256
- [x] Fixity Algorithms: SHA1, SHA256, SHA512, BLAKE2b-160, BLAKE2b-256, BLAKE2b-384, BLAKE2b-512, MD5
- [x] Concurrent checksum generation on ingest/extract (multi-threaded)
- [x] Minimized I/O (data is read and written only once on Object creation)
- [x] Update strategy echo (incl. deletions) and contribute
- [x] Deduplication (needs double read of all content files, switchable)
- [x] Nearly full coverage of validation errors and warnings
- [x] Content information
- [x] Extraction with version selection
- [x] Display of content via Webserver
- [x] Report generation
- [Community Extensions](https://github.com/OCFL/extensions/docs) 
  - [x] 0001-digest-algorithms
  - [x] 0002-flat-direct-storage-layout
  - [x] 0003-hash-and-id-n-tuple-storage-layout
  - [x] 0004-hashed-n-tuple-storage-layout
  - [ ] 0005-mutable-head
  - [x] 0006-flat-omit-prefix-storage-layout
  - [x] 0007-n-tuple-omit-prefix-storage-layout
  - [ ] 0008-schema-registry
- Local Extensions
  - [x] [NNNN-pairtree-storage-layout](https://pythonhosted.org/Pairtree/pairtree.pairtree_client.PairtreeStorageClient-class.html) 
  - [x] [NNNN-direct-clean-path-layout](docs/NNNN-direct-clean-path-layout.md)
  - [x] [NNNN-content-subpath](docs/NNNN-content-subpath.md) (integration of non-payload files in content)
  - [x] [NNNN-metafile](docs/NNNN-metafile.md) (integration of a metadata file)
  - [x] [NNNN-mets](docs/NNNN-mets.md) (generation of mets and premis files)
  - [x] [NNNN-indexer](docs/NNNN-indexer.md) (technical metadata indexing)
  - [x] [NNNN-migration](docs/NNNN-migration.md) (migration of file formats)
  - [x] [NNNN-gocfl-extension-manager](docs/NNNN-gocfl-extension-manager.md) (initial extension for sorted exclusion and sorted execution)
  - [x] [NNNN-filesystem](docs/NNNN-filesystem.md) (filesystem metadata extension)
  - [x] [NNNN-thumbnail](docs/NNNN-thumbnail.md) (generation of thumbnails)

## Command Line Interface

```
A fast and reliable OCFL creator, extractor and validator.
https://github.com/ocfl-archive/gocfl
Jürgen Enge (University Library Basel, juergen@info-age.net)
Version v1.0-beta.7

Usage:
  gocfl [flags]
  gocfl [command]

Available Commands:
  add         adds new object to existing ocfl structure
  completion  Generate the autocompletion script for the specified shell
  create      creates a new ocfl structure with initial content of one object
  display     show content of ocfl object in webbrowser
  extract     extract version of ocfl content
  extractmeta extract metadata from ocfl structure
  help        Help about any command
  init        initializes an empty ocfl structure
  stat        statistics of an ocfl structure
  update      update object in existing ocfl structure
  validate    validates an ocfl structure

Flags:
      --config string                 config file (default is $HOME/.gocfl.toml)
  -h, --help                          help for gocfl
      --log-file string               log output file (default is console)
      --log-level string              log level (CRITICAL|ERROR|WARNING|NOTICE|INFO|DEBUG) (default "ERROR")
      --s3-access-key-id string       Access Key ID for S3 Buckets
      --s3-endpoint string            Endpoint for S3 Buckets
      --s3-region string              Region for S3 Access
      --s3-secret-access-key string   Secret Access Key for S3 Buckets
      --with-indexer                  starts indexer as a local service

Use "gocfl [command] --help" for more information about a command.
```


# Go OCFL Implementation

This library supports the Oxford Common Filesystem Layout ([OCFL](https://ocfl.io/)) 
and focuses on creation, update, validation and extraction of ocfl StorageRoots and Objects.

## Why
There are several [OCFL tools & libraries](https://github.com/OCFL/spec/wiki/Implementations#code-libraries-validators-and-other-tools) 
which already exists. 

### Container & I/O Performance

Serialization of an OCFL Storage Root into a container format like ZIP should be possible 
without much overhead on disk I/O. Therefor generation of an OCFL Container should be possible 
with one Read and one Write Task. Including checksums and compression (ZIP). Without deduplication
this library supports this functionality. 

### Extensions

Extensions described in the OCFL Standard are quite open in their functionality and can 
belong to [Storage Root](https://ocfl.io/1.1/spec/#storage-root-extensions) or 
[Object](https://ocfl.io/1.1/spec/#object-extensions). Since there's no specification of 
a generic extension api, it's hard to integrate specific extension hooks into other 
libraries. This library identifies 7 different hooks for extensions till now. 


## Functionality

- [x] Supports local filesystems
- [x] Supports S3 Cloud Storage (via [MinIO Client SDK](https://github.com/minio/minio-go))
- [ ] SFTP Storage
- [ ] Google Cloud Storage
- [x] storage root in ZIP files (native)
- [x] Supports mixing of source and target storage systems
- [x] Non blocking validation (does not stop on validation errors)
- [x] Support for OCFL v1.0 and v1.1
- [ ] Documentation for API
- [x] Digest Algorithms for Manifest: SHA512, SHA256
- [x] Fixity Algorithms: SHA1, SHA256, SHA512, BLAKE2b-160, BLAKE2b-256, BLAKE2b-384, BLAKE2b-512, MD5
- [x] Concurrent checksum generation on ingest/extract (multi-threaded)
- [x] minimized I/O (data is read and written only once on Object creation)
- [x] update strategy echo (incl. deletions) and contribute
- [x] deduplication (needs double read of all content files, switchable)
- [x] nearly full coverage of validation errors and warnings
- [x] content information
- [x] extraction with version selection
- [Community Extensions](https://github.com/OCFL/extensions) 
  - [ ] 0001-digest-algorithms
  - [x] 0002-flat-direct-storage-layout
  - [x] 0003-hash-and-id-n-tuple-storage-layout
  - [x] 0004-hashed-n-tuple-storage-layout
  - [ ] 0005-mutable-head
  - [ ] 0006-flat-omit-prefix-storage-layout
  - [ ] 0007-n-tuple-omit-prefix-storage-layout
  - [ ] 0008-schema-registry
- Local Extensions
  - [x] [NNNN-pairtree-storage-layout](https://pythonhosted.org/Pairtree/pairtree.pairtree_client.PairtreeStorageClient-class.html) 
  - [x] NNNN-direct-clean-path-layout
  - [x] NNNN-content-subpath (integration of non-payload files in content)
  - [x] NNNN-metafile (integration of one file into extension folder)
  - [ ] NNNN-indexer (technical metadata indexing) 
  - [x] NNNN-gocfl-extension-manager (initial extension for sorted exclusion and sorted execution)

## Command Line Interface

```
An OCFL creator, extractor and validator.
      https://go.ub.unibas.ch/gocfl
      Jürgen Enge (University Library Basel, juergen@info-age.net)

Usage:
gocfl [flags]
gocfl [command]

Available Commands:
add         adds new object to existing ocfl structure
completion  Generate the autocompletion script for the specified shell
create      creates a new ocfl structure with initial content of one object
extract     extract version of ocfl content
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
      --s3-secret-access-key string   Secret Access Key for S3 Buckets

Use "gocfl [command] --help" for more information about a command.
```


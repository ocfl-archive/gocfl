# Go OCFL Implementation

This library supports the Oxford Common Filesystem Layout ([OCFL](https://ocfl.io/)) 
and focuses on creation, update, validation and extraction of ocfl StorageRoots and Objects.

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
- [x] minimized I/O (data is read and written only once on Object creation)#
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

### create
equals to storage root initialisation and addition auf one object (init/add).
Mainly used for storage root's with only one object (i.e. in zip files)

```
Usage:
  gocfl create [path to ocfl structure] [flags]

Examples:
gocfl create ./archive.zip /tmp/testdata --digest sha512 -u 'Jane Doe' -a 'mailto:user@domain' -m 'initial add' -object-id 'id:abc123'

Flags:
      --deduplicate                             set flag to force deduplication (slower)
      --default-object-extensions string        folder with initial extension configurations for new OCFL objects
      --default-storageroot-extensions string   folder with initial extension configurations for new OCFL Storage Root
  -d, --digest string                           digest to use for ocfl checksum
      --ext-NNNN-metafile-source string         url with metadata file. $ID will be replaced with object ID i.e. file:///c:/temp/$ID.json
  -f, --fixity string                           comma separated list of digest algorithms for fixity [blake2b-512 md5 sha1 sha256 sha512 blake2b-160 blake2b-256 blake2b-384]
  -h, --help                                    help for create
  -m, --message string                          message for new object version (required)
  -i, --object-id string                        object id to update (required)
      --ocfl-version string                     ocfl version for new storage root (default "1.1")
  -a, --user-address string                     user address for new object version (required)
  -u, --user-name string                        user name for new object version (required)

Global Flags:
      --config string                 config file (default is $HOME/.gocfl.toml)
      --log-file string               log output file (default is console)
      --log-level string              log level (CRITICAL|ERROR|WARNING|NOTICE|INFO|DEBUG) (default "ERROR")
      --s3-access-key-id string       Access Key ID for S3 Buckets
      --s3-endpoint string            Endpoint for S3 Buckets
      --s3-secret-access-key string   Secret Access Key for S3 Buckets
```

### stat

```
Usage:
  gocfl stat [path to ocfl structure] [flags]

Aliases:
  stat, info

Examples:
gocfl stat ./archive.zip

Flags:
  -h, --help                    help for stat
  -i, --object-id string        object id to show statistics for
  -p, --object-path string      object path to show statistics for
      --stat-info stringArray   info field to show. multiple use [ObjectFolders,ExtensionConfigs,Objects,ObjectVersions,ObjectVersionState,ObjectManifest,ObjectExtensionConfigs]

Global Flags:
      --config string                 config file (default is $HOME/.gocfl.toml)
      --log-file string               log output file (default is console)
      --log-level string              log level (CRITICAL|ERROR|WARNING|NOTICE|INFO|DEBUG) (default "ERROR")
      --s3-access-key-id string       Access Key ID for S3 Buckets
      --s3-endpoint string            Endpoint for S3 Buckets
      --s3-secret-access-key string   Secret Access Key for S3 Buckets
```

### update

```
Usage:
  gocfl update [path to ocfl structure] [flags]

Examples:
gocfl update ./archive.zip /tmp/testdata -u 'Jane Doe' -a 'mailto:user@domain' -m 'initial add' -object-id 'id:abc123'

Flags:
      --echo                              set flag to update strategy 'echo' (reflects deletions). if not set, update strategy is 'contribute'
      --ext-NNNN-metafile-source string   url with metadata file. $ID will be replaced with object ID i.e. file:///c:/temp/$ID.json
  -h, --help                              help for update
  -m, --message string                    message for new object version (required)
      --no-deduplicate                    set flag to disable deduplication (faster)
  -i, --object-id string                  object id to update (required)
  -a, --user-address string               user address for new object version (required)
  -u, --user-name string                  user name for new object version (required)

Global Flags:
      --config string                 config file (default is $HOME/.gocfl.toml)
      --log-file string               log output file (default is console)
      --log-level string              log level (CRITICAL|ERROR|WARNING|NOTICE|INFO|DEBUG) (default "ERROR")
      --s3-access-key-id string       Access Key ID for S3 Buckets
      --s3-endpoint string            Endpoint for S3 Buckets
      --s3-secret-access-key string   Secret Access Key for S3 Buckets
```

### extract

```
Usage:
  gocfl extract [path to ocfl structure] [path to target folder] [flags]

Examples:
gocfl extract ./archive.zip /tmp/archive

Flags:
      --ext-NNNN-content-subpath-area string   subpath for extraction (default: 'content'). 'all' for complete extraction
      --ext-NNNN-metafile-target string        url with metadata target folder
  -h, --help                                   help for extract
  -i, --object-id string                       object id to show statistics for
  -p, --object-path string                     object path to show statistics for
      --version string                         version to extract (default "latest")
      --with-manifest                          generate manifest file in object extraction folder

Global Flags:
      --config string                 config file (default is $HOME/.gocfl.toml)
      --log-file string               log output file (default is console)
      --log-level string              log level (CRITICAL|ERROR|WARNING|NOTICE|INFO|DEBUG) (default "ERROR")
      --s3-access-key-id string       Access Key ID for S3 Buckets
      --s3-endpoint string            Endpoint for S3 Buckets
      --s3-secret-access-key string   Secret Access Key for S3 Buckets
```

### validate

```
Usage:
  gocfl validate [path to ocfl structure] [flags]

Aliases:
  validate, check

Examples:
gocfl validate ./archive.zip

Flags:
  -h, --help                 help for validate
  -o, --object-path string   validate only the selected object in storage root

Global Flags:
      --config string                 config file (default is $HOME/.gocfl.toml)
      --log-file string               log output file (default is console)
      --log-level string              log level (CRITICAL|ERROR|WARNING|NOTICE|INFO|DEBUG) (default "ERROR")
      --s3-access-key-id string       Access Key ID for S3 Buckets
      --s3-endpoint string            Endpoint for S3 Buckets
      --s3-secret-access-key string   Secret Access Key for S3 Buckets
```
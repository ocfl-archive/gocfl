# GOCFL

Go OCFL implementation.

## Installation

### Via go ecosystem

`go install github.com/ocfl-archive/gocfl/v2/gocfl@latest`

### Via GitHub repository

* navigate to gocfl directory (you should see `main.go`).
* run `go tidy` to update local dependencies.
* run `go build` to create a locally compiled gocfl binary.

### Configuration

GOCFL relies on a configuration file to activate, among other things, indexing,
and migration capabilities of the GOCFL tool.

A simplified configuration file can be found at: at [gocfl2.toml][config-1].

[config-1]: ./config/gocfl2.toml

#### Pointing to a custom configuration

GOCL will be compiled with an embedded configuration file. GOCL expects this
configuration to be at `./config/default.toml`. The configuration can be
overwritten and compiled. Optionally, you can supply your own configuration
file, e.g. for an add command:

```sh
./gocfl add \
 ./storage_root /tmp/ocfltest1/ \
 -u "Jane Doe" \
 -a "mailto:user@domain" \
 -m "initial add" \
 --object-id 'id:abc123' \
 --config custom-config.toml
```

### Additional tools

GOCFL is optimised next to the following Windows utilities and you will find
refeences to them under Indexer and Migration settings in the config toml file:

* `convert.exe` via ImageMagick
* `identify.exe` via ImageMagick
* `ffmpeg.exe` via FFmpeg
* `ffprobe.exe` via FFmpeg
* `gswin64.exe` via Ghostscript
* `powershell.exe` via Windows Powershell

With the exception of Powershell (discussed below) you should be able to find
drop-in replacements in 'nix-like systems, e.g. `convert.exe` becomes `convert`
in Linux `identify.exe` becomes `identify` and `gswin64` becomes `gs`.

#### Use of Powershell

Powershell scripts are currently used to generate thumbnails for video and pdf.
They are found in the [./data/scripts][ps-1] folder. You can observe their
functionality to write equivalents for your own operating systen's shell.

[ps-1]: ./data/scripts/

### Invoking indexing and migration

Previous GOCFL implementations required a flag to invoke indexing. Now GOCFL
must be compiled with the ObjectExtensions setting configured in the
configuration toml. This line can be optionally commented out. It will look
like as follows:

```toml
[Add]

ObjectExtensions="./data/fullextensions/object"
```

Providing the additional tools are configured correctly, and their function
set to `Enabled=true` in the config, they will run during GOCFL activities such
as `add`.

## Go OCFL Implementation

This library supports the Oxford Common Filesystem Layout ([OCFL][ocfl-1]) and
focuses on `creation`, `update`, `validation` and `extraction` of OCFL
StorageRoots and Objects.

[ocfl-1]: https://ocfl.io/

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

### Why GOCFL

There are several [OCFL tools & libraries][ocfl-2] that already exist. This
software is built with the following motivation:

* I/O performance.
* Containers.
* Encryption.
* Extensions.
* Indexing.

[ocfl-2]: https://github.com/OCFL/spec/wiki/Implementations#code-libraries-validators-and-other-tools

#### I/O Performance

Regarding performance, Storage I/O generates the main performance issues.
Therefore, every file should be read and written only once. Only in case of
deduplication, the checksum of a file is calculated before ingest and a second
time while ingesting.

#### Containers

Serialization of an OCFL Storage Root into a container format like ZIP must not
generate overhead on disk I/O. Therefor generation of an OCFL Container is
possible without an intermediary OCFL Storage Root on a filesystem.

#### Encryption

For storing OCFL containers in low-security locations (cloud storage, etc.),
it's possible to create an AES-256 encrypted container on ingest.

#### Extensions

The extensions described in the OCFL standard are quite open in their
functionality and may belong to the [Storage Root][ext-1] or [Object][ext-2].
Since there's no specification of a generic extension api, it's difficult to
integrate specific extension hooks into other libraries. This library identifies
7 different extension hooks so far.

[ext-1]: https://ocfl.io/1.1/spec/#storage-root-extensions
[ext-2]: https://ocfl.io/1.1/spec/#object-extensions

#### Indexer

When content is ingested into OCFL objects, technical metadata should be
extracted and stored alongside the manifest data. This allows technical metadata
to be extracted alongside the content. Since the OCFL structure is quite rigid,
there's a need for a special extension to support this.

### GOCFL Functionality

<!--markdownlint-disable-->

* [x] Supports local filesystems
* [x] Supports S3 Cloud Storage (via [MinIO Client SDK](https://github.com/minio/minio-go))
* [ ] SFTP Storage
* [ ] Google Cloud Storage
* [x] Serialization into ZIP Container
* [x] AES Encryption of Container
* [x] Supports mixing of source and target storage systems
* [x] Non blocking validation (does not stop on validation errors)
* [x] Support for OCFL v1.0 and v1.1
* [ ] Documentation for API
* [x] Digest Algorithms for Manifest: SHA512, SHA256
* [x] Fixity Algorithms: SHA1, SHA256, SHA512, BLAKE2b-160, BLAKE2b-256, BLAKE2b-384, BLAKE2b-512, MD5
* [x] Concurrent checksum generation on ingest/extract (multi-threaded)
* [x] Minimized I/O (data is read and written only once on Object creation)
* [x] Update strategy echo (incl. deletions) and contribute
* [x] Deduplication (needs double read of all content files, switchable)
* [x] Nearly full coverage of validation errors and warnings
* [x] Content information
* [x] Extraction with version selection
* [x] Display of content via Webserver
* [x] Report generation
* [Community Extensions](https://github.com/OCFL/extensions/docs)
  * [x] 0001-digest-algorithms
  * [x] 0002-flat-direct-storage-layout
  * [x] 0003-hash-and-id-n-tuple-storage-layout
  * [x] 0004-hashed-n-tuple-storage-layout
  * [ ] 0005-mutable-head
  * [x] 0006-flat-omit-prefix-storage-layout
  * [x] 0007-n-tuple-omit-prefix-storage-layout
  * [ ] 0008-schema-registry
* Local Extensions
  * [x] [NNNN-pairtree-storage-layout](https://pythonhosted.org/Pairtree/pairtree.pairtree_client.PairtreeStorageClient-class.html)
  * [x] [NNNN-direct-clean-path-layout](docs/NNNN-direct-clean-path-layout.md)
  * [x] [NNNN-content-subpath](docs/NNNN-content-subpath.md) (integration of non-payload files in content)
  * [x] [NNNN-metafile](docs/NNNN-metafile.md) (integration of a metadata file)
  * [x] [NNNN-mets](docs/NNNN-mets.md) (generation of mets and premis files)
  * [x] [NNNN-indexer](docs/NNNN-indexer.md) (technical metadata indexing)
  * [x] [NNNN-migration](docs/NNNN-migration.md) (migration of file formats)
  * [x] [NNNN-gocfl-extension-manager](docs/NNNN-gocfl-extension-manager.md) (initial extension for sorted exclusion and sorted execution)
  * [x] [NNNN-filesystem](docs/NNNN-filesystem.md) (filesystem metadata extension)
  * [x] [NNNN-thumbnail](docs/NNNN-thumbnail.md) (generation of thumbnails)

<!--markdownlint-enable-->

### Command Line Interface

```text
A fast and reliable OCFL creator, extractor and validator.
https://github.com/ocfl-archive/gocfl
Jürgen Enge (University Library Basel, juergen@info-age.net)
Version v2.0.6

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
      --config string                 config file (default is embedded)
  -h, --help                          help for gocfl
      --log-file string               log output file (default is console)
      --log-level string              log level (CRITICAL|ERROR|WARNING|NOTICE|INFO|DEBUG)
      --s3-access-key-id string       Access Key ID for S3 Buckets
      --s3-endpoint string            Endpoint for S3 Buckets
      --s3-region string              Region for S3 Access
      --s3-secret-access-key string   Secret Access Key for S3 Buckets

Use "gocfl [command] --help" for more information about a command.
```

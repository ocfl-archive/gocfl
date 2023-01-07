# Init

The `init` command initializes an OCFL Storage Root. The [default extension configs](../data/defaultextensions/storageroot) are used for that. 

```
PS C:\daten\go\dev\gocfl> ../bin/gocfl.exe init --help
initializes an empty ocfl structure

Usage:
  gocfl init [path to ocfl structure] [flags]

Examples:
gocfl init ./archive.zip

Flags:
      --aes-iv string                           initialisation vector to use for encrypted container in hex format (32 charsempty: generate random vector
      --aes-key string                          key to use for encrypted container in hex format (64 chars, empty: generate random key
      --default-storageroot-extensions string   folder with initial extension configurations for new OCFL Storage Root
  -d, --digest string                           digest to use for ocfl checksum
      --encrypt-aes                             set flag to create encrypted container (only for container target)
  -h, --help                                    help for init
      --ocfl-version string                     ocfl version for new storage root (default "v")

Global Flags:
      --config string                 config file (default is $HOME/.gocfl.toml)
      --log-file string               log output file (default is console)
      --log-level string              log level (CRITICAL|ERROR|WARNING|NOTICE|INFO|DEBUG) (default "ERROR")
      --s3-access-key-id string       Access Key ID for S3 Buckets
      --s3-endpoint string            Endpoint for S3 Buckets
      --s3-secret-access-key string   Secret Access Key for S3 Buckets
```

## Examples

All Examples refer to the same [config file](../config/gocfl.toml)

### Storage Root on File System

```
PS C:\daten\go\dev\gocfl> ../bin/gocfl.exe init c:/temp/ocflroot --config ./config/gocfl.toml
Using config file: ./config/gocfl.toml
2023-01-07T18:51:47.523 cmd::doInit [init.go:104] > INFO - creating 'c:/temp/ocflroot'
2023-01-07T18:51:47.523 cmd::doInit [init.go:108] > INFO - creating 'c:/temp/ocflroot'

no errors found
2023-01-07T18:51:47.538 cmd::doInit.func1 [init.go:106] > INFO - Duration: 14.2393ms
PS C:\daten\go\dev\gocfl> dir /temp/ocflroot

    Directory: C:\temp\ocflroot

Mode                 LastWriteTime         Length Name
----                 -------------         ------ ----
d----          07.01.2023    18:51                extensions
-a---          07.01.2023    18:51              9 0=ocfl_1.1
-a---          07.01.2023    18:51            110 ocfl_layout.json
```

### Storage Root on ZIP File

```
PS C:\daten\go\dev\gocfl> ../bin/gocfl.exe init c:/temp/ocflroot.zip --config ./config/gocfl.toml
Using config file: ./config/gocfl.toml
2023-01-07T18:53:34.192 cmd::doInit [init.go:104] > INFO - creating 'c:/temp/ocflroot.zip'
2023-01-07T18:53:34.193 cmd::doInit [init.go:108] > INFO - creating 'c:/temp/ocflroot.zip'

no errors found
2023-01-07T18:53:34.197 cmd::doInit.func1 [init.go:106] > INFO - Duration: 4.009ms
PS C:\daten\go\dev\gocfl> dir /temp/ocflroot.zip*

    Directory: C:\temp

Mode                 LastWriteTime         Length Name
----                 -------------         ------ ----
-a---          07.01.2023    19:05           1648 ocflroot.zip
-a---          07.01.2023    19:05           1648 ocflroot.zip.aes
-a---          07.01.2023    19:05             32 ocflroot.zip.aes.iv
-a---          07.01.2023    19:05             64 ocflroot.zip.aes.key
-a---          07.01.2023    19:05            147 ocflroot.zip.aes.sha512
-a---          07.01.2023    19:05            143 ocflroot.zip.sha512
```



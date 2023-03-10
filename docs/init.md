# Init

The `init` command initializes an OCFL Storage Root.   
[Default extension configs](../data/defaultextensions/storageroot) are used. 

```
PS C:\daten\go\dev\gocfl> ../bin/gocfl.exe init --help
initializes an empty ocfl structure

Usage:
  gocfl init [path to ocfl structure] [flags]

Examples:
gocfl init ./archive.zip

Flags:
      --aes-iv string                           initialisation vector to use for encrypted container in hex format (32 chars, empty: generate random vector)
      --aes-key string                          key to use for encrypted container in hex format (64 chars, empty: generate random key)
      --default-storageroot-extensions string   folder with initial extension configurations for new OCFL Storage Root
  -d, --digest string                           digest to use for ocfl checksum
      --encrypt-aes                             create encrypted container (only for container target)
  -h, --help                                    help for init
      --no-compression                          do not compress data in zip file
      --ocfl-version string                     ocfl version for new storage root (default "1.1")

Global Flags:
      --config string                 config file (default is $HOME/.gocfl.toml)
      --log-file string               log output file (default is console)
      --log-level string              log level (CRITICAL|ERROR|WARNING|NOTICE|INFO|DEBUG) (default "ERROR")
      --s3-access-key-id string       Access Key ID for S3 Buckets
      --s3-endpoint string            Endpoint for S3 Buckets
      --s3-secret-access-key string   Secret Access Key for S3 Buckets
      --with-indexer                  starts indexer as a local service
```

## Examples

All Examples refer to the same [config file](../config/gocfl.toml)

### Storage Root on File System

```
PS C:\daten\go\dev\gocfl> ../bin/gocfl.exe init c:/temp/ocflroot --config ./config/gocfl.toml
Using config file: ./config/gocfl.toml
2023-01-08T13:27:31.770 cmd::doInit [init.go:104] > INFO - creating 'c:/temp/ocflroot'

no errors found
2023-01-08T13:27:31.775 cmd::doInit.func1 [init.go:106] > INFO - Duration: 5.3941ms

PS C:\Users\micro> Get-ChildItem /temp/ocflroot -recurse

    Directory: C:\temp\ocflroot

Mode                 LastWriteTime         Length Name
----                 -------------         ------ ----
d----          08.01.2023    13:27                extensions
-a---          08.01.2023    13:27              9 0=ocfl_1.1
-a---          08.01.2023    13:27            110 ocfl_layout.json

    Directory: C:\temp\ocflroot\extensions

Mode                 LastWriteTime         Length Name
----                 -------------         ------ ----
d----          08.01.2023    13:27                initial
d----          08.01.2023    13:27                NNNN-direct-clean-path-layout
d----          08.01.2023    13:27                NNNN-direct-path-layout

    Directory: C:\temp\ocflroot\extensions\initial

Mode                 LastWriteTime         Length Name
----                 -------------         ------ ----
-a---          08.01.2023    13:27            510 config.json

    Directory: C:\temp\ocflroot\extensions\NNNN-direct-clean-path-layout

Mode                 LastWriteTime         Length Name
----                 -------------         ------ ----
-a---          08.01.2023    13:27            298 config.json

    Directory: C:\temp\ocflroot\extensions\NNNN-direct-path-layout

Mode                 LastWriteTime         Length Name
----                 -------------         ------ ----
-a---          08.01.2023    13:27             50 config.json
```

### Storage Root on ZIP File

```
PS C:\daten\go\dev\gocfl> ../bin/gocfl.exe init c:/temp/ocfl_init.zip --config ./config/gocfl.toml
Using config file: ./config/gocfl.toml
2023-01-08T13:25:58.771 cmd::doInit [init.go:104] > INFO - creating 'c:/temp/ocfl_init.zip'

no errors found
2023-01-08T13:25:58.774 cmd::doInit.func1 [init.go:106] > INFO - Duration: 2.7303ms
PS C:\daten\go\dev\gocfl> dir /temp/ocfl_init.*

    Directory: C:\temp

Mode                 LastWriteTime         Length Name
----                 -------------         ------ ----
-a---          08.01.2023    13:25           1648 ocfl_init.zip
-a---          08.01.2023    13:25            144 ocfl_init.zip.sha512
```



# Validate

Validates an OCFL Storage Root with one or all Objects. Validation is non-blocking which allows to 
get a list of multiple errors (which may be follow-ups of previous ones).

```
PS C:\daten\go\dev\gocfl> ../bin/gocfl.exe validate --help
validates an ocfl structure

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
      --with-indexer                  starts indexer as a local service
      
```

## Fixtures (OCFL 1.1)
Evalution of the [OCFL fixtures](https://github.com/OCFL/fixtures/tree/main/1.1) result in the following output:

* [Bad Objects](bad-objects.txt)
* [Warn Objects](warn-objects.txt)
* [Good Objects](good-objects.txt)

## Examples

```
PS C:\daten\go\dev\gocfl> ../bin/gocfl.exe validate C:\temp\ocflroot --config ./config/gocfl.toml
Using config file: ./config/gocfl.toml
2023-01-09T16:24:36.152 cmd::validate [validate.go:46] > INFO - validating 'C:/temp/ocflroot'
2023-01-09T16:24:36.153 ocfl::(*StorageRootBase).Check [storagerootbase.go:397] > INFO - StorageRoot with version '1.1' found
object folder 'id=u003Ablah-blubb'
2023-01-09T16:24:36.154 ocfl::(*ObjectBase).Check [objectbase.go:1019] > INFO - object 'id:blah-blubb' with object version '1.1' found

[storage root 'file://C:/temp/ocflroot']
   #W013 - ‘In an OCFL Object, extension sub-directories SHOULD be named according to a registered extension name.’ [extension 'NNNN-direct-clean-path-layout' is not registered]
   #W013 - ‘In an OCFL Object, extension sub-directories SHOULD be named according to a registered extension name.’ [extension 'NNNN-direct-path-layout' is not registered]
   #W013 - ‘In an OCFL Object, extension sub-directories SHOULD be named according to a registered extension name.’ [extension 'NNNN-gocfl-extension-manager' is not registered]

[object 'file://C:/temp/ocflroot/id=u003Ablah-blubb' - '']
   #W013 - ‘In an OCFL Object, extension sub-directories SHOULD be named according to a registered extension name.’ [extension 'NNNN-direct-clean-path-layout' is not registered]
   #W013 - ‘In an OCFL Object, extension sub-directories SHOULD be named according to a registered extension name.’ [extension 'NNNN-gocfl-extension-manager' is not registered]

no errors found
2023-01-09T16:24:40.858 cmd::validate.func1 [validate.go:44] > INFO - Duration: 4.7068474s
```

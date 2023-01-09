# Stat

Shows statistic on content.

```
PS C:\daten\go\dev\gocfl> ../bin/gocfl.exe info --help
statistics of an ocfl structure

Usage:
  gocfl stat [path to ocfl structure] [flags]

Aliases:
  stat, info

Examples:
gocfl stat ./archive.zip

Flags:
  -h, --help                 help for stat
  -i, --object-id string     object id to show statistics for
  -p, --object-path string   object path to show statistics for
      --stat-info string     comma separated list of info fields to show [ObjectManifest,ObjectFolders,ExtensionConfigs,Objects,ObjectVersions,Extension,ObjectVersionState,ObjectExtension,ObjectExtensionConfigs]

Global Flags:
      --config string                 config file (default is $HOME/.gocfl.toml)
      --log-file string               log output file (default is console)
      --log-level string              log level (CRITICAL|ERROR|WARNING|NOTICE|INFO|DEBUG) (default "ERROR")
      --s3-access-key-id string       Access Key ID for S3 Buckets
      --s3-endpoint string            Endpoint for S3 Buckets
      --s3-secret-access-key string   Secret Access Key for S3 Buckets
```

## Examples

```
PS C:\daten\go\dev\gocfl> ../bin/gocfl.exe stat C:\temp\ocflroot --config ./config/gocfl.toml
Using config file: ./config/gocfl.toml
2023-01-09T16:21:38.441 cmd::doStat [stat.go:85] > INFO - opening 'C:/temp/ocflroot'
Storage Root
OCFL Version: 1.1
Initial Extension:
---
{
  "extensionName": "NNNN-gocfl-extension-manager",
  "sort": {
    "StorageRootPath": [
      "NNNN-direct-clean-path-layout"
    ]
  },
  "exclusion": {
    "StorageRootPath": [
      [
        "NNNN-direct-clean-path-layout",
        "0003-hash-and-id-n-tuple-storage-layout",
        "0004-hashed-n-tuple-storage-layout",
        "0002-flat-direct-storage-layout",
        "NNNN-pairtree-storage-layout",
        "NNNN-direct-path-layout"
      ]
    ]
  }
}
---
Extension Configurations:
---
{
  "extensionName": "NNNN-direct-clean-path-layout",
  "maxPathnameLen": 32000,
  "maxFilenameLen": 127,
  "replacementString": "_",
  "whitespaceReplacementString": " ",
  "utfEncode": true,
  "fallbackDigestAlgorithm": "sha512",
  "fallbackFolder": "fallback",
  "fallbackSubdirs": 0
}
---
{
  "extensionName": "NNNN-direct-path-layout"
}
Object Folders: id=u003Ablah-blubb
Initial Extension:
---
{
  "extensionName": "NNNN-gocfl-extension-manager",
  "sort": {
    "StorageRootPath": [
      "NNNN-direct-clean-path-layout"
    ]
  },
  "exclusion": {
    "StorageRootPath": [
      [
        "NNNN-direct-clean-path-layout",
        "0003-hash-and-id-n-tuple-storage-layout",
        "0004-hashed-n-tuple-storage-layout",
        "0002-flat-direct-storage-layout",
        "NNNN-pairtree-storage-layout",
        "NNNN-direct-path-layout"
      ]
    ]
  }
}
---
Extension Configurations:
---
{
  "extensionName": "NNNN-direct-clean-path-layout",
  "maxPathnameLen": 32000,
  "maxFilenameLen": 127,
  "replacementString": "_",
  "whitespaceReplacementString": " ",
  "utfEncode": true,
  "fallbackDigestAlgorithm": "sha512",
  "fallbackFolder": "fallback",
  "fallbackSubdirs": 0
}
---
{
  "extensionName": "NNNN-direct-path-layout"
}
Object: id=u003Ablah-blubb
[id:blah-blubb] Digest: sha512
[id:blah-blubb] Head: v2
[id:blah-blubb] Fixity: blake2b-384, md5, sha256
[id:blah-blubb] Manifest: 16 files (14 unique files)
[id:blah-blubb] Version v2
[id:blah-blubb]     User: Jürgen Enge (mailto:juergen@info-age.net)
[id:blah-blubb]     Created: 2023-01-09 16:12:08 +0100 CET
[id:blah-blubb]     Message: first update
[id:blah-blubb] Version v1
[id:blah-blubb]     User: Jürgen Enge (mailto:juergen@info-age.net)
[id:blah-blubb]     Created: 2023-01-09 16:11:46 +0100 CET
[id:blah-blubb]     Message: Initial commit
[id:blah-blubb] Initial Extension:
---
{
  "extensionName": "NNNN-gocfl-extension-manager",
  "sort": {
    "ObjectContentPath": [
      "NNNN-direct-clean-path-layout",
      "NNNN-content-subpath"
    ]
  },
  "exclusion": {
    "ObjectContentPath": [
      [
        "NNNN-direct-clean-path-layout",
        "NNNN-direct-path-layout"
      ]
    ]
  }
}
---
[id:blah-blubb] Extension Configurations:
---
{
  "extensionName": "0001-digest-algorithms"
}
---
{
  "extensionName": "NNNN-direct-clean-path-layout",
  "maxPathnameLen": 32000,
  "maxFilenameLen": 127,
  "replacementString": "_",
  "whitespaceReplacementString": " ",
  "utfEncode": false,
  "fallbackDigestAlgorithm": "sha512",
  "fallbackFolder": "fallback",
  "fallbackSubdirs": 0
}

[storage root 'file://C:/temp/ocflroot']
   #W013 - ‘In an OCFL Object, extension sub-directories SHOULD be named according to a registered extension name.’ [extension 'NNNN-direct-clean-path-layout' is not registered]
   #W013 - ‘In an OCFL Object, extension sub-directories SHOULD be named according to a registered extension name.’ [extension 'NNNN-direct-path-layout' is not registered]
   #W013 - ‘In an OCFL Object, extension sub-directories SHOULD be named according to a registered extension name.’ [extension 'NNNN-gocfl-extension-manager' is not registered]

[object 'file://C:/temp/ocflroot/id=u003Ablah-blubb' - '']
   #W013 - ‘In an OCFL Object, extension sub-directories SHOULD be named according to a registered extension name.’ [extension 'NNNN-direct-clean-path-layout' is not registered]
   #W013 - ‘In an OCFL Object, extension sub-directories SHOULD be named according to a registered extension name.’ [extension 'NNNN-gocfl-extension-manager' is not registered]

no errors found
2023-01-09T16:21:38.446 cmd::doStat.func1 [stat.go:83] > INFO - Duration: 4.45ms
```

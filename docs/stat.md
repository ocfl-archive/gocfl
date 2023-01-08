# Stat

Shows statistic on content.

```
```

## Examples

```
PS C:\daten\go\dev\gocfl> ../bin/gocfl.exe stat C:/temp/ocflroot --config ./config/gocfl.toml
Using config file: ./config/gocfl.toml
2023-01-08T16:57:13.151 cmd::doStat [stat.go:85] > INFO - opening 'C:/temp/ocflroot'
Storage Root
OCFL Version: 1.1
Object Folders: id=u003Ablah-blubb, id=u003Ablah-blubb-blubb
Object: id=u003Ablah-blubb
[id:blah-blubb] Digest: sha512
[id:blah-blubb] Head: v2
[id:blah-blubb] Manifest: 17 files (15 unique files)
[id:blah-blubb] Version v1
[id:blah-blubb]     User: Jürgen Enge (mailto:juergen@info-age.net)
[id:blah-blubb]     Created: 2023-01-08 14:53:37 +0100 CET
[id:blah-blubb]     Message: Initial commit
[id:blah-blubb] Version v2
[id:blah-blubb]     User: Jürgen Enge (mailto:juergen@info-age.net)
[id:blah-blubb]     Created: 2023-01-08 14:54:48 +0100 CET
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
Object: id=u003Ablah-blubb-blubb
[id:blah-blubb-blubb] Digest: sha512
[id:blah-blubb-blubb] Head: v2
[id:blah-blubb-blubb] Manifest: 16 files (14 unique files)
[id:blah-blubb-blubb] Version v1
[id:blah-blubb-blubb]     User: Jürgen Enge (mailto:juergen@info-age.net)
[id:blah-blubb-blubb]     Created: 2023-01-08 16:56:40 +0100 CET
[id:blah-blubb-blubb]     Message: Initial commit
[id:blah-blubb-blubb] Version v2
[id:blah-blubb-blubb]     User: Jürgen Enge (mailto:juergen@info-age.net)
[id:blah-blubb-blubb]     Created: 2023-01-08 16:56:56 +0100 CET
[id:blah-blubb-blubb]     Message: Initial commit
[id:blah-blubb-blubb] Initial Extension:
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
[id:blah-blubb-blubb] Extension Configurations:
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

no errors found
2023-01-08T16:57:13.162 cmd::doStat.func1 [stat.go:83] > INFO - Duration: 10.6617ms
```

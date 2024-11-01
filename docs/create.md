# Create

The `create` command is a combination of the `init` and `add` commands in one step.   
It's mainly used for single Object Storage Root in a serialized container (ZIP).   
Deduplication is disabled by default.

The [default extension configs](../data/defaultextensions/object) are used for that.

```text
PS C:\daten\go\dev\gocfl> ../bin/gocfl.exe create --help
initializes an empty ocfl structure and adds contents of a directory subtree to it
This command is a combination of init and add

Usage:
  gocfl create [path to ocfl structure] [path to content folder] [flags]

Examples:
gocfl create ./archive.zip /tmp/testdata --digest sha512 -u 'Jane Doe' -a 'mailto:user@domain' -m 'initial add' -object-id 'id:abc123'

Flags:
      --aes-iv string                               initialisation vector to use for encrypted container in hex format (32 char, sempty: generate random vector)
      --aes-key string                              key to use for encrypted container in hex format (64 chars, empty: generate random key)
      --deduplicate                                 force deduplication (slower)
      --default-area string                         default area for update or ingest (default: content)
      --default-object-extensions string            folder with initial extension configurations for new OCFL objects
      --default-storageroot-extensions string       folder with initial extension configurations for new OCFL Storage Root
  -d, --digest string                               digest to use for ocfl checksum
      --encrypt-aes                                 create encrypted container (only for container target)
      --ext-NNNN-metafile-source string             url with metadata file. $ID will be replaced with object ID i.e. file:///c:/temp/$ID.json
      --ext-NNNN-mets-descriptive-metadata string   reference to archived descriptive metadata (i.e. ead:metadata:ead.xml)
  -f, --fixity string                               comma separated list of digest algorithms for fixity [blake2b-512 md5 sha1 sha256 sha512 blake2b-160 blake2b-256 blake2b-384]
  -h, --help                                        help for create
      --keypass-entry string                        keypass2 entry to use for key encryption
      --keypass-file string                         file with keypass2 database
      --keypass-key string                          key to use for keypass2 database decryption
  -m, --message string                              message for new object version (required)
      --no-compress                                 do not compress data in zip file
  -i, --object-id string                            object id to update (required)
      --ocfl-version string                         ocfl version for new storage root (default "1.1")
  -a, --user-address string                         user address for new object version (required)
  -u, --user-name string                            user name for new object version (required)

Global Flags:
      --config string                 config file (default is $HOME/.gocfl.toml)
      --log-file string               log output file (default is console)
      --log-level string              log level (CRITICAL|ERROR|WARNING|NOTICE|INFO|DEBUG)
      --s3-access-key-id string       Access Key ID for S3 Buckets
      --s3-endpoint string            Endpoint for S3 Buckets
      --s3-region string              Region for S3 Access
      --s3-secret-access-key string   Secret Access Key for S3 Buckets
```
## Examples

All Examples refer to the same [config file](../config/gocfl.toml).

### Storage Root in ZIP file
```
PS C:\daten\go\dev\gocfl> ../bin/gocfl.exe create C:/temp/ocfl_create.zip C:/temp/ocfltest --config ./config/gocfl.toml -i "id:blah-blubb"
Using config file: ./config/gocfl.toml
2023-01-07T17:14:38.120 cmd::doCreate [create.go:167] > INFO - creating 'C:/temp/ocfl_create.zip'
creating 'C:/temp/ocfl_create.zip'
2023-01-07T17:14:38.123 ocfl::(*ObjectBase).AddFile [objectbase.go:599] > INFO - adding file content:2017-06-05_10-36-51_793.jpeg
2023-01-07T17:14:38.159 ocfl::(*ObjectBase).AddFile [objectbase.go:599] > INFO - adding file content:2017-07-25_20-15-18_980.jpeg
2023-01-07T17:14:38.180 ocfl::(*ObjectBase).AddFile [objectbase.go:599] > INFO - adding file content:DSC_0111.JPG
2023-01-07T17:14:38.201 ocfl::(*ObjectBase).AddFile [objectbase.go:599] > INFO - adding file content:Kopie von Kopie von bangbang 26_3_gemacht_V2.xlsx
2023-01-07T17:14:38.246 ocfl::(*ObjectBase).AddFile [objectbase.go:599] > INFO - adding file content:Kopie von bangbang 26_3_gemacht_V2.xlsx
2023-01-07T17:14:38.291 ocfl::(*ObjectBase).AddFile [objectbase.go:599] > INFO - adding file content:bangbang_0226_10078_naegelin_2015_prisoners_dilemma_model_plusminus_staged.mp4--web_master.mp4
2023-01-07T17:14:58.713 ocfl::(*ObjectBase).AddFile [objectbase.go:599] > INFO - adding file content:collage.png
2023-01-07T17:15:24.443 ocfl::(*ObjectBase).AddFile [objectbase.go:599] > INFO - adding file content:salon.json
2023-01-07T17:15:24.496 ocfl::(*ObjectBase).AddFile [objectbase.go:599] > INFO - adding file content:~blä blubb[in]/Modulhandbuch_MA_Gestaltung.pdf
2023-01-07T17:15:24.525 ocfl::(*ObjectBase).AddFile [objectbase.go:599] > INFO - adding file content:~blä blubb[in]/bangbang.csv
2023-01-07T17:15:24.614 ocfl::(*ObjectBase).AddFile [objectbase.go:599] > INFO - adding file content:~blä blubb[in]/bangbang26_3V2.csv
2023-01-07T17:15:24.706 ocfl::(*ObjectBase).AddFile [objectbase.go:599] > INFO - adding file content:~blä blubb[in]/bangbang_names.csv
2023-01-07T17:15:24.719 ocfl::(*ObjectBase).AddFile [objectbase.go:599] > INFO - adding file content:~blä blubb[in]/bla.pptx
2023-01-07T17:15:24.728 ocfl::(*ObjectBase).AddFile [objectbase.go:599] > INFO - adding file content:~blä blubb[in]/salon.json
2023-01-07T17:15:24.785 ocfl::(*ObjectBase).AddFile [objectbase.go:599] > INFO - adding file content:~blä blubb[in]/sizecalculation.xlsx
2023-01-07T17:15:24.791 ocfl::(*ObjectBase).Close [objectbase.go:450] > INFO - Closing object 'id:blah-blubb'

no errors found
2023-01-07T17:15:24.796 cmd::doCreate.func1 [create.go:165] > INFO - Duration: 46.6759594s

PS C:\daten\go\dev\gocfl> dir /temp/ocfl_create.*

    Directory: C:\temp

Mode                 LastWriteTime         Length Name
----                 -------------         ------ ----
-a---          07.01.2023    17:15     1589427609 ocfl_create.zip
-a---          07.01.2023    17:15            146 ocfl_create.zip.sha512
```

### Storage Root on S3 Storage

```
PS C:\daten\go\dev\gocfl> ../bin/gocfl.exe create 'bucket:testarchive/test3' C:/temp/ocfltest --config ./config/gocfl.toml -i "id:blah-blubb" --s3-secret-access-key "SWORDFISH"
Using config file: ./config/gocfl.toml
2023-01-07T18:28:45.238 cmd::doCreate [create.go:167] > INFO - creating 'bucket:testarchive/test3'
creating 'bucket:testarchive/test3'
2023-01-07T18:28:47.100 ocfl::(*ObjectBase).AddFile [objectbase.go:599] > INFO - adding file content:2017-06-05_10-36-51_793.jpeg
2023-01-07T18:28:47.381 ocfl::(*ObjectBase).AddFile [objectbase.go:599] > INFO - adding file content:2017-07-25_20-15-18_980.jpeg
2023-01-07T18:28:47.545 ocfl::(*ObjectBase).AddFile [objectbase.go:599] > INFO - adding file content:DSC_0111.JPG
2023-01-07T18:28:47.699 ocfl::(*ObjectBase).AddFile [objectbase.go:599] > INFO - adding file content:Kopie von Kopie von bangbang 26_3_gemacht_V2.xlsx
2023-01-07T18:28:47.897 ocfl::(*ObjectBase).AddFile [objectbase.go:599] > INFO - adding file content:Kopie von bangbang 26_3_gemacht_V2.xlsx
2023-01-07T18:28:48.113 ocfl::(*ObjectBase).AddFile [objectbase.go:599] > INFO - adding file content:bangbang_0226_10078_naegelin_2015_prisoners_dilemma_model_plusminus_staged.mp4--web_master.mp4
2023-01-07T18:29:17.531 ocfl::(*ObjectBase).AddFile [objectbase.go:599] > INFO - adding file content:collage.png
2023-01-07T18:29:46.964 ocfl::(*ObjectBase).AddFile [objectbase.go:599] > INFO - adding file content:salon.json
2023-01-07T18:29:47.347 ocfl::(*ObjectBase).AddFile [objectbase.go:599] > INFO - adding file content:~blä blubb[in]/Modulhandbuch_MA_Gestaltung.pdf
2023-01-07T18:29:47.522 ocfl::(*ObjectBase).AddFile [objectbase.go:599] > INFO - adding file content:~blä blubb[in]/bangbang.csv
2023-01-07T18:29:47.757 ocfl::(*ObjectBase).AddFile [objectbase.go:599] > INFO - adding file content:~blä blubb[in]/bangbang26_3V2.csv
2023-01-07T18:29:48.062 ocfl::(*ObjectBase).AddFile [objectbase.go:599] > INFO - adding file content:~blä blubb[in]/bangbang_names.csv
2023-01-07T18:29:48.241 ocfl::(*ObjectBase).AddFile [objectbase.go:599] > INFO - adding file content:~blä blubb[in]/bla.pptx
2023-01-07T18:29:48.386 ocfl::(*ObjectBase).AddFile [objectbase.go:599] > INFO - adding file content:~blä blubb[in]/salon.json
2023-01-07T18:29:48.836 ocfl::(*ObjectBase).AddFile [objectbase.go:599] > INFO - adding file content:~blä blubb[in]/sizecalculation.xlsx
2023-01-07T18:29:49.008 ocfl::(*ObjectBase).Close [objectbase.go:450] > INFO - Closing object 'id:blah-blubb'

no errors found
2023-01-07T18:29:50.478 cmd::doCreate.func1 [create.go:165] > INFO - Duration: 1m5.2408916s
```

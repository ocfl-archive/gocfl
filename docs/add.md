# Add

The `add` command add a new object to an  OCFL Storage Root.    
Deduplication is disabled by default.  
If the object already exists, an error will occur.  

The [default extension configs](../data/defaultextensions/object) are used for that.

```
PS C:\daten\go\dev\gocfl> ../bin/gocfl.exe add --help
opens an existing ocfl structure and adds a new object. if an object with the given id already exists, an error is produced

Usage:
  gocfl add [path to ocfl structure] [flags]

Examples:
gocfl add ./archive.zip /tmp/testdata -u 'Jane Doe' -a 'mailto:user@domain' -m 'initial add' -object-id 'id:abc123'

Flags:
      --aes-iv string                      initialisation vector to use for encrypted container in hex format (32 chars empty: generate random vector)
      --aes-key string                     key to use for encrypted container in hex format (64 chars, empty: generate random key)
      --deduplicate                        force deduplication (slower)
      --default-object-extensions string   folder with initial extension configurations for new OCFL objects
  -d, --digest string                      digest to use for ocfl checksum
      --encrypt-aes                        create encrypted container (only for container target)
      --ext-NNNN-metafile-source string    url with metadata file. $ID will be replaced with object ID i.e. file:///c:/temp/$ID.json
  -f, --fixity string                      comma separated list of digest algorithms for fixity
  -h, --help                               help for add
  -m, --message string                     message for new object version (required)
  -i, --object-id string                   object id to update (required)
  -a, --user-address string                user address for new object version (required)
  -u, --user-name string                   user name for new object version (required)

Global Flags:
      --config string                 config file (default is $HOME/.gocfl.toml)
      --log-file string               log output file (default is console)
      --log-level string              log level (CRITICAL|ERROR|WARNING|NOTICE|INFO|DEBUG) (default "ERROR")
      --s3-access-key-id string       Access Key ID for S3 Buckets
      --s3-endpoint string            Endpoint for S3 Buckets
      --s3-secret-access-key string   Secret Access Key for S3 Buckets
```

## Examples

All Examples refer to the same [config file](../config/gocfl.toml).

# Storage Root on local filesystem

```
PS C:\daten\go\dev\gocfl> ../bin/gocfl.exe add C:/temp/ocflroot C:/temp/ocfltest --config ./config/gocfl.toml -i "id:blah-blubb"
Using config file: ./config/gocfl.toml
opening 'C:/temp/ocflroot'
2023-01-08T13:37:00.814 cmd::doAdd [add.go:147] > INFO - opening 'C:/temp/ocflroot'
2023-01-08T13:37:00.817 ocfl::(*ObjectBase).AddFile [objectbase.go:599] > INFO - adding file content:2017-06-05_10-36-51_793.jpeg
2023-01-08T13:37:00.826 ocfl::(*ObjectBase).AddFile [objectbase.go:599] > INFO - adding file content:2017-07-25_20-15-18_980.jpeg
2023-01-08T13:37:00.831 ocfl::(*ObjectBase).AddFile [objectbase.go:599] > INFO - adding file content:DSC_0111.JPG
2023-01-08T13:37:00.837 ocfl::(*ObjectBase).AddFile [objectbase.go:599] > INFO - adding file content:Kopie von Kopie von bangbang 26_3_gemacht_V2.xlsx
2023-01-08T13:37:00.859 ocfl::(*ObjectBase).AddFile [objectbase.go:599] > INFO - adding file content:Kopie von bangbang 26_3_gemacht_V2.xlsx
2023-01-08T13:37:00.877 ocfl::(*ObjectBase).AddFile [objectbase.go:599] > INFO - adding file content:bangbang_0226_10078_naegelin_2015_prisoners_dilemma_model_plusminus_staged.mp4--web_master.mp4
2023-01-08T13:37:03.650 ocfl::(*ObjectBase).AddFile [objectbase.go:599] > INFO - adding file content:collage.png
2023-01-08T13:37:06.600 ocfl::(*ObjectBase).AddFile [objectbase.go:599] > INFO - adding file content:salon.json
2023-01-08T13:37:06.624 ocfl::(*ObjectBase).AddFile [objectbase.go:599] > INFO - adding file content:~blä blubb[in]/Modulhandbuch_MA_Gestaltung.pdf
2023-01-08T13:37:06.638 ocfl::(*ObjectBase).AddFile [objectbase.go:599] > INFO - adding file content:~blä blubb[in]/bangbang.csv
2023-01-08T13:37:06.652 ocfl::(*ObjectBase).AddFile [objectbase.go:599] > INFO - adding file content:~blä blubb[in]/bangbang26_3V2.csv
2023-01-08T13:37:06.667 ocfl::(*ObjectBase).AddFile [objectbase.go:599] > INFO - adding file content:~blä blubb[in]/bangbang_names.csv
2023-01-08T13:37:06.675 ocfl::(*ObjectBase).AddFile [objectbase.go:599] > INFO - adding file content:~blä blubb[in]/bla.pptx
2023-01-08T13:37:06.682 ocfl::(*ObjectBase).AddFile [objectbase.go:599] > INFO - adding file content:~blä blubb[in]/salon.json
2023-01-08T13:37:06.710 ocfl::(*ObjectBase).AddFile [objectbase.go:599] > INFO - adding file content:~blä blubb[in]/sizecalculation.xlsx
2023-01-08T13:37:06.716 ocfl::(*ObjectBase).Close [objectbase.go:450] > INFO - Closing object 'id:blah-blubb'

no errors found
2023-01-08T13:37:06.721 cmd::doAdd.func1 [add.go:144] > INFO - Duration: 5.9077761s
```


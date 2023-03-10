# Update

Updates an existing object in a Storage Root. If used with a serialized container, a temporary copy
will be created.  
Deduplication is enabled by default.

The [default extension configs](../data/defaultextensions/object) are used for that.

```
PS C:\daten\go\dev\gocfl> ../bin/gocfl.exe update --help
opens an existing ocfl structure and updates an object. if an object with the given id does not exist, an error is produced

Usage:
  gocfl update [path to ocfl structure] [flags]

Examples:
gocfl update ./archive.zip /tmp/testdata -u 'Jane Doe' -a 'mailto:user@domain' -m 'first update' -object-id 'id:abc123'

Flags:
      --aes-iv string                     initialisation vector to use for encrypted container in hex format (32 charsempty: generate random vector
      --aes-key string                    key to use for encrypted container in hex format (64 chars, empty: generate random key
  -d, --digest string                     digest to use for zip file checksum
      --echo                              update strategy 'echo' (reflects deletions). if not set, update strategy is 'contribute'
      --encrypt-aes                       set flag to create encrypted container (only for container target)
      --ext-NNNN-metafile-source string   url with metadata file. $ID will be replaced with object ID i.e. file:///c:/temp/$ID.json
  -h, --help                              help for update
  -m, --message string                    message for new object version (required)
      --no-compression                    do not compress data in zip file
      --no-deduplicate                    disable deduplication (faster)
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
      --with-indexer                  starts indexer as a local service
```

## Examples

All Examples refer to the same [config file](../config/gocfl.toml).

### Update on ZIP file
```
PS C:\daten\go\dev\gocfl> ../bin/gocfl.exe update C:/temp/ocfl.zip C:/temp/ocfltest1 --config ./config/gocfl.toml -i "id:blah-blubb" -u 'Jane Doe' -a 'mailto:user@domain' -m 'first update' 
Using config file: ./config/gocfl.toml
2023-01-08T15:41:06.718 cmd::doUpdate [update.go:147] > INFO - opening 'C:/temp/ocfl.zip'
2023-01-08T15:41:06.721 ocfl::(*ObjectBase).AddFile [objectbase.go:602] > INFO - adding file content:2017-07-25_20-15-18_980 - Kopie.jpeg
2023-01-08T15:41:06.722 ocfl::(*ObjectBase).AddFile [objectbase.go:661] > INFO - [id:blah-blubb] file with same content as '2017-07-25_20-15-18_980 - Kopie.jpeg' already exists. creating virtual copy
2023-01-08T15:41:06.722 ocfl::(*InventoryBase).CopyFile [inventorybase.go:925] > INFO - [id:blah-blubb] copying '5cb8c60eb3c7641561df988493acdd0fbc6b6325ec396a6eaf6a9cbc329e1790b006d61b4465371c21a105b0fb5a77dff9a219ed57ead6cd074d6b8a6e2be896' -> '2017-07-25_20-15-18_980 - Kopie.jpeg'
2023-01-08T15:41:06.722 ocfl::(*ObjectBase).AddFile [objectbase.go:602] > INFO - adding file content:2017-07-25_20-15-18_980XXX.jpeg
2023-01-08T15:41:06.723 ocfl::(*ObjectBase).AddFile [objectbase.go:661] > INFO - [id:blah-blubb] file with same content as '2017-07-25_20-15-18_980XXX.jpeg' already exists. creating virtual copy
2023-01-08T15:41:06.724 ocfl::(*InventoryBase).CopyFile [inventorybase.go:925] > INFO - [id:blah-blubb] copying '5cb8c60eb3c7641561df988493acdd0fbc6b6325ec396a6eaf6a9cbc329e1790b006d61b4465371c21a105b0fb5a77dff9a219ed57ead6cd074d6b8a6e2be896' -> '2017-07-25_20-15-18_980XXX.jpeg'
2023-01-08T15:41:06.724 ocfl::(*ObjectBase).AddFile [objectbase.go:602] > INFO - adding file content:DSC_0111.JPG
2023-01-08T15:41:06.726 ocfl::(*ObjectBase).AddFile [objectbase.go:656] > INFO - [id:blah-blubb] 'DSC_0111.JPG' already exists. ignoring
2023-01-08T15:41:06.726 ocfl::(*ObjectBase).AddFile [objectbase.go:602] > INFO - adding file content:Kopie von bangbang 26_3_gemacht_V2.xlsx
2023-01-08T15:41:06.730 ocfl::(*ObjectBase).AddFile [objectbase.go:656] > INFO - [id:blah-blubb] 'Kopie von bangbang 26_3_gemacht_V2.xlsx' already exists. ignoring
2023-01-08T15:41:06.730 ocfl::(*ObjectBase).AddFile [objectbase.go:602] > INFO - adding file content:bangbang_0226_10078_naegelin_2015_prisoners_dilemma_model_plusminus_staged.mp4--web_master.mp4
2023-01-08T15:41:08.109 ocfl::(*ObjectBase).AddFile [objectbase.go:656] > INFO - [id:blah-blubb] 'bangbang_0226_10078_naegelin_2015_prisoners_dilemma_model_plusminus_staged.mp4--web_master.mp4' already exists. ignoring
2023-01-08T15:41:08.109 ocfl::(*ObjectBase).AddFile [objectbase.go:602] > INFO - adding file content:collageX.png
2023-01-08T15:41:09.591 ocfl::(*ObjectBase).AddFile [objectbase.go:656] > INFO - [id:blah-blubb] 'collageX.png' already exists. ignoring
2023-01-08T15:41:09.591 ocfl::(*ObjectBase).AddFile [objectbase.go:602] > INFO - adding file content:convert.sh
2023-01-08T15:41:09.593 ocfl::(*ObjectBase).AddFile [objectbase.go:602] > INFO - adding file content:salon.json
2023-01-08T15:41:09.603 ocfl::(*ObjectBase).AddFile [objectbase.go:656] > INFO - [id:blah-blubb] 'salon.json' already exists. ignoring
2023-01-08T15:41:09.603 ocfl::(*ObjectBase).AddFile [objectbase.go:602] > INFO - adding file content:together - Kopie.png
2023-01-08T15:41:09.604 ocfl::(*ObjectBase).AddFile [objectbase.go:661] > INFO - [id:blah-blubb] file with same content as 'together - Kopie.png' already exists. creating virtual copy
2023-01-08T15:41:09.605 ocfl::(*InventoryBase).CopyFile [inventorybase.go:925] > INFO - [id:blah-blubb] copying '0dc7a8414ffe8ff4c0189091207e839d337412545401959cd0aa640043112941af00b79bcfa6b98f2b5f0236862774b76644eb5e105483e3e1591e54382319ea' -> 'together - Kopie.png'
2023-01-08T15:41:09.605 ocfl::(*ObjectBase).AddFile [objectbase.go:602] > INFO - adding file content:together.png
2023-01-08T15:41:09.606 ocfl::(*ObjectBase).AddFile [objectbase.go:656] > INFO - [id:blah-blubb] 'together.png' already exists. ignoring
2023-01-08T15:41:09.606 ocfl::(*ObjectBase).AddFile [objectbase.go:602] > INFO - adding file content:~blä blubb[in]/Modulhandbuch_MA_Gestaltung.pdf
2023-01-08T15:41:09.608 ocfl::(*ObjectBase).AddFile [objectbase.go:656] > INFO - [id:blah-blubb] '~blä blubb[in]/Modulhandbuch_MA_Gestaltung.pdf' already exists. ignoring
2023-01-08T15:41:09.608 ocfl::(*ObjectBase).AddFile [objectbase.go:602] > INFO - adding file content:~blä blubb[in]/bangbang.csv
2023-01-08T15:41:09.611 ocfl::(*ObjectBase).AddFile [objectbase.go:656] > INFO - [id:blah-blubb] '~blä blubb[in]/bangbang.csv' already exists. ignoring
2023-01-08T15:41:09.611 ocfl::(*ObjectBase).AddFile [objectbase.go:602] > INFO - adding file content:~blä blubb[in]/bangbang26_3V2.csv
2023-01-08T15:41:09.616 ocfl::(*ObjectBase).AddFile [objectbase.go:656] > INFO - [id:blah-blubb] '~blä blubb[in]/bangbang26_3V2.csv' already exists. ignoring
2023-01-08T15:41:09.616 ocfl::(*ObjectBase).AddFile [objectbase.go:602] > INFO - adding file content:~blä blubb[in]/bangbang_names.csv
2023-01-08T15:41:09.616 ocfl::(*ObjectBase).AddFile [objectbase.go:656] > INFO - [id:blah-blubb] '~blä blubb[in]/bangbang_names.csv' already exists. ignoring
2023-01-08T15:41:09.616 ocfl::(*ObjectBase).AddFile [objectbase.go:602] > INFO - adding file content:~blä blubb[in]/bla.pptx
2023-01-08T15:41:09.617 ocfl::(*ObjectBase).AddFile [objectbase.go:656] > INFO - [id:blah-blubb] '~blä blubb[in]/bla.pptx' already exists. ignoring
2023-01-08T15:41:09.617 ocfl::(*ObjectBase).AddFile [objectbase.go:602] > INFO - adding file content:~blä blubb[in]/salon.json
2023-01-08T15:41:09.627 ocfl::(*ObjectBase).AddFile [objectbase.go:656] > INFO - [id:blah-blubb] '~blä blubb[in]/salon.json' already exists. ignoring
2023-01-08T15:41:09.627 ocfl::(*ObjectBase).AddFile [objectbase.go:602] > INFO - adding file content:~blä blubb[in]/sizecalculation.xlsx
2023-01-08T15:41:09.627 ocfl::(*ObjectBase).AddFile [objectbase.go:656] > INFO - [id:blah-blubb] '~blä blubb[in]/sizecalculation.xlsx' already exists. ignoring
2023-01-08T15:41:09.627 ocfl::(*ObjectBase).Close [objectbase.go:453] > INFO - Closing object 'id:blah-blubb'
2023-01-08T15:41:09.629 ocfl::(*InventoryBase).DeleteFile [inventorybase.go:898] > INFO - [id:blah-blubb] removing '2017-07-25_20-15-18_980.jpeg' from state
2023-01-08T15:41:09.629 ocfl::(*InventoryBase).DeleteFile [inventorybase.go:898] > INFO - [id:blah-blubb] removing '~blä blubb[in]/2017-07-25_20-15-18_980 - Kopie.jpeg' from state

no errors found
2023-01-08T15:41:13.086 cmd::doUpdate.func1 [update.go:145] > INFO - Duration: 6.3676128s
```

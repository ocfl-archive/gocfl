# Single Object OCFL Storage Root
Single Object Storage Root is mainly used with serialized containers.

## Creation
The `create` command is a combination of the `init` and `add` commands in one step. 

All Examples use the same [config file](../config/gocfl.toml). Logging is set to `INFO`

### Example: ZIP file with encrypted copy
#### Creation
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
```
#### Validation of OCFL structure
```
PS C:\daten\go\dev\gocfl> ../bin/gocfl.exe validate C:/temp/ocfl_create.zip --config ./config/gocfl.toml
Using config file: ./config/gocfl.toml
2023-01-07T17:19:42.818 cmd::validate [validate.go:46] > INFO - validating 'C:/temp/ocfl_create.zip'
2023-01-07T17:19:42.819 ocfl::(*StorageRootBase).Check [storagerootbase.go:391] > INFO - StorageRoot with version '1.1' found
object folder 'id=u003Ablah-blubb'
2023-01-07T17:19:42.820 ocfl::(*ObjectBase).Check [objectbase.go:991] > INFO - object 'id:blah-blubb' with object version '1.1' found

no errors found
2023-01-07T17:19:47.981 cmd::validate.func1 [validate.go:44] > INFO - Duration: 5.1621485s
```
#### Validation of ZIP checksum using linux standard tools
```
$ sha512sum -c ocfl_create.zip.sha512
ocfl_create.zip: OK
```
#### Validation of encrypted ZIP file using openssl
1) Decrypt file with provided key and initialisation vector and generate checksum
2) Compare with provided ZIP file checksum
```
$ openssl enc -aes-256-ctr -nosalt -d -in ocfl_create.zip.aes -out - -K "`cat ocfl_create.zip.aes.key`" -iv "`cat ocfl_create.zip.aes.iv`" | sha512sum.exe -
71e506bbfaa65b84d27736538d2db5f6d1809ef7d2061b099adfc59320fbeb376b8113725434079a23da298befe1486408a00a0bc11ecca139181f9d350fe20e *-

$ cat ocfl_create.zip.sha512
71e506bbfaa65b84d27736538d2db5f6d1809ef7d2061b099adfc59320fbeb376b8113725434079a23da298befe1486408a00a0bc11ecca139181f9d350fe20e *ocfl_create.zip
```
#### Validation of encrypted ZIP file with checksum
```
$ sha512sum -c ocfl_create.zip.aes.sha512
ocfl_create.zip.aes: OK
```
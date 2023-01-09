# Extract

Extract is used for OCFL Object content extraction. Target filenames are named according to the Inventory 
state. While extractions, the Inventory Manifest digests are checked. 
Optionally a `shaXXXsum`  compatible manifest file is written to the output folder.

```
PS C:\daten\go\dev\gocfl> ../bin/gocfl.exe extract --help
extract version of ocfl content

Usage:
  gocfl extract [path to ocfl structure] [path to target folder] [flags]

Examples:
gocfl extract ./archive.zip /tmp/archive

Flags:
      --ext-NNNN-content-subpath-area string   subpath for extraction (default: 'content'). 'all' for complete extraction
      --ext-NNNN-metafile-target string        url with metadata target folder
  -h, --help                                   help for extract
  -i, --object-id string                       object id to extract
  -p, --object-path string                     object path to extract
      --version string                         version to extract (default "latest")
      --with-manifest                          generate manifest file in object extraction folder

Global Flags:
      --config string                 config file (default is $HOME/.gocfl.toml)
      --log-file string               log output file (default is console)
      --log-level string              log level (CRITICAL|ERROR|WARNING|NOTICE|INFO|DEBUG) (default "ERROR")
      --s3-access-key-id string       Access Key ID for S3 Buckets
      --s3-endpoint string            Endpoint for S3 Buckets
      --s3-secret-access-key string   Secret Access Key for S3 Buckets
```

# Examples

## All Objects including manifest

```
PS C:\daten\go\dev\gocfl> ../bin/gocfl.exe extract C:\temp\ocflroot c:/temp/ocflextract --config ./config/gocfl.toml --with-manifest
Using config file: ./config/gocfl.toml
2023-01-09T18:15:06.184 cmd::doExtract [extract.go:69] > INFO - creating 'C:/temp/ocflroot'
extraction done without errors

[storage root 'file://C:/temp/ocflroot']
   #W013 - ‘In an OCFL Object, extension sub-directories SHOULD be named according to a registered extension name.’ [extension 'NNNN-direct-clean-path-layout' is not registered]
   #W013 - ‘In an OCFL Object, extension sub-directories SHOULD be named according to a registered extension name.’ [extension 'NNNN-direct-path-layout' is not registered]
   #W013 - ‘In an OCFL Object, extension sub-directories SHOULD be named according to a registered extension name.’ [extension 'NNNN-gocfl-extension-manager' is not registered]

[object 'file://C:/temp/ocflroot/id=u003Ablah-blubb' - '']
   #W013 - ‘In an OCFL Object, extension sub-directories SHOULD be named according to a registered extension name.’ [extension 'NNNN-direct-clean-path-layout' is not registered]
   #W013 - ‘In an OCFL Object, extension sub-directories SHOULD be named according to a registered extension name.’ [extension 'NNNN-gocfl-extension-manager' is not registered]

no errors found
2023-01-09T18:15:10.426 cmd::doExtract.func1 [extract.go:67] > INFO - Duration: 4.2419016s
```

The resulting manifest.sha512 looks like this:
```
d57a8d5b8714fe9d7e7991e76541bac525b3eae412ed0f1753522620e03a56c50ed51d0e422202f66fb6925b57363ba6f7ba3f14edf5d1a85f351c5c48e38a50 ~blä blubb[in]/bangbang_names.csv
e803e5c53d76f24e261455a9515effbc868f75fd6edd4afa7c4c2712419c7615cc41bd8fd9ddd80d4e24ae5760ef87aad20312ebd999f2add4c3ad204a71d966 convert.sh
7b97552ee0de278b5d5b9b85a060e1443cfaa195ba910303ce861a804110f73d68464084b2bf3cb1794f39f09075d8f3cae264eb27a325aeaa019c8ca2a565ba ~blä blubb[in]/bangbang26_3V2.csv
9858c7b4232ae3b9c1c5117b16bc8f207c218e0a591540520af1b9c2a52be539a3663a9a894727e995586b5f58876b786e5e1bedf7d6c66161ab1d499c082a2d collageX.png
a4538e32f44a602ad217dc1bff40859ebb90f527b70ac991fb35dfb4f40784699dae003c330b47db813dd55254fdc9801f6ca91f8f92d528638fcc2fdd02d66a ~blä blubb[in]/bla.pptx
ba9b8346dd0a642411fd8671d8eb64c4e31851f4846e35a5177096a6a6d3e1a30392e6c6353e8aff1f5d2a28f1a5ed660fc5071b17042bc9d109cc55b976cba9 salon.json
ba9b8346dd0a642411fd8671d8eb64c4e31851f4846e35a5177096a6a6d3e1a30392e6c6353e8aff1f5d2a28f1a5ed660fc5071b17042bc9d109cc55b976cba9 ~blä blubb[in]/salon.json
c02e3ae78068bd5a02e425995337cf699e6c3e6146f74c76028cad4130154e8590e65cdc95e5950b7feaa4be275da4cf7948299711032c14b03098e7b26d3b9b ~blä blubb[in]/bangbang.csv
f07aa4bfcc0d2def4013022a499813e0507ae109afb308db66152adbe1b586fdb0f5dc77054de3c0d041c4a6727b3e27ac816469d02091449d24a9c6aeb41562 bangbang_0226_10078_naegelin_2015_prisoners_dilemma_model_plusminus_staged.mp4--web_master.mp4
0dc7a8414ffe8ff4c0189091207e839d337412545401959cd0aa640043112941af00b79bcfa6b98f2b5f0236862774b76644eb5e105483e3e1591e54382319ea together.png
0dc7a8414ffe8ff4c0189091207e839d337412545401959cd0aa640043112941af00b79bcfa6b98f2b5f0236862774b76644eb5e105483e3e1591e54382319ea together - Kopie.png
5cb8c60eb3c7641561df988493acdd0fbc6b6325ec396a6eaf6a9cbc329e1790b006d61b4465371c21a105b0fb5a77dff9a219ed57ead6cd074d6b8a6e2be896 2017-07-25_20-15-18_980 - Kopie.jpeg
5cb8c60eb3c7641561df988493acdd0fbc6b6325ec396a6eaf6a9cbc329e1790b006d61b4465371c21a105b0fb5a77dff9a219ed57ead6cd074d6b8a6e2be896 2017-07-25_20-15-18_980XXX.jpeg
77b918ce6a21852e49c47ed01409b3ff53362f4a06f2e771583a86111faa2403ba694945697a9625cbc03fc2fde41e7758003d7677d591b566e4bf95cae7918c DSC_0111.JPG
6c9fec902ab8ab76c4e8a7c1943d9b235f6d7f1232bb3ae5dfbf2e6c68e93054315647f1c257d401486b421c3eb7eeaecf27e0c4ec675cfe3f35a30bf82374e2 ~blä blubb[in]/Modulhandbuch_MA_Gestaltung.pdf
b49156ca2b5b71758d04a44581f904d737ef707c9a21343b7d293b821c6439b4611740877edaa2b53552b184df0d5928eb109af05e8e517f2b953946f2f6c8f6 ~blä blubb[in]/sizecalculation.xlsx
ef9360a4df757648fbe392f81a2cde82d00a379a0b9c992d6b93b165f105e6347c40c889db0be9a672b79254e918386add3f01ed6a030ffe1816f1dd85f446fb Kopie von bangbang 26_3_gemacht_V2.xlsx
```
To check the written content files 
```
je@LAPTOP-KOTV7LO0:/mnt/c/temp/ocflextract/id=u003Ablah-blubb$ sha512sum -c manifest.sha512
~blä blubb[in]/bangbang_names.csv: OK
convert.sh: OK
~blä blubb[in]/bangbang26_3V2.csv: OK
collageX.png: OK
~blä blubb[in]/bla.pptx: OK
salon.json: OK
~blä blubb[in]/salon.json: OK
~blä blubb[in]/bangbang.csv: OK
bangbang_0226_10078_naegelin_2015_prisoners_dilemma_model_plusminus_staged.mp4--web_master.mp4: OK
together.png: OK
together - Kopie.png: OK
2017-07-25_20-15-18_980 - Kopie.jpeg: OK
2017-07-25_20-15-18_980XXX.jpeg: OK
DSC_0111.JPG: OK
~blä blubb[in]/Modulhandbuch_MA_Gestaltung.pdf: OK
~blä blubb[in]/sizecalculation.xlsx: OK
Kopie von bangbang 26_3_gemacht_V2.xlsx: OK
```
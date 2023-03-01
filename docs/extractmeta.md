# Extractmeta

Extractmeta is used for OCFL Object content metadata extraction. In Addition to the Inventory data, 
extensions, which provide additional metadata are integrated.

```
PS C:\daten\go\dev\gocfl> ../bin/gocfl.exe extractmeta --help
extract metadata from ocfl structure

Usage:
  gocfl extractmeta [path to ocfl structure] [flags]

Examples:
gocfl extractmeta ./archive.zip --output-json ./archive_meta.json

Flags:
  -h, --help                 help for extractmeta
  -i, --object-id string     object id to extract
  -p, --object-path string   object path to extract
      --output-json string   path to json file with metadata
      --version string       version to extract (default "latest")

Global Flags:
      --config string                 config file (default is $HOME/.gocfl.toml)
      --log-file string               log output file (default is console)
      --log-level string              log level (CRITICAL|ERROR|WARNING|NOTICE|INFO|DEBUG) (default "ERROR")
      --s3-access-key-id string       Access Key ID for S3 Buckets
      --s3-endpoint string            Endpoint for S3 Buckets
      --s3-region string              Region for S3 Access
      --s3-secret-access-key string   Secret Access Key for S3 Buckets
      --with-indexer                  starts indexer as a local service

```

# Examples

## Write Metadata into json file

```
PS C:\daten\go\dev\gocfl\build> .\gocfl_windows_amd64.exe extractmeta c:/temp/ocfl_create.zip --config ../config/gocfl.toml --object-id 'id:blah-blubb' --output-json 'c:/temp/ocfl_create.json'
Using config file: ../config/gocfl.toml
2023-02-24T16:53:38.620 cmd::doExtractMeta [extractmeta.go:69] > INFO - extracting metadata from 'c:/temp/ocfl_create.zip'
metadata extraction done without errors

[storage root 'zipfs://']
   #W013 - ‘In an OCFL Object, extension sub-directories SHOULD be named according to a registered extension name.’ [extension 'NNNN-direct-clean-path-layout' is not registered]
   #W013 - ‘In an OCFL Object, extension sub-directories SHOULD be named according to a registered extension name.’ [extension 'NNNN-direct-path-layout' is not registered]
   #W013 - ‘In an OCFL Object, extension sub-directories SHOULD be named according to a registered extension name.’ [extension 'NNNN-gocfl-extension-manager' is not registered]

[object 'zipfs://id=u003Ablah-blubb' - '']
   #W013 - ‘In an OCFL Object, extension sub-directories SHOULD be named according to a registered extension name.’ [extension 'NNNN-content-subpath' is not registered]
   #W013 - ‘In an OCFL Object, extension sub-directories SHOULD be named according to a registered extension name.’ [extension 'NNNN-direct-clean-path-layout' is not registered]
   #W013 - ‘In an OCFL Object, extension sub-directories SHOULD be named according to a registered extension name.’ [extension 'NNNN-gocfl-extension-manager' is not registered]
   #W013 - ‘In an OCFL Object, extension sub-directories SHOULD be named according to a registered extension name.’ [extension 'NNNN-indexer' is not registered]

no errors found
2023-02-24T16:53:38.630 cmd::doExtractMeta.func1 [extractmeta.go:67] > INFO - Duration: 10.1231ms
```

If the [NNNN-indexer](NNNN-indexer.md) extension is used, the resulting ocfl_create.json looks like this:
```json
{
  "Objects": {
    "id:blah-blubb": {
      "ID": "id:blah-blubb",
      "DigestAlgorithm": "sha512",
      "Versions": {
        "v1": {
          "Created": "2023-02-24T14:30:37+01:00",
          "Message": "Initial commit",
          "Name": "Jürgen Enge",
          "Address": "mailto:email@host"
        },
        "v2": {
          "Created": "2023-02-24T14:47:39+01:00",
          "Message": "Update",
          "Name": "Jürgen Enge",
          "Address": "mailto:email@host"
        }
      },
      "Files": {
        "0dc7a8414ffe8ff4c0189091207e839d337412545401959cd0aa640043112941af00b79bcfa6b98f2b5f0236862774b76644eb5e105483e3e1591e54382319ea": {
          "Checksums": {
            "blake2b-384": "55f51e13c159062a3304f6700640be1da375fef157dec5d5c0ef80a6ebd4e36fcc27a9446f74b95ca3b32a1cbddc8fca",
            "md5": "b70affb13a953a4c0b3e7ec8fcfeac5f",
            "sha256": "fcfee4deae4fd5f9d0a6469450a02bc2b943b2df458b8b54aae6a1b382b1feda"
          },
          "InternalName": [
            "v1/content/data/together.png"
          ],
          "VersionName": {
            "v1": [
              "data/together.png"
            ],
            "v2": [
              "data/together.png",
              "data/together - Kopie.png"
            ]
          },
          "Extension": {
            "NNNN-indexer": {
              "errors": {},
              "height": 1080,
              "identify": {
                "image": {
                  "backgroundColor": "#FFFFFFFFFFFF",
                  "baseDepth": 8,
                  "baseType": "Undefined",
                  "borderColor": "#DFDFDFDFDFDF",
                  "channelDepth": {
                    "blue": 1,
                    "green": 8,
                    "red": 8
                  },
                  "channelStatistics": {
                    "blue": {
                      "entropy": 0.339137,
                      "kurtosis": 4.04701,
                      "max": 255,
                      "mean": 218.234,
                      "median": 237,
                      "min": 0,
                      "skewness": -2.38196,
                      "standardDeviation": 72.7394
                    },
                    "green": {
                      "entropy": 0.384724,
                      "kurtosis": 3.53214,
                      "max": 254,
                      "mean": 202.311,
                      "median": 211,
                      "min": 0,
                      "skewness": -2.20083,
                      "standardDeviation": 68.8518
                    },
                    "red": {
                      "entropy": 0.383274,
                      "kurtosis": 3.41873,
                      "max": 254,
                      "mean": 195.948,
                      "median": 209,
                      "min": 0,
                      "skewness": -2.1486,
                      "standardDeviation": 67.0297
                    }
                  },
                  "chromaticity": {
                    "bluePrimary": {
                      "x": 0.15,
                      "y": 0.06
                    },
                    "greenPrimary": {
                      "x": 0.3,
                      "y": 0.6
                    },
                    "redPrimary": {
                      "x": 0.64,
                      "y": 0.33
                    },
                    "whitePrimary": {
                      "x": 0.3127,
                      "y": 0.329
                    }
                  },
                  "class": "DirectClass",
                  "colorspace": "sRGB",
                  "compose": "Over",
                  "compression": "Zip",
                  "depth": 8,
                  "dispose": "Undefined",
                  "elapsedTime": "0:01.125",
                  "endianness": "Undefined",
                  "filesize": "645553B",
                  "format": "PNG",
                  "formatDescription": "Portable Network Graphics",
                  "gamma": 0.45455,
                  "geometry": {
                    "height": 1080,
                    "width": 1920,
                    "x": 0,
                    "y": 0
                  },
                  "imageStatistics": {
                    "Overall": {
                      "entropy": 0.369045,
                      "kurtosis": 3.43043,
                      "max": 255,
                      "mean": 205.498,
                      "median": 219,
                      "min": 0,
                      "skewness": -2.16984,
                      "standardDeviation": 69.5403
                    }
                  },
                  "intensity": "Undefined",
                  "interlace": "None",
                  "iterations": 0,
                  "matteColor": "#BDBDBDBDBDBD",
                  "mimeType": "image/png",
                  "name": "-",
                  "numberPixels": "2073600",
                  "orientation": "Undefined",
                  "pageGeometry": {
                    "height": 1080,
                    "width": 1920,
                    "x": 0,
                    "y": 0
                  },
                  "permissions": 666,
                  "pixels": 6220800,
                  "pixelsPerSecond": "16.5714MB",
                  "printSize": {
                    "x": 50.8071,
                    "y": 28.579
                  },
                  "properties": {
                    "date:create": "2023-02-24T13:30:42+00:00",
                    "date:modify": "2023-02-24T13:30:42+00:00",
                    "date:timestamp": "2023-02-24T13:30:42+00:00",
                    "png:IHDR.bit-depth-orig": "8",
                    "png:IHDR.bit_depth": "8",
                    "png:IHDR.color-type-orig": "2",
                    "png:IHDR.color_type": "2 (Truecolor)",
                    "png:IHDR.interlace_method": "0 (Not interlaced)",
                    "png:IHDR.width,height": "1920, 1080",
                    "png:cHRM": "chunk was found (see Chromaticity, above)",
                    "png:gAMA": "gamma=0.45455 (See Gamma, above)",
                    "png:pHYs": "x_res=3779, y_res=3779, units=1",
                    "png:sRGB": "intent=0 (Perceptual Intent)",
                    "signature": "adb42d7242ba1041007f5f2dad9d5d047d192dc34d2e53f1d4dfee5ec361f095"
                  },
                  "renderingIntent": "Perceptual",
                  "resolution": {
                    "x": 37.79,
                    "y": 37.79
                  },
                  "tainted": false,
                  "transparentColor": "#000000000000",
                  "type": "TrueColor",
                  "units": "PixelsPerCentimeter",
                  "userTime": "0.063u",
                  "version": "ImageMagick 7.1.0-57 Q16 x64 eadf378:20221230 https://imagemagick.org"
                },
                "version": "1.0"
              },
              "mimetype": "image/png",
              "mimetypes": [
                "image/png"
              ],
              "siegfried": [
                {
                  "Basis": [
                    "extension match png",
                    "byte match at [[0 16] [37 4] [645541 12]] (signature 3/3)"
                  ],
                  "ID": "fmt/12",
                  "MIME": "image/png",
                  "Name": "Portable Network Graphics",
                  "Namespace": "pronom",
                  "Version": "1.1",
                  "Warning": ""
                }
              ],
              "size": 645553,
              "width": 1920
            }
          }
        },
        "22c45b78f78f193a837a9245e3f440d3043af325e91595e28476a02ec396bfd2f08822d0fb49ed372083c2fb764dcae800229976b11c08b93c23ba7f2ec7ffd1": {
          "Checksums": {
            "blake2b-384": "e5ccbac0a912c27c3a163a1689fda486b311aa3d0c5f117bc60d2bfcadb4a574af95e425746858f462fb0f7c0df4ab3b",
            "md5": "3a5256bf071e3e880ae8fb23d62d8e4a",
            "sha256": "394f9bacce4959ece535fc64ced4eeb1038174e83f98e98bbf274211cc66fb8a"
          },
          "InternalName": [
            "v1/content/data/bangbang_0044_00244_2021-0701_kakophonie-bild-def.mp4--web_master.mp4"
          ],
          "VersionName": {
            "v1": [
              "data/bangbang_0044_00244_2021-0701_kakophonie-bild-def.mp4--web_master.mp4"
            ],
            "v2": []
          },
          "Extension": {
            "NNNN-indexer": {
              "duration": 119,
              "errors": {},
              "ffprobe": {
                "format": {
                  "Filename": "C:\\temp\\ocfltest0\\bangbang_0044_00244_2021-0701_kakophonie-bild-def.mp4--web_master.mp4",
                  "bit_rate": "2220297",
                  "duration": "119.084000",
                  "format_long_name": "QuickTime / MOV",
                  "format_name": "mov,mp4,m4a,3gp,3g2,mj2",
                  "nb_programs": 0,
                  "nb_streams": 2,
                  "probe_score": 100,
                  "size": "33050243",
                  "tags": {
                    "compatible_brands": "isomiso2avc1mp41",
                    "encoder": "Lavf58.38.100",
                    "major_brand": "isom",
                    "minor_version": "512"
                  }
                },
                "streams": [
                  {
                    "Index": 0,
                    "avg_frame_rate": "24/1",
                    "bit_rate": "2089299",
                    "chroma_location": "left",
                    "codec_long_name": "H.264 / AVC / MPEG-4 AVC / MPEG-4 part 10",
                    "codec_name": "h264",
                    "codec_tag": "0x31637661",
                    "codec_tag_string": "avc1",
                    "codec_time_base": "",
                    "codec_type": "video",
                    "coded_height": 720,
                    "coded_width": 1280,
[...]
```

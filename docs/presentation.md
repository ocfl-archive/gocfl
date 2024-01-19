# Default

## Create test dataset
```bash
gocfl create c:/temp/ocfl/ocfl_test0.zip c:/temp/ocfl/sip/payload/ --object-id "id:abc_123" -m "Initial Commit" -u "Juergen Enge" -a "mailto:juergen@info-age.net"
```

## Validate it
```bash
gocfl validate c:/temp/ocfl/ocfl_test0.zip 
```

## Extract Metadata
```bash
gocfl extractmeta c:/temp/ocfl/ocfl_test0.zip --output c:/temp/ocfl/ocfl_test0.json
```

## Look inside
```bash
gocfl display c:/temp/ocfl/ocfl_test0.zip 
```

# Extended (Extension) Version 

## Create test dataset
```bash
cd /temp/ocfl
gocfl create c:/temp/ocfl/ocfl_test1.zip c:/temp/ocfl/sip/payload metadata:c:/temp/ocfl/sip/meta --config c:/temp/ocfl/gocfl2.toml --ext-NNNN-metafile-source file://C:/temp/ocfl/sip/info.json --object-id "id:abc_123" -m "Initial Commit" -u "Juergen Enge" -a "mailto:juergen@info-age.net"
```

## Validate it
```bash
gocfl validate c:/temp/ocfl/ocfl_test1.zip --config c:/temp/ocfl/gocfl2.toml 
```

## Extract Metadata
```bash
gocfl extractmeta c:/temp/ocfl/ocfl_test1.zip --output c:/temp/ocfl/ocfl_test1.json
```

## Look inside
```bash
gocfl display c:/temp/ocfl/ocfl_test1.zip --config c:/temp/ocfl/gocfl2.toml 
```

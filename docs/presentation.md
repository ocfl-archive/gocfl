# Default

## Create test dataset
```bash
../build/gocfl_windows_amd64.exe create c:/temp/test/ocfl_test0.zip c:/temp/test/data/ --object-id "id:abc_123" -m "Initial Commit" -u "Juergen Enge" -a "mailto:juergen@info-age.net"
```

## Validate it
```bash
../build/gocfl_windows_amd64.exe validate c:/temp/test/ocfl_test0.zip 
```

## Extract Metadata
```bash
../build/gocfl_windows_amd64.exe extractmeta c:/temp/test/ocfl_test0.zip --output c:/temp/test/ocfl_test0.json
```

## Look inside
```bash
../build/gocfl_windows_amd64.exe display c:/temp/test/ocfl_test0.zip 
```

# Extended (Extension) Version 

## Create test dataset
```bash
../build/gocfl_windows_amd64.exe create c:/temp/test/ocfl_test1.zip --config ../config/gocfl.toml c:/temp/test/data/ --object-id "id:abc_123" -m "Initial Commit" -u "Juergen Enge" -a "mailto:juergen@info-age.net"
```

## Validate it
```bash
../build/gocfl_windows_amd64.exe validate c:/temp/test/ocfl_test1.zip --config ../config/gocfl.toml 
```

## Extract Metadata
```bash
../build/gocfl_windows_amd64.exe extractmeta c:/temp/test/ocfl_test1.zip --output c:/temp/test/ocfl_test1.json
```

## Look inside
```bash
../build/gocfl_windows_amd64.exe display c:/temp/test/ocfl_test1.zip --config ../config/gocfl.toml 
```

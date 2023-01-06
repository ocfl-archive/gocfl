# Facts & Figures for large object example

A complete digital museal exhibition will be used as large test object. 
The Exhibition contains mainly of PNG images and H.264 movies. This means, that there's no
space for good compression in a ZIP file.

```
Size:         326 GB (350’808’857’649 Bytes)
Size on Disk: 326 GB (350’996’570’112 Bytes)
Content:      91’774 Files, 5’842 Folder
```

## Device for testing
```
Lenovo Laptop
Processor:        11th Gen Intel(R) Core(TM) i7-11850H @ 2.50GHz   2.50 GHz
Memory:           64.0 GB
Operating System: Windows 11 Pro
OS Version:       22H2
Disk:             SSD
```

## Ingest

### Command
```
gocfl. exe create C:/temp/ocfl_create.zip C:/temp/bangbang metadata:C:/temp/standorte --config ./config/gocfl.toml -i "id:blah-blubb" --ext-NNNN-metafile-source "file:///C:/temp/WhatsApp Bild 2022-12-11 um 15.25.47.jpg"
```
This command will result in a ZIP File and an encrypted ZIP File. Both files will have a sidecar 
with SHA512 checksum and there will be a sidecar file with encryption key and one with encryption
initialisation vector (hex encoded).

This command generates six checksums
* SHA512 for Manifest
* MD5,SHA256,BLAKE2b-384 for Fixity
* SHA512 for ZIP File
* SHA512 for encrypted ZIP File

and encrypts the emerging ZIP with AES256 concurrently

### [Config File](largeobject.toml)

### Extensions

The [default extensions](../data/defaultextensions) for Storage Root and Objects are used in this case.

#### Ressource Usage

CPU Usage distributes nicely over the different CPU Cores. Six cores have more load because of the 
checksum generation.  
![CPU Monitor](largeobject_cpu.png)

Disk I/O is quite bad on this laptop, but it is clear, that CPU is not the limiting factor on
ingest.  
Memory Usage will stay under 200MB for the whole time.
![Task Manager](largeobject_taskmanager.png)

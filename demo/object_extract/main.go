package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"

	"github.com/je4/filesystem/v4/pkg/vfsrw"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v3/pkg/appendfs"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/initocfl"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/storageroot"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
	"github.com/rs/zerolog"
)

func main() {
	// 1. Parse command line parameters
	pathPtr := flag.String("path", "", "Path to the OCFL Storage Root")
	destPtr := flag.String("dest", "", "Destination path for extraction")
	idPtr := flag.String("id", "my-object-id", "OCFL Object ID")
	flag.Parse()

	if *pathPtr == "" || *destPtr == "" {
		fmt.Println("Usage: go run main.go -path <storage_root_path> -dest <destination_path> [-id <object_id>]")
		os.Exit(1)
	}
	srPath := *pathPtr
	destPath := *destPtr
	objID := *idPtr

	// 2. Setup Environment
	ctx := context.Background()
	out := zerolog.ConsoleWriter{Out: os.Stderr}
	zlogger := zerolog.New(out)
	var _zlogger zLogger.ZLogger = &zlogger
	logger := ocfllogger.NewOCFLLogger(ctx, &zlogger, nil, version.Version1_1, nil)

	cfg := vfsrw.Config{
		"local": &vfsrw.VFS{
			Name: "local",
			Type: "os",
			OS: &vfsrw.OS{
				BaseDir: srPath,
			},
		},
	}
	vfs, err := vfsrw.NewFS(cfg, _zlogger)
	if err != nil {
		log.Fatalf("failed to create vfs: %v", err)
	}
	defer vfs.Close()

	fsys := appendfs.FS(vfs)

	// 3. Setup Factories
	ocflVer := version.Version1_1
	_, srExtFactory, _ := initocfl.SetupExtensionManager[storageroot.ExtensionManager](nil, nil, logger)
	srFactory := initocfl.NewFactoryStorageRoot(ocflVer, srExtFactory, logger)

	objExtManager, objExtFactory, _ := initocfl.SetupExtensionManager[object.ExtensionManager](nil, nil, logger)
	objFactory := initocfl.NewFactoryObject(ocflVer, objExtFactory, logger)

	// 4. Initialize Storage Root and Create an Object with content (if not exists)
	sr := srFactory.NewStorageRoot(ctx).
		WithWriteFS(fsys).
		WithDigestAlgorithm(checksum.DigestSHA512)
	_ = sr.GetInitializer().Init()

	objFolder, _ := sr.IdToFolder(objID)
	objFS, _ := appendfs.Sub(fsys, objFolder)
	obj := objFactory.NewObject(ctx).WithExtensionManager(objExtManager)

	// Try to initialize. If it fails, we assume it exists and we'll try to load it.
	if err := obj.GetInitializer().WithFS(objFS).Init(objID, checksum.DigestSHA512, nil); err == nil {
		vw, _ := obj.StartUpdate(objFS, "initial version", "GOCFL", "mailto:ocfl@ocflarchive", false)
		_ = vw.AddData([]byte("Hello OCFL extraction"), "extract-me.txt", false, "", false, false)
		_ = vw.Close()
	}

	// --- Extracting Objects ---

	// Prepare destination filesystem (local)
	destCfg := vfsrw.Config{
		"dest": &vfsrw.VFS{
			Name: "dest",
			Type: "os",
			OS: &vfsrw.OS{
				BaseDir: destPath,
			},
		},
	}
	destVFS, err := vfsrw.NewFS(destCfg, _zlogger)
	if err != nil {
		log.Fatalf("failed to create destination vfs: %v", err)
	}
	defer destVFS.Close()
	destFS := appendfs.FS(destVFS)

	// Reload the object for reading
	readObjFS := fs.FS(objFS)
	loader := obj.GetLoader().WithFS(readObjFS)
	if err := loader.Load(); err != nil {
		log.Fatalf("failed to load object: %v", err)
	}

	// Get extractor
	extractor := obj.GetExtractor().WithFS(readObjFS, destFS)

	// Extract the head version
	err = extractor.Extract(nil, false, "")
	if err != nil {
		log.Fatalf("failed to extract object: %v", err)
	}

	fmt.Printf("Object '%s' (head version) from '%s' successfully extracted to '%s'.\n", objID, srPath, destPath)

	// Verify extraction
	_, err = fs.Stat(destFS, "extract-me.txt")
	if err != nil {
		fmt.Printf("Verification failed: %v\n", err)
	} else {
		fmt.Println("Verification: 'extract-me.txt' found in destination.")
	}
}

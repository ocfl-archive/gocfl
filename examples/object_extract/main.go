package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"

	"github.com/je4/filesystem/v4/pkg/vfsrw"
	"github.com/je4/filesystem/v4/pkg/writefs"
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v3/pkg/appendfs"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/initocfl"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
	"github.com/rs/zerolog"
)

func main() {
	// --- Step 1: Command Line Parameter Parsing ---
	// Define and parse the parameters for the OCFL Storage Root and the Object to be extracted.
	pathPtr := flag.String("path", "", "Path to the OCFL Storage Root")
	destPtr := flag.String("dest", "", "Destination path for extraction")
	idPtr := flag.String("id", "my-object-id", "OCFL Object ID")
	objectPathPtr := flag.String("objectpath", "", "Path to the OCFL Object")
	flag.Parse()

	if (*pathPtr == "" || *destPtr == "") && *objectPathPtr == "" {
		fmt.Println("Usage: go run main.go -path <storage_root_path> -dest <destination_path> [-id <object_id>]")
		fmt.Println("   or: go run main.go -objectpath <object_path> -dest <destination_path> [-id <object_id>]")
		os.Exit(1)
	}
	srPath := *pathPtr
	destPath := *destPtr
	objID := *idPtr

	// --- Step 2: Logging Infrastructure Setup ---
	// Initialize a context and a logger using zerolog and the OCFL-specific logger wrapper.
	ctx := context.Background()
	out := zerolog.ConsoleWriter{Out: os.Stderr}
	zlogger := zerolog.New(out)
	var _zlogger zLogger.ZLogger = &zlogger
	logger := ocfllogger.NewOCFLLogger(ctx, &zlogger, nil, version.Version1_1, nil)

	// --- Step 3: Virtual Filesystem (VFS) Configuration ---
	// OCFL operations are performed via a filesystem abstraction layer.
	cfg := vfsrw.Config{}
	vfs, err := vfsrw.NewFS(cfg, _zlogger)
	if err != nil {
		log.Fatalf("failed to create vfs: %v", err)
	}
	defer vfs.Close()

	// Register the local filesystem (OS-specific root) to the VFS.
	if err := vfsrw.AddLocal(vfs, nil); err != nil {
		logger.Fatal().Err(err).Msg("failed to add local filesystem")
	}

	// Define the filesystem for the Storage Root or Object directory.
	// We use appendfs here because it is required for OCFL operations,
	// providing the necessary functionality to access the OCFL structure.
	var storageRootFS fs.FS
	var objFolder string
	if *objectPathPtr != "" {
		objFolder = writefs.RealPath(vfs, *objectPathPtr)
	} else {
		storageRootFS, err = fs.Sub(vfs, writefs.RealPath(vfs, *pathPtr))
		if err != nil {
			log.Fatalf("failed to create subfs for storage root '%s': %v", *pathPtr, err)
		}
	}

	// --- Step 4: OCFL Version and Loading ---
	ocflVer := version.Version1_1

	// --- Step 5: Storage Root and Object Loading ---
	// If a Storage Root is provided, we determine the Object's folder.
	if storageRootFS != nil {
		sr, err := initocfl.LoadStorageRoot(ctx, storageRootFS, ocflVer, logger)
		if err != nil {
			logger.Fatal().Err(err).Msgf("failed to load storage root at '%v'", storageRootFS)
		}

		// Determine the folder path for the given Object ID within the Storage Root.
		objFolder, err = sr.IdToFolder(objID)
		if err != nil {
			log.Fatalf("failed to get folder for id '%s': %v", objID, err)
		}
		// Prepend storage root path to objFolder to get the absolute path.
		objFolder = writefs.RealPath(vfs, *pathPtr+"/"+objFolder)
	}

	// Create a sub-filesystem for the target object directory.
	objFS, err := fs.Sub(vfs, objFolder)
	if err != nil {
		logger.Fatal().Err(err).Msgf("failed to create subfs for object folder '%s'", objFolder)
	}

	// Load the existing object.
	obj, err := initocfl.LoadObject(ctx, objFS, ocflVer, logger)
	if err != nil {
		log.Fatalf("failed to load object '%s' at '%s': %v", objID, objFolder, err)
	}

	// --- Step 6: Extracting the OCFL Object ---
	// The objFS is used twice here for different purposes:
	// 1. In Step 5, the Loader used objFS to read the OCFL metadata (inventory, sidecar) to understand the object's structure.
	// 2. Here in Step 6, the Extractor uses objFS as the source to access the actual content files stored within the OCFL versions.

	// A: Prepare destination filesystem (local) where the object will be extracted to.
	destRealPath := writefs.RealPath(vfs, *destPtr)
	destFS, err := appendfs.Sub(vfs, destRealPath)
	if err != nil {
		log.Fatalf("failed to create destination fs: %v", err)
	}

	// B: Get the extractor for the object and specify the source (objFS) and destination (destFS) filesystems.
	extractor := obj.GetExtractor().WithDestFS(destFS)

	// C: Extract the head version of the object to the destination.
	err = extractor.Extract(nil, false, "")
	if err != nil {
		log.Fatalf("failed to extract object: %v", err)
	}

	fmt.Printf("Object '%s' (head version) from '%s' successfully extracted to '%s'.\n", objID, srPath, destPath)

	// --- Step 7: Verification (Optional) ---
	// In a real scenario, you would check if the extracted files exist in destPath.
	fmt.Println("Extraction completed.")
}

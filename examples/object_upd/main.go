package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/filesystem/pkg/appendfs"
	"github.com/ocfl-archive/filesystem/pkg/vfsrw"
	"github.com/ocfl-archive/filesystem/pkg/writefs"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/initocfl"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/util"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
	"github.com/rs/zerolog"
)

func main() {
	// --- Step 1: Command Line Parameter Parsing ---
	// Define and parse the parameters for the OCFL Storage Root and the Object to be updated.
	pathPtr := flag.String("path", "", "Path to the OCFL Storage Root")
	idPtr := flag.String("id", "my-object-id", "OCFL Object ID")
	objectPathPtr := flag.String("objectpath", "", "Path to the OCFL Object")
	flag.Parse()

	if (*pathPtr == "" || *idPtr == "") && *objectPathPtr == "" {
		fmt.Println("Usage: go run main.go -path <storage_root_path> -id <object_id>")
		fmt.Println("   or: go run main.go -objectpath <object_path> [-id <object_id>]")
		os.Exit(1)
	}

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
	// We use appendfs here because it is required for write operations in OCFL,
	// providing the necessary functionality to append to or create files within the OCFL structure.
	var storageRootFS appendfs.FS
	var objFolder string
	if *objectPathPtr != "" {
		objFolder = writefs.RealPath(vfs, *objectPathPtr)
	} else {
		srPath := writefs.RealPath(vfs, *pathPtr)
		storageRootFS, err = appendfs.Sub(vfs, srPath)
		if err != nil {
			log.Fatalf("failed to create subfs for storage root '%s': %v", srPath, err)
		}
	}

	// --- Step 4: OCFL Determination ---
	if storageRootFS != nil {
		// A: If an existing Storage Root exists, we read the OCFL version from it.
		_, err = util.GetStorageRootVersion(storageRootFS)
		if err != nil {
			logger.Fatal().Err(err).Msg("failed to get storage root version")
		}
	}

	// --- Step 5: Object Path Determination ---
	var objID = *idPtr
	if storageRootFS != nil {
		// A: Load existing Storage Root.
		sr, srCloser, err := initocfl.LoadStorageRoot(ctx, storageRootFS, nil, nil, logger)
		if err != nil {
			logger.Fatal().Err(err).Msgf("failed to load storage root at '%v'", storageRootFS)
		}
		defer srCloser.Close()

		// B: Determine the folder path for the given Object ID within the Storage Root.
		objFolder, err = sr.IdToFolder(objID)
		if err != nil {
			log.Fatalf("failed to get folder for id '%s': %v", objID, err)
		}
		// C: Prepend storage root path to objFolder to get the absolute path.
		objFolder = writefs.RealPath(vfs, *pathPtr+"/"+objFolder)
	}

	// --- Step 6: OCFL Object Loading ---
	// A: Create a sub-filesystem for the target object directory.
	// Again, appendfs is used to enable write operations within this sub-filesystem.
	objFS, err := appendfs.Sub(vfs, objFolder)
	if err != nil {
		logger.Fatal().Err(err).Msgf("failed to create subfs for object folder '%s'", objFolder)
	}

	// B: Instantiate and configure the Object.
	obj, objCloser, err := initocfl.LoadObject(ctx, objFS, nil, logger)
	if err != nil {
		log.Fatalf("failed to load object '%s' at '%s': %v", objID, objFolder, err)
	}
	defer objCloser.Close()

	fmt.Printf("OCFL Object '%s' successfully loaded from folder '%s'.\n", objID, objFolder)

	// --- Step 7: Update the Object with a New Version ---
	// A: Start a new version update for the object.
	// This returns a VersionWriter which allows adding, deleting, or renaming files.
	vw, err := obj.StartUpdate("Updating object: renaming, removing, and adding files", "GOCFL", "mailto:ocfl@ocflarchive", false)
	if err != nil {
		log.Fatalf("failed to start update for object '%s': %v", objID, err)
	}

	// B: Rename an existing file (from object_add example).
	// RenameFile(source, destination, digest) - passing empty string for digest to match by path.
	if err := vw.RenameFile("file1.txt", "README.txt", ""); err != nil {
		log.Fatalf("failed to rename file: %v", err)
	}
	fmt.Println("Renamed: file1.txt -> README.txt")

	// C: Remove an existing file.
	// DeleteFile(path, digest) - passing empty string for digest to match by path.
	if err := vw.DeleteFile("docs/note.md", ""); err != nil {
		log.Fatalf("failed to delete file: %v", err)
	}
	fmt.Println("Removed: docs/note.md")

	// D: Add a new file and update an existing one.
	filesToUpdate := map[string]string{
		"notes/summary.txt": "This is a new summary file added in the second version.",
		"config.json":       `{"status": "active", "version": 2, "updated": true}`,
	}

	for path, content := range filesToUpdate {
		// AddData(data, path, checkDuplicate, area, noExtensionHook, isDir)
		err = vw.AddData([]byte(content), path, true, "", false, false)
		if err != nil {
			log.Fatalf("failed to add/update file '%s': %v", path, err)
		}
		fmt.Printf("Added/Updated file: %s\n", path)
	}

	// E: Close the VersionWriter to commit the changes and finalize the new version.
	if err := vw.Close(); err != nil {
		log.Fatalf("failed to close version writer: %v", err)
	}

	vErrors := logger.ValidationErrors()
	if len(vErrors) > 0 {
		fmt.Printf("\nValidation issues found for object '%s':\n", objID)
		for _, vErr := range vErrors {
			fmt.Printf("- %s\n", vErr.Error())
		}
		fmt.Println()
	}

	fmt.Printf("Successfully updated OCFL Object '%s' to a new version.\n", objID)
}

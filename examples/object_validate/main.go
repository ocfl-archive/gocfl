package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"

	"github.com/je4/utils/v2/pkg/zLogger"
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
	ctx := context.Background()
	out := zerolog.ConsoleWriter{Out: os.Stderr}
	zlogger := zerolog.New(out)
	var _zlogger zLogger.ZLogger = &zlogger
	logger := ocfllogger.NewOCFLLogger(ctx, &zlogger, nil, version.Version1_1, nil)

	// --- Step 3: Virtual Filesystem (VFS) Configuration ---
	cfg := vfsrw.Config{}
	vfs, err := vfsrw.NewFS(cfg, _zlogger)
	if err != nil {
		log.Fatalf("failed to create vfs: %v", err)
	}
	defer vfs.Close()

	if err := vfsrw.AddLocal(vfs, nil); err != nil {
		logger.Fatal().Err(err).Msg("failed to add local filesystem")
	}

	var storageRootFS fs.FS
	var objFolder string
	if *objectPathPtr != "" {
		objFolder = writefs.RealPath(vfs, *objectPathPtr)
	} else {
		srPath := writefs.RealPath(vfs, *pathPtr)
		// No appendfs is needed here because we are only reading the storage root to find the object.
		storageRootFS, err = fs.Sub(vfs, srPath)
		if err != nil {
			log.Fatalf("failed to create subfs for storage root '%s': %v", srPath, err)
		}
	}

	// --- Step 4: OCFL Determination ---
	if storageRootFS != nil {
		_, err = util.GetStorageRootVersion(storageRootFS)
		if err != nil {
			logger.Fatal().Err(err).Msg("failed to get storage root version")
		}
	}

	// --- Step 5: Object Path Determination ---
	var objID = *idPtr
	if storageRootFS != nil {
		sr, srCloser, err := initocfl.LoadStorageRoot(ctx, storageRootFS, nil, nil, logger)
		if err != nil {
			logger.Fatal().Err(err).Msgf("failed to load storage root at '%v'", storageRootFS)
		}
		defer srCloser.Close()

		objFolder, err = sr.IdToFolder(objID)
		if err != nil {
			log.Fatalf("failed to get folder for id '%s': %v", objID, err)
		}
		objFolder = writefs.RealPath(vfs, *pathPtr+"/"+objFolder)
	}

	// --- Step 6: OCFL Object Loading ---
	// No appendfs is needed here because validation is a read-only operation.
	// appendfs is only required when we need to add new files or versions to an existing OCFL object.
	objFS, err := fs.Sub(vfs, objFolder)
	if err != nil {
		log.Fatalf("failed to create sub fs for object folder '%s': %v", objFolder, err)
	}

	obj, objCloser, err := initocfl.LoadObject(ctx, objFS, nil, logger)
	if err != nil {
		log.Fatalf("failed to load object '%s' at '%s': %v", objID, objFolder, err)
	}
	defer objCloser.Close()

	fmt.Printf("OCFL Object '%s' successfully loaded from folder '%s'.\n", objID, objFolder)

	// --- Step 7: Object Validation ---
	fmt.Printf("Validating object '%s'...\n", objID)
	checker := obj.GetChecker()
	err = checker.Check()

	vErrors := logger.ValidationErrors()
	if len(vErrors) > 0 {
		fmt.Printf("\nValidation issues found for object '%s':\n", objID)
		for _, vErr := range vErrors {
			fmt.Printf("- %s\n", vErr.Error())
		}
		fmt.Println()
	}

	if err != nil {
		fmt.Printf("Validation failed for object '%s': %v\n", objID, err)
		os.Exit(1)
	}

	fmt.Printf("OCFL Object '%s' is valid.\n", objID)

}

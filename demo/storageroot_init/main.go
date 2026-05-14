package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/je4/filesystem/v4/pkg/vfsrw"
	"github.com/je4/filesystem/v4/pkg/writefs"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v3/pkg/appendfs"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/initocfl"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/storageroot"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
	"github.com/rs/zerolog"
)

func main() {
	// --- Step 1: Command Line Parameter Parsing ---
	// Define and parse the path where the OCFL Storage Root will be initialized.
	pathPtr := flag.String("path", "", "Path to the OCFL Storage Root")
	flag.Parse()

	if *pathPtr == "" {
		fmt.Println("Usage: go run main.go -path <storage_root_path>")
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

	// Determine the absolute path and create a sub-filesystem for the Storage Root directory.
	// We use appendfs here because it is required for write operations in OCFL,
	// providing the necessary functionality to append to or create files within the OCFL structure.
	srPath := writefs.RealPath(vfs, *pathPtr)
	storageRootFS, err := appendfs.Sub(vfs, srPath)
	if err != nil {
		log.Fatalf("failed to create subfs of %v for folder '%s': %v", vfs, srPath, err)
	}

	// --- Step 4: OCFL Storage Root Initialization ---
	// Define the target OCFL version.
	ocflVer := version.Version1_1

	// A: Setup the Extension Manager to handle OCFL extensions.
	extManager, extFactory, err := initocfl.SetupExtensionManager[storageroot.ExtensionManager](nil, nil, logger)
	if err != nil {
		log.Fatalf("failed to create extension factory: %v", err)
	}

	// B: Create a Storage Root factory for the specified OCFL version.
	srFactory := initocfl.NewFactoryStorageRoot(ocflVer, extFactory, logger)

	// C: Instantiate and configure the Storage Root object.
	// We associate it with our filesystem, extension manager, and preferred digest algorithm.
	sr := srFactory.NewStorageRoot(ctx).
		WithWriteFS(storageRootFS).
		WithExtensionManager(extManager).
		WithDigestAlgorithm(checksum.DigestSHA512)

	// D: Execute the initialization process.
	// This creates the necessary OCFL structure (e.g., ocfl_layout.json, namaste file) on disk.
	initializer := sr.GetInitializer()
	err = initializer.Init()
	if err != nil {
		log.Fatalf("failed to initialize storage root at '%s': %v", srPath, err)
	}

	fmt.Printf("OCFL Storage Root successfully initialized at '%s'.\n", srPath)
}

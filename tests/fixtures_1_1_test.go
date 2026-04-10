package tests

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/ocfl-archive/gocfl/v2/gocfl/cmd"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/functions"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfllogger"
	"github.com/rs/zerolog"
)

func TestFixtures11(t *testing.T) {
	fixtureRoot := "../../../fixtures/1.1"
	absFixtureRoot, err := filepath.Abs(fixtureRoot)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(absFixtureRoot); os.IsNotExist(err) {
		t.Skipf("fixtures not found at %s", absFixtureRoot)
	}

	subdirs := []string{"bad-objects", "warn-objects", "good-objects"}

	reCode := regexp.MustCompile(`[WE]\d{3}`)

	for _, subdir := range subdirs {
		dirPath := filepath.Join(absFixtureRoot, subdir)
		entries, err := os.ReadDir(dirPath)
		if err != nil {
			t.Errorf("cannot read dir %s: %v", dirPath, err)
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			t.Run(filepath.Join(subdir, entry.Name()), func(t *testing.T) {
				expectedCodes := reCode.FindAllString(entry.Name(), -1)
				path := filepath.Join(dirPath, entry.Name())

				// Setup components
				ctx := validation.NewContextValidation(context.Background())
				logger := ocfllogger.NewOCFLLogger(ctx, new(zerolog.New(os.Stderr)), nil, version.Default)

				err := cmd.RegisterComplexExtensions(
					map[string]string{},
					"",
					false,
					nil,
					nil,
					nil,
					logger,
				)
				if err != nil {
					logger.Error().Err(err).Msg("cannot create extension factory")
					return
				}

				/*
					storageRootExtensionManager, objectExtensionManager, err := cmd.InitDefaultExtensions(version.Version1_1, extensionFactory, "extensions", "extensions", logger)
					if err != nil {
						logger.Error().Err(err).Msg("cannot initialize default extensions")
						return
					}
					defer func() {
						if err := objectExtensionManager.Terminate(); err != nil {
							logger.Error().Err(err).Msg("cannot terminate object extension manager")
						}
						if err := storageRootExtensionManager.Terminate(); err != nil {
							logger.Error().Err(err).Msg("cannot terminate storage root extension manager")
						}
					}()

				*/

				fsys := os.DirFS(path)

				// Run validation
				err = functions.CheckObject(ctx, fsys, extensionFactory, logger)
				// We ignore the error from CheckObject as we are interested in the validation status
				if err != nil {
					t.Logf("CheckObject returned error (expected for bad-objects): %v", err)
				}

				status, err := validation.GetValidationStatus(ctx)
				if err != nil {
					t.Fatalf("cannot get validation status: %v", err)
				}

				foundCodes := make(map[string]bool)
				for _, vErr := range status.Errors {
					foundCodes[string(vErr.Code)] = true
				}

				// If we have no expected codes, it's a good object or at least we don't expect specific errors
				for _, expected := range expectedCodes {
					if !foundCodes[expected] {
						var found []string
						for c := range foundCodes {
							found = append(found, c)
						}
						t.Errorf("expected code %s not found. Found codes: %v", expected, found)
					}
				}

				if subdir == "good-objects" {
					hasError := false
					var found []string
					for _, vErr := range status.Errors {
						// Filter out W000 and E001 (which is noise due to missing extensions folder in these fixtures)
						if vErr.Code != "W000" && vErr.Code != "E001" && (vErr.Code[0] == 'E' || vErr.Code[0] == 'W') {
							hasError = true
							found = append(found, string(vErr.Code))
						}
					}
					if hasError {
						t.Errorf("good object should have no errors/warnings (except W000/E001), but found: %v", found)
					}
				}
			})
		}
	}
}

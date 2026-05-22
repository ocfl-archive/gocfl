//nolint:all
package inventoryimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v3/pkg/initocfl"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
	"github.com/rs/zerolog"
)

type fileData struct {
	stateFilename    []string
	manifestFilename string
	checksumData     map[checksum.DigestAlgorithm]string
}

type inventoryTest struct {
	testData          string
	fileData          fileData
	modified          bool
	getChecksum       string
	expectedFiles     []string
	expectedFilesFlat []string
	// Brittle and prone to break with i18n but might be useful in
	// these early tests to discuss.
	manifestText  string
	expectedError error
}

// Basic inventory that can be partially loaded into memory. Ideally
// we would do more with the version and fixity field but they're not
// used that this point in the code.
//
// The field : "versions": {"v1": {
// cannot be loaded into memory at this point but I could be
// initializing a portion of the code incorrectly.
//
// Only Manifest is updated in this test.
var inventoryEx1 = `
{
   "id": "ex:1",
   "type": "https://ocfl.io/1.1/spec/#inventory",
   "digestAlgorithm": "sha512",
   "head": "v1",
   "manifest": {
      "1a873f7ee602d253faa48a3fb64dde4202194611264f5f33541daaec4400a808c2100b90cda575c1f1a8b0465fa1f8d8c8c4336decd6845d26fbb904060937ea": [
         "v1/content/test.file1"
      ],
      "2a873f7ee602d253faa48a3fb64dde4202194611264f5f33541daaec4400a808c2100b90cda575c1f1a8b0465fa1f8d8c8c4336decd6845d26fbb904060937ea": [
         "v1/content/test.file2"
      ]
   },
   "fixity": {}
}
`

var inventoryTests = []inventoryTest{
	// Adds duplicate with different filename. 3 files, 2 unique.
	{
		testData: inventoryEx1,
		fileData: fileData{
			stateFilename:    []string{"add.file1"},
			manifestFilename: "v1/content/add.file1",
			checksumData: map[checksum.DigestAlgorithm]string{
				checksum.DigestSHA512: "1a873f7ee602d253faa48a3fb64dde4202194611264f5f33541daaec4400a808c2100b90cda575c1f1a8b0465fa1f8d8c8c4336decd6845d26fbb904060937ea",
			},
		},
		modified:          true,
		getChecksum:       "1a873f7ee602d253faa48a3fb64dde4202194611264f5f33541daaec4400a808c2100b90cda575c1f1a8b0465fa1f8d8c8c4336decd6845d26fbb904060937ea",
		expectedFiles:     []string{"v1/content/test.file1", "v1/content/add.file1"},
		expectedFilesFlat: []string{"v1/content/test.file2", "v1/content/test.file1", "v1/content/add.file1"},
		manifestText:      "3 files (2 unique)",
	},
	// Adds new filename filename. 3 files, 3 unique.
	{
		testData: inventoryEx1,
		fileData: fileData{
			stateFilename:    []string{"add.file1"},
			manifestFilename: "v1/content/add.file1",
			checksumData: map[checksum.DigestAlgorithm]string{
				checksum.DigestSHA512: "3a873f7ee602d253faa48a3fb64dde4202194611264f5f33541daaec4400a808c2100b90cda575c1f1a8b0465fa1f8d8c8c4336decd6845d26fbb904060937ea",
			},
		},
		modified:          true,
		getChecksum:       "3a873f7ee602d253faa48a3fb64dde4202194611264f5f33541daaec4400a808c2100b90cda575c1f1a8b0465fa1f8d8c8c4336decd6845d26fbb904060937ea",
		expectedFiles:     []string{"v1/content/add.file1"},
		expectedFilesFlat: []string{"v1/content/test.file2", "v1/content/test.file1", "v1/content/add.file1"},
		manifestText:      "3 files (3 unique)",
	},
	// Tries to add an identical file with same checksum and name.
	// 2 files 2 unique.
	{
		testData: inventoryEx1,
		fileData: fileData{
			stateFilename:    []string{"test.file1"},
			manifestFilename: "v1/content/test.file1",
			checksumData: map[checksum.DigestAlgorithm]string{
				checksum.DigestSHA512: "1a873f7ee602d253faa48a3fb64dde4202194611264f5f33541daaec4400a808c2100b90cda575c1f1a8b0465fa1f8d8c8c4336decd6845d26fbb904060937ea",
			},
		},
		modified:          false,
		getChecksum:       "1a873f7ee602d253faa48a3fb64dde4202194611264f5f33541daaec4400a808c2100b90cda575c1f1a8b0465fa1f8d8c8c4336decd6845d26fbb904060937ea",
		expectedFiles:     []string{"v1/content/test.file1"},
		expectedFilesFlat: []string{"v1/content/test.file2", "v1/content/test.file1"},
		manifestText:      "2 files (2 unique)",
	},
}

func addFileTest(t *testing.T, test inventoryTest) {
	ctx := context.TODO()
	var zlogger zLogger.ZLogger = new(zerolog.New(os.Stderr).With().Str("timestamp", time.Now().String()).Logger())
	var logger = initocfl.NewOCFLLogger(ctx, zlogger, nil, version.Default, nil)

	// Initial state.
	testState := &stateBase{
		State: map[string][]string{
			"1a873f7ee602d253faa48a3fb64dde4202194611264f5f33541daaec4400a808c2100b90cda575c1f1a8b0465fa1f8d8c8c4336decd6845d26fbb904060937ea": []string{"test.file1"},
			"2a873f7ee602d253faa48a3fb64dde4202194611264f5f33541daaec4400a808c2100b90cda575c1f1a8b0465fa1f8d8c8c4336decd6845d26fbb904060937ea": []string{"test.file2"},
		},
	}

	testVersion := &versionBase{
		Created: inventory.NewOCFLTime(time.Now()),
		Message: inventory.NewOCFLString(""),
		State:   testState,
	}

	testVersions := &versionsBase{
		// It is unclear why it isn't possible to load the JSON version
		// into memory so we do it manually.
		versions:    map[int]inventory.Version{0: testVersion},
		versionInts: map[string]int{"v1": 0},
	}

	testInventory := InventoryBase{
		Manifest: &ManifestBase{
			manifest: map[string][]string{},
		},
		Versions: testVersions,
		Fixity: &FixityBase{
			fixity: map[checksum.DigestAlgorithm]map[string][]string{},
		},
	}

	err := json.Unmarshal([]byte(test.testData), &testInventory)
	if err != nil {
		t.Fatal("error unmarshalling JSON from test:", err)
	}

	testInventory.logger = logger
	err = testInventory.AddFile(
		test.fileData.stateFilename,
		test.fileData.manifestFilename,
		test.fileData.checksumData,
	)
	if err != nil {
		t.Errorf("error is not nil: %s", err)
	}

	if testInventory.modified != test.modified {
		t.Errorf(
			"modified state incorrectly updated as '%t', expected: '%t'",
			testInventory.modified,
			test.modified,
		)
	}

	resManifest := testInventory.Manifest
	if fmt.Sprintf("%s", resManifest) != test.manifestText {
		t.Errorf(
			"didn't receive the correct manifest text: '%s' expectedd '%s'",
			resManifest,
			test.manifestText,
		)
	}

	files, err := resManifest.GetFiles(test.getChecksum)
	if err != nil {
		t.Errorf("retrieving from manifest should be nil: %s", err)
	}

	if !reflect.DeepEqual(files, test.expectedFiles) {
		t.Errorf(
			"manifest is incorrect for checksum, retrieved '%s', expected '%s'",
			files,
			test.expectedFiles,
		)
	}

	fileList := resManifest.GetFilesFlat()
	for file := range fileList {
		if !slices.Contains(test.expectedFilesFlat, strings.TrimSpace(string(file))) {
			log.Println(file)
			t.Errorf(
				"file '%s' not in expected file list, expected '%s'",
				file,
				test.expectedFilesFlat,
			)
		}
	}
}

func TestAddFile(t *testing.T) {
	for _, test := range inventoryTests {
		addFileTest(t, test)
	}
}

var inventoryEx2 = `
{
   "id": "ex:1",
   "type": "https://ocfl.io/1.1/spec/#inventory",
   "digestAlgorithm": "sha512",
   "head": "v1",
   "manifest": {
      "1a873f7ee602d253faa48a3fb64dde4202194611264f5f33541daaec4400a808c2100b90cda575c1f1a8b0465fa1f8d8c8c4336decd6845d26fbb904060937ea": [
         "v1/content/test.file1"
      ],
      "2a873f7ee602d253faa48a3fb64dde4202194611264f5f33541daaec4400a808c2100b90cda575c1f1a8b0465fa1f8d8c8c4336decd6845d26fbb904060937ea": [
         "v1/content/test.file2"
      ]
   },
   "fixity": {}
}
`

var inventoryEx3 = `
{
   "id": "ex:1",
   "type": "https://ocfl.io/1.1/spec/#inventory",
   "digestAlgorithm": "md5",
   "head": "v1",
   "manifest": {
      "1a873f7ee602d253faa48a3fb64dde4202194611264f5f33541daaec4400a808c2100b90cda575c1f1a8b0465fa1f8d8c8c4336decd6845d26fbb904060937ea": [
         "v1/content/test.file1"
      ],
      "2a873f7ee602d253faa48a3fb64dde4202194611264f5f33541daaec4400a808c2100b90cda575c1f1a8b0465fa1f8d8c8c4336decd6845d26fbb904060937ea": [
         "v1/content/test.file2"
      ]
   },
   "fixity": {}
}
`

const DigestFake checksum.DigestAlgorithm = "UnknownDigest"

var inventoryTestsChecksumError = []inventoryTest{
	// Uses an unknown checksum type. NB. also an invalid checkum string.
	{
		testData: inventoryEx2,
		fileData: fileData{
			stateFilename:    []string{""},
			manifestFilename: "",
			checksumData: map[checksum.DigestAlgorithm]string{
				DigestFake: "checksum123",
			},
		},
		modified:      false,
		expectedError: errors.New("no digest for 'sha512' in checksums"),
	},
	// Tries to add SHA512 checksum to MD5 manifest. NB. also an
	// invalid checkum string.
	{
		testData: inventoryEx3,
		fileData: fileData{
			stateFilename:    []string{""},
			manifestFilename: "",
			checksumData: map[checksum.DigestAlgorithm]string{
				checksum.DigestSHA512: "checksumSHA512",
			},
		},
		modified:      false,
		expectedError: errors.New("no digest for 'md5' in checksums"),
	},
}

func addFileTestChecksumError(t *testing.T, test inventoryTest) {
	ctx := context.TODO()
	var zlogger zLogger.ZLogger = new(zerolog.New(os.Stderr).With().Str("timestamp", time.Now().String()).Logger())
	var logger = initocfl.NewOCFLLogger(ctx, zlogger, nil, version.Default, nil)

	// Initial state.
	testState := &stateBase{
		State: map[string][]string{
			"1a873f7ee602d253faa48a3fb64dde4202194611264f5f33541daaec4400a808c2100b90cda575c1f1a8b0465fa1f8d8c8c4336decd6845d26fbb904060937ea": []string{"test.file1"},
			"2a873f7ee602d253faa48a3fb64dde4202194611264f5f33541daaec4400a808c2100b90cda575c1f1a8b0465fa1f8d8c8c4336decd6845d26fbb904060937ea": []string{"test.file2"},
		},
	}

	testVersion := &versionBase{
		Created: inventory.NewOCFLTime(time.Now()),
		Message: inventory.NewOCFLString(""),
		State:   testState,
	}

	testVersions := &versionsBase{
		// It is unclear why it isn't possible to load the JSON version
		// into memory so we do it manually.
		versions:    map[int]inventory.Version{0: testVersion},
		versionInts: map[string]int{"v1": 0},
	}

	testInventory := InventoryBase{
		Manifest: &ManifestBase{
			manifest: map[string][]string{},
		},
		Versions: testVersions,
		Fixity: &FixityBase{
			fixity: map[checksum.DigestAlgorithm]map[string][]string{},
		},
	}

	err := json.Unmarshal([]byte(test.testData), &testInventory)
	if err != nil {
		t.Fatal("error unmarshalling JSON from test:", err)
	}

	testInventory.logger = logger
	err = testInventory.AddFile(
		test.fileData.stateFilename,
		test.fileData.manifestFilename,
		test.fileData.checksumData,
	)
	if err == nil {
		t.Fatal("error is nil, expected a checksum validation error")
	}

	if testInventory.modified != test.modified {
		t.Fatalf(
			"inventory modified incorrectly: '%t', expected: '%t'",
			testInventory.modified,
			test.modified,
		)
	}

	if err.Error() != test.expectedError.Error() {
		t.Fatalf("expectedd error '%s', got '%s'", err, test.expectedError)
	}
}

func TestAddFileChecksumErrors(t *testing.T) {
	for _, test := range inventoryTestsChecksumError {
		addFileTestChecksumError(t, test)
	}
}

var inventoryEx4 = `
{
   "id": "ex:1",
   "type": "https://ocfl.io/1.1/spec/#inventory",
   "digestAlgorithm": "sha512",
   "head": "v1",
   "manifest": {
      "1a873f7ee602d253faa48a3fb64dde4202194611264f5f33541daaec4400a808c2100b90cda575c1f1a8b0465fa1f8d8c8c4336decd6845d26fbb904060937ea": [
         "v1/content/test.file1"
      ],
      "2a873f7ee602d253faa48a3fb64dde4202194611264f5f33541daaec4400a808c2100b90cda575c1f1a8b0465fa1f8d8c8c4336decd6845d26fbb904060937ea": [
         "v1/content/test.file2"
      ]
   },
   "fixity": {}
}
`

var inventoryTestVersionErrors = []inventoryTest{
	// Try adding a duplicate file but the version doesn't exist in
	// the base inventory.
	{
		testData: inventoryEx1,
		fileData: fileData{
			stateFilename:    []string{"test.file2"},
			manifestFilename: "v1/content/test.file2",
			checksumData: map[checksum.DigestAlgorithm]string{
				checksum.DigestSHA512: "2a873f7ee602d253faa48a3fb64dde4202194611264f5f33541daaec4400a808c2100b90cda575c1f1a8b0465fa1f8d8c8c4336decd6845d26fbb904060937ea",
			},
		},
		modified: false,
		expectedError: errors.New(
			"cannot add for duplicate of '[test.file2]' [2a873f7ee602d253faa48a3fb64dde4202194611264f5f33541daaec4400a808c2100b90cda575c1f1a8b0465fa1f8d8c8c4336decd6845d26fbb904060937ea]: checksum for version  does not exist",
		),
	},
	// Checksum for a new version does not exist.
	{
		testData: inventoryEx1,
		fileData: fileData{
			stateFilename:    []string{"test.file2"},
			manifestFilename: "v1/content/test.file2",
			checksumData: map[checksum.DigestAlgorithm]string{
				checksum.DigestSHA512: "9a873f7ee602d253faa48a3fb64dde4202194611264f5f33541daaec4400a808c2100b90cda575c1f1a8b0465fa1f8d8c8c4336decd6845d26fbb904060937ea",
			},
		},
		// TODO: This looks like an invalid condition as we get an error
		// and it is marked as modified which it is.
		modified: true,
		expectedError: errors.New(
			"cannot add for duplicate of '[test.file2]' [9a873f7ee602d253faa48a3fb64dde4202194611264f5f33541daaec4400a808c2100b90cda575c1f1a8b0465fa1f8d8c8c4336decd6845d26fbb904060937ea]: checksum for version  does not exist",
		),
	},
}

func addFileTestVersionErrors(t *testing.T, test inventoryTest) {
	ctx := context.TODO()
	var zlogger zLogger.ZLogger = new(zerolog.New(os.Stderr).With().Str("timestamp", time.Now().String()).Logger())
	var logger = initocfl.NewOCFLLogger(ctx, zlogger, nil, version.Default, nil)

	// Initial state.
	testState := &stateBase{
		State: map[string][]string{
			"1a873f7ee602d253faa48a3fb64dde4202194611264f5f33541daaec4400a808c2100b90cda575c1f1a8b0465fa1f8d8c8c4336decd6845d26fbb904060937ea": []string{"test.file1"},
			"2a873f7ee602d253faa48a3fb64dde4202194611264f5f33541daaec4400a808c2100b90cda575c1f1a8b0465fa1f8d8c8c4336decd6845d26fbb904060937ea": []string{"test.file2"},
		},
	}

	vers := &inventory.VersionNumber{}
	vers.Init(99, "v1")

	testVersion := &versionBase{
		version: vers,
		Created: inventory.NewOCFLTime(time.Now()),
		Message: inventory.NewOCFLString(""),
		State:   testState,
	}

	testVersions := &versionsBase{
		// It is unclear why it isn't possible to load the JSON version
		// into memory so we do it manually.
		versions:    map[int]inventory.Version{0: testVersion},
		versionInts: map[string]int{"v1": 0},
	}

	testInventory := InventoryBase{
		Manifest: &ManifestBase{
			manifest: map[string][]string{},
		},
		Versions: testVersions,
		Fixity: &FixityBase{
			fixity: map[checksum.DigestAlgorithm]map[string][]string{},
		},
	}

	err := json.Unmarshal([]byte(test.testData), &testInventory)
	if err != nil {
		t.Fatal("error unmarshalling JSON from test:", err)
	}

	testInventory.logger = logger
	err = testInventory.AddFile(
		test.fileData.stateFilename,
		test.fileData.manifestFilename,
		test.fileData.checksumData,
	)
	if err == nil {
		t.Errorf("error is nil but we expect an error")
	}

	if err.Error() != test.expectedError.Error() {
		t.Errorf(
			"error '%s' is not expected error '%s'",
			err.Error(),
			test.expectedError.Error(),
		)
	}

	if testInventory.modified != test.modified {
		t.Errorf(
			"modified state incorrectly updated as '%t', expected: '%t'",
			testInventory.modified,
			test.modified,
		)
	}
}

func TestAddFileVersionErrors(t *testing.T) {
	for _, test := range inventoryTestVersionErrors {
		addFileTestVersionErrors(t, test)
	}
}

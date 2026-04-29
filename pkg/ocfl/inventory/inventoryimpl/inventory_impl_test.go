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

	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
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
	manifestText string
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
	zerologger := zerolog.New(os.Stderr).With().Str("timestamp", time.Now().String()).Logger()
	var zlogger zLogger.ZLogger = &zerologger
	var logger = ocfllogger.NewOCFLLogger(ctx, zlogger, nil, version.Default, nil)

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

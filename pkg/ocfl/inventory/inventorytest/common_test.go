package inventorytest

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/initocfl"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/validation"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
	"github.com/rs/zerolog"
)

var ctx = context.Background()
var logger = ocfllogger.NewOCFLLogger(
	ctx,
	new(zerolog.New(zerolog.NewConsoleWriter())),
	nil,
	version.Version1_1,
	nil,
)

func getFactory(ver version.OCFLVersion) inventory.Factory {
	return initocfl.NewFactoryObject(ver, nil, logger)
}

func genericExampleState(ver version.OCFLVersion, cnt int, t *testing.T) inventory.State {
	f := getFactory(ver)
	var state = f.NewState(context.Background())
	for i := 0; i < cnt; i++ {
		modified, err := state.AddFile(fmt.Sprintf("file%03d", i), fmt.Sprintf("digest%03d", i))
		if err != nil {
			t.Fatalf("Error adding file %d: %v", i, err)
		}
		if !modified {
			t.Fatalf("File %d has not modified state", i)
		}
		modified, err = state.AddFile(fmt.Sprintf("file%03dx", i), fmt.Sprintf("digest%03d", i))
		if err != nil {
			t.Fatalf("Error adding file%03dx: %v", i, err)
		}
		if !modified {
			t.Fatalf("file%03dc has not modified state", i)
		}
	}
	return state
}

func genericStateJSONMarshal(ver version.OCFLVersion, t *testing.T) {
	f := getFactory(ver)
	var cnt = 4
	var state = genericExampleState(ver, cnt, t)
	var num int
	var ps int
	for _, paths := range state.Iterate() {
		num++
		ps += len(paths)
	}
	if num != cnt {
		t.Errorf("Number of digests does not match expected number of files. Expected %d, got %d", cnt, num)
	}
	if ps != cnt*2 {
		t.Errorf("Number of files does not match expected number of files. Expected %d, got %d", cnt*2, num)
	}
	bytes, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("Error marshalling state: %v", err)
	}
	state2 := f.NewState(context.Background())
	if err := json.Unmarshal(bytes, state2); err != nil {
		t.Fatalf("Error unmarshalling state: %v", err)
	}
	if !reflect.DeepEqual(state, state2) {
		t.Errorf("States are not equal. Expected %v, got %v", state, state2)
	}
}

func genericStateJSONUnmarshal(ver version.OCFLVersion, t *testing.T) {
	f := getFactory(ver)
	var bytes = []byte(`{
  "digest000" : [ "file000", "file000x" ],
  "digest001" : [ "file001", "file001x" ],
  "digest002" : [ "file002", "file002x" ],
  "digest003" : [ "file003", "file003x" ]
}`)
	var state = f.NewState(context.Background())
	if err := json.Unmarshal(bytes, state); err != nil {
		t.Fatalf("Error unmarshalling state: %v", err)
	}
	var num int
	var ps int
	for _, paths := range state.Iterate() {
		num++
		ps += len(paths)
	}
	if num != 4 {
		t.Errorf("Number of digests does not match expected number of files. Expected %d, got %d", 4, num)
	}
	if ps != 4*2 {
		t.Errorf("Number of files does not match expected number of files. Expected %d, got %d", 4*2, num)
	}
}

func genericStateJSONUnmarshalError(ver version.OCFLVersion, t *testing.T) {
	f := getFactory(ver)
	var bytes = []byte(`{
  "digest000" : [ "file000", "file000x" ],
  "digest001" : [ "file001", "file001x" ],
  "digest002" : [ "file002", "file002x" ],
  "digest003" : "file003"
}`)
	var state = f.NewState(context.Background())
	if err := json.Unmarshal(bytes, state); err != nil {
		t.Errorf("Error unmarshalling state: %v", err)
	}
	if state.Err() == nil {
		t.Error("state.Err() should have returned an error")
	}
}

func genericStateDelete(ver version.OCFLVersion, t *testing.T) {
	f := getFactory(ver)
	var state = f.NewState(context.Background())
	state.AddFile("file1", "digest1")
	state.AddFile("file2", "digest1")
	state.AddFile("file3", "digest2")

	if _, err := state.DeleteFile("file1"); err != nil {
		t.Fatalf("Error deleting file1: %v", err)
	}
	if d := state.FileChecksum("file1"); d != "" {
		t.Errorf("file1 should have been deleted")
	}
	if d := state.FileChecksum("file2"); d != "digest1" {
		t.Errorf("file2 should still exist with digest1")
	}

	if _, err := state.DeleteFile("file2"); err != nil {
		t.Fatalf("Error deleting file2: %v", err)
	}
	if d := state.FileChecksum("file2"); d != "" {
		t.Errorf("file2 should have been deleted")
	}
	// digest1 should be gone now as no file points to it
	var found bool
	for d := range state.Iterate() {
		if d == "digest1" {
			found = true
			break
		}
	}
	if found {
		t.Errorf("digest1 should have been removed from state")
	}
}

func genericStateRename(ver version.OCFLVersion, t *testing.T) {
	f := getFactory(ver)
	var state = f.NewState(context.Background())
	state.AddFile("file1", "digest1")

	if _, err := state.RenameFile("file1", "file2"); err != nil {
		t.Fatalf("Error renaming file1 to file2: %v", err)
	}
	if d := state.FileChecksum("file1"); d != "" {
		t.Errorf("file1 should have been renamed")
	}
	if d := state.FileChecksum("file2"); d != "digest1" {
		t.Errorf("file2 should exist with digest1")
	}
}

func genericStateUniqueness(ver version.OCFLVersion, t *testing.T) {
	f := getFactory(ver)
	var state = f.NewState(context.Background())
	state.AddFile("file1", "digest1")
	state.AddFile("file2", "digest2")

	// adding same file with same digest -> no change
	modified, err := state.AddFile("file1", "digest1")
	if err != nil {
		t.Fatalf("Error adding file1: %v", err)
	}
	if modified {
		t.Errorf("Adding same file/digest should not result in modification")
	}

	// adding same file with different digest -> change
	modified, err = state.AddFile("file1", "digest2")
	if err != nil {
		t.Fatalf("Error adding file1: %v", err)
	}
	if !modified {
		t.Errorf("Adding same file with different digest should result in modification")
	}
	if d := state.FileChecksum("file1"); d != "digest2" {
		t.Errorf("file1 should have digest2")
	}

	// digest1 should be gone
	for d := range state.Iterate() {
		if d == "digest1" {
			t.Errorf("digest1 should have been removed")
		}
	}
}

// User Tests
func genericUserJSONMarshal(ver version.OCFLVersion, t *testing.T) {
	f := getFactory(ver)
	user := f.NewUser(context.Background()).WithName("Alice").WithAddress("mailto:alice@example.org")
	bytes, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("Failed to marshal user: %s", err)
	}
	user2 := f.NewUser(context.Background())
	if err := json.Unmarshal(bytes, user2); err != nil {
		t.Fatalf("Failed to unmarshal user: %s", err)
	}
	if !reflect.DeepEqual(user, user2) {
		t.Fatalf("Failed to unmarshal user - not equal")
	}
}

func genericUserJSONUnmarshal(ver version.OCFLVersion, t *testing.T) {
	f := getFactory(ver)
	var jsonData = []byte(`
{
	"address": "mailto:alice@example.org",
	"name": "Alice"
}
`)
	user := f.NewUser(context.Background())
	if err := json.Unmarshal(jsonData, user); err != nil {
		t.Errorf("unmarshal error: %v", err)
	}
	name := user.GetName()
	if name != "Alice" {
		t.Errorf("name error: %v", name)
	}
	address := user.GetAddress()
	if address != "mailto:alice@example.org" {
		t.Errorf("address error: %v", address)
	}
}

func genericUserInvalidJSON(ver version.OCFLVersion, t *testing.T) {
	f := getFactory(ver)
	var jsonData = []byte(`
{
	"address": 42,
	"name": "Alice"
}
`)
	user := f.NewUser(context.Background())
	if err := json.Unmarshal(jsonData, user); err != nil {
		t.Errorf("unmarshal error: %v", err)
	}
	if user.Err() == nil {
		t.Error("should get an error")
	}
}

func genericUserInvalidAddressCheck(ver version.OCFLVersion, t *testing.T) {
	f := getFactory(ver)
	user := f.NewUser(context.Background()).WithAddress("xxx")
	if err := user.Check(inventory.NewVersionNumber().WithString("v1")); err != nil {
		t.Fatalf("check error: %v", err)
	}
	vErrors := logger.ValidationErrors()
	var hasW009 bool
	for _, w := range vErrors {
		if w.Code == validation.W009 {
			hasW009 = true
			break
		}
	}
	if !hasW009 {
		t.Errorf("no warning '%s' in %v", validation.W009, vErrors)
	}
}

// Version Tests
func genericExampleVersion(ver version.OCFLVersion, stateFileCnt int, t *testing.T) inventory.Version {
	f := getFactory(ver)
	var user = f.NewUser(context.Background()).WithName("Test User").WithAddress("test@example.com")
	var state = f.NewState(context.Background())
	for i := 0; i < stateFileCnt; i++ {
		modified, err := state.AddFile(fmt.Sprintf("file%03d", i), fmt.Sprintf("digest%03d", i))
		if err != nil {
			t.Fatalf("Error adding file %d: %v", i, err)
		}
		if !modified {
			t.Fatalf("File %d has not modified state", i)
		}
	}
	var version = f.NewVersion(context.Background()).
		WithState(state).
		WithMessage("Test version message").
		WithUser(user).
		WithCreated(time.Now().UTC().Truncate(time.Second))
	return version
}

func genericVersionJSONMarshal(ver version.OCFLVersion, t *testing.T) {
	f := getFactory(ver)
	var cnt = 3
	var v = genericExampleVersion(ver, cnt, t)
	bytes, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("Error marshalling version: %v", err)
	}
	v2 := f.NewVersion(context.Background())
	if err := json.Unmarshal(bytes, v2); err != nil {
		t.Fatalf("Error unmarshalling version: %v", err)
	}
	if !reflect.DeepEqual(v, v2) {
		t.Errorf("versions are not equal. Expected %v, got %v", v, v2)
	}
}

func genericVersionJSONUnmarshal(ver version.OCFLVersion, t *testing.T) {
	f := getFactory(ver)
	var bytes = []byte(`{
  "created": "2024-01-15T10:30:00Z",
  "message": "Initial commit",
  "user": {
    "name": "John Doe",
    "address": "john@example.com"
  },
  "state": {
    "digest001": ["file001"],
    "digest002": ["file002"]
  }
}`)
	var v = f.NewVersion(context.Background())
	if err := json.Unmarshal(bytes, v); err != nil {
		t.Fatalf("Error unmarshalling version: %v", err)
	}
	created := v.GetCreated()
	expectedTime, _ := time.Parse(time.RFC3339, "2024-01-15T10:30:00Z")
	if !created.Equal(expectedTime) {
		t.Errorf("Created time mismatch. Expected %v, got %v", expectedTime, created)
	}
	if v.GetMessage() != "Initial commit" {
		t.Errorf("Message mismatch. Expected 'Initial commit', got '%s'", v.GetMessage())
	}
	if v.GetUser().GetName() != "John Doe" {
		t.Errorf("User name mismatch. Expected 'John Doe', got '%s'", v.GetUser().GetName())
	}
}

func genericVersionJSONUnmarshalError(ver version.OCFLVersion, t *testing.T) {
	f := getFactory(ver)
	var bytes = []byte(`{
  "created": "invalid-date",
  "message": "Test message",
  "state": {
    "digest001": "invalid-not-array"
  }
}`)
	var v = f.NewVersion(context.Background())
	if err := json.Unmarshal(bytes, v); err != nil {
		t.Errorf("Error unmarshalling version: %v", err)
	}
	if v.Err() == nil {
		t.Error("version.Err() should have returned an error")
	}
}

func genericVersionState(ver version.OCFLVersion, t *testing.T) {
	var v = genericExampleVersion(ver, 2, t)
	state := v.GetState()
	if state == nil {
		t.Fatalf("Version.GetState() should not have returned a nil state")
	}
	var num int
	for range state.Iterate() {
		num++
	}
	if num != 2 {
		t.Errorf("Number of files in state does not match. Expected 2, got %d", num)
	}
}

func genericVersionMessage(ver version.OCFLVersion, t *testing.T) {
	f := getFactory(ver)
	var v = f.NewVersion(context.Background()).WithMessage("Test message")
	if v.GetMessage() != "Test message" {
		t.Errorf("Message mismatch. Expected 'Test message', got '%s'", v.GetMessage())
	}
}

func genericVersionCreated(ver version.OCFLVersion, t *testing.T) {
	f := getFactory(ver)
	now := time.Now().UTC().Truncate(time.Second)
	var v = f.NewVersion(context.Background()).WithCreated(now)
	if !v.GetCreated().Equal(now) {
		t.Errorf("Created time mismatch. Expected %v, got %v", now, v.GetCreated())
	}
}

// Versions Tests
func genericVersionsJSONMarshal(ver version.OCFLVersion, t *testing.T) {
	f := getFactory(ver)
	versions := f.NewVersions(context.Background())
	v := f.NewVersion(context.Background()).WithMessage("test version")
	versions.SetVersion(inventory.NewVersionNumber().WithString("v1"), v)
	bytes, err := json.Marshal(versions)
	if err != nil {
		t.Fatalf("Failed to marshal versions: %s", err)
	}
	versions2 := f.NewVersions(context.Background())
	if err := json.Unmarshal(bytes, versions2); err != nil {
		t.Fatalf("Failed to unmarshal versions: %s", err)
	}
	if !versions.Equals(versions2) {
		t.Fatalf("Failed to unmarshal versions - not equal")
	}
}

func genericVersionsJSONUnmarshal(ver version.OCFLVersion, t *testing.T) {
	f := getFactory(ver)
	var jsonData = []byte(`
{
	"v1": {
		"created": "2023-01-01T12:00:00Z",
		"message": "initial version",
		"state": {},
		"user": {
			"name": "Alice",
			"address": "mailto:alice@example.org"
		}
	}
}
`)
	versions := f.NewVersions(context.Background())
	if err := json.Unmarshal(jsonData, versions); err != nil {
		t.Errorf("unmarshal error: %v", err)
	}
	v := versions.GetVersion(inventory.NewVersionNumber().WithString("v1"))
	if v == nil {
		t.Error("version v1 not found")
	} else if v.GetMessage() != "initial version" {
		t.Errorf("message error: %v", v.GetMessage())
	}
}

func genericVersionsInvalidJSON(ver version.OCFLVersion, t *testing.T) {
	f := getFactory(ver)
	var jsonData = []byte(`
{
	"v1": {
		"created": "invalid-date",
		"message": "test",
		"state": {},
		"user": {}
	}
}
`)
	versions := f.NewVersions(context.Background())
	if err := json.Unmarshal(jsonData, versions); err != nil {
		t.Errorf("unmarshal error: %v", err)
	}
	if versions.Err() == nil {
		t.Error("should get an error")
	}
}

func genericVersionsCheck(ver version.OCFLVersion, t *testing.T) {
	f := getFactory(ver)
	versions := f.NewVersions(context.Background())
	v := f.NewVersion(context.Background()).WithMessage("test")
	versions.SetVersion(inventory.NewVersionNumber().WithString("v1"), v)

	if err := versions.Check([]string{}); err != nil {
		t.Fatalf("check error: %v", err)
	}
	vErrors := logger.ValidationErrors()
	var hasE049 bool
	for _, e := range vErrors {
		if e.Code == validation.E049 {
			hasE049 = true
			break
		}
	}
	if hasE049 {
		t.Errorf("unexpected error '%s' in %v", validation.E049, vErrors)
	}
}

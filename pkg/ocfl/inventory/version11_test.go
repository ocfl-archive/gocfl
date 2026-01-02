package inventory

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/types"
)

func exampleVersion(stateFileCnt int, t *testing.T) types.Version {
	var user = ocfl.f11.NewUser().WithName("Test User").WithAddress("test@example.com")
	var state = ocfl.f11.NewState()
	for i := 0; i < stateFileCnt; i++ {
		modified, err := state.AddFile(fmt.Sprintf("file%03d", i), fmt.Sprintf("digest%03d", i))
		if err != nil {
			t.Fatalf("Error adding file %d: %v", i, err)
		}
		if !modified {
			t.Fatalf("File %d has not modified state", i)
		}
	}
	var version = ocfl.f11.NewVersion().
		WithState(state).
		WithMessage("Test version message").
		WithUser(user).
		WithCreated(time.Now().UTC().Truncate(time.Second))
	return version
}

func Test_VersionJSONMarshal(t *testing.T) {
	var cnt = 3
	var version = exampleVersion(cnt, t)

	bytes, err := json.Marshal(version)
	if err != nil {
		t.Fatalf("Error marshalling version: %v", err)
	}

	version2 := ocfl.f11.NewVersion()
	if err := json.Unmarshal(bytes, version2); err != nil {
		t.Fatalf("Error unmarshalling version: %v", err)
	}

	if !reflect.DeepEqual(version, version2) {
		t.Errorf("versions are not equal. Expected %v, got %v", version, version2)
	}
}

func Test_VersionJSONUnmarshal(t *testing.T) {
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
	var version = ocfl.f11.NewVersion()
	if err := json.Unmarshal(bytes, version); err != nil {
		t.Fatalf("Error unmarshalling version: %v", err)
	}

	created := version.GetCreated()
	expectedTime, _ := time.Parse(time.RFC3339, "2024-01-15T10:30:00Z")
	if !created.Equal(expectedTime) {
		t.Errorf("Created time mismatch. Expected %v, got %v", expectedTime, created)
	}

	message := version.GetMessage()
	if message != "Initial commit" {
		t.Errorf("Message mismatch. Expected 'Initial commit', got '%s'", message)
	}

	user := version.GetUser()
	name := user.GetName()
	if name != "John Doe" {
		t.Errorf("User name mismatch. Expected 'John Doe', got '%s'", name)
	}
}

func Test_VersionJSONUnmarshalError(t *testing.T) {
	var bytes = []byte(`{
  "created": "invalid-date",
  "message": "Test message",
  "state": {
    "digest001": "invalid-not-array"
  }
}`)
	var version = ocfl.f11.NewVersion()
	if err := json.Unmarshal(bytes, version); err != nil {
		t.Errorf("Error unmarshalling version: %v", err)
	}
	if version.Err() == nil {
		t.Error("version.Err() should have returned an error")
	}
}

func Test_VersionState(t *testing.T) {
	var version = exampleVersion(2, t)

	state := version.GetState()
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

func Test_VersionMessage(t *testing.T) {
	var version = ocfl.f11.NewVersion().WithMessage("Test message")

	message := version.GetMessage()
	if message != "Test message" {
		t.Errorf("Message mismatch. Expected 'Test message', got '%s'", message)
	}
}

func Test_VersionCreated(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	var version = ocfl.f11.NewVersion().WithCreated(now)

	created := version.GetCreated()
	if !created.Equal(now) {
		t.Errorf("Created time mismatch. Expected %v, got %v", now, created)
	}
}

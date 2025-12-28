package inventory

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
)

func Test_VersionsJSONMarshal(t *testing.T) {
	versions := f11.NewVersions()
	version := f11.NewVersion().WithMessage("test version")
	versions.AddVersion("v1", version)
	bytes, err := json.Marshal(versions)
	if err != nil {
		t.Fatalf("Failed to marshal versions: %s", err)
	}
	versions2 := f11.NewVersions()
	if err := json.Unmarshal(bytes, versions2); err != nil {
		t.Fatalf("Failed to unmarshal versions: %s", err)
	}
	if !reflect.DeepEqual(versions, versions2) {
		t.Fatalf("Failed to unmarshal versions - not equal")
	}
}

func Test_VersionsJSONUnmarshal(t *testing.T) {
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
	versions := f11.NewVersions()
	if err := json.Unmarshal(jsonData, versions); err != nil {
		t.Errorf("unmarshal error: %v", err)
	}
	version := versions.GetVersion("v1")
	if version == nil {
		t.Error("version v1 not found")
	}
	if version.GetMessage() != "initial version" {
		t.Errorf("message error: %v", version.GetMessage())
	}
}

func Test_VersionsInvalidJSON(t *testing.T) {
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
	versions := f11.NewVersions()
	if err := json.Unmarshal(jsonData, versions); err != nil {
		t.Errorf("unmarshal error: %v", err)
	}
	if versions.Err() == nil {
		t.Error("should get an error")
	}
}

func Test_VersionsCheck(t *testing.T) {
	val := NewDummyValidation()
	versions := f11.NewVersions()
	version := f11.NewVersion().WithMessage("test")
	versions.AddVersion("v1", version)

	if err := versions.Check(val, []string{}, []string{}); err != nil {
		t.Fatalf("check error: %v", err)
	}

	if val.HasError(validation.E049) {
		t.Errorf("unexpected error '%s' in %v", validation.E049, val.Error)
	}
}

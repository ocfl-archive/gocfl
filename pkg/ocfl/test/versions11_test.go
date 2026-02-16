package test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/factory/factoryimpl"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/inventory/inventoryimpl"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
	"github.com/rs/zerolog"
)

func Test_VersionsJSONMarshal(t *testing.T) {
	var logger = zerolog.New(zerolog.NewConsoleWriter())
	var f = factoryimpl.NewFactory(version.Version1_1, nil, nil, &logger)
	versions := f.NewVersions(context.Background())
	version := f.NewVersion(context.Background()).WithMessage("test version")
	versions.SetVersion(inventory.NewVersionNumber().WithString("v1"), version)
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

func Test_VersionsJSONUnmarshal(t *testing.T) {
	var logger = zerolog.New(zerolog.NewConsoleWriter())
	var f = factoryimpl.NewFactory(version.Version1_1, nil, nil, &logger)
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
	version := versions.GetVersion(inventory.NewVersionNumber().WithString("v1"))
	if version == nil {
		t.Error("version v1 not found")
	}
	if version.GetMessage() != "initial version" {
		t.Errorf("message error: %v", version.GetMessage())
	}
}

func Test_VersionsInvalidJSON(t *testing.T) {
	var logger = zerolog.New(zerolog.NewConsoleWriter())
	var f = factoryimpl.NewFactory(version.Version1_1, nil, nil, &logger)
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

func Test_VersionsCheck(t *testing.T) {
	var logger = zerolog.New(zerolog.NewConsoleWriter())
	var f = factoryimpl.NewFactory(version.Version1_1, nil, nil, &logger)
	val := inventoryimpl.NewDummyValidation()
	versions := f.NewVersions(context.Background())
	version := f.NewVersion(context.Background()).WithMessage("test")
	versions.SetVersion(inventory.NewVersionNumber().WithString("v1"), version)

	if err := versions.Check(val, []string{}); err != nil {
		t.Fatalf("check error: %v", err)
	}

	if val.HasError(validation.E049) {
		t.Errorf("unexpected error '%s' in %v", validation.E049, val.Error)
	}
}

package test

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/factory/factoryimpl"
	inventorytypes "github.com/ocfl-archive/gocfl/v2/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/inventory/inventoryimpl"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
	"github.com/rs/zerolog"
)

func Test_UserJSONMarshal(t *testing.T) {
	var logger = zerolog.New(zerolog.NewConsoleWriter())
	var f = factoryimpl.NewFactory(version.Version1_1, nil)
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

func Test_UserJSONUnmarshal(t *testing.T) {
	var logger = zerolog.New(zerolog.NewConsoleWriter())
	var f = factoryimpl.NewFactory(version.Version1_1, nil)
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

func Test_UserInvalidJSON(t *testing.T) {
	var logger = zerolog.New(zerolog.NewConsoleWriter())
	var f = factoryimpl.NewFactory(version.Version1_1, nil)
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

func Test_UserInvalidAddressCheck(t *testing.T) {
	var logger = zerolog.New(zerolog.NewConsoleWriter())
	var f = factoryimpl.NewFactory(version.Version1_1, nil)

	val := inventoryimpl.NewDummyValidation()
	user := f.NewUser(context.Background()).WithAddress("xxx")
	if err := user.Check(val, inventorytypes.NewVersionNumber().WithString("v1")); err != nil {
		t.Fatalf("check error: %v", err)
	}

	if !val.HasWarning(validation.W009) {
		t.Errorf("no warning '%s' in %v", validation.W009, val.Warning)
	}
}

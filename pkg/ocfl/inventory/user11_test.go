package inventory

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/types"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
)

func Test_UserJSONMarshal(t *testing.T) {
	user := ocfl.f11.NewUser().WithName("Alice").WithAddress("mailto:alice@example.org")
	bytes, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("Failed to marshal user: %s", err)
	}
	user2 := ocfl.f11.NewUser()
	if err := json.Unmarshal(bytes, user2); err != nil {
		t.Fatalf("Failed to unmarshal user: %s", err)
	}
	if !reflect.DeepEqual(user, user2) {
		t.Fatalf("Failed to unmarshal user - not equal")
	}
}

func Test_UserJSONUnmarshal(t *testing.T) {
	var jsonData = []byte(`
{
	"address": "mailto:alice@example.org",
	"name": "Alice"
}
`)
	user := ocfl.f11.NewUser()
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
	var jsonData = []byte(`
{
	"address": 42,
	"name": "Alice"
}
`)
	user := ocfl.f11.NewUser()
	if err := json.Unmarshal(jsonData, user); err != nil {
		t.Errorf("unmarshal error: %v", err)
	}
	if user.Err() == nil {
		t.Error("should get an error")
	}
}

func Test_UserInvalidAddressCheck(t *testing.T) {
	val := NewDummyValidation()
	user := ocfl.f11.NewUser().WithAddress("xxx")
	if err := user.Check(val, types.NewVersionNumber().WithString("v1")); err != nil {
		t.Fatalf("check error: %v", err)
	}

	if !val.HasWarning(validation.W009) {
		t.Errorf("no warning '%s' in %v", validation.W009, val.Warning)
	}
}

package inventory

import (
	"encoding/json"
	"testing"
)

var factory = NewFactory11()

func Test_UserJSON(t *testing.T) {
	var jsonData = []byte(`
{
	"address": "mailto:alice@example.org",
	"name": "Alice"
}
`)
	user := factory.NewUser()
	if err := json.Unmarshal(jsonData, user); err != nil {
		t.Errorf("unmarshal error: %v", err)
	}

}

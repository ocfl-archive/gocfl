package config

import (
	"fmt"
	"testing"

	"github.com/BurntSushi/toml"
)

func TestMiniConfig_MarshalText(t *testing.T) {
	mc := MiniConfig{
		"database.host":      "localhost",
		"server.api.timeout": "30s",
		"server.api.debug":   true,
		"logging.level":      "info",
		"Database.port":      5432,
		"simple":             42,
	}

	data, err := mc.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText failed: %v", err)
	}

	fmt.Printf("Generated TOML:\n%s\n", string(data))

	// Unmarshal back to check hierarchy
	var result map[string]any
	if err := toml.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal generated TOML: %v", err)
	}

	// Verify database (Note: MarshalText uses strings.ToLower on keys)
	db, ok := result["database"].(map[string]any)
	if !ok {
		t.Fatalf("expected database section to be a map, got %T", result["database"])
	}
	if db["host"] != "localhost" {
		t.Errorf("expected database.host 'localhost', got '%v'", db["host"])
	}
	if db["port"] != int64(5432) { // TOML unmarshals numbers as int64
		t.Errorf("expected database.port 5432, got %v (%T)", db["port"], db["port"])
	}

	// Verify server.api
	server, ok := result["server"].(map[string]any)
	if !ok {
		t.Fatalf("expected server section to be a map")
	}
	api, ok := server["api"].(map[string]any)
	if !ok {
		t.Fatalf("expected server.api section to be a map")
	}
	if api["timeout"] != "30s" {
		t.Errorf("expected server.api.timeout '30s', got '%v'", api["timeout"])
	}
	if api["debug"] != true {
		t.Errorf("expected server.api.debug true, got %v", api["debug"])
	}

	// Verify logging
	logging, ok := result["logging"].(map[string]any)
	if !ok {
		t.Fatalf("expected logging section to be a map")
	}
	if logging["level"] != "info" {
		t.Errorf("expected logging.level 'info', got '%v'", logging["level"])
	}

	// Verify simple
	if result["simple"] != int64(42) {
		t.Errorf("expected simple 42, got %v", result["simple"])
	}
}

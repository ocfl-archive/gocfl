package test

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"testing"

	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/factory/factoryimpl"
	inventorytypes "github.com/ocfl-archive/gocfl/v2/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
	"github.com/rs/zerolog"
)

func exampleState(cnt int, t *testing.T) inventorytypes.State {
	var logger = zerolog.New(zerolog.NewConsoleWriter())
	var f = factoryimpl.NewFactory(version.Version1_1, nil, &logger)
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

func Test_StateJSONMarshal(t *testing.T) {
	var logger = zerolog.New(zerolog.NewConsoleWriter())
	var f = factoryimpl.NewFactory(version.Version1_1, nil, &logger)
	var cnt = 4
	var state = exampleState(cnt, t)
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

func Test_StateJSONUnmarshal(t *testing.T) {
	var logger = zerolog.New(zerolog.NewConsoleWriter())
	var f = factoryimpl.NewFactory(version.Version1_1, nil, &logger)
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

func Test_StateJSONUnmarshalError(t *testing.T) {
	var logger = zerolog.New(zerolog.NewConsoleWriter())
	var f = factoryimpl.NewFactory(version.Version1_1, nil, &logger)
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

func Test_StateDelete(t *testing.T) {
	var cnt = 4
	var state = exampleState(cnt, t)
	modified, err := state.DeleteFile("file000")
	if err != nil {
		t.Fatalf("Error deleting file000: %v", err)
	}
	if !modified {
		t.Error("file000 has not modified state")
	}
	modified, err = state.DeleteFile("file000x")
	if err != nil {
		t.Fatalf("Error deleting file000x: %v", err)
	}
	if !modified {
		t.Error("file000x has not modified state")
	}
	modified, err = state.DeleteFile("file000")
	if err != nil {
		t.Fatalf("Error deleting file000: %v", err)
	}
	if modified {
		t.Error("second deletion of file000 has modified state")
	}
	var num int
	var ps int
	for _, paths := range state.Iterate() {
		num++
		ps += len(paths)
	}
	if num != cnt-1 {
		t.Errorf("Number of digests does not match expected number of files. Expected %d, got %d", cnt-1, num)
	}
	if ps != (cnt-1)*2 {
		t.Errorf("Number of files does not match expected number of files. Expected %d, got %d", (cnt-1)*2, num)
	}
}

func Test_StateRename(t *testing.T) {
	var cnt = 4
	var state = exampleState(cnt, t)
	modified, err := state.RenameFile("file000x", "file000y")
	if err != nil {
		t.Fatalf("Error renaming file000x: %v", err)
	}
	if !modified {
		t.Error("rename file000x to file000y has not modified state")
	}
	paths, err := state.GetFiles("digest000")
	if err != nil {
		t.Fatalf("Error getting files for digest digest000: %v", err)
	}
	if len(paths) != 2 {
		t.Errorf("Number of digests does not match expected number of files. Expected %d, got %d", 2, len(paths))
	}
	if !slices.Contains(paths, "file000y") {
		t.Error("file000y not in digest000")
	}
}

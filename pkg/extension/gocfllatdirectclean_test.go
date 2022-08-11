package extension

import (
	"fmt"
	"testing"
)

func TestFlatCleanDirectory(t *testing.T) {
	l, err := NewFlatDirectClean(&FlatDirectCleanConfig{
		Config: &Config{ExtensionName: FlatDirectCleanName},
		MaxLen: 255,
	})
	objectID := "object-01"
	testResult := "object-01"
	rootPath, err := l.ID2Path(objectID)
	if err != nil {
		t.Errorf("cannot convert %s", objectID)
	}
	if rootPath != testResult {
		t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
	}
	fmt.Printf("FlatDirectClean(%s) -> %s\n", objectID, rootPath)

	objectID = "..hor_rib:lé-$id"
	testResult = "..hor_rib_lé-$id"
	rootPath, err = l.ID2Path(objectID)
	if err != nil {
		t.Errorf("cannot convert %s - %v", objectID, err)
	} else {
		if rootPath != testResult {
			t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
		} else {
			fmt.Printf("FlatDirectClean(%s) -> %s\n", objectID, rootPath)
		}
	}

	// https://ocfl.github.io/extensions/0002-flat-direct-storage-layout.html
	// Example 2
	objectID = "info:fedora/object-01"
	testResult = "info_fedora/object-01"
	rootPath, err = l.ID2Path(objectID)
	if err != nil {
		t.Errorf("cannot convert %s - %v", objectID, err)
	} else {
		if rootPath != testResult {
			t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
		} else {
			fmt.Printf("FlatDirectClean(%s) -> %s\n", objectID, rootPath)
		}
	}

	objectID = "~ info:fedora/-obj#ec@t-\"01 "
	testResult = "info_fedora/obj_ec_t-_01"
	rootPath, err = l.ID2Path(objectID)
	if err != nil {
		t.Errorf("cannot convert %s - %v", objectID, err)
	} else {
		if rootPath != testResult {
			t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
		} else {
			fmt.Printf("FlatDirectClean(%s) -> %s\n", objectID, rootPath)
		}
	}

	objectID = "abcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghij"
	testResult = "abcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghij"
	rootPath, err = l.ID2Path(objectID)
	if err != nil {
		fmt.Printf("FlatDirectClean(%s) -> %v\n", objectID, err)
	} else {
		t.Errorf("%s -> should have error too long", objectID)
	}

}

package extension

import (
	"fmt"
	"testing"
)

func TestFlatDirectory(t *testing.T) {
	// https://ocfl.github.io/extensions/0002-flat-direct-storage-layout.html
	// Example 1
	l := StorageLayoutFlatDirect{}
	objectID := "object-01"
	testResult := "object-01"
	rootPath, err := l.BuildStoragerootPath(nil, objectID)
	if err != nil {
		t.Errorf("cannot convert %s", objectID)
	}
	if rootPath != testResult {
		t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
	}
	fmt.Printf("StorageLayoutFlatDirect(%s) -> %s\n", objectID, rootPath)

	objectID = "..hor_rib:lé-$id"
	testResult = "..hor_rib:lé-$id"
	rootPath, err = l.BuildStoragerootPath(nil, objectID)
	if err != nil {
		t.Errorf("cannot convert %s - %v", objectID, err)
	} else {
		if rootPath != testResult {
			t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
		} else {
			fmt.Printf("StorageLayoutFlatDirect(%s) -> %s\n", objectID, rootPath)
		}
	}

	// https://ocfl.github.io/extensions/0002-flat-direct-storage-layout.html
	// Example 2
	objectID = "info:fedora/object-01"
	testResult = "info:fedora/object-01"
	rootPath, err = l.BuildStoragerootPath(nil, objectID)
	if err != nil {
		t.Errorf("cannot convert %s - %v", objectID, err)
	} else {
		if rootPath != testResult {
			t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
		} else {
			fmt.Printf("StorageLayoutFlatDirect(%s) -> %s\n", objectID, rootPath)
		}
	}

	objectID = "abcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghij"
	testResult = "abcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghij"
	rootPath, err = l.BuildStoragerootPath(nil, objectID)
	if err != nil {
		t.Errorf("cannot convert %s - %v", objectID, err)
	} else {
		if rootPath != testResult {
			t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
		} else {
			fmt.Printf("StorageLayoutFlatDirect(%s) -> %s\n", objectID, rootPath)
		}
	}

}

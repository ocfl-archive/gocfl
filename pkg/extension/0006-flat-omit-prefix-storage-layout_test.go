package extension

import (
	"fmt"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	"testing"
)

func TestFlatOmitPrefixStorageLayout(t *testing.T) {
	// https://github.com/OCFL/extensions/blob/main/docs/0006-flat-omit-prefix-storage-layout.md
	// Example 1
	l := FlatOmitPrefixStorageLayout{
		FlatOmitPrefixStorageLayoutConfig: &FlatOmitPrefixStorageLayoutConfig{
			ExtensionConfig: &ocfl.ExtensionConfig{ExtensionName: "0006-flat-omit-prefix-storage-layout"},
			Delimiter:       ":",
		},
	}
	objectID := "namespace:12887296"
	testResult := "12887296"
	rootPath, err := l.BuildStorageRootPath(nil, objectID)
	if err != nil {
		t.Errorf("cannot convert %s", objectID)
	}
	if rootPath != testResult {
		t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
	}
	fmt.Printf("FlatOmitPrefixStorageLayout(%s) -> %s\n", objectID, rootPath)

	objectID = "urn:uuid:6e8bc430-9c3a-11d9-9669-0800200c9a66"
	testResult = "6e8bc430-9c3a-11d9-9669-0800200c9a66"
	rootPath, err = l.BuildStorageRootPath(nil, objectID)
	if err != nil {
		t.Errorf("cannot convert %s - %v", objectID, err)
	} else {
		if rootPath != testResult {
			t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
		} else {
			fmt.Printf("FlatOmitPrefixStorageLayout(%s) -> %s\n", objectID, rootPath)
		}
	}

	// Example 1
	l = FlatOmitPrefixStorageLayout{
		FlatOmitPrefixStorageLayoutConfig: &FlatOmitPrefixStorageLayoutConfig{
			ExtensionConfig: &ocfl.ExtensionConfig{ExtensionName: "0006-flat-omit-prefix-storage-layout"},
			Delimiter:       "edu/",
		},
	}
	objectID = "https://institution.edu/3448793"
	testResult = "3448793"
	rootPath, err = l.BuildStorageRootPath(nil, objectID)
	if err != nil {
		t.Errorf("cannot convert %s", objectID)
	}
	if rootPath != testResult {
		t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
	}
	fmt.Printf("FlatOmitPrefixStorageLayout(%s) -> %s\n", objectID, rootPath)

	objectID = "https://institution.edu/abc/edu/f8.05v"
	testResult = "f8.05v"
	rootPath, err = l.BuildStorageRootPath(nil, objectID)
	if err != nil {
		t.Errorf("cannot convert %s - %v", objectID, err)
	} else {
		if rootPath != testResult {
			t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
		} else {
			fmt.Printf("FlatOmitPrefixStorageLayout(%s) -> %s\n", objectID, rootPath)
		}
	}

}

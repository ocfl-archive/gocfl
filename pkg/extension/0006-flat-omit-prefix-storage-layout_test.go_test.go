package extension

import (
	"fmt"

	"testing"

	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"
)

func TestNTupleOmitPrefixStorageLayout(t *testing.T) {
	// https://github.com/OCFL/extensions/blob/main/docs/0006-flat-omit-prefix-storage-layout.md
	// Example 1
	l := NTupleOmitPrefixStorageLayout{
		NTupleOmitPrefixStorageLayoutConfig: &NTupleOmitPrefixStorageLayoutConfig{
			ExtensionConfig:   &extension.ExtensionConfig{ExtensionName: "0006-flat-omit-prefix-storage-layout"},
			Delimiter:         ":",
			TupleSize:         4,
			NumberOfTuples:    2,
			ZeroPadding:       "left",
			ReverseObjectRoot: true,
		},
	}
	objectID := "namespace:12887296"
	testResult := "6927/8821/12887296"
	rootPath, err := l.BuildStorageRootPath(nil, objectID)
	if err != nil {
		t.Errorf("cannot convert %s", objectID)
	}
	if rootPath != testResult {
		t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
	}
	fmt.Printf("NTupleOmitPrefixStorageLayout(%s) -> %s\n", objectID, rootPath)

	objectID = "urn:uuid:6e8bc430-9c3a-11d9-9669-0800200c9a66"
	testResult = "66a9/c002/6e8bc430-9c3a-11d9-9669-0800200c9a66"
	rootPath, err = l.BuildStorageRootPath(nil, objectID)
	if err != nil {
		t.Errorf("cannot convert %s - %v", objectID, err)
	} else {
		if rootPath != testResult {
			t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
		} else {
			fmt.Printf("NTupleOmitPrefixStorageLayout(%s) -> %s\n", objectID, rootPath)
		}
	}

	objectID = "abc123"
	testResult = "321c/ba00/abc123"
	rootPath, err = l.BuildStorageRootPath(nil, objectID)
	if err != nil {
		t.Errorf("cannot convert %s - %v", objectID, err)
	} else {
		if rootPath != testResult {
			t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
		} else {
			fmt.Printf("NTupleOmitPrefixStorageLayout(%s) -> %s\n", objectID, rootPath)
		}
	}

	// Example 1
	l = NTupleOmitPrefixStorageLayout{
		NTupleOmitPrefixStorageLayoutConfig: &NTupleOmitPrefixStorageLayoutConfig{
			ExtensionConfig:   &extension.ExtensionConfig{ExtensionName: "0006-flat-omit-prefix-storage-layout"},
			Delimiter:         "edu/",
			TupleSize:         3,
			NumberOfTuples:    3,
			ZeroPadding:       "right",
			ReverseObjectRoot: false,
		},
	}
	objectID = "https://institution.edu/3448793"
	testResult = "344/879/300/3448793"
	rootPath, err = l.BuildStorageRootPath(nil, objectID)
	if err != nil {
		t.Errorf("cannot convert %s", objectID)
	}
	if rootPath != testResult {
		t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
	}
	fmt.Printf("NTupleOmitPrefixStorageLayout(%s) -> %s\n", objectID, rootPath)

	objectID = "https://institution.edu/abc/edu/f8.05v"
	testResult = "f8./05v/000/f8.05v"
	rootPath, err = l.BuildStorageRootPath(nil, objectID)
	if err != nil {
		t.Errorf("cannot convert %s - %v", objectID, err)
	} else {
		if rootPath != testResult {
			t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
		} else {
			fmt.Printf("NTupleOmitPrefixStorageLayout(%s) -> %s\n", objectID, rootPath)
		}
	}

}

package extension

import (
	"fmt"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl"
	"testing"
)

func TestHashAndIdNTuple(t *testing.T) {
	// https://ocfl.github.io/extensions/0003-hash-and-id-n-tuple-storage-layout.html#encapsulation-directory
	// Example 1
	l, err := NewStorageLayoutHashAndIdNTuple(&StorageLayoutHashAndIdNTupleConfig{
		ExtensionConfig: &ocfl.ExtensionConfig{ExtensionName: "0003-hash-and-id-n-tuple-storage-layout"},
		DigestAlgorithm: string(checksum.DigestSHA256),
		TupleSize:       3,
		NumberOfTuples:  3,
	})
	if err != nil {
		t.Errorf("error calling NewStorageLayoutHashAndIdNTuple(%s, %v, %v) - %v", checksum.DigestSHA256, 3, 3, err)
		return
	}
	fmt.Printf("\nNewStorageLayoutHashAndIdNTuple(%s, %v, %v)\n", checksum.DigestSHA256, 3, 3)
	objectID := "object-01"
	testResult := "3c0/ff4/240/object-01"
	rootPath, err := l.BuildStorageRootPath(nil, objectID)
	if err != nil {
		t.Errorf("cannot convert %s", objectID)
	}
	if rootPath != testResult {
		t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
	}
	fmt.Printf("StorageLayoutHashAndIdNTuple(%s) -> %s\n", objectID, rootPath)

	objectID = "..hor/rib:le-$id"
	testResult = "487/326/d8c/%2e%2ehor%2frib%3ale-%24id"
	rootPath, err = l.BuildStorageRootPath(nil, objectID)
	if err != nil {
		t.Errorf("cannot convert %s", objectID)
	}
	if rootPath != testResult {
		t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
	}
	fmt.Printf("StorageLayoutHashAndIdNTuple(%s) -> %s\n", objectID, rootPath)

	// https://ocfl.github.io/extensions/0003-hash-and-id-n-tuple-storage-layout.html#encapsulation-directory
	// Example 2
	l, err = NewStorageLayoutHashAndIdNTuple(&StorageLayoutHashAndIdNTupleConfig{
		ExtensionConfig: &ocfl.ExtensionConfig{ExtensionName: "0003-hash-and-id-n-tuple-storage-layout"},
		DigestAlgorithm: string(checksum.DigestMD5),
		TupleSize:       2,
		NumberOfTuples:  15,
	})
	if err != nil {
		t.Errorf("error calling NewStorageLayoutHashAndIdNTuple(%s, %v, %v) - %v", checksum.DigestMD5, 2, 15, err)
		return
	}
	fmt.Printf("\nNewStorageLayoutHashAndIdNTuple(%s, %v, %v)\n", checksum.DigestMD5, 2, 15)
	objectID = "object-01"
	testResult = "ff/75/53/44/92/48/5e/ab/b3/9f/86/35/67/28/88/object-01"
	rootPath, err = l.BuildStorageRootPath(nil, objectID)
	if err != nil {
		t.Errorf("cannot convert %s", objectID)
	}
	if rootPath != testResult {
		t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
	}
	fmt.Printf("StorageLayoutHashAndIdNTuple(%s) -> %s\n", objectID, rootPath)

	objectID = "..hor/rib:le-$id"
	testResult = "08/31/97/66/fb/6c/29/35/dd/17/5b/94/26/77/17/%2e%2ehor%2frib%3ale-%24id"
	rootPath, err = l.BuildStorageRootPath(nil, objectID)
	if err != nil {
		t.Errorf("cannot convert %s", objectID)
	}
	if rootPath != testResult {
		t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
	}
	fmt.Printf("StorageLayoutHashAndIdNTuple(%s) -> %s\n", objectID, rootPath)

	// https://ocfl.github.io/extensions/0003-hash-and-id-n-tuple-storage-layout.html#encapsulation-directory
	// Example 3
	l, err = NewStorageLayoutHashAndIdNTuple(&StorageLayoutHashAndIdNTupleConfig{
		ExtensionConfig: &ocfl.ExtensionConfig{ExtensionName: "0003-hash-and-id-n-tuple-storage-layout"},
		DigestAlgorithm: string(checksum.DigestSHA256),
		TupleSize:       0,
		NumberOfTuples:  0,
	})
	if err != nil {
		t.Errorf("error calling NewStorageLayoutHashAndIdNTuple(%s, %v, %v) - %v", checksum.DigestSHA256, 0, 0, err)
		return
	}
	fmt.Printf("\nNewStorageLayoutHashAndIdNTuple(%s, %v, %v)\n", checksum.DigestSHA256, 0, 0)
	objectID = "object-01"
	testResult = "object-01"
	rootPath, err = l.BuildStorageRootPath(nil, objectID)
	if err != nil {
		t.Errorf("cannot convert %s", objectID)
	}
	if rootPath != testResult {
		t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
	}
	fmt.Printf("StorageLayoutHashAndIdNTuple(%s) -> %s\n", objectID, rootPath)

	objectID = "..hor/rib:le-$id"
	testResult = "%2e%2ehor%2frib%3ale-%24id"
	rootPath, err = l.BuildStorageRootPath(nil, objectID)
	if err != nil {
		t.Errorf("cannot convert %s", objectID)
	}
	if rootPath != testResult {
		t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
	}
	fmt.Printf("StorageLayoutHashAndIdNTuple(%s) -> %s\n", objectID, rootPath)

}

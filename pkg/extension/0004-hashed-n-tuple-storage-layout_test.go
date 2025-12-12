package extension

import (
	"fmt"

	"testing"

	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"
)

func TestHashedNTuple(t *testing.T) {
	// https://ocfl.github.io/extensions/0004-hashed-n-tuple-storage-layout.html
	// Example 1
	l, err := NewStorageLayoutHashedNTuple(&StorageLayoutHashedNTupleConfig{
		ExtensionConfig: &extension.ExtensionConfig{ExtensionName: "0004-hashed-n-tuple-storage-layout"},
		DigestAlgorithm: string(checksum.DigestSHA256),
		TupleSize:       3,
		NumberOfTuples:  3,
		ShortObjectRoot: false,
	},
	)
	if err != nil {
		t.Errorf("error calling NewStorageLayoutHashedNTuple(%s, %v, %v, %v) - %v", checksum.DigestSHA256, 3, 3, false, err)
		return
	}
	fmt.Printf("\nNewStorageLayoutHashedNTuple(%s, %v, %v, %v) - %v\n", checksum.DigestSHA256, 3, 3, false, err)
	objectID := "object-01"
	testResult := "3c0/ff4/240/3c0ff4240c1e116dba14c7627f2319b58aa3d77606d0d90dfc6161608ac987d4"
	rootPath, err := l.BuildStorageRootPath(nil, objectID)
	if err != nil {
		t.Errorf("cannot convert %s", objectID)
	}
	if rootPath != testResult {
		t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
	}
	fmt.Printf("StorageLayoutHashAndIdNTuple(%s) -> %s\n", objectID, rootPath)

	objectID = "..hor/rib:le-$id"
	testResult = "487/326/d8c/487326d8c2a3c0b885e23da1469b4d6671fd4e76978924b4443e9e3c316cda6d"
	rootPath, err = l.BuildStorageRootPath(nil, objectID)
	if err != nil {
		t.Errorf("cannot convert %s", objectID)
	}
	if rootPath != testResult {
		t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
	}
	fmt.Printf("StorageLayoutHashAndIdNTuple(%s) -> %s\n", objectID, rootPath)

	// https://ocfl.github.io/extensions/0004-hashed-n-tuple-storage-layout.html
	// Example 2
	l, err = NewStorageLayoutHashedNTuple(&StorageLayoutHashedNTupleConfig{
		ExtensionConfig: &extension.ExtensionConfig{ExtensionName: "0004-hashed-n-tuple-storage-layout"},
		DigestAlgorithm: string(checksum.DigestMD5),
		TupleSize:       2,
		NumberOfTuples:  15,
		ShortObjectRoot: true,
	})
	if err != nil {
		t.Errorf("error calling NewStorageLayoutHashedNTuple(%s, %v, %v, %v) - %v", checksum.DigestMD5, 2, 15, true, err)
		return
	}
	fmt.Printf("\nNewStorageLayoutHashedNTuple(%s, %v, %v, %v)\n", checksum.DigestMD5, 2, 15, true)
	objectID = "object-01"
	testResult = "ff/75/53/44/92/48/5e/ab/b3/9f/86/35/67/28/88/4e"
	rootPath, err = l.BuildStorageRootPath(nil, objectID)
	if err != nil {
		t.Errorf("cannot convert %s", objectID)
	}
	if rootPath != testResult {
		t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
	}
	fmt.Printf("StorageLayoutHashedNTuple(%s) -> %s\n", objectID, rootPath)

	objectID = "..hor/rib:le-$id"
	testResult = "08/31/97/66/fb/6c/29/35/dd/17/5b/94/26/77/17/e0"
	rootPath, err = l.BuildStorageRootPath(nil, objectID)
	if err != nil {
		t.Errorf("cannot convert %s", objectID)
	}
	if rootPath != testResult {
		t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
	}
	fmt.Printf("StorageLayoutHashedNTuple(%s) -> %s\n", objectID, rootPath)

	// https://ocfl.github.io/extensions/0004-hashed-n-tuple-storage-layout.html
	// Example 3
	l, err = NewStorageLayoutHashedNTuple(&StorageLayoutHashedNTupleConfig{
		ExtensionConfig: &extension.ExtensionConfig{ExtensionName: "0004-hashed-n-tuple-storage-layout"},
		DigestAlgorithm: string(checksum.DigestSHA256),
		TupleSize:       0,
		NumberOfTuples:  0,
		ShortObjectRoot: false,
	},
	)
	if err != nil {
		t.Errorf("error calling NewStorageLayoutHashedNTuple(%s, %v, %v, %v) - %v", checksum.DigestSHA256, 0, 0, false, err)
		return
	}
	fmt.Printf("\nNewStorageLayoutHashedNTuple(%s, %v, %v, %v)\n", checksum.DigestSHA256, 0, 0, false)
	objectID = "object-01"
	testResult = "3c0ff4240c1e116dba14c7627f2319b58aa3d77606d0d90dfc6161608ac987d4"
	rootPath, err = l.BuildStorageRootPath(nil, objectID)
	if err != nil {
		t.Errorf("cannot convert %s", objectID)
	}
	if rootPath != testResult {
		t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
	}
	fmt.Printf("StorageLayoutHashedNTuple(%s) -> %s\n", objectID, rootPath)

	objectID = "..hor/rib:le-$id"
	testResult = "487326d8c2a3c0b885e23da1469b4d6671fd4e76978924b4443e9e3c316cda6d"
	rootPath, err = l.BuildStorageRootPath(nil, objectID)
	if err != nil {
		t.Errorf("cannot convert %s", objectID)
	}
	if rootPath != testResult {
		t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
	}
	fmt.Printf("StorageLayoutHashedNTuple(%s) -> %s\n", objectID, rootPath)

}

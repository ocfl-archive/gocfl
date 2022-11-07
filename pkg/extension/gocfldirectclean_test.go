package extension

import (
	"fmt"
	"go.ub.unibas.ch/gocfl/v2/pkg/extension/storageroot"
	"testing"
)

func TestFlatCleanDirectoryWithoutUTFEncode(t *testing.T) {
	l, err := NewStorageLayoutDirectClean(&DirectCleanConfig{
		Config:                      &storageroot.Config{ExtensionName: DirectCleanName},
		MaxPathnameLen:              32000,
		MaxFilenameLen:              127,
		WhitespaceReplacementString: " ",
		ReplacementString:           "_",
		UTFEncode:                   false,
	})
	objectID := "object-01"
	testResult := "object-01"
	rootPath, err := l.ExecuteID(objectID)
	if err != nil {
		t.Errorf("cannot convert %s", objectID)
	}
	if rootPath != testResult {
		t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
	}
	fmt.Printf("DirectClean(%s) -> %s\n", objectID, rootPath)

	objectID = "..hor_rib:lé-$id"
	testResult = "..hor_rib_lé-$id"
	rootPath, err = l.ExecuteID(objectID)
	if err != nil {
		t.Errorf("cannot convert %s - %v", objectID, err)
	} else {
		if rootPath != testResult {
			t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
		} else {
			fmt.Printf("DirectClean(%s) -> %s\n", objectID, rootPath)
		}
	}

	// https://ocfl.github.io/extensions/0002-flat-direct-storage-layout.html
	// Example 2
	objectID = "info:fedora/object-01"
	testResult = "info_fedora/object-01"
	rootPath, err = l.ExecuteID(objectID)
	if err != nil {
		t.Errorf("cannot convert %s - %v", objectID, err)
	} else {
		if rootPath != testResult {
			t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
		} else {
			fmt.Printf("DirectClean(%s) -> %s\n", objectID, rootPath)
		}
	}

	objectID = "~ info:fedora/-obj#ec@t-\"01 "
	testResult = "info_fedora/obj_ec_t-_01"
	rootPath, err = l.ExecuteID(objectID)
	if err != nil {
		t.Errorf("cannot convert %s - %v", objectID, err)
	} else {
		if rootPath != testResult {
			t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
		} else {
			fmt.Printf("DirectClean(%s) -> %s\n", objectID, rootPath)
		}
	}

	objectID = "/test/ ~/.../blah"
	testResult = "test/_../blah"
	rootPath, err = l.ExecuteID(objectID)
	if err != nil {
		t.Errorf("cannot convert %s - %v", objectID, err)
	} else {
		if rootPath != testResult {
			t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
		} else {
			fmt.Printf("DirectClean(%s) -> %s\n", objectID, rootPath)
		}
	}

	objectID = "https://hdl.handle.net/XXXXX/test/bl ah"
	testResult = "https_/hdl.handle.net/XXXXX/test/bl ah"
	rootPath, err = l.ExecuteID(objectID)
	if err != nil {
		t.Errorf("cannot convert %s - %v", objectID, err)
	} else {
		if rootPath != testResult {
			t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
		} else {
			fmt.Printf("DirectClean(%s) -> %s\n", objectID, rootPath)
		}
	}

	objectID = "abcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghij"
	testResult = "abcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghij"
	rootPath, err = l.ExecuteID(objectID)
	if err != nil {
		fmt.Printf("DirectClean(%s) -> %v\n", objectID, err)
	} else {
		t.Errorf("%s -> should have error too long", objectID)
	}

}

func TestFlatCleanDirectoryWithUTFEncode(t *testing.T) {
	l, err := NewStorageLayoutDirectClean(&DirectCleanConfig{
		Config:                      &storageroot.Config{ExtensionName: DirectCleanName},
		MaxPathnameLen:              32000,
		MaxFilenameLen:              127,
		WhitespaceReplacementString: " ",
		ReplacementString:           "_",
		UTFEncode:                   true,
	})
	objectID := "object-01"
	testResult := "object-01"
	rootPath, err := l.ExecuteID(objectID)
	if err != nil {
		t.Errorf("cannot convert %s", objectID)
	}
	if rootPath != testResult {
		t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
	}
	fmt.Printf("DirectClean(%s) -> %s\n", objectID, rootPath)

	objectID = "..hor_rib:lé-$id"
	testResult = "..hor_rib=u003Alé-$id"
	rootPath, err = l.ExecuteID(objectID)
	if err != nil {
		t.Errorf("cannot convert %s - %v", objectID, err)
	} else {
		if rootPath != testResult {
			t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
		} else {
			fmt.Printf("DirectClean(%s) -> %s\n", objectID, rootPath)
		}
	}

	// https://ocfl.github.io/extensions/0002-flat-direct-storage-layout.html
	// Example 2
	objectID = "info:fedora/object-01"
	testResult = "info=u003Afedora/object-01"
	rootPath, err = l.ExecuteID(objectID)
	if err != nil {
		t.Errorf("cannot convert %s - %v", objectID, err)
	} else {
		if rootPath != testResult {
			t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
		} else {
			fmt.Printf("DirectClean(%s) -> %s\n", objectID, rootPath)
		}
	}

	objectID = "~ info:fedora/-obj#ec@t-\"01 "
	testResult = "=u007E=u0020info=u003Afedora/-obj=u0023ec=u0040t-=u002201=u0020"
	rootPath, err = l.ExecuteID(objectID)
	if err != nil {
		t.Errorf("cannot convert %s - %v", objectID, err)
	} else {
		if rootPath != testResult {
			t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
		} else {
			fmt.Printf("DirectClean(%s) -> %s\n", objectID, rootPath)
		}
	}

	objectID = "/test/ ~/.../blah"
	testResult = "test/=u0020~/=u002E../blah"
	rootPath, err = l.ExecuteID(objectID)
	if err != nil {
		t.Errorf("cannot convert %s - %v", objectID, err)
	} else {
		if rootPath != testResult {
			t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
		} else {
			fmt.Printf("DirectClean(%s) -> %s\n", objectID, rootPath)
		}
	}

	objectID = "https://hdl.handle.net/XXXXX/test/bl ah"
	testResult = "https=u003A/hdl.handle.net/XXXXX/test/bl=u0020ah"
	rootPath, err = l.ExecuteID(objectID)
	if err != nil {
		t.Errorf("cannot convert %s - %v", objectID, err)
	} else {
		if rootPath != testResult {
			t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
		} else {
			fmt.Printf("DirectClean(%s) -> %s\n", objectID, rootPath)
		}
	}

	objectID = "abcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghij"
	testResult = "abcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghij"
	rootPath, err = l.ExecuteID(objectID)
	if err != nil {
		fmt.Printf("DirectClean(%s) -> %v\n", objectID, err)
	} else {
		t.Errorf("%s -> should have error too long", objectID)
	}

}

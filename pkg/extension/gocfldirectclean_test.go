package extension

import (
	"fmt"
	"go.ub.unibas.ch/gocfl/v2/pkg/checksum"
	"go.ub.unibas.ch/gocfl/v2/pkg/ocfl"
	"testing"
)

func TestFlatCleanDirectoryWithoutUTFEncode(t *testing.T) {
	l := DirectClean{
		&DirectCleanConfig{
			ExtensionConfig:             &ocfl.ExtensionConfig{ExtensionName: DirectCleanName},
			MaxPathnameLen:              32000,
			MaxFilenameLen:              127,
			WhitespaceReplacementString: " ",
			ReplacementString:           "_",
			UTFEncode:                   false,
			FallbackSubFolders:          2,
			FallbackDigestAlgorithm:     "md5",
			FallbackFolder:              "fallback",
		},
	}
	l.hash, _ = checksum.GetHash(l.FallbackDigestAlgorithm)
	objectID := "object-01"
	testResult := "object-01"
	rootPath, err := l.BuildStoragerootPath(nil, objectID)
	if err != nil {
		t.Errorf("cannot convert %s", objectID)
	}
	if rootPath != testResult {
		t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
	}
	fmt.Printf("DirectClean(%s) -> %s\n", objectID, rootPath)

	objectID = "..hor_rib:lé-$id"
	testResult = "..hor_rib_lé-$id"
	rootPath, err = l.BuildStoragerootPath(nil, objectID)
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
	rootPath, err = l.BuildStoragerootPath(nil, objectID)
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
	rootPath, err = l.BuildStoragerootPath(nil, objectID)
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
	rootPath, err = l.BuildStoragerootPath(nil, objectID)
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
	rootPath, err = l.BuildStoragerootPath(nil, objectID)
	if err != nil {
		t.Errorf("cannot convert %s - %v", objectID, err)
	} else {
		if rootPath != testResult {
			t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
		} else {
			fmt.Printf("DirectClean(%s) -> %s\n", objectID, rootPath)
		}
	}

	objectID = "abcdefghijabcdefghij abcdefghijabcdefghij abcdefghijabcdefghij abcdefghijabcdefghij abcdefghijabcdefghij abcdefghijabcdefghij abcdefghijabcdefghij abcdefghijabcdefghij abcdefghijabcdefghij abcdefghijabcdefghij abcdefghijabcdefghij abcdefghijabcdefghij abcdefghijabcdefghij"
	testResult = "fallback/0/e/0eafabb38fa7f1583d1461afe980ebdc"
	rootPath, err = l.BuildStoragerootPath(nil, objectID)
	if err != nil {
		t.Errorf("cannot convert %s - %v", objectID, err)
	} else {
		if rootPath != testResult {
			t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
		} else {
			fmt.Printf("DirectClean(%s) -> %s\n", objectID, rootPath)
		}
	}

}

func TestFlatCleanDirectoryWithUTFEncode(t *testing.T) {
	l := DirectClean{
		&DirectCleanConfig{
			ExtensionConfig:             &ocfl.ExtensionConfig{ExtensionName: DirectCleanName},
			MaxPathnameLen:              32000,
			MaxFilenameLen:              127,
			WhitespaceReplacementString: " ",
			ReplacementString:           "_",
			UTFEncode:                   true,
			FallbackSubFolders:          2,
			FallbackDigestAlgorithm:     "sha512",
			FallbackFolder:              "fallback",
		},
	}
	l.hash, _ = checksum.GetHash(l.FallbackDigestAlgorithm)
	objectID := "object-01"
	testResult := "object-01"
	rootPath, err := l.BuildStoragerootPath(nil, objectID)
	if err != nil {
		t.Errorf("cannot convert %s", objectID)
	}
	if rootPath != testResult {
		t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
	}
	fmt.Printf("DirectClean(%s) -> %s\n", objectID, rootPath)

	objectID = "object=u123a-01"
	testResult = "object=u003Du123a-01"
	rootPath, err = l.BuildStoragerootPath(nil, objectID)
	if err != nil {
		t.Errorf("cannot convert %s", objectID)
	}
	if rootPath != testResult {
		t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
	}
	fmt.Printf("DirectClean(%s) -> %s\n", objectID, rootPath)

	objectID = "object=u123a-01"
	testResult = "object=u003Du123a-01"
	rootPath, err = l.BuildStoragerootPath(nil, objectID)
	if err != nil {
		t.Errorf("cannot convert %s", objectID)
	}
	if rootPath != testResult {
		t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
	}
	fmt.Printf("DirectClean(%s) -> %s\n", objectID, rootPath)

	objectID = "..hor_rib:lé-$id"
	testResult = "..hor_rib=u003Alé-$id"
	rootPath, err = l.BuildStoragerootPath(nil, objectID)
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
	rootPath, err = l.BuildStoragerootPath(nil, objectID)
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
	rootPath, err = l.BuildStoragerootPath(nil, objectID)
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
	rootPath, err = l.BuildStoragerootPath(nil, objectID)
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
	rootPath, err = l.BuildStoragerootPath(nil, objectID)
	if err != nil {
		t.Errorf("cannot convert %s - %v", objectID, err)
	} else {
		if rootPath != testResult {
			t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
		} else {
			fmt.Printf("DirectClean(%s) -> %s\n", objectID, rootPath)
		}
	}

	objectID = "abcdefghijabcdefghij abcdefghijabcdefghij abcdefghijabcdefghij abcdefghijabcdefghij abcdefghijabcdefghij abcdefghijabcdefghij abcdefghijabcdefghij abcdefghijabcdefghij abcdefghijabcdefghij abcdefghijabcdefghij abcdefghijabcdefghij abcdefghijabcdefghij abcdefghijabcdefghij"
	testResult = "fallback/b/8/b8acda4abac53237afa03d6bbb078e1bf46b40438bb256df79b8d9ff0e57b32a688156ad21755363ea19953c160c4dd6d4db175b71e9aa87d68937181a9f69d/9"
	rootPath, err = l.BuildStoragerootPath(nil, objectID)
	if err != nil {
		t.Errorf("cannot convert %s - %v", objectID, err)
	} else {
		if rootPath != testResult {
			t.Errorf("%s -> %s != %s", objectID, rootPath, testResult)
		} else {
			fmt.Printf("DirectClean(%s) -> %s\n", objectID, rootPath)
		}
	}

}

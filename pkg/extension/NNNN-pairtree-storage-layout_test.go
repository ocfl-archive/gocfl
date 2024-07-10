package extension

import (
	"fmt"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl"
	"testing"
)

func TestPairtreeIDEncode(t *testing.T) {
	fmt.Printf("(NewPairTreeStorageLayout(%s, %s, %v, %s)\n", "", "", 2, checksum.DigestSHA256)

	ptsl, err := NewStorageLayoutPairTree(&StorageLayoutPairTreeConfig{
		ExtensionConfig: &ocfl.ExtensionConfig{ExtensionName: "gocfl-pairtree"},
		UriBase:         "",
		StoreDir:        "",
		ShortyLength:    2,
		DigestAlgorithm: string(checksum.DigestSHA256),
	})
	if err != nil {
		t.Errorf("instantiate failed - %v", err)
		return
	}

	sourceID := "ark:/13030/xt12t3"
	testResult := "ar/k+/=1/30/30/=x/t1/2t/3"
	dest, _ := ptsl.BuildStorageRootPath(nil, sourceID)
	if dest != testResult {
		t.Errorf("IDEncode(%s) => %s != %s", sourceID, dest, testResult)
	} else {
		fmt.Printf("IDEncode(%s) => %s\n", sourceID, dest)
	}

	/* wrong example????
	sourceID = "http://n2t.info/urn:nbn:se:kb:repos-1"
	testResult = "ht/tp/+=/=n/2t/,i/nf/o=/ur/n+/n/bn/+s/e+/kb/+/re/p/OS/-1"
	dest = ptsl.idToDirpath(sourceID)
	if dest != testResult {
		t.Errorf("IDEncode(%s) => %s != %s", sourceID, dest, testResult)
	} else {
		fmt.Printf("IDEncode(%s) => %s\n", sourceID, dest)
	}
	*/

	sourceID = "what-the-*@?#!^!?"
	testResult = "wh/at/-t/he/-^/2a/@^/3f/#!/^5/e!/^3/f"
	dest, _ = ptsl.BuildStorageRootPath(nil, sourceID)
	if dest != testResult {
		t.Errorf("IDEncode(%s) => %s != %s", sourceID, dest, testResult)
	} else {
		fmt.Printf("IDEncode(%s) => %s\n", sourceID, dest)
	}
}

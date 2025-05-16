package main

import (
	"bytes"
	"emperror.dev/emperror"
	"emperror.dev/errors"
	"encoding/json"
	"flag"
	"github.com/je4/utils/v2/pkg/encrypt"
	"github.com/je4/utils/v2/pkg/keepass2kms"
	"github.com/tink-crypto/tink-go/v2/keyset"
	"github.com/tink-crypto/tink-go/v2/streamingaead"
	"io"
	"os"
	"path/filepath"
)

var inputFile = flag.String("input", "", "input file")
var kdbxFile = flag.String("kdbx", "", "kdbx file")
var kdbxSecret = flag.String("kdbx-secret", "", "kdbx secret")
var kdbxKey = flag.String("kdbx-key", "", "kdbx key path")
var outputFile = flag.String("output", "", "output file")

func main() {
	flag.Parse()
	if inputFile == nil {
		panic("input file not set")
	}
	encFile := *inputFile
	keyfile := *inputFile + ".key.json"
	keyData, err := os.ReadFile(keyfile)
	if err != nil {
		emperror.Panic(errors.Errorf("cannot read key file '%s': %v", keyfile, err))
	}
	kStruct := &encrypt.KeyStruct{}
	if err := json.Unmarshal(keyData, kStruct); err != nil {
		emperror.Panic(errors.Errorf("cannot unmarshal key file '%s': %v", keyfile, err))
	}

	db, err := keepass2kms.LoadKeePassDBFromFile(*kdbxFile, *kdbxSecret)
	if err != nil {
		emperror.Panic(errors.Errorf("cannot load keepass2 database '%s': %v", *kdbxFile, err))
	}
	client, err := keepass2kms.NewClient(db, filepath.Base(*kdbxFile))
	if err != nil {
		emperror.Panic(errors.Errorf("cannot create keepass2 client: %v", err))
	}
	//	registry.RegisterKMSClient(client)

	aead, err := client.GetAEAD(*kdbxKey)
	if err != nil {
		emperror.Panic(errors.Errorf("cannot get aead '%s': %v", *kdbxKey, err))
	}

	kh, err := keyset.Read(keyset.NewBinaryReader(bytes.NewBuffer(kStruct.EncryptedKey)), aead)
	if err != nil {
		emperror.Panic(errors.Errorf("cannot read keyset: %v", err))
	}

	fp, err := os.Open(encFile)
	if err != nil {
		emperror.Panic(errors.Errorf("cannot open file '%s': %v", encFile, err))
	}
	defer fp.Close()

	stream, err := streamingaead.New(kh)
	if err != nil {
		emperror.Panic(errors.Errorf("cannot create streamingaead: %v", err))
	}
	dec, err := stream.NewDecryptingReader(fp, kStruct.Aad)
	if err != nil {
		emperror.Panic(errors.Errorf("cannot create decrypting reader: %v", err))
	}

	var out io.Writer
	if outputFile != nil && *outputFile != "" {
		outFP, err := os.Create(*outputFile)
		if err != nil {
			emperror.Panic(errors.Errorf("cannot create output file '%s': %v", *outputFile, err))
		}
		defer outFP.Close()
		out = outFP
	} else {
		out = os.Stdout
	}

	if _, err := io.Copy(out, dec); err != nil {
		emperror.Panic(errors.Errorf("cannot write to stdout: %v", err))
	}
}

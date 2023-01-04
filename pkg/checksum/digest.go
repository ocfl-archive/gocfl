package checksum

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/exp/maps"
	"hash"
)

type DigestAlgorithm string

const (
	DigestMD5        DigestAlgorithm = "md5"
	DigestSHA1       DigestAlgorithm = "sha1"
	DigestSHA256     DigestAlgorithm = "sha256"
	DigestSHA512     DigestAlgorithm = "sha512"
	DigestBlake2b160 DigestAlgorithm = "blake2b-160"
	DigestBlake2b256 DigestAlgorithm = "blake2b-256"
	DigestBlake2b384 DigestAlgorithm = "blake2b-384"
	DigestBlake2b512 DigestAlgorithm = "blake2b-512"
)

var hashFunc = map[DigestAlgorithm]func() hash.Hash{
	DigestMD5:    md5.New,
	DigestSHA1:   sha1.New,
	DigestSHA256: sha256.New,
	DigestSHA512: sha512.New,
	DigestBlake2b160: func() hash.Hash {
		h, err := blake2b.New(20, nil)
		if err != nil {
			panic(err)
		}
		return h
	},
	DigestBlake2b256: func() hash.Hash {
		h, err := blake2b.New256(nil)
		if err != nil {
			panic(err)
		}
		return h
	},
	DigestBlake2b384: func() hash.Hash {
		h, err := blake2b.New384(nil)
		if err != nil {
			panic(err)
		}
		return h
	},
	DigestBlake2b512: func() hash.Hash {
		h, err := blake2b.New512(nil)
		if err != nil {
			panic(err)
		}
		return h
	},
}

var DigestsNames = maps.Keys(hashFunc)

func HashExists(csType DigestAlgorithm) bool {
	_, ok := hashFunc[csType]
	return ok
}

func GetHash(csType DigestAlgorithm) (hash.Hash, error) {
	f, ok := hashFunc[csType]
	if !ok {
		return nil, fmt.Errorf("unknown checksum %s", csType)
	}
	sink := f()
	return sink, nil
}

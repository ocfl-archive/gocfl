package premis

import (
	"encoding/xml"
)

func NewStringPlusAuthority(str string, authority string, authorityURI string, valueURI string) *StringPlusAuthority {
	return &StringPlusAuthority{
		AuthorityAttr:    authority,
		AuthorityURIAttr: authorityURI,
		ValueURIAttr:     valueURI,
		Value:            str,
	}
}

var locCryptoFunctions = map[string]string{
	"crc32":  "CRC32",
	"md5":    "MD5",
	"sha1":   "SHA-1",
	"sha256": "SHA-256",
	"sha384": "SHA-384",
	"sha512": "SHA-512",
}

var locCryptoFunctionsURI = map[string]string{
	"crc32":  "https://id.loc.gov/vocabulary/preservation/cryptographicHashFunctions/crc32",
	"md5":    "https://id.loc.gov/vocabulary/preservation/cryptographicHashFunctions/md5",
	"sha1":   "https://id.loc.gov/vocabulary/preservation/cryptographicHashFunctions/sha1",
	"sha256": "https://id.loc.gov/vocabulary/preservation/cryptographicHashFunctions/sha256",
	"sha384": "https://id.loc.gov/vocabulary/preservation/cryptographicHashFunctions/sha384",
	"sha512": "https://id.loc.gov/vocabulary/preservation/cryptographicHashFunctions/sha512",
}

func NewFixityComplexType(digestAlg, checksum, originator string) *FixityComplexType {
	fct := &FixityComplexType{
		XMLName: xml.Name{},
		MessageDigestAlgorithm: &StringPlusAuthority{
			Value: digestAlg,
		},
		MessageDigest: checksum,
		MessageDigestOriginator: &StringPlusAuthority{
			//XMLName:                 xml.Name{},
			Value: originator,
		},
	}
	if a, ok := locCryptoFunctions[digestAlg]; ok {
		fct.MessageDigestAlgorithm.AuthorityAttr = "cryptographicHashFunctions"
		fct.MessageDigestAlgorithm.AuthorityURIAttr = "http://id.loc.gov/vocabulary/preservation/cryptographicHashFunctions"
		fct.MessageDigestAlgorithm.ValueURIAttr = locCryptoFunctionsURI[digestAlg]
		fct.MessageDigestAlgorithm.Value = a
	}
	return fct
}

func NewSignificantPropertiesComplexType(name, value string) *SignificantPropertiesComplexType {
	return &SignificantPropertiesComplexType{
		XMLName:                        xml.Name{},
		SignificantPropertiesType:      NewStringPlusAuthority(name, "", "", ""),
		SignificantPropertiesValue:     value,
		SignificantPropertiesExtension: nil,
	}
}

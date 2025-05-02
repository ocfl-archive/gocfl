# OCFL Community Extension NNNN: Signature

* __Extension Name:__ NNNN-timestamp
* **Authors:** Jürgen Enge (Basel)
* **Minimum OCFL Version:** 1.0
* **OCFL Community Extensions Version:** 1.0
* **Obsoletes:** n/a
* **Obsoleted by:** n/a

## Overview

Adding trusted signatures (timestamps) to OCFL objects is a common requirement for
digital preservation. This extension allows the storage of a signature file in within 
the OCFL object.  

### Usage Scenario

Trusted Archives require the storage of a signature file within the OCFL object. 
This extension supports the generation of multiple trusted timestamps based on RFC 3161. 
For every timestamp a request and a response file is stored within the extension folder.


## Parameters

### Summary

* **Name:** `Authority`
    * **Description:** a map of url's containing a name and the url of the trusted timestamp authority
    * **Type:** string
    * **Default:**

* **Name:** `CertChain`
    * **Description:** requests full certificate chain within the result of the trusted timestamp authority
    * **Type:** Boolean
    * **Default:**

## Procedure

After finalizing an OCFL version, the extension gets the inventory checksum, creates a 
timestamp request and sends it to the trusted timestamp authority. The response is
stored together with the request within the extension folder

## Examples

### Parameters

```json
{
  "extensionName": "NNNN-timestamp",
  "Authority": {
    "freeTSA": "https://freetsa.org/tsr",
    "BIT": "http://tsa.pki.admin.ch/tsa"
  },
  "CertChain": false
}
```

### Result 

```
data/BIT.v1.tsq
data/BIT.v1.tsr
data/freeTSA.v1.tsq
data/freeTSA.v1.tsr
```
Files with extension `.tsq` are the request files, files with extension `.tsr` are the response files.
`v1` is the corresponding version of the OCFL object.

### Verification

The signature files can be verified using the `openssl` command line tool. 
Normally, the TSA certificate and the CA certificate are required to verify the signature.

For verification get the checksum from inventory.json.sha512 of the version you want to verify.

```bash
openssl ts -verify -in <TSA Response file> -digest <checksum> -CAfile <cacrt>.pem -untrusted <tsacert>.pem
```
The command will return `Verification: OK` if the signature is valid.
```bash
je@LAPTOP-KOTV7LO0:/mnt/c/temp/ocfl$ openssl ts -verify -in freeTSA.v1.tsr -digest 19cb0859b65edc2db3bc5c6dcf01e832ca66ea3ecb55c6a0e66573accb8bde317a4ee29ca4f841d60b25ec9fe331e533f54249a1e9c97ba9724810b7c64101cf -CAfile cacert.pem -untruste
d tsa.crt
Using configuration from /usr/lib/ssl/openssl.cnf
Warning: certificate from 'tsa.crt' with subject '/O=Free TSA/OU=TSA/description=This certificate digitally signs documents and time stamp requests made using the freetsa.org online services/CN=www.freetsa.org/emailAddress=busilezas@gmail.com/L=Wuerzburg/C=DE/ST=Bayern' is not a CA cert
Verification: OK
```

If the verification is valid, then check the timestamp with the command
```bash
openssl ts -verify -in <TSA Response file> -text 
```

The command will return the content of the response in human-readable format. 
```bash
je@LAPTOP-KOTV7LO0:/mnt/c/temp/ocfl$ openssl ts -reply -in freeTSA.v1.tsr -text
Using configuration from /usr/lib/ssl/openssl.cnf
Status info:
Status: Granted.
Status description: unspecified
Failure info: unspecified

TST info:
Version: 1
Policy OID: tsa_policy1
Hash Algorithm: sha512
Message data:
    0000 - 19 cb 08 59 b6 5e dc 2d-b3 bc 5c 6d cf 01 e8 32   ...Y.^.-..\m...2
    0010 - ca 66 ea 3e cb 55 c6 a0-e6 65 73 ac cb 8b de 31   .f.>.U...es....1
    0020 - 7a 4e e2 9c a4 f8 41 d6-0b 25 ec 9f e3 31 e5 33   zN....A..%...1.3
    0030 - f5 42 49 a1 e9 c9 7b a9-72 48 10 b7 c6 41 01 cf   .BI...{.rH...A..
Serial number: 0x06658FD0
Time stamp: Apr 26 11:14:11 2025 GMT
Accuracy: unspecified
Ordering: yes
Nonce: unspecified
TSA: DirName:/O=Free TSA/OU=TSA/description=This certificate digitally signs documents and time stamp requests made using the freetsa.org online services/CN=www.freetsa.org/emailAddress=busilezas@gmail.com/L=Wuerzburg/C=DE/ST=Bayern
Extensions:
```
The `Time` should correspond to the creation timestamp of the OCFL version (not identical).
```json
{
   "id": "ub:test01",
   "type": "https://ocfl.io/1.1/spec/#inventory",
   "digestAlgorithm": "sha512",
   "head": "v1",
   "manifest": {...},
   "versions": {
      "v1": {
         "created": "2025-04-26T13:13:41+02:00",
         "message": "initial commit",
         "state": {...},
         "user": {...}
      }
   }
}
```

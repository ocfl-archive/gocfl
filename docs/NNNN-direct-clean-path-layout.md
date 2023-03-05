# OCFL Community Extension NNNN: Direct Clean Path Layout

* **Extension Name:** NNNN-direct-clean-path-layout
* **Authors:** Jürgen Enge (Basel)
* **Minimum OCFL Version:** 1.0
* **OCFL Community Extensions Version:** 1.0
* **Obsoletes:** n/a
* **Obsoleted by:** n/a

## Overview

This extension can be used to either as a storage layout extension that maps OCFL
object identifiers to storage paths or as an object extension that maps logical paths to content paths.
This is done by replacing or removing "dangerous characters" from names.

This functionality is based on David A. Wheeler's essay "Fixing Unix/Linux/POSIX Filenames" (https://www.dwheeler.com/essays/fixing-unix-linux-filenames.html)

### Usage Scenario

This extension is intended to be used when you want the internal OCFL layout to match
the logical structure of the stored objects as closely as possible, while also
ensuring all file and directory names do not include problematic characters.

One example could be complete filesystems, which come from estates, and are not
under control of the archivists.

### Caveat

#### Projection
This extension provides injective (one-to-one) projections only if
`encodeUTF == true`.  
If you don't want to use the verbose UTF-Code replacements (`encodeUTF == false`), first check your source to make sure, that there's no chance of
different names to be mapped on one target. This happens on whitespace replacement as well as on
the removal of leading `-`, `~` and ` ` (blank) or trailing ` ` (blank)
Software which generates the OCFL
structure should raise an error in this case.
##### Example names, that would map to the same target
* `~file` => `file`
* `-file` => `file`
* ` file` => `file`
* `file` => `file`
* `file ` => `file`

#### Object Identifiers
When using this extension as a storage layout, you must ensure that none of your
object identifiers are prefixes of other identifiers. For example,
`https://hdl.handle.net/XXXXX/test` and `https://hdl.handle.net/XXXXX/test/blah`.
This is a problem because it would result in the storage paths
`https/hdl.handle.net/XXXXX/test` and `https/hdl.handle.net/XXXXX/test/blah`, which is
invalid because the first object contains the second.


## Parameters

### Summary

* **Name:** `encodeUTF`
    * **Description:** Decides whether "dangerous" characters will be replaced by a
      defined replacement string or it's utf code (e.g. `=u0020` for blank char)
    * **Type:** bool
    * **Default:** false
* **Name:** `maxFilenameLen`
    * **Description:** Determines the maximum number of characters within parts of the full pathname separated by `/` (files or folders).   
      A result with more characters will raise an error.
    * **Type:** number
    * **Constraints:** An integer greater than 0
    * **Default:** 127
* **Name:** `maxPathnameLen`
    * **Description:** Determines the maximum number of result characters.
      A result with more characters will raise an error.
    * **Type:** number
    * **Constraints:** An integer greater than 0
    * **Default:** 32000
* **Name:** `replacementString`
    * **Description:** String that is used to replace non-whitespace characters
      which need replacement.  
      If `encodeUTF == true` only used for replacement of non-UTF8 characters.
    * **Type:** string
    * **Default:** "_"
* **Name:** `whitespaceReplacementString`
    * **Description:** String that is used to replace [whitespaces](https://en.wikipedia.org/wiki/Template:Whitespace_(Unicode)).  
      Only used if `encodeUTF == false`.  
      **Hint:** if you want to remove (inner) whitespaces, just use the empty string
    * **Type:** string
    * **Default:** " " (U+0020)
* **Name:** `fallbackDigestAlgorithm`
    * **Description:** Name of the digest algorithm to use for generating fallback filenames.
      Restricted to Algorithms supported by OCFL
    * **Type:** string
    * **Default:** `md5`
* **Name:** `fallbackFolder`
    * **Description:** Name of the folder, the fallback files will appear
    * **Type:** string
    * **Default:** `fallback`
* **Name:** `fallbackSubdirs`
    * **Description:** Number of sub-folder build from prefix of fallback digest.  
      **Hint:** only needed if a large number of fallback entries is expected
    * **Type:** string
    * **Default:** `0`

### Definition of terms
* **`maxFilenameLen`/`maxPathnameLen`:**
  Several filesystems (e.g. Ext2/3/4) have byte restrictions on filename length.
  When using UTF16 (i.e. NTFS) or UTF32 characters in the filesystem, the byte length is double or quad of the character length.
  For filesystems which are using UTF8 character sets, the length of a name in bytes
  can be calculated only by building the UTF8 string and count the byte length afterwards.

  **Hint:** UTF8 has always less or equal bytes than UTF32 which means, that
  assuming UTF32 instead of UTF8 for length calculation is safe, but would give you
  only 63 characters on 255 byte restrictions.

## Procedure

The following is an outline to the steps for mapping an identifier/filepath.

[UTF Replacement Character List](#utf-replacement-character-list)
is a list of UTF characters mentioned below.

### When `utfEncode` is `true`

1. Replace all non-UTF8 characters with `replacementString`
2. Split the string at path separator `/`
3. For each part do the following
    1. Replace `=` with `=u003D` if it is followed by `u` and four hex digits
    2. Replace any character from this list with its utf code in the form `=uXXXX`
       where `XXXX` is the code:
       `U+0000-U+001F` `U+007F` `U+0020` `U+0085` `U+00A0` `U+1680` `U+2000-U+200F`
       `U+2028` `U+2029` `U+202F` `U+205F` `U+3000` `\n` `\t` `*` `?` `:` `[` `]` `"`
       `<` `>` `|` `(` `)` `{` `}` `&` `'` `!` `;` `#` `@`
    3. If part only contains periods, replace first period (`.`), with UTF Code (`U+002E`)
    4. Remove part completely, if its length is 0
    5. When length of part is larger than `maxFilenameLen` use fallback function and return
4. Join the parts with path separator `/`
5. When length of result is larger than `maxPathnameLen` use fallback function and return

### When `utfEncode` is `false`

1. Replace all non-UTF8 characters with `replacementString`
2. Split the string at path separator `/`
3. For each part do the following
    1. Replace any whitespace character from this list with `whitespaceReplacementString`:
       `U+0009` `U+000A-U+000D` `U+0020` `U+0085` `U+00A0` `U+1680` `U+2000-U+200F`
       `U+2028` `U+2029` `U+202F` `U+205F` `U+3000`
    2. Replace any character from this list with `replacementString`:
       `U+0000-U+001F` `U+007f` `*` `?` `:` `[` `]` `"`
       `<` `>` `|` `(` `)` `{` `}` `&` `'` `!` `;` `#` `@`
    3. Remove leading spaces, `-` and `~` / remove trailing spaces
    4. If part only contains periods, replace first period (`.`) with `replacementString`
    5. Remove part completely, if its length is 0
    6. When length of part is larger than `maxFilenameLen` use fallback function and return
4. Join the parts with path separator `/`
5. When length of result is larger than `maxPathnameLen` use fallback function and return

### Fallback function

1. Create digest with `fallbackDigestAlgorithm` from initial parameter as lower case hex string
2. When digest is longer than `maxFilenameLen` add folder separator `/` after `maxFilenameLen` (character or byte based)
   until all parts are smaller or equal `maxFilenameLen`
3. Prepend `fallbackSubdirs` number of digest characters as prefix folders with separator `/`
4. Prepend `fallbackFolder` with separator `/`


## Examples

### Parameters

It is not necessary to specify any parameters to use the default configuration.
However, if you were to do so, it would look like the following:

```json
{
    "extensionName": "NNNN-direct-clean-path-layout",
    "maxFilenameLen": 127,
    "maxFilenameLen": 32000,
    "utfEncode": false,
    "replacementString": "_",
    "whitespaceReplacementString": " ",
    "fallbackDigestAlgorithm": "md5",
    "fallbackFolder": "fallback",
    "fallbackSubdirs": 0
}
```

### Mappings

#### #1 `utfEncode == false`

```json
{
    "extensionName": "NNNN-direct-clean-path-layout",
    "maxFilenameLen": 127,
    "maxFilenameLen": 32000,
    "utfEncode": false,
    "replacementString": "_",
    "whitespaceReplacementString": " ",
    "fallbackDigestAlgorithm": "md5",
    "fallbackFolder": "fallback",
    "fallbackSubdirs": 2
}
```

| ID or Path                                                                                                                                                                                                                                                                         | Result                                          |
|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------------------------------------------|
| `..hor_rib:lé-$id`                                                                                                                                                                                                                                                                 | `..hor_rib_lé-$id`                              |
| `info:fedora/object-01`                                                                                                                                                                                                                                                            | `info_fedora/object-01`                         |
| `~ info:fedora/-obj#ec@t-"01 `                                                                                                                                                                                                                                                     | `info_fedora/obj_ec_t-_01`                      |
| `/test/ ~/.../blah`                                                                                                                                                                                                                                                                | `test/_../blah`                                 |
| `https://hdl.handle.net/XXXXX/test/bl ah`                                                                                                                                                                                                                                          | `https_/hdl.handle.net/XXXXX/test/bl ah`        |
| `abcdefghijabcdefghij abcdefghijabcdefghij abcdefghijabcdefghij abcdefghijabcdefghij abcdefghijabcdefghij abcdefghijabcdefghij abcdefghijabcdefghij abcdefghijabcdefghij abcdefghijabcdefghij abcdefghijabcdefghij abcdefghijabcdefghij abcdefghijabcdefghij abcdefghijabcdefghij` | `fallback/0/e/0eafabb38fa7f1583d1461afe980ebdc` |

#### #2 `utfEncode == true`

```json
{
    "extensionName": "NNNN-direct-clean-path-layout",
    "maxFilenameLen": 127,
    "maxFilenameLen": 32000,
    "utfEncode": true,
    "replacementString": "_",
    "whitespaceReplacementString": " ",
    "fallbackDigestAlgorithm": "sha512",
    "fallbackFolder": "fallback",
    "fallbackSubdirs": 2
}
```

| ID or Path                                                                                                                                                                                                                                                                         | Result                                                                                                                                           |
|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------|
| `..hor_rib:lé-$id`                                                                                                                                                                                                                                                                 | `..hor_rib=u003Alé-$id`                                                                                                                          |
| `object=u123a-01`                                                                                                                                                                                                                                                                  | `object=u003Du123a-01`                                                                                                                           |
| `object=u13a-01`                                                                                                                                                                                                                                                                   | `object=u13a-01`                                                                                                                                 |
| `info:fedora/object-01`                                                                                                                                                                                                                                                            | `info=u003Afedora/object-01`                                                                                                                     |
| `~ info:fedora/-obj#ec@t-"01 `                                                                                                                                                                                                                                                     | `=u007E=u0020info=u003Afedora/-obj=u0023ec=u0040t-=u002201=u0020`                                                                                |
| `/test/ ~/.../blah`                                                                                                                                                                                                                                                                | `test/=u0020~/=u002E../blah`                                                                                                                     |
| `https://hdl.handle.net/XXXXX/test/bl ah`                                                                                                                                                                                                                                          | `https=u003A/hdl.handle.net/XXXXX/test/bl=u0020ah`                                                                                               |
| `abcdefghijabcdefghij abcdefghijabcdefghij abcdefghijabcdefghij abcdefghijabcdefghij abcdefghijabcdefghij abcdefghijabcdefghij abcdefghijabcdefghij abcdefghijabcdefghij abcdefghijabcdefghij abcdefghijabcdefghij abcdefghijabcdefghij abcdefghijabcdefghij abcdefghijabcdefghij` | `fallback/b/8/b8acda4abac53237afa03d6bbb078e1bf46b40438bb256df79b8d9ff0e57b32a688156ad21755363ea19953c160c4dd6d4db175b71e9aa87d68937181a9f69d/9` |
### Implementation

#### GO

```go
package cleanpath

import (
	"emperror.dev/errors"
	"regexp"
	"strings"
	[...]
)


var directCleanRuleAll = regexp.MustCompile("[\u0000-\u001f\u007f\u0020\u0085\u00a0\u1680\u2000-\u200f\u2028\u2029\u202f\u205f\u3000\n\t*?:\\[\\]\"<>|(){}&'!\\;#@]")
var directCleanRuleWhitespace = regexp.MustCompile("[\u0009\u000a-\u000d\u0020\u0085\u00a0\u1680\u2000-\u200f\u2028\u2029\u202f\u205f\u3000]")
var directCleanRuleEqual = regexp.MustCompile("=(u[a-zA-Z0-9]{4})")
var directCleanRule_1_5 = regexp.MustCompile("[\u0000-\u001F\u007F\n\r\t*?:\\[\\]\"<>|(){}&'!\\;#@]")
var directCleanRule_2_4_6 = regexp.MustCompile("^[\\-~\u0009\u000a-\u000d\u0020\u0085\u00a0\u1680\u2000-\u200f\u2028\u2029\u202f\u205f\u3000]*(.*?)[\u0009\u000a-\u000d\u0020\u0085\u00a0\u1680\u2000-\u20a0\u2028\u2029\u202f\u205f\u3000]*$")
var directCleanRulePeriods = regexp.MustCompile("^\\.+$")

var directCleanErrFilenameTooLong = errors.New("filename too long")
var directCleanErrPathnameTooLong = errors.New("pathname too long")

type DirectCleanConfig struct {
    [...]
	MaxPathnameLen              int                      `json:"maxPathnameLen"`
	MaxFilenameLen              int                      `json:"maxFilenameLen"`
	ReplacementString           string                   `json:"replacementString"`
	WhitespaceReplacementString string                   `json:"whitespaceReplacementString"`
	UTFEncode                   bool                     `json:"utfEncode"`
	FallbackDigestAlgorithm     checksum.DigestAlgorithm `json:"fallbackDigestAlgorithm"`
	FallbackFolder              string                   `json:"fallbackFolder"`
	FallbackSubFolders          int                      `json:"fallbackSubdirs"`
	hash                        hash.Hash                `json:"-"`
	hashMutex                   sync.Mutex               `json:"-"`
}

[...]


func encodeUTFCode(s string) string {
	return "=u" + strings.Trim(fmt.Sprintf("%U", []rune(s)), "U+[]")
}

func (sl *DirectClean) fallback(fname string) (string, error) {
	// internal mutex for reusing the hash object
	sl.hashMutex.Lock()

	// reset hash function
	sl.hash.Reset()
	// add path
	if _, err := sl.hash.Write([]byte(fname)); err != nil {
		sl.hashMutex.Unlock()
		return "", errors.Wrapf(err, "cannot hash path '%s'", fname)
	}
	sl.hashMutex.Unlock()

	// get digest and encode it
	digestString := hex.EncodeToString(sl.hash.Sum(nil))

	// check whether digest fits in filename length
	parts := len(digestString) / sl.MaxFilenameLen
	rest := len(digestString) % sl.MaxFilenameLen
	if rest > 0 {
		parts++
	}
	// cut the digest if it's too long for filename length
	result := ""
	for i := 0; i < parts; i++ {
		result = filepath.Join(result, digestString[i*sl.MaxFilenameLen:min((i+1)*sl.MaxFilenameLen, len(digestString))])
	}

	// add all necessary subfolders
	for i := 0; i < sl.FallbackSubFolders; i++ {
		// paranoia, but safe
		result = filepath.Join(string(([]rune(digestString))[sl.FallbackSubFolders-i-1]), result)
	}
	/*
		result = filepath.Join(sl.FallbackFolder, result)
		result = filepath.Clean(result)
		result = filepath.ToSlash(result)
		result = strings.TrimLeft(result, "/")
	*/
	result = strings.TrimLeft(filepath.ToSlash(filepath.Clean(filepath.Join(sl.FallbackFolder, result))), "/")
	if len(result) > sl.MaxPathnameLen {
		return result, errors.Errorf("result has length of %d which is more than max allowed length of %d", len(result), sl.MaxPathnameLen)
	}
	return result, nil
}

func (sl *DirectClean) BuildPath(fname string) (string, error) {

	// 1. Replace all non-UTF8 characters with replacementString
	fname = strings.ToValidUTF8(fname, sl.ReplacementString)

	// 2. Split the string at path separator /
	names := strings.Split(fname, "/")
	result := []string{}

	for _, n := range names {
		if len(n) == 0 {
			continue
		}
		if sl.UTFEncode {
			// Replace `=` with `=u003D` if it is followed by `u` and four hex digits 
			n = directCleanRuleEqual.ReplaceAllString(n, "=u003D$1")
			n = directCleanRuleAll.ReplaceAllStringFunc(n, encodeUTFCode)
			if n[0] == '~' || directCleanRulePeriods.MatchString(n) {
				n = encodeUTFCode(string(n[0])) + n[1:]
			}
		} else {
			// Replace any whitespace character from this list with whitespaceReplacementString: U+0009 U+000A-U+000D U+0020 U+0085 U+00A0 U+1680 U+2000-U+200F U+2028 U+2029 U+202F U+205F U+3000
			n = directCleanRuleWhitespace.ReplaceAllString(n, sl.WhitespaceReplacementString)
			n = directCleanRule_1_5.ReplaceAllString(n, sl.ReplacementString)
			n = directCleanRule_2_4_6.ReplaceAllString(n, "$1")
			if directCleanRulePeriods.MatchString(n) {
				n = sl.ReplacementString + n[1:]
			}
		}

		lenN := len(n)
		if lenN > sl.MaxFilenameLen {
			return sl.fallback(fname)
			//return "", errors.Wrapf(directCleanErrFilenameTooLong, "filename: %s", n)
		}

		if lenN > 0 {
			result = append(result, n)
		}
	}

	fname = strings.Join(result, "/")

	if len(fname) > sl.MaxPathnameLen {
		return sl.fallback(fname)
		//return "", errors.Wrapf(directCleanErrPathnameTooLong, "pathname: %s", fname)
	}

	return fname, nil
}

[...]
```

# Appendix

## UTF Replacement Character List

| Code	  | Decimal	 | Octal     | 	Description                   | 	Abbreviation / Key       |
|--------|----------|-----------|--------------------------------|---------------------------|
| U+0000 | 0        | 0         | Null character                 | NUL                       |
| U+0001 | 1        | 1         | Start of Heading               | SOH / Ctrl-A              |
| U+0002 | 2        | 2         | Start of Text                  | STX / Ctrl-B              |
| U+0003 | 3        | 3         | End-of-text character          | ETX / Ctrl-C1             |
| U+0004 | 4        | 4         | End-of-transmission character  | EOT / Ctrl-D2             |
| U+0005 | 5        | 5         | Enquiry character              | ENQ / Ctrl-E              |
| U+0006 | 6        | 6         | Acknowledge character          | ACK / Ctrl-F              |
| U+0007 | 7        | 7         | Bell character                 | BEL / Ctrl-G3             |
| U+0008 | 8        | 10        | Backspace                      | BS / Ctrl-H               |
| U+0009 | 9        | 11        | Horizontal tab                 | HT / Ctrl-I               |
| U+000A | 10       | 12        | Line feed                      | LF / Ctrl-J4              |
| U+000B | 11       | 13        | Vertical tab                   | VT / Ctrl-K               |
| U+000C | 12       | 14        | Form feed                      | FF / Ctrl-L               |
| U+000D | 13       | 15        | Carriage return                | CR / Ctrl-M5              |
| U+000E | 14       | 16        | Shift Out                      | SO / Ctrl-N               |
| U+000F | 15       | 17        | Shift In                       | SI / Ctrl-O6              |
| U+0010 | 16       | 20        | Data Link Escape               | DLE / Ctrl-P              |
| U+0011 | 17       | 21        | Device Control 1               | DC1 / Ctrl-Q7             |
| U+0012 | 18       | 22        | Device Control 2               | DC2 / Ctrl-R              |
| U+0013 | 19       | 23        | Device Control 3               | DC3 / Ctrl-S8             |
| U+0014 | 20       | 24        | Device Control 4               | DC4 / Ctrl-T              |
| U+0015 | 21       | 25        | Negative-acknowledge character | NAK / Ctrl-U9             |
| U+0016 | 22       | 26        | Synchronous Idle               | SYN / Ctrl-V              |
| U+0017 | 23       | 27        | End of Transmission Block      | ETB / Ctrl-W              |
| U+0018 | 24       | 30        | Cancel character               | CAN / Ctrl-X10            |
| U+0019 | 25       | 31        | End of Medium                  | EM / Ctrl-Y               |
| U+001A | 26       | 32        | Substitute character           | SUB / Ctrl-Z11            |
| U+001B | 27       | 33        | Escape character               | ESC                       |
| U+001C | 28       | 34        | File Separator                 | FS                        |
| U+001D | 29       | 35        | Group Separator                | GS                        |
| U+001E | 30       | 36        | Record Separator               | RS                        |
| U+001F | 31       | 37        | Unit Separator                 | US                        |
| U+001F | 31       | 37        | Unit Separator                 | US                        |
| U+0020 | 32       | 40        | Space                          |                           |
| U+007F | 127      | 177       | Delete                         | DEL                       |
| U+0085 | 133      | 0302 0205 | Next Line                      | NEL                       |
| U+00A0 | 160      | 0302 0240 | &nbsp;                         | Non-breaking space        |
| U+1680 | 5760     | 13200     |                                | OGHAM SPACE MARK          |
| U+2000 | 8192     | 20000     |                                | EN QUAD                   |
| U+2001 | 8193     | 20001     |                                | EM QUAD                   |
| U+2002 | 8194     | 20002     |                                | EN SPACE                  |
| U+2003 | 8195     | 20003     |                                | EM SPACE                  |
| U+2004 | 8196     | 20004     |                                | THREE-PER-EM SPACE        |
| U+2005 | 8197     | 20005     |                                | FOUR-PER-EM SPACE         |
| U+2006 | 8198     | 20006     |                                | SIX-PER-EM SPACE          |
| U+2007 | 8199     | 20007     |                                | FIGURE SPACE              |
| U+2008 | 8200     | 20010     |                                | PUNCTUATION SPACE         |
| U+2009 | 8201     | 20011     |                                | THIN SPACE                |
| U+200A | 8202     | 20012     |                                | HAIR SPACE                |
| U+200B | 8203     | 20013     |                                | ZERO WIDTH SPACE          |
| U+200C | 8204     | 20014     |                                | ZERO WIDTH NON-JOINER     |
| U+200D | 8205     | 20015     |                                | ZERO WIDTH JOINER         |
| U+200E | 8206     | 20016     |                                | LEFT-TO-RIGHT MARK        |
| U+200F | 8207     | 20017     |                                | RIGHT-TO-LEFT MARK        |
| U+2028 | 8232     | 20050     |                                | LINE SEPARATOR            |
| U+2029 | 8233     | 20051     |                                | PARAGRAPH SEPARATOR       |
| U+205F | 8287     | 20137     |                                | MEDIUM MATHEMATICAL SPACE |
| U+3000 | 12288    | 30000     |                                | IDEOGRAPHIC SPACE         |


* Wikipedia contributors. (2022, November 3). List of Unicode characters. In Wikipedia, The Free Encyclopedia. Retrieved 14:00, November 4, 2022, from https://en.wikipedia.org/w/index.php?title=List_of_Unicode_characters&oldid=1119877694
* Unicode/Character reference/2000-2FFF. (2021, September 27). Wikibooks, The Free Textbook Project. Retrieved 14:29, November 4, 2022 from https://en.wikibooks.org/w/index.php?title=Unicode/Character_reference/2000-2FFF&oldid=3991460.
* Unicode/Character reference/3000-3FFF. (2020, March 18). Wikibooks, The Free Textbook Project. Retrieved 14:50, November 4, 2022 from https://en.wikibooks.org/w/index.php?title=Unicode/Character_reference/3000-3FFF&oldid=3668212.
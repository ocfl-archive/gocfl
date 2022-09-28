package ocfl

var OCFLValidationError1_1 = map[ValidationErrorCode]*ValidationError{
	E001: {Code: E001, Description: "‘The OCFL Object Root must not contain files or directories other than those specified in the following sections.’", Ref: "https://ocfl.io/draft/spec/#E001"},
	E002: {Code: E002, Description: "‘The version declaration must be formatted according to the NAMASTE specification.’", Ref: "https://ocfl.io/draft/spec/#E002"},
	E003: {Code: E003, Description: "‘There must be exactly one version declaration file in the base directory of the OCFL Object Root giving the OCFL version in the filename.’", Ref: "https://ocfl.io/draft/spec/#E003"},
	E004: {Code: E004, Description: "‘The [version declaration] filename MUST conform to the pattern T=dvalue, where T must be 0, and dvalue must be ocfl_object_, followed by the OCFL specification version number.’", Ref: "https://ocfl.io/draft/spec/#E004"},
	E005: {Code: E005, Description: "‘The [version declaration] filename must conform to the pattern T=dvalue, where T MUST be 0, and dvalue must be ocfl_object_, followed by the OCFL specification version number.’", Ref: "https://ocfl.io/draft/spec/#E005"},
	E006: {Code: E006, Description: "‘The [version declaration] filename must conform to the pattern T=dvalue, where T must be 0, and dvalue MUST be ocfl_object_, followed by the OCFL specification version number.’", Ref: "https://ocfl.io/draft/spec/#E006"},
	E007: {Code: E007, Description: "‘The text contents of the [version declaration] file must be the same as dvalue, followed by a newline (\n).’", Ref: "https://ocfl.io/draft/spec/#E007"},
	E008: {Code: E008, Description: "‘OCFL Object content must be stored as a sequence of one or more versions.’", Ref: "https://ocfl.io/draft/spec/#E008"},
	E009: {Code: E009, Description: "‘The version number sequence MUST start at 1 and must be continuous without missing integers.’", Ref: "https://ocfl.io/draft/spec/#E009"},
	E010: {Code: E010, Description: "‘The version number sequence must start at 1 and MUST be continuous without missing integers.’", Ref: "https://ocfl.io/draft/spec/#E010"},
	E011: {Code: E011, Description: "‘If zero-padded version directory numbers are used then they must start with the prefix v and then a zero.’", Ref: "https://ocfl.io/draft/spec/#E011"},
	E012: {Code: E012, Description: "‘All version directories of an object must use the same naming convention: either a non-padded version directory number, or a zero-padded version directory number of consistent length.’", Ref: "https://ocfl.io/draft/spec/#E012"},
	E013: {Code: E013, Description: "‘Operations that add a new version to an object must follow the version directory naming convention established by earlier versions.’", Ref: "https://ocfl.io/draft/spec/#E013"},
	E014: {Code: E014, Description: "‘In all cases, references to files inside version directories from inventory files must use the actual version directory names.’", Ref: "https://ocfl.io/draft/spec/#E014"},
	E015: {Code: E015, Description: "‘There must be no other files as children of a version directory, other than an inventory file and a inventory digest.’", Ref: "https://ocfl.io/draft/spec/#E015"},
	E016: {Code: E016, Description: "‘Version directories must contain a designated content sub-directory if the version contains files to be preserved, and should not contain this sub-directory otherwise.’", Ref: "https://ocfl.io/draft/spec/#E016"},
	E017: {Code: E017, Description: "‘The contentDirectory value MUST NOT contain the forward slash (/) path separator and must not be either one or two periods (. or ..).’", Ref: "https://ocfl.io/draft/spec/#E017"},
	E018: {Code: E018, Description: "‘The contentDirectory value must not contain the forward slash (/) path separator and MUST NOT be either one or two periods (. or ..).’", Ref: "https://ocfl.io/draft/spec/#E018"},
	E019: {Code: E019, Description: "‘If the key contentDirectory is set, it MUST be set in the first version of the object and must not change between versions of the same object.’", Ref: "https://ocfl.io/draft/spec/#E019"},
	E020: {Code: E020, Description: "‘If the key contentDirectory is set, it must be set in the first version of the object and MUST NOT change between versions of the same object.’", Ref: "https://ocfl.io/draft/spec/#E020"},
	E021: {Code: E021, Description: "‘If the key contentDirectory is not present in the inventory file then the name of the designated content sub-directory must be content.’", Ref: "https://ocfl.io/draft/spec/#E021"},
	E022: {Code: E022, Description: "‘OCFL-compliant tools (including any validators) must ignore all directories in the object version directory except for the designated content directory.’", Ref: "https://ocfl.io/draft/spec/#E022"},
	E023: {Code: E023, Description: "‘Every file within a version's content directory must be referenced in the manifest section of the inventory.’", Ref: "https://ocfl.io/draft/spec/#E023"},
	E024: {Code: E024, Description: "‘There must not be empty directories within a version's content directory.’", Ref: "https://ocfl.io/draft/spec/#E024"},
	E025: {Code: E025, Description: "‘For content-addressing, OCFL Objects must use either sha512 or sha256, and should use sha512.’", Ref: "https://ocfl.io/draft/spec/#E025"},
	E026: {Code: E026, Description: "‘For storage of additional fixity values, or to support legacy content migration, implementers must choose from the following controlled vocabulary of digest algorithms, or from a list of additional algorithms given in the [Digest-Algorithms-Extension].’", Ref: "https://ocfl.io/draft/spec/#E026"},
	E027: {Code: E027, Description: "‘OCFL clients must support all fixity algorithms given in the table below, and may support additional algorithms from the extensions.’", Ref: "https://ocfl.io/draft/spec/#E027"},
	E028: {Code: E028, Description: "‘Optional fixity algorithms that are not supported by a client must be ignored by that client.’", Ref: "https://ocfl.io/draft/spec/#E028"},
	E029: {Code: E029, Description: "‘SHA-1 algorithm defined by [FIPS-180-4] and must be encoded using hex (base16) encoding [RFC4648].’", Ref: "https://ocfl.io/draft/spec/#E029"},
	E030: {Code: E030, Description: "‘SHA-256 algorithm defined by [FIPS-180-4] and must be encoded using hex (base16) encoding [RFC4648].’", Ref: "https://ocfl.io/draft/spec/#E030"},
	E031: {Code: E031, Description: "‘SHA-512 algorithm defined by [FIPS-180-4] and must be encoded using hex (base16) encoding [RFC4648].’", Ref: "https://ocfl.io/draft/spec/#E031"},
	E032: {Code: E032, Description: "‘[blake2b-512] must be encoded using hex (base16) encoding [RFC4648].’", Ref: "https://ocfl.io/draft/spec/#E032"},
	E033: {Code: E033, Description: "‘An OCFL Object Inventory MUST follow the [JSON] structure described in this section and must be named inventory.json.’", Ref: "https://ocfl.io/draft/spec/#E033"},
	E034: {Code: E034, Description: "‘An OCFL Object Inventory must follow the [JSON] structure described in this section and MUST be named inventory.json.’", Ref: "https://ocfl.io/draft/spec/#E034"},
	E035: {Code: E035, Description: "‘The forward slash (/) path separator must be used in content paths in the manifest and fixity blocks within the inventory.’", Ref: "https://ocfl.io/draft/spec/#E035"},
	E036: {Code: E036, Description: "‘An OCFL Object Inventory must include the following keys: [id, type, digestAlgorithm, head]’", Ref: "https://ocfl.io/draft/spec/#E036"},
	E037: {Code: E037, Description: "‘[id] must be unique in the local context, and should be a URI [RFC3986].’", Ref: "https://ocfl.io/draft/spec/#E037"},
	E038: {Code: E038, Description: "‘In the object root inventory [the type value] must be the URI of the inventory section of the specification version matching the object conformance declaration.’", Ref: "https://ocfl.io/draft/spec/#E038"},
	E039: {Code: E039, Description: "‘[digestAlgorithm] must be the algorithm used in the manifest and state blocks.’", Ref: "https://ocfl.io/draft/spec/#E039"},
	E040: {Code: E040, Description: "[head] must be the version directory name with the highest version number.’", Ref: "https://ocfl.io/draft/spec/#E040"},
	E041: {Code: E041, Description: "‘In addition to these keys, there must be two other blocks present, manifest and versions, which are discussed in the next two sections.’", Ref: "https://ocfl.io/draft/spec/#E041"},
	E042: {Code: E042, Description: "‘Content paths within a manifest block must be relative to the OCFL Object Root.’", Ref: "https://ocfl.io/draft/spec/#E042"},
	E043: {Code: E043, Description: "‘An OCFL Object Inventory must include a block for storing versions.’", Ref: "https://ocfl.io/draft/spec/#E043"},
	E044: {Code: E044, Description: "‘This block MUST have the key of versions within the inventory, and it must be a JSON object.’", Ref: "https://ocfl.io/draft/spec/#E044"},
	E045: {Code: E045, Description: "‘This block must have the key of versions within the inventory, and it MUST be a JSON object.’", Ref: "https://ocfl.io/draft/spec/#E045"},
	E046: {Code: E046, Description: "‘The keys of [the versions object] must correspond to the names of the version directories used.’", Ref: "https://ocfl.io/draft/spec/#E046"},
	E047: {Code: E047, Description: "‘Each value [of the versions object] must be another JSON object that characterizes the version, as described in the 3.5.3.1 Version section.’", Ref: "https://ocfl.io/draft/spec/#E047"},
	E048: {Code: E048, Description: "‘A JSON object to describe one OCFL Version, which must include the following keys: [created, state]’", Ref: "https://ocfl.io/draft/spec/#E048"},
	E049: {Code: E049, Description: "‘[the value of the “created” key] must be expressed in the Internet Date/Time Format defined by [RFC3339].’", Ref: "https://ocfl.io/draft/spec/#E049"},
	E050: {Code: E050, Description: "‘The keys of [the “state” JSON object] are digest values, each of which must correspond to an entry in the manifest of the inventory.’", Ref: "https://ocfl.io/draft/spec/#E050"},
	E051: {Code: E051, Description: "‘The logical path [value of a “state” digest key] must be interpreted as a set of one or more path elements joined by a / path separator.’", Ref: "https://ocfl.io/draft/spec/#E051"},
	E052: {Code: E052, Description: "‘[logical] Path elements must not be ., .., or empty (//).’", Ref: "https://ocfl.io/draft/spec/#E052"},
	E053: {Code: E053, Description: "‘Additionally, a logical path must not begin or end with a forward slash (/).’", Ref: "https://ocfl.io/draft/spec/#E053"},
	E054: {Code: E054, Description: "‘The value of the user key must contain a user name key, “name” and should contain an address key, “address”.’", Ref: "https://ocfl.io/draft/spec/#E054"},
	E055: {Code: E055, Description: "‘If present, [the fixity] block must have the key of fixity within the inventory.’", Ref: "https://ocfl.io/draft/spec/#E055"},
	E056: {Code: E056, Description: "‘The fixity block must contain keys corresponding to the controlled vocabulary given in the digest algorithms listed in the Digests section, or in a table given in an Extension.’", Ref: "https://ocfl.io/draft/spec/#E056"},
	E057: {Code: E057, Description: "‘The value of the fixity block for a particular digest algorithm must follow the structure of the manifest block; that is, a key corresponding to the digest value, and an array of content paths that match that digest.’", Ref: "https://ocfl.io/draft/spec/#E057’"},
	E058: {Code: E058, Description: "‘Every occurrence of an inventory file must have an accompanying sidecar file stating its digest.’", Ref: "https://ocfl.io/draft/spec/#E058"},
	E059: {Code: E059, Description: "‘This value must match the value given for the digestAlgorithm key in the inventory.’", Ref: "https://ocfl.io/draft/spec/#E059"},
	E060: {Code: E060, Description: "‘The digest sidecar file must contain the digest of the inventory file.’", Ref: "https://ocfl.io/draft/spec/#E060"},
	E061: {Code: E061, Description: "‘[The digest sidecar file] must follow the format: DIGEST inventory.json’", Ref: "https://ocfl.io/draft/spec/#E061"},
	E062: {Code: E062, Description: "‘The digest of the inventory must be computed only after all changes to the inventory have been made, and thus writing the digest sidecar file is the last step in the versioning process.’", Ref: "https://ocfl.io/draft/spec/#E062"},
	E063: {Code: E063, Description: "‘Every OCFL Object must have an inventory file within the OCFL Object Root, corresponding to the state of the OCFL Object at the current version.’", Ref: "https://ocfl.io/draft/spec/#E063"},
	E064: {Code: E064, Description: "‘Where an OCFL Object contains inventory.json in version directories, the inventory file in the OCFL Object Root must be the same as the file in the most recent version.’", Ref: "https://ocfl.io/draft/spec/#E064"},
	E066: {Code: E066, Description: "‘Each version block in each prior inventory file must represent the same object state as the corresponding version block in the current inventory file.’", Ref: "https://ocfl.io/draft/spec/#E066"},
	E067: {Code: E067, Description: "‘The extensions directory must not contain any files or sub-directories other than extension sub-directories.’", Ref: "https://ocfl.io/draft/spec/#E067"},
	E069: {Code: E069, Description: "‘An OCFL Storage Root MUST contain a Root Conformance Declaration identifying it as such.’", Ref: "https://ocfl.io/draft/spec/#E069"},
	E070: {Code: E070, Description: "‘If present, [the ocfl_layout.json document] MUST include the following two keys in the root JSON object: [extension, description]’", Ref: "https://ocfl.io/draft/spec/#E070"},
	E071: {Code: E071, Description: "‘The value of the [ocfl_layout.json] extension key must be the registered extension name for the extension defining the arrangement under the storage root.’", Ref: "https://ocfl.io/draft/spec/#E071"},
	E072: {Code: E072, Description: "‘The directory hierarchy used to store OCFL Objects MUST NOT contain files that are not part of an OCFL Object.’", Ref: "https://ocfl.io/draft/spec/#E072"},
	E073: {Code: E073, Description: "‘Empty directories MUST NOT appear under a storage root.’", Ref: "https://ocfl.io/draft/spec/#E073"},
	E074: {Code: E074, Description: "‘Although implementations may require multiple OCFL Storage Roots - that is, several logical or physical volumes, or multiple “buckets” in an object store - each OCFL Storage Root MUST be independent.’", Ref: "https://ocfl.io/draft/spec/#E074"},
	E075: {Code: E075, Description: "‘The OCFL version declaration MUST be formatted according to the NAMASTE specification.’", Ref: "https://ocfl.io/draft/spec/#E075"},
	E076: {Code: E076, Description: "‘There must be exactly one version declaration file in the base directory of the OCFL Storage Root giving the OCFL version in the filename.’", Ref: "https://ocfl.io/draft/spec/#E076"},
	E077: {Code: E077, Description: "‘[The OCFL version declaration filename] MUST conform to the pattern T=dvalue, where T must be 0, and dvalue must be ocfl_, followed by the OCFL specification version number.’", Ref: "https://ocfl.io/draft/spec/#E077"},
	E078: {Code: E078, Description: "‘[The OCFL version declaration filename] must conform to the pattern T=dvalue, where T MUST be 0, and dvalue must be ocfl_, followed by the OCFL specification version number.’", Ref: "https://ocfl.io/draft/spec/#E078"},
	E079: {Code: E079, Description: "‘[The OCFL version declaration filename] must conform to the pattern T=dvalue, where T must be 0, and dvalue MUST be ocfl_, followed by the OCFL specification version number.’", Ref: "https://ocfl.io/draft/spec/#E079"},
	E080: {Code: E080, Description: "‘The text contents of [the OCFL version declaration file] MUST be the same as dvalue, followed by a newline (\n).’", Ref: "https://ocfl.io/draft/spec/#E080"},
	E081: {Code: E081, Description: "‘OCFL Objects within the OCFL Storage Root also include a conformance declaration which MUST indicate OCFL Object conformance to the same or earlier version of the specification.’", Ref: "https://ocfl.io/draft/spec/#E081"},
	E082: {Code: E082, Description: "‘OCFL Object Roots MUST be stored either as the terminal resource at the end of a directory storage hierarchy or as direct children of a containing OCFL Storage Root.’", Ref: "https://ocfl.io/draft/spec/#E082"},
	E083: {Code: E083, Description: "‘There MUST be a deterministic mapping from an object identifier to a unique storage path.’", Ref: "https://ocfl.io/draft/spec/#E083"},
	E084: {Code: E084, Description: "‘Storage hierarchies MUST NOT include files within intermediate directories.’", Ref: "https://ocfl.io/draft/spec/#E084"},
	E085: {Code: E085, Description: "‘Storage hierarchies MUST be terminated by OCFL Object Roots.’", Ref: "https://ocfl.io/draft/spec/#E085"},
	E087: {Code: E087, Description: "‘An OCFL validator MUST ignore any files in the storage root it does not understand.’", Ref: "https://ocfl.io/draft/spec/#E087"},
	E088: {Code: E088, Description: "‘An OCFL Storage Root MUST NOT contain directories or sub-directories other than as a directory hierarchy used to store OCFL Objects or for storage root extensions.’", Ref: "https://ocfl.io/draft/spec/#E088"},
	E089: {Code: E089, Description: "‘If the preservation of non-OCFL-compliant features is required then the content MUST be wrapped in a suitable disk or filesystem image format which OCFL can treat as a regular file.’", Ref: "https://ocfl.io/draft/spec/#E089"},
	E090: {Code: E090, Description: "‘Hard and soft (symbolic) links are not portable and MUST NOT be used within OCFL Storage hierarchies.’", Ref: "https://ocfl.io/draft/spec/#E090"},
	E091: {Code: E091, Description: "‘Filesystems MUST preserve the case of OCFL filepaths and filenames.’", Ref: "https://ocfl.io/draft/spec/#E091"},
	E092: {Code: E092, Description: "‘The value for each key in the manifest must be an array containing the content paths of files in the OCFL Object that have content with the given digest.’", Ref: "https://ocfl.io/draft/spec/#E092"},
	E093: {Code: E093, Description: "‘Where included in the fixity block, the digest values given must match the digests of the files at the corresponding content paths.’", Ref: "https://ocfl.io/draft/spec/#E093"},
	E094: {Code: E094, Description: "‘The value of [the message] key is freeform text, used to record the rationale for creating this version. It must be a JSON string.’", Ref: "https://ocfl.io/draft/spec/#E094"},
	E095: {Code: E095, Description: "‘Within a version, logical paths must be unique and non-conflicting, so the logical path for a file cannot appear as the initial part of another logical path.’", Ref: "https://ocfl.io/draft/spec/#E095"},
	E096: {Code: E096, Description: "‘As JSON keys are case sensitive, while digests may not be, there is an additional requirement that each digest value must occur only once in the manifest regardless of case.’", Ref: "https://ocfl.io/draft/spec/#E096"},
	E097: {Code: E097, Description: "‘As JSON keys are case sensitive, while digests may not be, there is an additional requirement that each digest value must occur only once in the fixity block for any digest algorithm, regardless of case.’", Ref: "https://ocfl.io/draft/spec/#E097"},
	E098: {Code: E098, Description: "‘The content path must be interpreted as a set of one or more path elements joined by a / path separator.’", Ref: "https://ocfl.io/draft/spec/#E098"},
	E099: {Code: E099, Description: "‘[content] path elements must not be ., .., or empty (//).’", Ref: "https://ocfl.io/draft/spec/#E099"},
	E100: {Code: E100, Description: "‘A content path must not begin or end with a forward slash (/).’", Ref: "https://ocfl.io/draft/spec/#E100"},
	E101: {Code: E101, Description: "‘Within an inventory, content paths must be unique and non-conflicting, so the content path for a file cannot appear as the initial part of another content path.’", Ref: "https://ocfl.io/draft/spec/#E101"},
	E102: {Code: E102, Description: "‘An inventory file must not contain keys that are not specified.’", Ref: "https://ocfl.io/draft/spec/#E102"},
	E103: {Code: E103, Description: "‘Each version directory within an OCFL Object MUST conform to either the same or a later OCFL specification version as the preceding version directory.’", Ref: "https://ocfl.io/draft/spec/#E103"},
	E104: {Code: E104, Description: "‘Version directory names MUST be constructed by prepending v to the version number.’", Ref: "https://ocfl.io/draft/spec/#E104"},
	E105: {Code: E105, Description: "‘The version number MUST be taken from the sequence of positive, base-ten integers: 1, 2, 3, etc.’", Ref: "https://ocfl.io/draft/spec/#E105"},
	E106: {Code: E106, Description: "‘The value of the manifest key MUST be a JSON object.’", Ref: "https://ocfl.io/draft/spec/#E106"},
	E107: {Code: E107, Description: "‘The value of the manifest key must be a JSON object, and each key MUST correspond to a digest value key found in one or more state blocks of the current and/or previous version blocks of the OCFL Object.’", Ref: "https://ocfl.io/draft/spec/#E107"},
	E108: {Code: E108, Description: "‘The contentDirectory value MUST represent a direct child directory of the version directory in which it is found.’", Ref: "https://ocfl.io/draft/spec/#E108"},
	E110: {Code: E110, Description: "‘A unique identifier for the OCFL Object MUST NOT change between versions of the same object.’", Ref: "https://ocfl.io/draft/spec/#E110"},
	E111: {Code: E111, Description: "‘If present, [the value of the fixity key] MUST be a JSON object, which may be empty.’", Ref: "https://ocfl.io/draft/spec/#E111"},
	E112: {Code: E112, Description: "‘The extensions directory must not contain any files or sub-directories other than extension sub-directories.’", Ref: "https://ocfl.io/draft/spec/#E112"},
	W001: {Code: W001, Description: "‘Implementations SHOULD use version directory names constructed without zero-padding the version number, ie. v1, v2, v3, etc.’’", Ref: "https://ocfl.io/draft/spec/#W001"},
	W002: {Code: W002, Description: "‘The version directory SHOULD NOT contain any directories other than the designated content sub-directory. Once created, the contents of a version directory are expected to be immutable.’", Ref: "https://ocfl.io/draft/spec/#W002"},
	W003: {Code: W003, Description: "‘Version directories must contain a designated content sub-directory if the version contains files to be preserved, and SHOULD NOT contain this sub-directory otherwise.’", Ref: "https://ocfl.io/draft/spec/#W003"},
	W004: {Code: W004, Description: "‘For content-addressing, OCFL Objects SHOULD use sha512.’", Ref: "https://ocfl.io/draft/spec/#W004"},
	W005: {Code: W005, Description: "‘The OCFL Object Inventory id SHOULD be a URI.’", Ref: "https://ocfl.io/draft/spec/#W005"},
	W007: {Code: W007, Description: "‘In the OCFL Object Inventory, the JSON object describing an OCFL Version, SHOULD include the message and user keys.’", Ref: "https://ocfl.io/draft/spec/#W007"},
	W008: {Code: W008, Description: "‘In the OCFL Object Inventory, in the version block, the value of the user key SHOULD contain an address key, address.’", Ref: "https://ocfl.io/draft/spec/#W008"},
	W009: {Code: W009, Description: "‘In the OCFL Object Inventory, in the version block, the address value SHOULD be a URI: either a mailto URI [RFC6068] with the e-mail address of the user or a URL to a personal identifier, e.g., an ORCID iD.’", Ref: "https://ocfl.io/draft/spec/#W009"},
	W010: {Code: W010, Description: "‘In addition to the inventory in the OCFL Object Root, every version directory SHOULD include an inventory file that is an Inventory of all content for versions up to and including that particular version.’", Ref: "https://ocfl.io/draft/spec/#W010"},
	W011: {Code: W011, Description: "‘In the case that prior version directories include an inventory file, the values of the created, message and user keys in each version block in each prior inventory file SHOULD have the same values as the corresponding keys in the corresponding version block in the current inventory file.’", Ref: "https://ocfl.io/draft/spec/#W011"},
	W012: {Code: W012, Description: "‘Implementers SHOULD use the logs directory, if present, for storing files that contain a record of actions taken on the object.’", Ref: "https://ocfl.io/draft/spec/#W012"},
	W013: {Code: W013, Description: "‘In an OCFL Object, extension sub-directories SHOULD be named according to a registered extension name.’", Ref: "https://ocfl.io/draft/spec/#W013"},
	W014: {Code: W014, Description: "‘Storage hierarchies within the same OCFL Storage Root SHOULD use just one layout pattern.’", Ref: "https://ocfl.io/draft/spec/#W014"},
	W015: {Code: W015, Description: "‘Storage hierarchies within the same OCFL Storage Root SHOULD consistently use either a directory hierarchy of OCFL Objects or top-level OCFL Objects.’", Ref: "https://ocfl.io/draft/spec/#W015"},
	W016: {Code: W016, Description: "‘In the Storage Root, extension sub-directories SHOULD be named according to a registered extension name.’", Ref: "https://ocfl.io/draft/spec/#W016"},
}
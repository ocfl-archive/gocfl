package ocfl

import "fmt"

type ErrorCode string

const (
	E001 = "E001"
	E002 = "E002"
	E003 = "E003"
	E004 = "E004"
	E005 = "E005"
	E006 = "E006"
	E007 = "E007"
	E008 = "E008"
	E009 = "E009"
	E010 = "E010"
	E011 = "E011"
	E012 = "E012"
	E013 = "E013"
	E014 = "E014"
	E015 = "E015"
	E016 = "E016"
	E017 = "E017"
	E018 = "E018"
	E019 = "E019"
	E020 = "E020"
	E021 = "E021"
	E022 = "E022"
	E023 = "E023"
	E024 = "E024"
	E025 = "E025"
	E026 = "E026"
	E027 = "E027"
	E028 = "E028"
	E029 = "E029"
	E030 = "E030"
	E031 = "E031"
	E032 = "E032"
	E033 = "E033"
	E034 = "E034"
	E035 = "E035"
	E036 = "E036"
	E037 = "E037"
	E038 = "E038"
	E039 = "E039"
	E040 = "E040"
	E041 = "E041"
	E042 = "E042"
	E043 = "E043"
	E044 = "E044"
	E045 = "E045"
	E046 = "E046"
	E047 = "E047"
	E048 = "E048"
	E049 = "E049"
	E050 = "E050"
	E051 = "E051"
	E052 = "E052"
	E053 = "E053"
	E054 = "E054"
	E055 = "E055"
	E056 = "E056"
	E057 = "E057"
	E058 = "E058"
	E059 = "E059"
	E060 = "E060"
	E061 = "E061"
	E062 = "E062"
	E063 = "E063"
	E064 = "E064"
	E066 = "E066"
	E067 = "E067"
	E068 = "E068"
	E069 = "E069"
	E070 = "E070"
	E071 = "E071"
	E072 = "E072"
	E073 = "E073"
	E074 = "E074"
	E075 = "E075"
	E076 = "E076"
	E077 = "E077"
	E078 = "E078"
	E079 = "E079"
	E080 = "E080"
	E081 = "E081"
	E082 = "E082"
	E083 = "E083"
	E084 = "E084"
	E085 = "E085"
	E086 = "E086"
	E087 = "E087"
	E088 = "E088"
	E089 = "E089"
	E090 = "E090"
	E091 = "E091"
	E092 = "E092"
	E093 = "E093"
	E094 = "E094"
	E095 = "E095"
	E096 = "E096"
	E097 = "E097"
	E098 = "E098"
	E099 = "E099"
	E100 = "E100"
	E101 = "E101"
	E102 = "E102"
)

type ValidationError struct {
	Code        ErrorCode
	Description string
	Ref         string
	Err         error
}

func (verr *ValidationError) Error() string {
	if verr.Err == nil {
		return fmt.Sprintf("Validation Error #%s - %s (%s)", verr.Code, verr.Description, verr.Ref)
	} else {
		return fmt.Sprintf("Validation Error #%s - %s (%s): %v", verr.Code, verr.Description, verr.Ref, verr.Err)
	}
}

var ErrorE001 = &ValidationError{Code: E001, Description: "The OCFL Object Root must not contain files or directories other than those specified in the following sections.", Ref: "https://ocfl.io/1.0/spec/#E001"}
var ErrorE002 = &ValidationError{Code: E002, Description: "The version declaration must be formatted according to the NAMASTE specification.", Ref: "https://ocfl.io/1.0/spec/#E002"}
var ErrorE003 = &ValidationError{Code: E003, Description: "[The version declaration] must be a file in the base directory of the OCFL Object Root giving the OCFL version in the filename.", Ref: "https://ocfl.io/1.0/spec/#E003"}
var ErrorE004 = &ValidationError{Code: E004, Description: "The [version declaration] filename MUST conform to the pattern T=dvalue, where T must be 0, and dvalue must be ocfl_object_, followed by the OCFL specification version number.", Ref: "https://ocfl.io/1.0/spec/#E004"}
var ErrorE005 = &ValidationError{Code: E005, Description: "The [version declaration] filename must conform to the pattern T=dvalue, where T MUST be 0, and dvalue must be ocfl_object_, followed by the OCFL specification version number.", Ref: "https://ocfl.io/1.0/spec/#E005"}
var ErrorE006 = &ValidationError{Code: E006, Description: "The [version declaration] filename must conform to the pattern T=dvalue, where T must be 0, and dvalue MUST be ocfl_object_, followed by the OCFL specification version number.", Ref: "https://ocfl.io/1.0/spec/#E006"}
var ErrorE007 = &ValidationError{Code: E007, Description: "The text contents of the [version declaration] file must be the same as dvalue, followed by a newline (\n).", Ref: "https://ocfl.io/1.0/spec/#E007"}
var ErrorE008 = &ValidationError{Code: E008, Description: "OCFL Object content must be stored as a sequence of one or more versions.", Ref: "https://ocfl.io/1.0/spec/#E008"}
var ErrorE009 = &ValidationError{Code: E009, Description: "The version number sequence MUST start at 1 and must be continuous without missing integers.", Ref: "https://ocfl.io/1.0/spec/#E009"}
var ErrorE010 = &ValidationError{Code: E010, Description: "The version number sequence must start at 1 and MUST be continuous without missing integers.", Ref: "https://ocfl.io/1.0/spec/#E010"}
var ErrorE011 = &ValidationError{Code: E011, Description: "If zero-padded version directory numbers are used then they must start with the prefix v and then a zero.", Ref: "https://ocfl.io/1.0/spec/#E011"}
var ErrorE012 = &ValidationError{Code: E012, Description: "All version directories of an object must use the same naming convention: either a non-padded version directory number, or a zero-padded version directory number of consistent length.", Ref: "https://ocfl.io/1.0/spec/#E012"}
var ErrorE013 = &ValidationError{Code: E013, Description: "Operations that add a new version to an object must follow the version directory naming convention established by earlier versions.", Ref: "https://ocfl.io/1.0/spec/#E013"}
var ErrorE014 = &ValidationError{Code: E014, Description: "In all cases, references to files inside version directories from inventory files must use the actual version directory names.", Ref: "https://ocfl.io/1.0/spec/#E014"}
var ErrorE015 = &ValidationError{Code: E015, Description: "There must be no other files as children of a version directory, other than an inventory file and a inventory digest.", Ref: "https://ocfl.io/1.0/spec/#E015"}
var ErrorE016 = &ValidationError{Code: E016, Description: "Version directories must contain a designated content sub-directory if the version contains files to be preserved, and should not contain this sub-directory otherwise.", Ref: "https://ocfl.io/1.0/spec/#E016"}
var ErrorE017 = &ValidationError{Code: E017, Description: "The contentDirectory value MUST NOT contain the forward slash (/) path separator and must not be either one or two periods (. or ..).", Ref: "https://ocfl.io/1.0/spec/#E017"}
var ErrorE018 = &ValidationError{Code: E018, Description: "The contentDirectory value must not contain the forward slash (/) path separator and MUST NOT be either one or two periods (. or ..).", Ref: "https://ocfl.io/1.0/spec/#E018"}
var ErrorE019 = &ValidationError{Code: E019, Description: "If the key contentDirectory is set, it MUST be set in the first version of the object and must not change between versions of the same object.", Ref: "https://ocfl.io/1.0/spec/#E019"}
var ErrorE020 = &ValidationError{Code: E020, Description: "If the key contentDirectory is set, it must be set in the first version of the object and MUST NOT change between versions of the same object.", Ref: "https://ocfl.io/1.0/spec/#E020"}
var ErrorE021 = &ValidationError{Code: E021, Description: "If the key contentDirectory is not present in the inventory file then the name of the designated content sub-directory must be content.", Ref: "https://ocfl.io/1.0/spec/#E021"}
var ErrorE022 = &ValidationError{Code: E022, Description: "OCFL-compliant tools (including any validators) must ignore all directories in the object version directory except for the designated content directory.", Ref: "https://ocfl.io/1.0/spec/#E022"}
var ErrorE023 = &ValidationError{Code: E023, Description: "Every file within a version's content directory must be referenced in the manifest section of the inventory.", Ref: "https://ocfl.io/1.0/spec/#E023"}
var ErrorE024 = &ValidationError{Code: E024, Description: "There must not be empty directories within a version's content directory.", Ref: "https://ocfl.io/1.0/spec/#E024"}
var ErrorE025 = &ValidationError{Code: E025, Description: "For content-addressing, OCFL Objects must use either sha512 or sha256, and should use sha512.", Ref: "https://ocfl.io/1.0/spec/#E025"}
var ErrorE026 = &ValidationError{Code: E026, Description: "For storage of additional fixity values, or to support legacy content migration, implementers must choose from the following controlled vocabulary of digest algorithms, or from a list of additional algorithms given in the [Digest-Algorithms-Extension].", Ref: "https://ocfl.io/1.0/spec/#E026"}
var ErrorE027 = &ValidationError{Code: E027, Description: "OCFL clients must support all fixity algorithms given in the table below, and may support additional algorithms from the extensions.", Ref: "https://ocfl.io/1.0/spec/#E027"}
var ErrorE028 = &ValidationError{Code: E028, Description: "Optional fixity algorithms that are not supported by a client must be ignored by that client.", Ref: "https://ocfl.io/1.0/spec/#E028"}
var ErrorE029 = &ValidationError{Code: E029, Description: "SHA-1 algorithm defined by [FIPS-180-4] and must be encoded using hex (base16) encoding [RFC4648].", Ref: "https://ocfl.io/1.0/spec/#E029"}
var ErrorE030 = &ValidationError{Code: E030, Description: "SHA-256 algorithm defined by [FIPS-180-4] and must be encoded using hex (base16) encoding [RFC4648].", Ref: "https://ocfl.io/1.0/spec/#E030"}
var ErrorE031 = &ValidationError{Code: E031, Description: "SHA-512 algorithm defined by [FIPS-180-4] and must be encoded using hex (base16) encoding [RFC4648].", Ref: "https://ocfl.io/1.0/spec/#E031"}
var ErrorE032 = &ValidationError{Code: E032, Description: "[blake2b-512] must be encoded using hex (base16) encoding [RFC4648].", Ref: "https://ocfl.io/1.0/spec/#E032"}
var ErrorE033 = &ValidationError{Code: E033, Description: "An OCFL Object Inventory MUST follow the [JSON] structure described in this section and must be named inventory.json.", Ref: "https://ocfl.io/1.0/spec/#E033"}
var ErrorE034 = &ValidationError{Code: E034, Description: "An OCFL Object Inventory must follow the [JSON] structure described in this section and MUST be named inventory.json.", Ref: "https://ocfl.io/1.0/spec/#E034"}
var ErrorE035 = &ValidationError{Code: E035, Description: "The forward slash (/) path separator must be used in content paths in the manifest and fixity blocks within the inventory.", Ref: "https://ocfl.io/1.0/spec/#E035"}
var ErrorE036 = &ValidationError{Code: E036, Description: "An OCFL Object Inventory must include the following keys: [id, type, digestAlgorithm, head]", Ref: "https://ocfl.io/1.0/spec/#E036"}
var ErrorE037 = &ValidationError{Code: E037, Description: "[id] must be unique in the local context, and should be a URI [RFC3986].", Ref: "https://ocfl.io/1.0/spec/#E037"}
var ErrorE038 = &ValidationError{Code: E038, Description: "In the object root inventory [the type value] must be the URI of the inventory section of the specification version matching the object conformance declaration.", Ref: "https://ocfl.io/1.0/spec/#E038"}
var ErrorE039 = &ValidationError{Code: E039, Description: "[digestAlgorithm] must be the algorithm used in the manifest and state blocks.", Ref: "https://ocfl.io/1.0/spec/#E039"}
var ErrorE040 = &ValidationError{Code: E040, Description: "[head] must be the version directory name with the highest version number.", Ref: "https://ocfl.io/1.0/spec/#E040"}
var ErrorE041 = &ValidationError{Code: E041, Description: "In addition to these keys, there must be two other blocks present, manifest and versions, which are discussed in the next two sections.", Ref: "https://ocfl.io/1.0/spec/#E041"}
var ErrorE042 = &ValidationError{Code: E042, Description: "Content paths within a manifest block must be relative to the OCFL Object Root.", Ref: "https://ocfl.io/1.0/spec/#E042"}
var ErrorE043 = &ValidationError{Code: E043, Description: "An OCFL Object Inventory must include a block for storing versions.", Ref: "https://ocfl.io/1.0/spec/#E043"}
var ErrorE044 = &ValidationError{Code: E044, Description: "This block MUST have the key of versions within the inventory, and it must be a JSON object.", Ref: "https://ocfl.io/1.0/spec/#E044"}
var ErrorE045 = &ValidationError{Code: E045, Description: "This block must have the key of versions within the inventory, and it MUST be a JSON object.", Ref: "https://ocfl.io/1.0/spec/#E045"}
var ErrorE046 = &ValidationError{Code: E046, Description: "The keys of [the versions object] must correspond to the names of the version directories used.", Ref: "https://ocfl.io/1.0/spec/#E046"}
var ErrorE047 = &ValidationError{Code: E047, Description: "Each value [of the versions object] must be another JSON object that characterizes the version, as described in the 3.5.3.1 Version section.", Ref: "https://ocfl.io/1.0/spec/#E047"}
var ErrorE048 = &ValidationError{Code: E048, Description: "A JSON object to describe one OCFL Version, which must include the following keys: [created, state]", Ref: "https://ocfl.io/1.0/spec/#E048"}
var ErrorE049 = &ValidationError{Code: E049, Description: "[the value of the “created” key] must be expressed in the Internet Date/Time Format defined by [RFC3339].", Ref: "https://ocfl.io/1.0/spec/#E049"}
var ErrorE050 = &ValidationError{Code: E050, Description: "The keys of [the “state” JSON object] are digest values, each of which must correspond to an entry in the manifest of the inventory.", Ref: "https://ocfl.io/1.0/spec/#E050"}
var ErrorE051 = &ValidationError{Code: E051, Description: "The logical path [value of a “state” digest key] must be interpreted as a set of one or more path elements joined by a / path separator.", Ref: "https://ocfl.io/1.0/spec/#E051"}
var ErrorE052 = &ValidationError{Code: E052, Description: "[logical] Path elements must not be ., .., or empty (//).", Ref: "https://ocfl.io/1.0/spec/#E052"}
var ErrorE053 = &ValidationError{Code: E053, Description: "Additionally, a logical path must not begin or end with a forward slash (/).", Ref: "https://ocfl.io/1.0/spec/#E053"}
var ErrorE054 = &ValidationError{Code: E054, Description: "The value of the user key must contain a user name key, “name” and should contain an address key, “address”.", Ref: "https://ocfl.io/1.0/spec/#E054"}
var ErrorE055 = &ValidationError{Code: E055, Description: "This block must have the key of fixity within the inventory.", Ref: "https://ocfl.io/1.0/spec/#E055"}
var ErrorE056 = &ValidationError{Code: E056, Description: "The fixity block must contain keys corresponding to the controlled vocabulary given in the digest algorithms listed in the Digests section, or in a table given in an Extension.", Ref: "https://ocfl.io/1.0/spec/#E056"}
var ErrorE057 = &ValidationError{Code: E057, Description: "The value of the fixity block for a particular digest algorithm must follow the structure of the manifest block; that is, a key corresponding to the digest value, and an array of content paths that match that digest.", Ref: "https://ocfl.io/1.0/spec/#E057"}
var ErrorE058 = &ValidationError{Code: E058, Description: "Every occurrence of an inventory file must have an accompanying sidecar file stating its digest.", Ref: "https://ocfl.io/1.0/spec/#E058"}
var ErrorE059 = &ValidationError{Code: E059, Description: "This value must match the value given for the digestAlgorithm key in the inventory.", Ref: "https://ocfl.io/1.0/spec/#E059"}
var ErrorE060 = &ValidationError{Code: E060, Description: "The digest sidecar file must contain the digest of the inventory file.", Ref: "https://ocfl.io/1.0/spec/#E060"}
var ErrorE061 = &ValidationError{Code: E061, Description: "[The digest sidecar file] must follow the format: DIGEST inventory.json", Ref: "https://ocfl.io/1.0/spec/#E061"}
var ErrorE062 = &ValidationError{Code: E062, Description: "The digest of the inventory must be computed only after all changes to the inventory have been made, and thus writing the digest sidecar file is the last step in the versioning process.", Ref: "https://ocfl.io/1.0/spec/#E062"}
var ErrorE063 = &ValidationError{Code: E063, Description: "Every OCFL Object must have an inventory file within the OCFL Object Root, corresponding to the state of the OCFL Object at the current version.", Ref: "https://ocfl.io/1.0/spec/#E063"}
var ErrorE064 = &ValidationError{Code: E064, Description: "Where an OCFL Object contains inventory.json in version directories, the inventory file in the OCFL Object Root must be the same as the file in the most recent version.", Ref: "https://ocfl.io/1.0/spec/#E064"}
var ErrorE066 = &ValidationError{Code: E066, Description: "Each version block in each prior inventory file must represent the same object state as the corresponding version block in the current inventory file.", Ref: "https://ocfl.io/1.0/spec/#E066"}
var ErrorE067 = &ValidationError{Code: E067, Description: "The extensions directory must not contain any files, and no sub-directories other than extension sub-directories.", Ref: "https://ocfl.io/1.0/spec/#E067"}
var ErrorE068 = &ValidationError{Code: E068, Description: "The specific structure and function of the extension, as well as a declaration of the registered extension name must be defined in one of the following locations: The OCFL Extensions repository OR The Storage Root, as a plain text document directly in the Storage Root.", Ref: "https://ocfl.io/1.0/spec/#E068"}
var ErrorE069 = &ValidationError{Code: E069, Description: "An OCFL Storage Root MUST contain a Root Conformance Declaration identifying it as such.", Ref: "https://ocfl.io/1.0/spec/#E069"}
var ErrorE070 = &ValidationError{Code: E070, Description: "If present, [the ocfl_layout.json document] MUST include the following two keys in the root JSON object: [key, description]", Ref: "https://ocfl.io/1.0/spec/#E070"}
var ErrorE071 = &ValidationError{Code: E071, Description: "The value of the [ocfl_layout.json] extension key must be the registered extension name for the extension defining the arrangement under the storage root.", Ref: "https://ocfl.io/1.0/spec/#E071"}
var ErrorE072 = &ValidationError{Code: E072, Description: "The directory hierarchy used to store OCFL Objects MUST NOT contain files that are not part of an OCFL Object.", Ref: "https://ocfl.io/1.0/spec/#E072"}
var ErrorE073 = &ValidationError{Code: E073, Description: "Empty directories MUST NOT appear under a storage root.", Ref: "https://ocfl.io/1.0/spec/#E073"}
var ErrorE074 = &ValidationError{Code: E074, Description: "Although implementations may require multiple OCFL Storage Roots - that is, several logical or physical volumes, or multiple “buckets” in an object store - each OCFL Storage Root MUST be independent.", Ref: "https://ocfl.io/1.0/spec/#E074"}
var ErrorE075 = &ValidationError{Code: E075, Description: "The OCFL version declaration MUST be formatted according to the NAMASTE specification.", Ref: "https://ocfl.io/1.0/spec/#E075"}
var ErrorE076 = &ValidationError{Code: E076, Description: "[The OCFL version declaration] MUST be a file in the base directory of the OCFL Storage Root giving the OCFL version in the filename.", Ref: "https://ocfl.io/1.0/spec/#E076"}
var ErrorE077 = &ValidationError{Code: E077, Description: "[The OCFL version declaration filename] MUST conform to the pattern T=dvalue, where T must be 0, and dvalue must be ocfl_, followed by the OCFL specification version number.", Ref: "https://ocfl.io/1.0/spec/#E077"}
var ErrorE078 = &ValidationError{Code: E078, Description: "[The OCFL version declaration filename] must conform to the pattern T=dvalue, where T MUST be 0, and dvalue must be ocfl_, followed by the OCFL specification version number.", Ref: "https://ocfl.io/1.0/spec/#E078"}
var ErrorE079 = &ValidationError{Code: E079, Description: "[The OCFL version declaration filename] must conform to the pattern T=dvalue, where T must be 0, and dvalue MUST be ocfl_, followed by the OCFL specification version number.", Ref: "https://ocfl.io/1.0/spec/#E079"}
var ErrorE080 = &ValidationError{Code: E080, Description: "The text contents of [the OCFL version declaration file] MUST be the same as dvalue, followed by a newline (\n).", Ref: "https://ocfl.io/1.0/spec/#E080"}
var ErrorE081 = &ValidationError{Code: E081, Description: "OCFL Objects within the OCFL Storage Root also include a conformance declaration which MUST indicate OCFL Object conformance to the same or earlier version of the specification.", Ref: "https://ocfl.io/1.0/spec/#E081"}
var ErrorE082 = &ValidationError{Code: E082, Description: "OCFL Object Roots MUST be stored either as the terminal resource at the end of a directory storage hierarchy or as direct children of a containing OCFL Storage Root.", Ref: "https://ocfl.io/1.0/spec/#E082"}
var ErrorE083 = &ValidationError{Code: E083, Description: "There MUST be a deterministic mapping from an object identifier to a unique storage path.", Ref: "https://ocfl.io/1.0/spec/#E083"}
var ErrorE084 = &ValidationError{Code: E084, Description: "Storage hierarchies MUST NOT include files within intermediate directories.", Ref: "https://ocfl.io/1.0/spec/#E084"}
var ErrorE085 = &ValidationError{Code: E085, Description: "Storage hierarchies MUST be terminated by OCFL Object Roots.", Ref: "https://ocfl.io/1.0/spec/#E085"}
var ErrorE086 = &ValidationError{Code: E086, Description: "The storage root extensions directory MUST conform to the same guidelines and limitations as those defined for object extensions.", Ref: "https://ocfl.io/1.0/spec/#E086"}
var ErrorE087 = &ValidationError{Code: E087, Description: "An OCFL validator MUST ignore any files in the storage root it does not understand.", Ref: "https://ocfl.io/1.0/spec/#E087"}
var ErrorE088 = &ValidationError{Code: E088, Description: "An OCFL Storage Root MUST NOT contain directories or sub-directories other than as a directory hierarchy used to store OCFL Objects or for storage root extensions.", Ref: "https://ocfl.io/1.0/spec/#E088"}
var ErrorE089 = &ValidationError{Code: E089, Description: "If the preservation of non-OCFL-compliant features is required then the content MUST be wrapped in a suitable disk or filesystem image format which OCFL can treat as a regular file.", Ref: "https://ocfl.io/1.0/spec/#E089"}
var ErrorE090 = &ValidationError{Code: E090, Description: "Hard and soft (symbolic) links are not portable and MUST NOT be used within OCFL Storage hierachies.", Ref: "https://ocfl.io/1.0/spec/#E090"}
var ErrorE091 = &ValidationError{Code: E091, Description: "Filesystems MUST preserve the case of OCFL filepaths and filenames.", Ref: "https://ocfl.io/1.0/spec/#E091"}
var ErrorE092 = &ValidationError{Code: E092, Description: "The value for each key in the manifest must be an array containing the content paths of files in the OCFL Object that have content with the given digest.", Ref: "https://ocfl.io/1.0/spec/#E092"}
var ErrorE093 = &ValidationError{Code: E093, Description: "Where included in the fixity block, the digest values given must match the digests of the files at the corresponding content paths.", Ref: "https://ocfl.io/1.0/spec/#E093"}
var ErrorE094 = &ValidationError{Code: E094, Description: "The value of [the message] key is freeform text, used to record the rationale for creating this version. It must be a JSON string.", Ref: "https://ocfl.io/1.0/spec/#E094"}
var ErrorE095 = &ValidationError{Code: E095, Description: "Within a version, logical paths must be unique and non-conflicting, so the logical path for a file cannot appear as the initial part of another logical path.", Ref: "https://ocfl.io/1.0/spec/#E095"}
var ErrorE096 = &ValidationError{Code: E096, Description: "As JSON keys are case sensitive, while digests may not be, there is an additional requirement that each digest value must occur only once in the manifest regardless of case.", Ref: "https://ocfl.io/1.0/spec/#E096"}
var ErrorE097 = &ValidationError{Code: E097, Description: "As JSON keys are case sensitive, while digests may not be, there is an additional requirement that each digest value must occur only once in the fixity block for any digest algorithm, regardless of case.", Ref: "https://ocfl.io/1.0/spec/#E097"}
var ErrorE098 = &ValidationError{Code: E098, Description: "The content path must be interpreted as a set of one or more path elements joined by a / path separator.", Ref: "https://ocfl.io/1.0/spec/#E098"}
var ErrorE099 = &ValidationError{Code: E099, Description: "[content] path elements must not be ., .., or empty (//).", Ref: "https://ocfl.io/1.0/spec/#E099"}
var ErrorE100 = &ValidationError{Code: E100, Description: "A content path must not begin or end with a forward slash (/).", Ref: "https://ocfl.io/1.0/spec/#E100"}
var ErrorE101 = &ValidationError{Code: E101, Description: "Within an inventory, content paths must be unique and non-conflicting, so the content path for a file cannot appear as the initial part of another content path.", Ref: "https://ocfl.io/1.0/spec/#E101"}
var ErrorE102 = &ValidationError{Code: E102, Description: "An inventory file must not contain keys that are not specified.", Ref: "https://ocfl.io/1.0/spec/#E102"}

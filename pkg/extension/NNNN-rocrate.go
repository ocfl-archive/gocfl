package extension

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"slices"
	"strings"

	"emperror.dev/errors"

	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl"
	"github.com/ocfl-archive/gocfl/v2/pkg/rocrate"
)

// ROCrateFileName ...
const ROCrateFileName = "NNNN-ro-crate"
const ROCrateEnabled = "enabled"

// ext prefix.
const extPrefix = "ext"

// Additional parameters for RO-CRATE metafile.
const rOCrateParamOrg = "organisation"
const rOCrateParamOrgID = "organisationID"
const rOCrateParamUser = "user"
const rOCrateParamAddress = "address"

// RoCrateFileDescription ...
const RoCrateFileDescription = "Description for RO-Crate extension"

// registered ...
const registered = false

// ROCrateFileConfig ...
type ROCrateFileConfig struct {
	*ocfl.ExtensionConfig
	// StorageType describes the location type where the technical
	// metadata is stored.
	StorageType string `json:"storageType"`
	// StorageName describes the storage location within the specified
	// storage type.
	StorageName string `json:"storageName"`
	// MetaName describes the name of the metadata file created by
	// this extension. This must be the same as NNNN-metafile as it
	// will override that extension.
	MetaName string `json:"metaFileName,omitempty"`
}

// ROCrateFile provides a way to record configuration and other
// metadata.
type ROCrateFile struct {
	// combination of the config and other metadata.
	*ROCrateFileConfig
	fsys       fs.FS
	stored     bool
	info       map[string][]byte
	enabled    bool
	userConfig *userConfig
}

// userConfig stores the additional metadata a user needs to provide to
// complete the info.json object.
type userConfig struct {
	organisation   string
	organisationID string
	user           string
	address        string
}

// GetROCrateFileParams enables the retrieval of configured values from
// the GOCFL configuration.
func GetROCrateFileParams() []*ocfl.ExtensionExternalParam {
	return []*ocfl.ExtensionExternalParam{
		{
			ExtensionName: ROCrateFileName,
			Functions:     []string{"add", "create"},
			Param:         ROCrateEnabled,
			Description:   "replace metafile extension functionality if enabled and map RO-CRATE metadata",
			Default:       "false",
		},
		{
			ExtensionName: ROCrateFileName,
			Functions:     []string{"add", "create"},
			Param:         rOCrateParamOrg,
			Description:   "provide insitutional organisation to the RO-CRATE extension metadata",
			Default:       "",
		},
		{
			ExtensionName: ROCrateFileName,
			Functions:     []string{"add", "create"},
			Param:         rOCrateParamOrgID,
			Description:   "provide insitutional organisation ID to the RO-CRATE extension metadata",
			Default:       "",
		},
		{
			ExtensionName: ROCrateFileName,
			Functions:     []string{"add", "create"},
			Param:         rOCrateParamUser,
			Description:   "provide insitutional user to the RO-CRATE extension metadata",
			Default:       "",
		},
		{
			ExtensionName: ROCrateFileName,
			Functions:     []string{"add", "create"},
			Param:         rOCrateParamAddress,
			Description:   "provide user address to the RO-CRATE extension metadata",
			Default:       "",
		},
	}
}

// NewROCrateFileFS returns a new ROCrateFile objet providing access
// to configuration and other parameters requirerd by this extension.
func NewROCrateFileFS(fsys fs.FS) (*ROCrateFile, error) {
	data, err := fs.ReadFile(fsys, "config.json")
	if err != nil {
		return nil, errors.Wrap(err, "cannot read config.json")
	}
	var config = &ROCrateFileConfig{
		ExtensionConfig: &ocfl.ExtensionConfig{ExtensionName: ROCrateFileName},
		StorageType:     "extension",
		StorageName:     "metadata",
	}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal DirectCleanConfig '%s'", string(data))
	}
	rcFile, err := NewROCrateFile(config)
	return rcFile, err
}

// NewROCrateFile provides a helper to create a new object that helps us
// to understand the internals of the extension
func NewROCrateFile(config *ROCrateFileConfig) (*ROCrateFile, error) {
	rcFile := &ROCrateFile{
		ROCrateFileConfig: config,
		info:              map[string][]byte{},
	}
	// check internal extension name is correct..
	if config.ExtensionName != rcFile.GetName() {
		return nil, errors.New(
			fmt.Sprintf(
				"invalid extension name'%s'for extension %s",
				config.ExtensionName,
				rcFile.GetName(),
			),
		)
	}
	return rcFile, nil
}

// Terminate ...
func (rcFile *ROCrateFile) Terminate() error {
	// not implemented.
	return nil
}

// GetFS provides a helper to retrieve the configurred filesystem
// object.
func (rcFile *ROCrateFile) GetFS() fs.FS {
	return rcFile.fsys
}

// GetConfig provides a helper to retrieve the RO-CRATE configuration
// object.
func (rcFile *ROCrateFile) GetConfig() any {
	return rcFile.ROCrateFileConfig
}

// IsRegistered describes whether this is an official GOCL extension.
func (rcFile *ROCrateFile) IsRegistered() bool {
	return registered
}

// ParamROCrateEnabled provides a mechanism for retrieivng the enabled
// parameter from the user's configuration, e.g. given a param map,
// call `param[ParamRoCrateEnabled()]“.
func ParamROCrateEnabled() string {
	return fmt.Sprintf("%s-%s-%s", extPrefix, ROCrateFileName, ROCrateEnabled)
}

// paramROCrateOrg enables the consistent retrieval of the
// configured "organisation" parameter.
func paramROCrateOrg() string {
	return fmt.Sprintf("%s-%s-%s", extPrefix, ROCrateFileName, rOCrateParamOrg)
}

// paramROCrateOrgID enables the consistent retrieval of the
// configured "organisation id" parameter.
func paramROCrateOrgID() string {
	return fmt.Sprintf("%s-%s-%s", extPrefix, ROCrateFileName, rOCrateParamOrgID)
}

// paramROCrateUser enables the consistent retrieval of the
// configured "user" parameter.
func paramROCrateUser() string {
	return fmt.Sprintf("%s-%s-%s", extPrefix, ROCrateFileName, rOCrateParamUser)
}

// paramROCrateAddress enables the consistent retrieval of the
// configured "address" parameter.
func paramROCrateAddress() string {
	return fmt.Sprintf("%s-%s-%s", extPrefix, ROCrateFileName, rOCrateParamAddress)
}

// SetParams allows us to set parameters provided to the extension via
// the config, e.g. CLI (or TOML?)
func (rcFile *ROCrateFile) SetParams(params map[string]string) error {
	if params == nil {
		// this is unlikely to happen.
		errors.New("extension parameters are blank")
	}
	enabled := params[ParamROCrateEnabled()]
	if strings.ToLower(enabled) != "true" {
		rcFile.enabled = false
		return nil
	}
	rcFile.enabled = true
	userConfig := userConfig{}
	rcFile.userConfig = &userConfig
	rcFile.userConfig.organisation = params[paramROCrateOrg()]
	rcFile.userConfig.organisationID = params[paramROCrateOrgID()]
	rcFile.userConfig.user = params[paramROCrateUser()]
	rcFile.userConfig.address = params[paramROCrateAddress()]
	return nil
}

// SetFS provides a helper to set the filesystem object to be used
// by this extension.
func (rcFile *ROCrateFile) SetFS(fsys fs.FS, create bool) {
	rcFile.fsys = fsys
}

// GetName returns the name of this extension to the caller.
func (rcFile *ROCrateFile) GetName() string {
	return ROCrateFileName
}

// WriteConfig ...
func (rcFile *ROCrateFile) WriteConfig() error {
	// not implemented.
	return nil
}

// UpdateObjectBefore (before a new version of an OCFL object is
// created...) TODO...
func (rcFile *ROCrateFile) UpdateObjectBefore(object ocfl.Object) error {
	// not implemented.
	return nil
}

// UpdateObjectAfter (after all content to the new version is written)
func (rcFile *ROCrateFile) UpdateObjectAfter(object ocfl.Object) error {
	// not implemented.
	return nil

}

// GetMetadata (is called by any tool, which wants to report about
// content, e.g. data will be retrieved from here for use in METS.xml
// or PREMIS.xml.
func (rcFile *ROCrateFile) GetMetadata(object ocfl.Object) (map[string]any, error) {
	if !rcFile.enabled {
		return nil, nil
	}
	var err error
	var result = map[string]any{}
	inventory := object.GetInventory()
	versions := inventory.GetVersionStrings()
	slices.Reverse(versions)
	var metadata []byte
	for _, ver := range versions {
		var ok bool
		if metadata, ok = rcFile.info[ver]; ok {
			break
		}
		if metadata, err = ocfl.ReadFile(
			object,
			rcFile.MetaName,
			ver,
			rcFile.StorageType,
			rcFile.StorageName,
			rcFile.fsys); err == nil {
			break
		}
	}
	if metadata == nil {
		return nil, errors.Wrapf(err, "cannot read %s", rcFile.MetaName)
	}
	var metaStruct = map[string]any{}
	if err := json.Unmarshal(metadata, &metaStruct); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal '%s'", rcFile.MetaName)
	}
	result[""] = metaStruct
	return result, nil
}

// findROCrateMeta looks for the RO-CRATE metadata file within the
// objects spplied to the function.
func (rcFile *ROCrateFile) findROCrateMeta(stateFiles []string) bool {
	f := stateFiles[0]
	if f != "data/ro-crate-metadata.json" {
		return false
	}
	return true
}

// copyStream allows StreamObject to make a copy of a reader so that it
// can be given back safely to the caller and other stream functions
// can be performed on the object.
func copyStream(reader io.Reader) (io.Reader, error) {
	buf := &bytes.Buffer{}
	_, err := io.Copy(buf, reader)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

// infoJSONExists provides a guard that ensures we know what we're
// doing with info.json and its replacement when driven by the
// RO-CRATE extension.
func infoJSONExists(object ocfl.Object, metaFileName string) bool {
	inventory := object.GetInventory()
	for _, v := range inventory.GetManifest() {
		s := strings.Split(v[0], "/")
		if s[len(s)-1] == metaFileName {
			return true
		}
	}
	return false
}

// writeMetafile outputs the configured metadata file and content to
// the metadata storage location.
func (rcFile *ROCrateFile) writeMetafile(object ocfl.Object, rcMeta string) error {
	if !rcFile.enabled {
		return nil
	}
	if infoJSONExists(object, rcFile.MetaName) {
		return fmt.Errorf("%s exists, ensure metafile extension is not configured as on", rcFile.MetaName)
	}
	mappingFile := rcFile.MetaName
	log.Println(mappingFile)
	data := []byte(rcMeta)
	if _, err := object.AddReader(
		io.NopCloser(
			bytes.NewBuffer(data),
		),
		[]string{mappingFile},
		rcFile.StorageName,
		true,
		false,
	); err != nil {
		log.Println("there was an error")
		return err
	}
	return nil
}

// writeConfigValues completes a ro-crate info.json output by adding
// user or insitution supplied values to the structure.
func (rcFile *ROCrateFile) writeConfigValues(rcMeta *rocrate.GocflSummary) {
	rcMeta.Organisation = rcFile.userConfig.organisation
	rcMeta.OrganisationID = rcFile.userConfig.organisationID
	rcMeta.User = rcFile.userConfig.user
	rcMeta.Address = rcFile.userConfig.address
}

// StreamObject implements an interface object enabling the access of
// RO-CRATE content as a a stream so it can be utilised by this
// extension as required, i.e. in the case of RO-CRATE its metadata
// is accessed and mapped into a GOCFL compatible metadadata object.
func (rcFile *ROCrateFile) StreamObject(
	object ocfl.Object,
	reader io.Reader,
	stateFiles []string,
	dest string,
) error {
	if !rcFile.enabled {
		return nil
	}
	if !rcFile.findROCrateMeta(stateFiles) {
		return nil
	}
	inventory := object.GetInventory()
	if inventory == nil {
		return errors.New("no inventory available")
	}
	// copy file so that it can then be sent to another interface to
	// be read. In this case a ro-crate-metadata json reader.
	metaCopy, err := copyStream(reader)
	if err != nil {
		return err
	}
	processed, err := rocrate.ProcessMetadataStream(metaCopy)
	if err != nil {
		return errors.Wrapf(err, "cannot process RO-CRATE metadata")
	}
	rcMeta, _ := processed.GOCFLSummary()
	rcFile.writeConfigValues(&rcMeta)
	metaString := rcMeta.String()
	if metaString == rocrate.StringerError {
		return errors.New("problem creating GOCFL metadata JSON")
	}
	rcFile.writeMetafile(object, rcMeta.String())
	rcFile.info[inventory.GetHead()] = []byte(rcMeta.String())
	return nil
}

// check interface satisfaction
var (
	_ ocfl.Extension             = &ROCrateFile{}
	_ ocfl.ExtensionObjectChange = &ROCrateFile{}
	_ ocfl.ExtensionMetadata     = &ROCrateFile{}
)

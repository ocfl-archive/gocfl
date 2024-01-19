package extension

import (
	"bytes"
	"crypto/sha512"
	"emperror.dev/errors"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"github.com/je4/filesystem/v2/pkg/writefs"
	"github.com/je4/gocfl/v2/data/specs"
	"github.com/je4/gocfl/v2/pkg/dilcis/mets"
	"github.com/je4/gocfl/v2/pkg/dilcis/premis"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	"github.com/je4/gocfl/v2/version"
	"github.com/je4/indexer/v2/pkg/indexer"
	"github.com/op/go-logging"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"io"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const METSName = "NNNN-mets"
const METSDescription = "METS/EAD3/PREMIS metadata"

type metsInternalFiledata struct {
	ingestVersion string
	uuid          string
	cs            string
}

var metsMDTypes = []string{
	"MARC",
	"MODS",
	"EAD",
	"DC",
	"NISOIMG",
	"LC-AV",
	"VRA",
	"TEIHDR",
	"DDI",
	"FGDC",
	"LOM",
	"PREMIS",
	"PREMIS:OBJECT",
	"PREMIS:AGENT",
	"PREMIS:RIGHTS",
	"PREMIS:EVENT",
	"TEXTMD",
	"METSRIGHTS",
	"ISO 19115:2003 NAP",
	"EAC-CPF",
	"LIDO",
	"OTHER",
}

func GetMetsParams() []*ocfl.ExtensionExternalParam {
	return []*ocfl.ExtensionExternalParam{
		{
			ExtensionName: METSName,
			Functions:     []string{"add", "update", "create"},
			Param:         "descriptive-metadata",
			Description:   "reference to archived descriptive metadata (i.e. ead:metadata:ead.xml)",
		},
	}
}

func NewMetsFS(fsys fs.FS, logger *logging.Logger) (*Mets, error) {
	data, err := fs.ReadFile(fsys, "config.json")
	if err != nil {
		return nil, errors.Wrap(err, "cannot read config.json")
	}

	var config = &MetsConfig{
		ExtensionConfig:            &ocfl.ExtensionConfig{ExtensionName: METSName},
		StorageType:                "area",
		StorageName:                "metadata",
		PrimaryDescriptiveMetadata: "metadata:info.json",
		MetsFile:                   "mets.xml",
		PremisFile:                 "premis.xml",
	}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal DirectCleanConfig '%s'", string(data))
	}

	return NewMets(config, logger)
}
func NewMets(config *MetsConfig, logger *logging.Logger) (*Mets, error) {
	me := &Mets{
		MetsConfig: config,
		logger:     logger,
	}
	if config.ExtensionName != me.GetName() {
		return nil, errors.New(fmt.Sprintf("invalid extension name'%s'for extension %s", config.ExtensionName, me.GetName()))
	}
	return me, nil
}

type MetsConfig struct {
	*ocfl.ExtensionConfig
	StorageType                string `json:"storageType"`
	StorageName                string `json:"storageName"`
	PrimaryDescriptiveMetadata string `json:"primaryDescriptiveMetadata,omitempty"`
	MetsFile                   string `json:"metsFile,omitempty"`
	PremisFile                 string `json:"premisFile,omitempty"`
}
type Mets struct {
	*MetsConfig
	fsys   fs.FS
	logger *logging.Logger
	//	descriptiveMetadata     string
	//	descriptiveMetadataType string
}

func (me *Mets) GetFS() fs.FS {
	return me.fsys
}

func (me *Mets) GetConfig() any {
	return me.MetsConfig
}

func (me *Mets) IsRegistered() bool {
	return false
}

func (me *Mets) SetParams(params map[string]string) error {
	if params != nil {
		name := fmt.Sprintf("ext-%s-%s", METSName, "descriptive-metadata")
		if str, ok := params[name]; ok {
			me.PrimaryDescriptiveMetadata = str
		}
	}
	return nil
}

func (me *Mets) SetFS(fsys fs.FS) {
	me.fsys = fsys
}

func (me *Mets) GetName() string { return METSName }

func (me *Mets) WriteConfig() error {
	if me.fsys == nil {
		return errors.New("no filesystem set")
	}
	configWriter, err := writefs.Create(me.fsys, "config.json")
	if err != nil {
		return errors.Wrap(err, "cannot open config.json")
	}
	defer configWriter.Close()
	jenc := json.NewEncoder(configWriter)
	jenc.SetIndent("", "   ")
	if err := jenc.Encode(me.MetsConfig); err != nil {
		return errors.Wrapf(err, "cannot encode config to file")
	}

	return nil
}

func (me *Mets) UpdateObjectBefore(object ocfl.Object) error {
	return nil
}

var regexpIntPath = regexp.MustCompile(`´(v[0-9]+)/content/(.+)/.+`)

func checksumTypeToMets(t string) string {
	// Adler-32 CRC32 HAVAL MD5 MNP SHA-1 SHA-256 SHA-384 SHA-512 TIGER WHIRLPOOL
	t = strings.ToUpper(t)
	switch t {
	case "SHA512":
		return "SHA-512"
	case "SHA384":
		return "SHA-384"
	case "SHA256":
		return "SHA-256"
	case "SHA1":
		return "SHA-1"
	case "ADLER32":
		return "ADLER-32"
	case "CRC32", "MD5", "MNP", "TIGER", "WHIRLPOOL":
		return t
	default:
		return ""
	}
}

/*
type metaFileBase struct {
	Address             string   `json:"address,omitempty"`
	AlternativeTitles   []string `json:"alternative_titles,omitempty"`
	Collection          string   `json:"collection,omitempty"`
	CollectionId        string   `json:"collection_id,omitempty"`
	Created             string   `json:"created,omitempty"`
	Deprecates          string   `json:"deprecates,omitempty"`
	Description         string   `json:"description,omitempty"`
	Identifiers         []string `json:"identifiers,omitempty"`
	IngestWorkflow      string   `json:"ingest_workflow,omitempty"`
	Keywords            []string `json:"keywords,omitempty"`
	LastChanged         string   `json:"last_changed,omitempty"`
	Organisation        string   `json:"organisation,omitempty"`
	OrganisationAddress string   `json:"organisation_address,omitempty"`
	OrganisationId      string   `json:"organisation_id,omitempty"`
	References          []string `json:"references,omitempty"`
	Sets                []string `json:"sets,omitempty"`
	Signature           string   `json:"signature,omitempty"`
	Title               string   `json:"title,omitempty"`
	User                string   `json:"user,omitempty"`
}

*/

func (me *Mets) UpdateObjectAfter(object ocfl.Object) error {
	inventory := object.GetInventory()
	metadata, err := object.GetMetadata()
	if err != nil {
		return errors.Wrap(err, "cannot get metadata from object")
	}

	head := inventory.GetHead()
	versions := inventory.GetVersions()
	v, ok := versions[head]
	if !ok {
		return errors.Wrapf(err, "object has no version %s", head)
	}

	var contentSubPath = map[string]ContentSubPathEntry{}
	var metafileOrgAgentName string
	var metafileOrgAgentIdentifier string
	var metafilePersonAgentIdentifier string
	var metafilePersonAgentName string
	if extensionMap, _ := metadata.Extension.(map[string]any); extensionMap != nil {
		if contentSubPathAny, ok := extensionMap[ContentSubPathName]; ok {
			contentSubPath, _ = contentSubPathAny.(map[string]ContentSubPathEntry)
		}
		if metaFileAny, ok := extensionMap[MetaFileName]; ok {
			if metaFile, ok := metaFileAny.(map[string]any); ok {
				if str, ok := metaFile["organisation"]; ok {
					metafileOrgAgentName, _ = str.(string)
				}
				if str, ok := metaFile["organisation_address"]; ok {
					metafileOrgAgentIdentifier, _ = str.(string)
				}
				if str, ok := metaFile["address"]; ok {
					metafilePersonAgentIdentifier, _ = str.(string)
				}
				if str, ok := metaFile["user"]; ok {
					metafilePersonAgentName, _ = str.(string)
				}
			}
		}
	}
	var metsNames, premisNames *ocfl.NamesStruct
	var internalRelativePath, externalRelativePath, internalRelativePathCurrentVersion string
	switch strings.ToLower(me.StorageType) {
	case "area":
		metsNames, err = object.BuildNames([]string{me.MetsFile}, me.StorageName)
		if err != nil {
			return errors.Wrapf(err, "cannot build names for %s", me.MetsFile)
		}
		premisNames, err = object.BuildNames([]string{me.PremisFile}, me.StorageName)
		if err != nil {
			return errors.Wrapf(err, "cannot build names for %s", me.PremisFile)
		}
	case "path":
		path, err := object.GetAreaPath("content")
		if err != nil {
			return errors.Wrapf(err, "cannot get area path for '%s'", "content")
		}
		metsName := strings.TrimLeft(filepath.ToSlash(filepath.Join(path, me.StorageName, me.MetsFile)), "/")
		metsNames, err = object.BuildNames([]string{metsName}, "")
		if err != nil {
			return errors.Wrapf(err, "cannot build names for %s", metsName)
		}
		premisName := strings.TrimLeft(filepath.ToSlash(filepath.Join(path, me.StorageName, me.PremisFile)), "/")
		premisNames, err = object.BuildNames([]string{premisName}, "")
		if err != nil {
			return errors.Wrapf(err, "cannot build names for %s", premisName)
		}
	case "extension":
		metsName := strings.TrimLeft(filepath.ToSlash(filepath.Join(me.StorageName, fmt.Sprintf(me.MetsFile, object.GetVersion()))), "/")
		metsNames = &ocfl.NamesStruct{
			ExternalPaths: []string{me.MetsFile},
			InternalPath:  metsName,
			ManifestPath:  "",
		}
		premisName := strings.TrimLeft(filepath.ToSlash(filepath.Join(me.StorageName, fmt.Sprintf(me.PremisFile, object.GetVersion()))), "/")
		premisNames = &ocfl.NamesStruct{
			ExternalPaths: []string{me.PremisFile},
			InternalPath:  premisName,
			ManifestPath:  "",
		}
	default:
		return errors.Errorf("unsupported storage type '%s'", me.StorageType)
	}
	if len(premisNames.ExternalPaths) == 0 {
		return errors.Errorf("no external path for premis file")
	}
	if len(metsNames.ExternalPaths) == 0 {
		return errors.Errorf("no external path for mets file")
	}
	parts := strings.Split(metsNames.ExternalPaths[0], "/")
	for i := 1; i < len(parts); i++ {
		externalRelativePath += "../"
	}
	parts = strings.Split(metsNames.InternalPath, "/")
	for i := 1; i < len(parts)+2; i++ {
		internalRelativePath += "../"
	}
	for i := 1; i < len(parts); i++ {
		internalRelativePathCurrentVersion += "../"
	}

	metsFiles := map[string][]*mets.FileType{
		"schemas":  []*mets.FileType{},
		"content":  []*mets.FileType{},
		"metadata": []*mets.FileType{},
	}
	premisFiles := []*premis.File{}
	premisEvents := []*premis.EventComplexType{}
	dmdSecs := []*mets.MdSecType{}
	fileGrpUUID := map[string]string{
		"schemas":  uuid.NewString(),
		"content":  uuid.NewString(),
		"metadata": uuid.NewString(),
	}
	//metaFolder, _ := contentSubPath["metadata"]

	//internalPrefix := fmt.Sprintf("%s/content/", head)
	structPhysical := map[string][]string{
		"content":  []string{},
		"schemas":  []string{},
		"metadata": []string{},
	}
	/*
		structSemantical := map[string][]string{
			"Payload": []string{},
		}
	*/
	internalFiledata := map[string]metsInternalFiledata{}
	for cName, _ := range contentSubPath {
		//		structSemantical[cse.Description] = []string{}
		structPhysical[cName] = []string{}
		metsFiles[cName] = []*mets.FileType{}
		fileGrpUUID[cName] = uuid.NewString()
	}

	// get ingest versions
	for cs, metaFile := range metadata.Files {
		if extNames, ok := metaFile.VersionName[head]; ok {
			for _, extPath := range extNames {

				val := metsInternalFiledata{
					uuid: "uuid-" + uuid.NewString(),
					cs:   cs,
				}
				stateVersions := maps.Keys(metaFile.VersionName)
				for _, vStr := range inventory.GetVersionStrings() {
					if slices.Contains(stateVersions, vStr) {
						val.ingestVersion = vStr
						break
					}
				}
				internalFiledata[extPath] = val
			}
		}
	}

	for cs, metaFile := range metadata.Files {
		if extNames, ok := metaFile.VersionName[head]; ok {
			for _, extPath := range extNames {
				uuidString := internalFiledata[extPath].uuid
				var size int64
				var creationString string

				//		var fLocat = []*mets.FLocat{}
				if ext, ok := metaFile.Extension[FilesystemName]; ok {
					extFSL, ok := ext.(map[string][]*FileSystemLine)
					if !ok {
						return errors.Errorf("invalid type: %v", ext)
					}
					for _, ver := range inventory.GetVersionStrings() {
						if verHead, ok := extFSL[ver]; ok {
							if len(verHead) > 0 {
								creationString = verHead[0].Meta.CTime.Format("2006-01-02T15:04:05")
								size = int64(verHead[0].Meta.Size)
							}
						}
					}

				}

				metsFile := &mets.FileType{
					XMLName: xml.Name{},
					FILECORE: &mets.FILECORE{
						MIMETYPEAttr:     "application/octet-stream",
						SIZEAttr:         size,
						CREATEDAttr:      creationString,
						CHECKSUMAttr:     cs,
						CHECKSUMTYPEAttr: "SHA-512",
					},
					IDAttr:  uuidString,
					USEAttr: "Datafile",
					FLocat:  []*mets.FLocat{},
				}
				for verString, _ := range metaFile.VersionName {
					ver, ok := metadata.Versions[verString]
					if !ok {
						return errors.Errorf("cannot find version '%s'", verString)
					}
					premisEventPackage := &premis.EventComplexType{
						XMLName:     xml.Name{},
						XmlIDAttr:   "",
						VersionAttr: "",
						EventIdentifier: &premis.EventIdentifierComplexType{
							XMLName:              xml.Name{},
							SimpleLinkAttr:       "",
							EventIdentifierType:  premis.NewStringPlusAuthority("uuid", "", "", ""),
							EventIdentifierValue: "uuid-" + uuid.NewString(),
						},
						EventType:     premis.NewStringPlusAuthority("packing", "eventType", "https://id.loc.gov/vocabulary/preservation/eventType", "https://id.loc.gov/vocabulary/preservation/eventType/pac.html"),
						EventDateTime: ver.Created.Format(time.RFC3339),
						EventOutcomeInformation: []*premis.EventOutcomeInformationComplexType{
							&premis.EventOutcomeInformationComplexType{
								XMLName:      xml.Name{},
								EventOutcome: premis.NewStringPlusAuthority("success", "eventOutcome", "https://id.loc.gov/vocabulary/preservation/eventOutcome", "https://id.loc.gov/vocabulary/preservation/eventOutcome/suc.html"),
							},
						},
						LinkingAgentIdentifier: []*premis.LinkingAgentIdentifierComplexType{
							&premis.LinkingAgentIdentifierComplexType{
								XMLName:                     xml.Name{},
								LinkAgentXmlIDAttr:          "",
								SimpleLinkAttr:              "",
								LinkingAgentIdentifierType:  premis.NewStringPlusAuthority("local", "", "", ""),
								LinkingAgentIdentifierValue: ver.Address,
							},
							&premis.LinkingAgentIdentifierComplexType{
								XMLName:                     xml.Name{},
								LinkAgentXmlIDAttr:          "",
								SimpleLinkAttr:              "",
								LinkingAgentIdentifierType:  premis.NewStringPlusAuthority("local", "", "", ""),
								LinkingAgentIdentifierValue: metafileOrgAgentIdentifier,
							},
							&premis.LinkingAgentIdentifierComplexType{
								XMLName:                     xml.Name{},
								LinkAgentXmlIDAttr:          "",
								SimpleLinkAttr:              "",
								LinkingAgentIdentifierType:  premis.NewStringPlusAuthority("local", "", "", ""),
								LinkingAgentIdentifierValue: "https://github.com/je4/gocfl",
							},
						},
						LinkingObjectIdentifier: []*premis.LinkingObjectIdentifierComplexType{
							&premis.LinkingObjectIdentifierComplexType{
								XMLName:                      xml.Name{},
								LinkingObjectIdentifierType:  premis.NewStringPlusAuthority("uuid", "", "", ""),
								LinkingObjectIdentifierValue: uuidString,
							},
						},
					}
					premisEvents = append(premisEvents, premisEventPackage)
				}
				premisFile := &premis.File{
					XMLName:   xml.Name{},
					XSIType:   "file",
					XmlIDAttr: uuidString,
					ObjectIdentifier: []*premis.ObjectIdentifierComplexType{
						&premis.ObjectIdentifierComplexType{
							XMLName:               xml.Name{},
							ObjectIdentifierType:  premis.NewStringPlusAuthority("uuid", "", "", ""),
							ObjectIdentifierValue: uuidString,
						},
					},
					SignificantProperties: []*premis.SignificantPropertiesComplexType{},
					ObjectCharacteristics: []*premis.ObjectCharacteristicsComplexType{},
					Storage:               []*premis.StorageComplexType{},
				}
				var mimeType string
				if ext, ok := metaFile.Extension[IndexerName]; ok {
					extIndexer, ok := ext.(*indexer.ResultV2)
					if !ok {
						return errors.Wrapf(err, "invalid type: %v", ext)
					}
					metsFile.FILECORE.MIMETYPEAttr = extIndexer.Mimetype
					objectCharacter := &premis.ObjectCharacteristicsComplexType{
						XMLName:          xml.Name{},
						CompositionLevel: nil,
						Fixity: []*premis.FixityComplexType{
							premis.NewFixityComplexType(string(metadata.DigestAlgorithm), cs, "gocfl "+version.VERSION),
						},
						Size:   int64(extIndexer.Size),
						Format: []*premis.FormatComplexType{},
					}
					if extIndexer.Mimetype != "" {
						objectCharacter.Format = append(objectCharacter.Format, &premis.FormatComplexType{
							XMLName: xml.Name{},
							FormatDesignation: &premis.FormatDesignationComplexType{
								XMLName:    xml.Name{},
								FormatName: premis.NewStringPlusAuthority(extIndexer.Mimetype, "", "", ""),
							},
							FormatNote: []string{"IANA MIME-type"},
						})
					}
					if extIndexer.Pronom != "" {
						sfAny, _ := extIndexer.Metadata["siegfried"]
						sfAnyList, ok := sfAny.([]any)
						if ok {
							for _, sfEntry := range sfAnyList {
								sfMap, ok := sfEntry.(map[string]any)
								if !ok {
									continue
								}
								fct := &premis.FormatComplexType{
									XMLName:    xml.Name{},
									FormatNote: []string{"siegfried"},
								}
								if sfBasisAny, ok := sfMap["Basis"]; ok {
									if sfBasisAnyList, ok := sfBasisAny.([]any); ok {
										for _, sfBasisEntryAny := range sfBasisAnyList {
											if sfBasisString, ok := sfBasisEntryAny.(string); ok {
												fct.FormatNote = append(fct.FormatNote, "Basis: "+sfBasisString)
											}
										}
									}
								}

								if designationAny, ok := sfMap["Name"]; ok {
									if designation, ok := designationAny.(string); ok {
										fct.FormatDesignation = &premis.FormatDesignationComplexType{
											XMLName:    xml.Name{},
											FormatName: premis.NewStringPlusAuthority(designation, "", "", ""),
										}
									}
								}
								if idAny, ok := sfMap["ID"]; ok {
									if id, ok := idAny.(string); ok {
										fct.FormatRegistry = &premis.FormatRegistryComplexType{
											XMLName:            xml.Name{},
											FormatRegistryName: premis.NewStringPlusAuthority("PRONOM", "", "", ""),
											FormatRegistryKey:  premis.NewStringPlusAuthority(id, "", "", ""),
											FormatRegistryRole: premis.NewStringPlusAuthority(
												"specification",
												"http://id.loc.gov/vocabulary/preservation/formatRegistryRole",
												"http://id.loc.gov/vocabulary/preservation/formatRegistryRole",
												"http://id.loc.gov/vocabulary/preservation/formatRegistryRole/spe",
											),
										}
									}
									objectCharacter.Format = append(objectCharacter.Format, fct)
								}
							}
						}
					}
					for digest, checksum := range metaFile.Checksums {
						objectCharacter.Fixity = append(objectCharacter.Fixity,
							premis.NewFixityComplexType(string(digest), checksum, "gocfl "+version.VERSION),
						)
					}
					if extIndexer != nil {
						if extIndexer.Width > 0 {
							premisFile.SignificantProperties = append(premisFile.SignificantProperties,
								premis.NewSignificantPropertiesComplexType("width", fmt.Sprintf("%v", extIndexer.Width)),
							)
							premisFile.SignificantProperties = append(premisFile.SignificantProperties,
								premis.NewSignificantPropertiesComplexType("height", fmt.Sprintf("%v", extIndexer.Height)),
							)
						}
						if extIndexer.Duration > 0 {
							premisFile.SignificantProperties = append(premisFile.SignificantProperties,
								premis.NewSignificantPropertiesComplexType("duration", fmt.Sprintf("%v", extIndexer.Duration)),
							)
						}
						if extIndexer.Size > 0 {
							objectCharacter.Size = int64(extIndexer.Size)
						}
					}
					premisFile.ObjectCharacteristics = append(premisFile.ObjectCharacteristics, objectCharacter)
				}
				if len(premisFile.ObjectCharacteristics) == 0 {
					premisFile.ObjectCharacteristics = append(premisFile.ObjectCharacteristics, &premis.ObjectCharacteristicsComplexType{
						XMLName:          xml.Name{},
						CompositionLevel: nil,
						Fixity:           []*premis.FixityComplexType{},
						Size:             size,
						Format: []*premis.FormatComplexType{
							&premis.FormatComplexType{
								XMLName: xml.Name{},
								FormatDesignation: &premis.FormatDesignationComplexType{
									XMLName:    xml.Name{},
									FormatName: premis.NewStringPlusAuthority("application/octet-stream", "", "", ""),
								},
								FormatNote: []string{"IANA MIME-type"},
							},
						},
					})
				}
				for _, intPath := range metaFile.InternalName {
					parts := strings.Split(intPath, "/")
					if len(parts) <= 2 {
						return errors.Wrapf(err, "invalid path %s", intPath)
					}
					if parts[1] != "content" {
						return errors.Wrapf(err, "no content in %s", intPath)
					}
					var intArea = "content"
					var isSchema bool
					//					var intSemantic = "Other Payload"
					if len(parts) > 3 {
						if contentSubPath != nil {
							intArea = parts[2]
							// intSemantic = ""
							for area, cse := range contentSubPath {
								if cse.Path == intArea {
									intArea = area
									// intSemantic = cse.Description
									isSchema = parts[3] == "schemas"
									break
								}
							}
						}
					}
					if intArea == "metadata" && !isSchema {
						dmdSecs = append(dmdSecs, newMDSec(
							fmt.Sprintf("dmdSec-int-%s", uuidString),
							"area-metadata",
							intPath,
							"OTHER",
							"URL:internal",
							mimeType,
							creationString,
							size,
							"",
							"",
							cs,
							string(inventory.GetDigestAlgorithm())))
						continue
					}

					if isSchema {
						structPhysical["schemas"] = append(structPhysical["schemas"], uuidString)
					} else {
						structPhysical[intArea] = append(structPhysical[intArea], uuidString)
					}
					/*
						if intSemantic != "" {
							structSemantical[intSemantic] = append(structSemantical[intSemantic], uuidString)
						}
					*/
					href := intPath
					metsFile.FLocat = append(metsFile.FLocat, &mets.FLocat{
						LOCATION: &mets.LOCATION{
							LOCTYPEAttr:      "OTHER",
							OTHERLOCTYPEAttr: "URL:internal",
						},
						SimpleLink: &mets.SimpleLink{
							//XMLName:          xml.Name{},
							TypeAttr:      "simple",
							XlinkHrefAttr: href,
						},
						IDAttr:  "",
						USEAttr: "",
					})
					/*
						premisFile.Storage = append(premisFile.Storage, &premis.StorageComplexType{
							XMLName: xml.Name{},
							ContentLocation: &premis.ContentLocationComplexType{
								XMLName:              xml.Name{},
								SimpleLinkAttr:       "",
								ContentLocationType:  premis.NewStringPlusAuthority("internal", "", "", ""),
								ContentLocationValue: href,
							},
							StorageMedium: premis.NewStringPlusAuthority("OCFL Object Root", "", "", ""),
						})
					*/
				}
				//		if extNames, ok := metaFile.VersionName[head]; ok {
				//			for _, extPath := range extNames {
				parts := strings.Split(extPath, "/")
				var extArea = "content"
				var isSchema bool
				if len(parts) > 1 {
					if contentSubPath != nil {
						extArea = parts[0]
						for area, cse := range contentSubPath {
							if cse.Path == extArea {
								extArea = area
								isSchema = parts[1] == "schemas"
								break
							}
						}
					}
				}
				if extArea == "metadata" && !isSchema {
					if !slices.Contains([]string{metsNames.ExternalPaths[0], premisNames.ExternalPaths[0]}, extPath) {
						dmdSecs = append(dmdSecs, newMDSec(
							fmt.Sprintf("dmdSec-ext-%s", uuidString),
							"area-metadata",
							extPath,
							"URL",
							"",
							mimeType,
							creationString,
							size,
							"",
							"",
							cs,
							string(inventory.GetDigestAlgorithm())))
					}
					continue
				}

				metsFile.FLocat = append(metsFile.FLocat, &mets.FLocat{
					LOCATION: &mets.LOCATION{
						LOCTYPEAttr: "URL",
					},
					SimpleLink: &mets.SimpleLink{
						//XMLName:          xml.Name{},
						TypeAttr: "simple",
						XlinkHrefAttr:/* externalRelativePath + */ extPath,
					},
				})
				/*
					premisFile.Storage = append(premisFile.Storage, &premis.StorageComplexType{
						XMLName: xml.Name{},
						ContentLocation: &premis.ContentLocationComplexType{
							XMLName:             xml.Name{},
							SimpleLinkAttr:      "",
							ContentLocationType: premis.NewStringPlusAuthority("external", "", "", ""),
							ContentLocationValue: extPath,
						},
						StorageMedium: premis.NewStringPlusAuthority("extracted OCFL", "", "", ""),
					})
				*/

				//			}
				//		}
				/*
									var ingestTime time.Time
									var ingestVersion string
					//				_ = ingestVersion
									if internal, ok := internalFiledata[extPath]; ok {
										if internal.ingestVersion != "" {
											ingestVersion = internal.ingestVersion
											if versionData, ok := metadata.Versions[internal.ingestVersion]; ok {
												ingestTime = versionData.Created
											}
										}
									}
				*/
				if migrationAny, ok := metaFile.Extension[MigrationName]; ok {
					ver, ok := metadata.Versions[head]
					if !ok {
						return errors.Errorf("cannot find head version '%s'", head)
					}
					migration, ok := migrationAny.(*MigrationResult)
					if !ok {
						return errors.Wrapf(err, "invalid type for migration of '%s': %v", cs, migrationAny)
					}
					eventMigration := &premis.EventComplexType{
						XMLName:     xml.Name{},
						XmlIDAttr:   "",
						VersionAttr: "",
						EventIdentifier: &premis.EventIdentifierComplexType{
							XMLName:              xml.Name{},
							EventIdentifierType:  premis.NewStringPlusAuthority("uuid", "", "", ""),
							EventIdentifierValue: "uuid-" + uuid.NewString(), //migration.ID,
						},
						EventType:               premis.NewStringPlusAuthority("migration", "eventType", "https://id.loc.gov/vocabulary/preservation/eventType", "https://id.loc.gov/vocabulary/preservation/eventType/mig.html"),
						EventDateTime:           ver.Created.Format(time.RFC3339),
						EventDetailInformation:  nil,
						EventOutcomeInformation: []*premis.EventOutcomeInformationComplexType{},
						LinkingAgentIdentifier: []*premis.LinkingAgentIdentifierComplexType{
							&premis.LinkingAgentIdentifierComplexType{
								XMLName:                     xml.Name{},
								LinkAgentXmlIDAttr:          "",
								SimpleLinkAttr:              "",
								LinkingAgentIdentifierType:  premis.NewStringPlusAuthority("local", "", "", ""),
								LinkingAgentIdentifierValue: migration.ID,
								LinkingAgentRole:            nil,
							},
						},
						LinkingObjectIdentifier: []*premis.LinkingObjectIdentifierComplexType{},
					}
					var sourcePath string
					for intExtPath, val := range internalFiledata {
						if val.cs == migration.Source {
							sourcePath = intExtPath
							break
						}
					}
					if internal, ok := internalFiledata[sourcePath]; ok {
						if internal.uuid != "" {
							eventMigration.LinkingObjectIdentifier = append(eventMigration.LinkingObjectIdentifier, &premis.LinkingObjectIdentifierComplexType{
								XMLName:                      xml.Name{},
								LinkingObjectIdentifierType:  premis.NewStringPlusAuthority("uuid", "", "", ""),
								LinkingObjectIdentifierValue: internal.uuid,
								LinkingObjectRole: []*premis.StringPlusAuthority{
									premis.NewStringPlusAuthority("source", "eventRelatedObjectRole", "https://id.loc.gov/vocabulary/preservation/eventRelatedObjectRole", "https://id.loc.gov/vocabulary/preservation/eventRelatedObjectRole/sou.html"),
								},
							})
						}
					}

					if migration.Error == "" {
						eventMigration.LinkingObjectIdentifier = append(eventMigration.LinkingObjectIdentifier,
							&premis.LinkingObjectIdentifierComplexType{
								XMLName:                      xml.Name{},
								LinkingObjectIdentifierType:  premis.NewStringPlusAuthority("uuid", "", "", ""),
								LinkingObjectIdentifierValue: uuidString,
								LinkingObjectRole: []*premis.StringPlusAuthority{
									premis.NewStringPlusAuthority("outcome", "eventRelatedObjectRole", "https://id.loc.gov/vocabulary/preservation/eventRelatedObjectRole", "https://id.loc.gov/vocabulary/preservation/eventRelatedObjectRole/out.html"),
								},
							})
						eventMigration.EventOutcomeInformation = append(eventMigration.EventOutcomeInformation, &premis.EventOutcomeInformationComplexType{
							XMLName:      xml.Name{},
							EventOutcome: premis.NewStringPlusAuthority("success", "eventOutcome", "https://id.loc.gov/vocabulary/preservation/eventOutcome", "https://id.loc.gov/vocabulary/preservation/eventOutcome/suc.html"),
						})
					} else {
						eventMigration.EventOutcomeInformation = append(eventMigration.EventOutcomeInformation, &premis.EventOutcomeInformationComplexType{
							XMLName:      xml.Name{},
							EventOutcome: premis.NewStringPlusAuthority("fail", "eventOutcome", "https://id.loc.gov/vocabulary/preservation/eventOutcome/fai.html", "http://id.loc.gov/vocabulary/preservation/eventOutcome/fai"),
							EventOutcomeDetail: []*premis.EventOutcomeDetailComplexType{
								&premis.EventOutcomeDetailComplexType{
									XMLName:                xml.Name{},
									EventOutcomeDetailNote: migration.Error,
								},
							},
						})
					}
					premisEvents = append(premisEvents, eventMigration)
				}
				if len(metsFile.FLocat) > 0 {
					a := extArea
					if isSchema {
						a = "schemas"
					}
					metsFiles[a] = append(metsFiles[a], metsFile)
				}
				premisFiles = append(premisFiles, premisFile)
			}
		}
	}

	if me.PrimaryDescriptiveMetadata != "" {
		var metaFilename string
		var metaType string
		var metaArea string

		parts := strings.Split(me.PrimaryDescriptiveMetadata, ":")
		switch len(parts) {
		case 2:
			metaType = parts[0]
			metaArea = "content"
			if len(contentSubPath) == 0 {
				metaFilename = filepath.ToSlash(filepath.Clean(parts[1]))
			} else {
				if path, ok := contentSubPath[metaArea]; ok {
					metaFilename = filepath.ToSlash(filepath.Join(path.Path, parts[1]))
				} else {
					return errors.Errorf("cannot find content sub path '%s' for file '%s'", metaArea, me.PrimaryDescriptiveMetadata)
				}
			}
		case 3:
			metaType = parts[0]
			metaArea = parts[1]
			if path, ok := contentSubPath[metaArea]; ok {
				metaFilename = filepath.ToSlash(filepath.Join(path.Path, parts[2]))
			} else {
				return errors.Errorf("cannot find content sub path '%s' for file '%s'", metaArea, me.PrimaryDescriptiveMetadata)
			}
		default:
			return errors.Errorf("invalid descriptive metadata '%s'", me.PrimaryDescriptiveMetadata)
		}
		var found *ocfl.FileMetadata
		var foundChecksum string
		for checksum, metaFile := range metadata.Files {
			if ver, ok := metaFile.VersionName[head]; ok {
				for _, name := range ver {
					if name == metaFilename {
						found = metaFile
						foundChecksum = checksum
						break
					}
				}
				if found != nil {
					break
				}
			}
		}

		if found == nil {
			return errors.Errorf("cannot find descriptive metadata file '%s'", me.PrimaryDescriptiveMetadata)
		}
		metaType = strings.ToUpper(metaType)
		var otherMetaType string
		if !slices.Contains(metsMDTypes, metaType) {
			otherMetaType = metaType
			metaType = "OTHER"
		}
		/*
			switch metaType {
			case "MARC":
				foundMimetype = "application/marc"
			case "MARCXML":
				foundMimetype = "application/marcxml+xml"
			case "JSON":
				foundMimetype = "text/json"
			case "XML":
				foundMimetype = "text/xml"
			}
		*/

		// remove any existing mdSecs with the same checksum
		// todo: do it for internal and external name separately
		mdSecs2 := dmdSecs
		dmdSecs = make([]*mets.MdSecType, 0, len(mdSecs2))
		for i := 0; i < len(mdSecs2); i++ {
			if mdSecs2[i].MdRef.CHECKSUMAttr != foundChecksum {
				dmdSecs = append(dmdSecs, mdSecs2[i])
				structPhysical["metadata"] = append(structPhysical["metadata"], mdSecs2[i].IDAttr)
			}
		}
		mdSecs2 = nil

		var foundCreationString string
		var foundSize int64
		if ext, ok := found.Extension[FilesystemName]; ok {
			extFSL, ok := ext.(map[string][]*FileSystemLine)
			if !ok {
				return errors.Errorf("invalid type: %v", ext)
			}
			for _, ver := range inventory.GetVersionStrings() {
				if verHead, ok := extFSL[ver]; ok {
					if len(verHead) > 0 {
						foundCreationString = verHead[0].Meta.CTime.Format("2006-01-02T15:04:05")
						foundSize = int64(verHead[0].Meta.Size)
					}
				}
			}
		}
		var foundMimetype string = "application/octet-stream"

		if ext, ok := found.Extension[IndexerName]; ok {
			extIndexer, ok := ext.(*indexer.ResultV2)
			if !ok {
				return errors.Wrapf(err, "invalid type: %v", ext)
			}
			foundMimetype = extIndexer.Mimetype
		}

		if len(found.InternalName) > 0 {
			id := fmt.Sprintf("dmdSec-int-%s-%s", slug.Make(object.GetID()), head)
			dmdSecs = append(dmdSecs, newMDSec(
				id,
				"primary-metadata",
				found.InternalName[0],
				"OTHER",
				"URL:internal",
				foundMimetype,
				foundCreationString,
				foundSize,
				metaType,
				otherMetaType,
				foundChecksum,
				string(inventory.GetDigestAlgorithm())))
			structPhysical["metadata"] = append(structPhysical["metadata"], id)
		}
		if len(found.VersionName[head]) > 0 {
			id := fmt.Sprintf("dmdSec-ext-%s-%s", slug.Make(object.GetID()), head)
			dmdSecs = append(dmdSecs, newMDSec(
				id,
				"primary-metadata",
				found.VersionName[head][0],
				"URL",
				"",
				foundMimetype,
				foundCreationString,
				foundSize,
				metaType,
				otherMetaType,
				foundChecksum,
				string(inventory.GetDigestAlgorithm())))
			structPhysical["metadata"] = append(structPhysical["metadata"], id)
		}
	}

	agentIdentifier := metafilePersonAgentIdentifier
	agentName := metafilePersonAgentName
	if agentName == "" {
		agentName = v.User.Name.String()
		agentIdentifier = v.User.Address.String()
	}

	premisStruct := &premis.PremisComplexType{
		XMLName:           xml.Name{},
		XMLNS:             "http://www.loc.gov/premis/v3",
		XMLXLinkNS:        "http://www.w3.org/1999/xlink",
		XMLNSXSI:          "http://www.w3.org/2001/XMLSchema-instance",
		XSISchemaLocation: "http://www.loc.gov/premis/v3\nschemas/premis.xsd\nhttp://www.w3.org/1999/xlink\nschemas/xlink.xsd",
		VersionAttr:       "3.0",
		Object:            premisFiles,
		Event:             premisEvents,
		Agent: []*premis.AgentComplexType{
			&premis.AgentComplexType{
				XMLName:     xml.Name{},
				XmlIDAttr:   "",
				VersionAttr: "",
				AgentIdentifier: []*premis.AgentIdentifierComplexType{
					&premis.AgentIdentifierComplexType{
						XMLName:              xml.Name{},
						SimpleLinkAttr:       "",
						AgentIdentifierType:  premis.NewStringPlusAuthority("organization", "agentType", "https://id.loc.gov/vocabulary/preservation/agentType", "https://id.loc.gov/vocabulary/preservation/agentType/org.html"),
						AgentIdentifierValue: metafileOrgAgentIdentifier,
					}},
				AgentName: []*premis.StringPlusAuthority{
					premis.NewStringPlusAuthority(metafileOrgAgentName, "", "", ""),
				},
			},
			&premis.AgentComplexType{
				XMLName:     xml.Name{},
				XmlIDAttr:   "",
				VersionAttr: "",
				AgentIdentifier: []*premis.AgentIdentifierComplexType{
					&premis.AgentIdentifierComplexType{
						XMLName:              xml.Name{},
						SimpleLinkAttr:       "",
						AgentIdentifierType:  premis.NewStringPlusAuthority("local", "", "", ""),
						AgentIdentifierValue: agentIdentifier,
					}},
				AgentName: []*premis.StringPlusAuthority{
					premis.NewStringPlusAuthority(agentName, "", "", ""),
				},
			},
			&premis.AgentComplexType{
				XMLName:     xml.Name{},
				XmlIDAttr:   "",
				VersionAttr: "",
				AgentIdentifier: []*premis.AgentIdentifierComplexType{
					&premis.AgentIdentifierComplexType{
						XMLName:              xml.Name{},
						SimpleLinkAttr:       "",
						AgentIdentifierType:  premis.NewStringPlusAuthority("software", "agentType", "https://id.loc.gov/vocabulary/preservation/agentType", "https://id.loc.gov/vocabulary/preservation/agentType/sof.html"),
						AgentIdentifierValue: "https://github.com/je4/gocfl",
					}},
				AgentName: []*premis.StringPlusAuthority{
					premis.NewStringPlusAuthority(fmt.Sprintf("gocfl %s - Go OCFL implementation", version.VERSION), "", "", ""),
				},
			},
		},
		Rights: []*premis.RightsComplexType{},
	}

	premisBytes, err := xml.MarshalIndent(premisStruct, "", "  ")
	if err != nil {
		return errors.Wrap(err, "cannot marshal PREMIS")
	}

	premisChecksum := fmt.Sprintf("%x", sha512.Sum512(premisBytes))

	metsFileGrps := []*mets.FileGrp{}
	for a, files := range metsFiles {
		if len(files) == 0 {
			continue
		}
		metsFileGrps = append(metsFileGrps, &mets.FileGrp{
			XMLName: xml.Name{},
			FileGrpType: &mets.FileGrpType{
				XMLName: xml.Name{},
				IDAttr:  "uuid-" + fileGrpUUID[a],
				USEAttr: a,
				File:    files,
			},
		})
	}

	var amdSecs = []*mets.AmdSecType{}
	if premisNames != nil {
		id := "amdSec-int-uuid-" + uuid.NewString()
		sec := &mets.AmdSecType{
			XMLName: xml.Name{},
			IDAttr:  "uuid-" + uuid.NewString(),
			DigiprovMD: []*mets.MdSecType{
				newMDSec(
					id,
					"",
					premisNames.InternalPath,
					"OTHER",
					"URL:internal",
					"application/xml",
					"",
					int64(len(premisBytes)),
					"PREMIS",
					"",
					premisChecksum,
					"SHA-512"),
			},
		}
		structPhysical["metadata"] = append(structPhysical["metadata"], id)
		for _, ext := range premisNames.ExternalPaths {
			id := "amdSec-ext-uuid-" + uuid.NewString()
			sec.DigiprovMD = append(sec.DigiprovMD, newMDSec(
				id,
				"",
				ext,
				"URL",
				"",
				"application/xml",
				"", int64(len(premisBytes)),
				"PREMIS",
				"",
				premisChecksum,
				"SHA-512"))
			structPhysical["metadata"] = append(structPhysical["metadata"], id)
		}
		amdSecs = append(amdSecs, sec)
	}

	structMapPhysicalId := uuid.New()
	structMapPhysicalIdString := "urn:uuid:" + structMapPhysicalId.String()
	structMapPhysical := &mets.StructMapType{
		XMLName:   xml.Name{},
		IDAttr:    "",
		TYPEAttr:  "physical",
		LABELAttr: "AIP structMap",
		Div: &mets.DivType{
			XMLName: xml.Name{},
			ORDERLABELS: &mets.ORDERLABELS{
				ORDERAttr:      0,
				ORDERLABELAttr: "",
				LABELAttr:      structMapPhysicalIdString,
			},
			IDAttr: "uuid-" + structMapPhysicalId.String() + "-structMap-div",
			Div:    []*mets.DivType{},
		},
	}

	structMaps := []*mets.StructMapType{}

	for area, fileList := range structPhysical {
		if len(fileList) == 0 {
			continue
		}

		div := &mets.DivType{
			XMLName: xml.Name{},
			ORDERLABELS: &mets.ORDERLABELS{
				LABELAttr: area,
			},
			IDAttr: "uuid-" + uuid.New().String() + "-structMap-div",
			Fptr:   make([]*mets.Fptr, 0),
		}
		for _, u := range fileList {
			div.Fptr = append(div.Fptr, &mets.Fptr{
				XMLName:    xml.Name{},
				FILEIDAttr: u,
			})
		}
		if len(div.Fptr) > 0 {
			structMapPhysical.Div.Div = append(structMapPhysical.Div.Div, div)
		}

		//	structMapPhysical.Div.Div = append(structMapPhysical.Div.Div, structMapPhysicalDivVer)
	}
	structMaps = append(structMaps, structMapPhysical)

	structMapSemantical := &mets.StructMapType{
		XMLName:   xml.Name{},
		TYPEAttr:  "logical",
		LABELAttr: "AIP Structure",
		Div: &mets.DivType{
			XMLName: xml.Name{},
			ORDERLABELS: &mets.ORDERLABELS{
				LABELAttr: "Package Structure",
			},
			IDAttr: "uuid-" + uuid.New().String() + "-structMap-div",
			Div:    []*mets.DivType{},
		},
	}

	for area, uuids := range structPhysical {
		if len(uuids) == 0 {
			continue
		}

		var label = "Other Data"
		if cse, ok := contentSubPath[area]; ok {
			label = cse.Description
		} else {
			switch area {
			case "schemas":
				label = "Metadata Schema"
			case "metadata":
				label = "Metadata Files"
			}
		}
		div := &mets.DivType{
			XMLName: xml.Name{},
			ORDERLABELS: &mets.ORDERLABELS{
				LABELAttr: label,
			},
			IDAttr: "uuid-" + uuid.NewString() + "-structMap-div",
			Fptr:   make([]*mets.Fptr, 0),
		}
		for _, u := range uuids {
			div.Fptr = append(div.Fptr, &mets.Fptr{
				XMLName:    xml.Name{},
				FILEIDAttr: u,
			})
		}
		structMapSemantical.Div.Div = append(structMapSemantical.Div.Div, div)
	}
	structMaps = append(structMaps, structMapSemantical)
	archivistName := metafilePersonAgentName
	archivistAddress := metafilePersonAgentIdentifier
	if archivistName == "" {
		archivistName = v.User.Name.String()
		archivistAddress = v.User.Address.String()
	}

	m := &mets.Mets{
		XMLNS:             "http://www.loc.gov/METS/",
		XMLXLinkNS:        "http://www.w3.org/1999/xlink",
		XMLNSXSI:          "http://www.w3.org/2001/XMLSchema-instance",
		XSISchemaLocation: "http://www.loc.gov/METS/\nschemas/mets.xsd\nhttp://www.w3.org/1999/xlink\nschemas/xlink.xsd",
		MetsType: &mets.MetsType{
			XMLName:     xml.Name{},
			OBJIDAttr:   metadata.ID,
			LABELAttr:   fmt.Sprintf("METS Container for Object %s version %s - %s", metadata.ID, head, v.Message),
			TYPEAttr:    "AIP",
			PROFILEAttr: "http://www.ra.ee/METS/v01/IP.xml",
			MetsHdr: &mets.MetsHdr{
				XMLName:          xml.Name{},
				CREATEDATEAttr:   v.Created.Format("2006-01-02T15:04:05"),
				RECORDSTATUSAttr: "NEW",
				Agent: []*mets.Agent{
					&mets.Agent{
						XMLName:       xml.Name{},
						ROLEAttr:      "CREATOR",
						TYPEAttr:      "OTHER",
						OTHERTYPEAttr: "SOFTWARE",
						Name:          "gocfl",
						Note: []*mets.Note{
							&mets.Note{
								XMLName: xml.Name{},
								Value:   fmt.Sprintf("Version %s", version.VERSION),
							},
						},
					},
					&mets.Agent{
						XMLName:  xml.Name{},
						ROLEAttr: "ARCHIVIST",
						Name:     archivistName,
						Note: []*mets.Note{
							&mets.Note{
								XMLName: xml.Name{},
								Value:   archivistAddress,
							},
						},
					},
				},
				AltRecordID: nil,
				MetsDocumentID: &mets.MetsDocumentID{
					XMLName: xml.Name{},
					Value:   "mets.xml",
				},
			},
			DmdSec: dmdSecs,
			AmdSec: amdSecs,
			FileSec: &mets.FileSec{
				XMLName: xml.Name{},
				IDAttr:  "uuid-" + uuid.NewString(),
				FileGrp: metsFileGrps,
			},
			StructMap: structMaps,
		}}

	metsBytes, err := xml.MarshalIndent(m, "", "  ")
	if err != nil {
		return errors.Wrap(err, "cannot marshal METS")
	}

	switch strings.ToLower(me.StorageType) {
	case "area":
		metsName := me.MetsFile
		premisName := me.PremisFile

		//		if err := object.AddReader(io.NopCloser(bytes.NewBuffer(metsBytes)), []string{metsName}, area, true, false); err != nil {
		if err := object.AddData(metsBytes, metsName, false, me.StorageName, true, false); err != nil {
			return errors.Wrapf(err, "cannot write '%s'", metsName)
		}
		if err := object.AddReader(io.NopCloser(bytes.NewBuffer(premisBytes)), []string{premisName}, me.StorageName, true, false); err != nil {
			return errors.Wrapf(err, "cannot write '%s'", premisName)
		}
		//		if err := object.AddReader(io.NopCloser(bytes.NewBuffer(specs.METSXSD)), []string{"schemas/mets.xsd"}, area, true, false); err != nil {
		if err := object.AddData(specs.METSXSD, "schemas/mets.xsd", true, me.StorageName, true, false); err != nil {
			return errors.Wrapf(err, "cannot write '%s'", "schemas/mets.xsd")
		}
		//		if err := object.AddReader(io.NopCloser(bytes.NewBuffer(specs.PremisXSD)), []string{"schemas/premis.xsd"}, area, true, false); err != nil {
		if err := object.AddData(specs.PremisXSD, "schemas/premis.xsd", true, me.StorageName, true, false); err != nil {
			return errors.Wrapf(err, "cannot write '%s'", "schemas/premis.xsd")
		}
		//if err := object.AddReader(io.NopCloser(bytes.NewBuffer(specs.XLinkXSD)), []string{"schemas/xlink.xsd"}, area, true, false); err != nil {
		if err := object.AddData(specs.XLinkXSD, "schemas/xlink.xsd", true, me.StorageName, true, false); err != nil {
			return errors.Wrapf(err, "cannot write '%s'", "schemas/xlink.xsd")
		}
	case "path":
		path, err := object.GetAreaPath("content")
		if err != nil {
			return errors.Wrapf(err, "cannot get area path for '%s'", "content")
		}
		metsName := strings.TrimLeft(filepath.ToSlash(filepath.Join(path, me.StorageName, me.MetsFile)), "/")
		premisName := strings.TrimLeft(filepath.ToSlash(filepath.Join(path, me.StorageName, me.PremisFile)), "/")

		//		if err := object.AddReader(io.NopCloser(bytes.NewBuffer(metsBytes)), []string{metsName}, area, true, false); err != nil {
		if err := object.AddData(metsBytes, metsName, false, "", true, false); err != nil {
			return errors.Wrapf(err, "cannot write '%s'", metsName)
		}
		if err := object.AddReader(io.NopCloser(bytes.NewBuffer(premisBytes)), []string{premisName}, "", true, false); err != nil {
			return errors.Wrapf(err, "cannot write '%s'", premisName)
		}
		//		if err := object.AddReader(io.NopCloser(bytes.NewBuffer(specs.METSXSD)), []string{"schemas/mets.xsd"}, area, true, false); err != nil {
		if err := object.AddData(specs.METSXSD, "schemas/mets.xsd", true, "", true, false); err != nil {
			return errors.Wrapf(err, "cannot write '%s'", "schemas/mets.xsd")
		}
		//		if err := object.AddReader(io.NopCloser(bytes.NewBuffer(specs.PremisXSD)), []string{"schemas/premis.xsd"}, area, true, false); err != nil {
		if err := object.AddData(specs.PremisXSD, "schemas/premis.xsd", true, "", true, false); err != nil {
			return errors.Wrapf(err, "cannot write '%s'", "schemas/premis.xsd")
		}
		//if err := object.AddReader(io.NopCloser(bytes.NewBuffer(specs.XLinkXSD)), []string{"schemas/xlink.xsd"}, area, true, false); err != nil {
		if err := object.AddData(specs.XLinkXSD, "schemas/xlink.xsd", true, "", true, false); err != nil {
			return errors.Wrapf(err, "cannot write '%s'", "schemas/xlink.xsd")
		}
	case "extension":
		metsName := strings.TrimLeft(filepath.ToSlash(filepath.Join(me.StorageName, fmt.Sprintf(me.MetsFile, object.GetVersion()))), "/")
		premisName := strings.TrimLeft(filepath.ToSlash(filepath.Join(me.StorageName, fmt.Sprintf(me.PremisFile, object.GetVersion()))), "/")
		if err := writefs.WriteFile(me.fsys, metsName, metsBytes); err != nil {
			return errors.Wrapf(err, "cannot write file '%v/%s'", me.fsys, metsName)
		}
		if err := writefs.WriteFile(me.fsys, premisName, metsBytes); err != nil {
			return errors.Wrapf(err, "cannot write file '%v/%s'", me.fsys, premisName)
		}
		if err := writefs.WriteFile(me.fsys, "schemas/mets.xsd", specs.METSXSD); err != nil {
			return errors.Wrapf(err, "cannot write file '%v/%s'", me.fsys, "schemas/mets.xsd")
		}
		if err := writefs.WriteFile(me.fsys, "schemas/premis.xsd", specs.PremisXSD); err != nil {
			return errors.Wrapf(err, "cannot write file '%v/%s'", me.fsys, "schemas/premis.xsd")
		}
		if err := writefs.WriteFile(me.fsys, "schemas/xlink.xsd", specs.XLinkXSD); err != nil {
			return errors.Wrapf(err, "cannot write file '%v/%s'", me.fsys, "schemas/xlink.xsd")
		}
	default: // cannot happen here
		return errors.Errorf("unsupported storage type '%s'", me.StorageType)
	}

	return nil
}

func newMDSec(id, groupid, href, loctype, otherloctype, mimetype, created string, size int64, mdType, othermdtype, checksum, checksumType string) *mets.MdSecType {
	if mdType == "" {
		mdType = "OTHER"
	}
	if mdType == "OTHER" && othermdtype == "" {
		othermdtype = "UNKNOWN"
	}
	return &mets.MdSecType{
		IDAttr:      id,
		GROUPIDAttr: groupid,
		CREATEDAttr: time.Now().Format("2006-01-02T15:04:05"),
		MdRef: &mets.MdRef{
			XMLName:          xml.Name{},
			TypeAttr:         "simple",
			XlinkHrefAttr:    href,
			LOCTYPEAttr:      loctype,
			OTHERLOCTYPEAttr: otherloctype,
			MDTYPEAttr:       mdType,
			OTHERMDTYPEAttr:  othermdtype,
			MIMETYPEAttr:     mimetype,
			SIZEAttr:         size,
			CHECKSUMAttr:     checksum,
			CHECKSUMTYPEAttr: checksumTypeToMets(checksumType),
			CREATEDAttr:      created,
		},
		MdWrap: nil,
	}
}

// check interface satisfaction
var (
	_ ocfl.ExtensionObjectChange = &Mets{}
)

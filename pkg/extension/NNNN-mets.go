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
	var extensionMap map[string]any
	extensionMap, _ = metadata.Extension.(map[string]any)
	if extensionMap != nil {
		if contentSubPathAny, ok := extensionMap[ContentSubPathName]; ok {
			contentSubPath, _ = contentSubPathAny.(map[string]ContentSubPathEntry)
		}
	}

	var metsName, premisName string
	var extMetsName, extPremisName string
	var area string
	var metsNames, premisNames *ocfl.NamesStruct
	var internalRelativePath, externalRelativePath, internalRelativePathCurrentVersion string
	switch strings.ToLower(me.StorageType) {
	case "area":
		metsName = me.MetsFile
		premisName = me.PremisFile
		area = me.StorageName
		metsNames, err = object.BuildNames([]string{metsName}, area)
		if err != nil {
			return errors.Wrapf(err, "cannot build names for %s", metsName)
		}
		premisNames, err = object.BuildNames([]string{premisName}, area)
		if err != nil {
			return errors.Wrapf(err, "cannot build names for %s", premisName)
		}
	case "path":
		path, err := object.GetAreaPath("content")
		if err != nil {
			return errors.Wrapf(err, "cannot get area path for '%s'", "content")
		}
		metsName = strings.TrimLeft(filepath.ToSlash(filepath.Join(path, me.StorageName, me.MetsFile)), "/")
		premisName = strings.TrimLeft(filepath.ToSlash(filepath.Join(path, me.StorageName, me.PremisFile)), "/")
		area = ""
		metsNames, err = object.BuildNames([]string{metsName}, area)
		if err != nil {
			return errors.Wrapf(err, "cannot build names for %s", metsName)
		}
		premisNames, err = object.BuildNames([]string{premisName}, area)
		if err != nil {
			return errors.Wrapf(err, "cannot build names for %s", premisName)
		}
	case "extension":
		metsName = strings.TrimLeft(filepath.ToSlash(filepath.Join(me.StorageName, fmt.Sprintf(me.MetsFile, object.GetVersion()))), "/")
		premisName = strings.TrimLeft(filepath.ToSlash(filepath.Join(me.StorageName, fmt.Sprintf(me.PremisFile, object.GetVersion()))), "/")
		metsNames = &ocfl.NamesStruct{
			ExternalPaths: []string{},
			InternalPath:  metsName,
			ManifestPath:  "",
		}
		premisNames = &ocfl.NamesStruct{
			ExternalPaths: []string{},
			InternalPath:  premisName,
			ManifestPath:  "",
		}
	default:
		return errors.Errorf("unsupported storage type '%s'", me.StorageType)
	}
	if len(premisNames.ExternalPaths) > 0 {
		extPremisName = premisNames.ExternalPaths[0]
	}
	if len(metsNames.ExternalPaths) > 1 {
		return errors.Errorf("multiple external paths for mets file not supported - %v", metsNames.ExternalPaths)
	}
	if len(metsNames.ExternalPaths) == 1 {
		extMetsName = metsNames.ExternalPaths[0]
		parts := strings.Split(metsNames.ExternalPaths[0], "/")
		for i := 1; i < len(parts); i++ {
			externalRelativePath += "../"
		}
	}
	parts := strings.Split(metsNames.InternalPath, "/")
	for i := 1; i < len(parts)+2; i++ {
		internalRelativePath += "../"
	}
	for i := 1; i < len(parts); i++ {
		internalRelativePathCurrentVersion += "../"
	}

	metsFiles := map[string][]*mets.FileType{}
	premisFiles := []*premis.File{}
	premisEvents := []*premis.EventComplexType{}
	structMaps := []*mets.StructMapType{}
	dmdSecs := []*mets.MdSecType{}
	fileGrpUUID := map[string]string{}
	//metaFolder, _ := contentSubPath["metadata"]

	//internalPrefix := fmt.Sprintf("%s/content/", head)
	structPhysical := map[string][]string{}
	structSemantical := map[string][]string{}
	internalFiledata := map[string]struct {
		ingestVersion string
		uuid          string
		cs            string
	}{}
	//	ingest := map[string]string{}
	// file section
	if contentSubPath != nil {
		for cName, cse := range contentSubPath {
			structSemantical[cse.Description] = []string{}
			metsFiles[cName] = []*mets.FileType{}
			fileGrpUUID[cName] = uuid.NewString()
			if err != nil {
				return errors.Wrap(err, "cannot create uuid")
			}
		}
	} else {
		metsFiles["content"] = []*mets.FileType{}
		structSemantical["Payload"] = []string{}
		fileGrpUUID["content"] = uuid.NewString()
	}
	metsFiles["schemas"] = []*mets.FileType{}
	fileGrpUUID["schemas"] = uuid.NewString()
	versionStrings := inventory.GetVersionStrings()
	if contentSubPath != nil {
		for area, _ := range contentSubPath {
			structPhysical[area] = []string{}
		}
	} else {
		structPhysical["content"] = []string{}
	}
	structPhysical["schemas"] = []string{}

	// get ingest versions
	for cs, metaFile := range metadata.Files {

		if extNames, ok := metaFile.VersionName[head]; ok {
			for _, extPath := range extNames {

				val := struct {
					ingestVersion string
					uuid          string
					cs            string
				}{
					uuid: "uuid-" + uuid.New().String(),
					cs:   cs,
				}
				stateVersions := maps.Keys(metaFile.VersionName)
				for _, vStr := range versionStrings {
					if slices.Contains(stateVersions, vStr) {
						val.ingestVersion = vStr
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
						return errors.Wrapf(err, "invalid type: %v", ext)
					}
					for _, ver := range versionStrings {
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
					IDAttr:        uuidString,
					SEQAttr:       0,
					OWNERIDAttr:   "",
					ADMIDAttr:     nil,
					DMDIDAttr:     nil,
					GROUPIDAttr:   "",
					USEAttr:       "Datafile",
					BEGINAttr:     "",
					ENDAttr:       "",
					BETYPEAttr:    "",
					FLocat:        []*mets.FLocat{},
					FContent:      nil,
					Stream:        nil,
					TransformFile: nil,
					File:          nil,
				}
				premisFile := &premis.File{
					XMLName:     xml.Name{},
					XSIType:     "file",
					XmlIDAttr:   uuidString,
					VersionAttr: "",
					ObjectIdentifier: []*premis.ObjectIdentifierComplexType{
						&premis.ObjectIdentifierComplexType{
							XMLName:               xml.Name{},
							SimpleLinkAttr:        "",
							ObjectIdentifierType:  premis.NewStringPlusAuthority("uuid", "", "", ""),
							ObjectIdentifierValue: uuidString,
						},
					},
					PreservationLevel:                nil,
					SignificantProperties:            []*premis.SignificantPropertiesComplexType{},
					ObjectCharacteristics:            []*premis.ObjectCharacteristicsComplexType{},
					OriginalName:                     nil,
					Storage:                          []*premis.StorageComplexType{},
					SignatureInformation:             nil,
					Relationship:                     nil,
					LinkingEventIdentifier:           nil,
					LinkingRightsStatementIdentifier: nil,
					ObjectComplexType:                nil,
				}
				var mimeType string
				if ext, ok := metaFile.Extension[IndexerName]; ok {
					extIndexer, ok := ext.(*indexer.ResultV2)
					if !ok {
						return errors.Wrapf(err, "invalid type: %v", ext)
					}
					mimeType = extIndexer.Mimetype
					metsFile.FILECORE.MIMETYPEAttr = mimeType
					objectCharacter := &premis.ObjectCharacteristicsComplexType{
						XMLName:          xml.Name{},
						CompositionLevel: nil,
						Fixity: []*premis.FixityComplexType{
							premis.NewFixityComplexType(string(metadata.DigestAlgorithm), cs, "gocfl "+version.VERSION),
						},
						Size:                           0,
						Format:                         []*premis.FormatComplexType{},
						CreatingApplication:            nil,
						Inhibitors:                     nil,
						ObjectCharacteristicsExtension: nil,
					}
					if extIndexer.Mimetype != "" {
						objectCharacter.Format = append(objectCharacter.Format, &premis.FormatComplexType{
							XMLName: xml.Name{},
							FormatDesignation: &premis.FormatDesignationComplexType{
								XMLName:       xml.Name{},
								FormatName:    premis.NewStringPlusAuthority(extIndexer.Mimetype, "", "", ""),
								FormatVersion: "",
							},
							FormatRegistry: nil,
							FormatNote:     []string{"IANA MIME-type"},
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
									XMLName:           xml.Name{},
									FormatDesignation: nil,
									FormatRegistry:    nil,
									FormatNote:        []string{"siegfried"},
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
											XMLName:       xml.Name{},
											FormatName:    premis.NewStringPlusAuthority(designation, "", "", ""),
											FormatVersion: "",
										}
									}
								}
								if idAny, ok := sfMap["ID"]; ok {
									if id, ok := idAny.(string); ok {
										fct.FormatRegistry = &premis.FormatRegistryComplexType{
											XMLName:            xml.Name{},
											SimpleLinkAttr:     "",
											FormatRegistryName: premis.NewStringPlusAuthority("PRONOM", "", "", ""),
											FormatRegistryKey:  premis.NewStringPlusAuthority(id, "", "", ""),
											FormatRegistryRole: premis.NewStringPlusAuthority(
												"specification",
												"http://id.loc.gov/vocabulary/preservation/formatRegistryRole",
												"",
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
					var intSemantic = "Other Payload"
					if len(parts) > 3 {
						if contentSubPath != nil {
							intArea = parts[2]
							intSemantic = ""
							for area, cse := range contentSubPath {
								if cse.Path == intArea {
									intArea = area
									intSemantic = cse.Description
									isSchema = parts[3] == "schemas"
									break
								}
							}
						}
					}
					if intArea == "metadata" && !isSchema {
						dmdSecs = append(dmdSecs, newMDSec(fmt.Sprintf("dmdSec-int-%s", uuidString), "area-metadata", intPath, "OTHER", "URL:internal", mimeType, 0, "", cs, string(inventory.GetDigestAlgorithm())))
						continue
					}

					if isSchema {
						structPhysical["schemas"] = append(structPhysical["schemas"], uuidString)
					} else {
						structPhysical[intArea] = append(structPhysical[intArea], uuidString)
					}
					if intSemantic != "" {
						structSemantical[intSemantic] = append(structSemantical[intSemantic], uuidString)
					}
					/*
						href := internalRelativePath + intPath
						if strings.HasPrefix(intPath, internalPrefix) {
							href = internalRelativePathCurrentVersion + intPath[len(internalPrefix):]
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
							TypeAttr:         "simple",
							XlinkHrefAttr:    href,
							XlinkRoleAttr:    "",
							XlinkArcroleAttr: "",
							XlinkTitleAttr:   "",
							XlinkShowAttr:    "",
							XlinkActuateAttr: "",
						},
						IDAttr:  "",
						USEAttr: "",
					})
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
					if !slices.Contains([]string{extMetsName, extPremisName}, extPath) {
						dmdSecs = append(dmdSecs, newMDSec(
							fmt.Sprintf("dmdSec-ext-%s", uuidString),
							"area-metadata",
							extPath,
							"URL",
							"",
							mimeType,
							0,
							"",
							cs,
							string(inventory.GetDigestAlgorithm()),
						))
					}
					continue
				}

				metsFile.FLocat = append(metsFile.FLocat, &mets.FLocat{
					LOCATION: &mets.LOCATION{
						LOCTYPEAttr:      "URL",
						OTHERLOCTYPEAttr: "",
					},
					SimpleLink: &mets.SimpleLink{
						//XMLName:          xml.Name{},
						TypeAttr: "simple",
						XlinkHrefAttr:/* externalRelativePath + */ extPath,
						XlinkRoleAttr:    "",
						XlinkArcroleAttr: "",
						XlinkTitleAttr:   "",
						XlinkShowAttr:    "",
						XlinkActuateAttr: "",
					},
					IDAttr:  "",
					USEAttr: "",
				})
				premisFile.Storage = append(premisFile.Storage, &premis.StorageComplexType{
					XMLName: xml.Name{},
					ContentLocation: &premis.ContentLocationComplexType{
						XMLName:             xml.Name{},
						SimpleLinkAttr:      "",
						ContentLocationType: premis.NewStringPlusAuthority("external", "", "", ""),
						ContentLocationValue:/*externalRelativePath + */ extPath,
					},
					StorageMedium: premis.NewStringPlusAuthority("extracted OCFL", "", "", ""),
				})

				//			}
				//		}
				var ingestTime time.Time
				var ingestVersion string
				_ = ingestVersion
				if internal, ok := internalFiledata[extPath]; ok {
					if internal.ingestVersion != "" {
						ingestVersion = internal.ingestVersion
						if versionData, ok := metadata.Versions[internal.ingestVersion]; ok {
							ingestTime = versionData.Created
						}
					}
				}

				eventIngest := &premis.EventComplexType{
					XMLName:     xml.Name{},
					XmlIDAttr:   "",
					VersionAttr: "",
					EventIdentifier: &premis.EventIdentifierComplexType{
						XMLName:              xml.Name{},
						SimpleLinkAttr:       "",
						EventIdentifierType:  premis.NewStringPlusAuthority("local", "", "", ""),
						EventIdentifierValue: "ingest-" + cs,
					},
					EventType:               premis.NewStringPlusAuthority("MIGRATION", "", "", ""),
					EventDateTime:           ingestTime.Format(time.RFC3339),
					EventDetailInformation:  nil,
					EventOutcomeInformation: nil,
					LinkingAgentIdentifier:  nil,
					LinkingObjectIdentifier: nil,
				}
				if migrationAny, ok := metaFile.Extension[MigrationName]; ok {
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
							SimpleLinkAttr:       "",
							EventIdentifierType:  premis.NewStringPlusAuthority("local", "", "", ""),
							EventIdentifierValue: migration.ID,
						},
						EventType:               premis.NewStringPlusAuthority("MIGRATION", "", "", ""),
						EventDateTime:           ingestTime.Format(time.RFC3339),
						EventDetailInformation:  nil,
						EventOutcomeInformation: []*premis.EventOutcomeInformationComplexType{},
						LinkingAgentIdentifier:  nil,
						LinkingObjectIdentifier: []*premis.LinkingObjectIdentifierComplexType{
							&premis.LinkingObjectIdentifierComplexType{
								XMLName:                      xml.Name{},
								LinkObjectXmlIDAttr:          "",
								SimpleLinkAttr:               "",
								LinkingObjectIdentifierType:  premis.NewStringPlusAuthority("uuid", "", "", ""),
								LinkingObjectIdentifierValue: uuidString,
								LinkingObjectRole:            []*premis.StringPlusAuthority{premis.NewStringPlusAuthority("TARGET", "", "", "")},
							},
						},
					}
					if migration.Error == "" {
						eventMigration.EventOutcomeInformation = append(eventMigration.EventOutcomeInformation, &premis.EventOutcomeInformationComplexType{
							XMLName:            xml.Name{},
							EventOutcome:       premis.NewStringPlusAuthority("success", "", "", ""),
							EventOutcomeDetail: nil,
						})
						var sourcePath string
						for extPath, val := range internalFiledata {
							if val.cs == migration.Source {
								sourcePath = extPath
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
										premis.NewStringPlusAuthority("SOURCE", "", "", ""),
									},
								})
							}
						}
					} else {
						eventMigration.EventOutcomeInformation = append(eventMigration.EventOutcomeInformation, &premis.EventOutcomeInformationComplexType{
							XMLName:      xml.Name{},
							EventOutcome: premis.NewStringPlusAuthority("error", "", "", ""),
							EventOutcomeDetail: []*premis.EventOutcomeDetailComplexType{
								&premis.EventOutcomeDetailComplexType{
									XMLName:                     xml.Name{},
									EventOutcomeDetailNote:      migration.Error,
									EventOutcomeDetailExtension: nil,
								},
							},
						})
					}
					premisEvents = append(premisEvents, eventMigration)
				}
				_ = eventIngest
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
			IDAttr:         "uuid-" + structMapPhysicalId.String() + "-structMap-div",
			DMDIDAttr:      nil,
			ADMIDAttr:      nil,
			TYPEAttr:       "",
			CONTENTIDSAttr: nil,
			XlinkLabelAttr: nil,
			Mptr:           nil,
			Fptr:           nil,
			Div:            []*mets.DivType{},
		},
	}
	for area, fileList := range structPhysical {
		/*
			structMapPhysicalDivVer := &mets.DivType{
				XMLName: xml.Name{},
				ORDERLABELS: &mets.ORDERLABELS{
					ORDERAttr:      0,
					ORDERLABELAttr: "",
					LABELAttr:      "Version " + ver,
				},
				Div: make([]*mets.DivType, 0),
			}
		*/
		if len(fileList) == 0 {
			continue
		}

		div := &mets.DivType{
			XMLName: xml.Name{},
			ORDERLABELS: &mets.ORDERLABELS{
				ORDERAttr:      0,
				ORDERLABELAttr: "",
				LABELAttr:      area,
			},
			IDAttr:         "uuid-" + uuid.New().String() + "-structMap-div",
			DMDIDAttr:      nil,
			ADMIDAttr:      nil,
			TYPEAttr:       "",
			CONTENTIDSAttr: nil,
			XlinkLabelAttr: nil,
			Mptr:           nil,
			Fptr:           make([]*mets.Fptr, 0),
			Div:            nil,
		}
		for _, u := range fileList {
			div.Fptr = append(div.Fptr, &mets.Fptr{
				XMLName:        xml.Name{},
				IDAttr:         "",
				FILEIDAttr:     u,
				CONTENTIDSAttr: nil,
				Par:            nil,
				Seq:            nil,
				Area:           nil,
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
		IDAttr:    "",
		TYPEAttr:  "logical",
		LABELAttr: "AIP Structure",
		Div: &mets.DivType{
			XMLName: xml.Name{},
			ORDERLABELS: &mets.ORDERLABELS{
				ORDERAttr:      0,
				ORDERLABELAttr: "",
				LABELAttr:      "Package Structure",
			},
			IDAttr:         "uuid-" + uuid.New().String() + "-structMap-div",
			DMDIDAttr:      nil,
			ADMIDAttr:      nil,
			TYPEAttr:       "",
			CONTENTIDSAttr: nil,
			XlinkLabelAttr: nil,
			Mptr:           nil,
			Fptr:           nil,
			Div:            []*mets.DivType{},
		},
	}
	for area, uuids := range structSemantical {
		if len(uuids) == 0 {
			continue
		}

		div := &mets.DivType{
			XMLName: xml.Name{},
			ORDERLABELS: &mets.ORDERLABELS{
				ORDERAttr:      0,
				ORDERLABELAttr: "",
				LABELAttr:      area,
			},
			IDAttr:         "uuid-" + uuid.New().String() + "-structMap-div",
			DMDIDAttr:      nil,
			ADMIDAttr:      nil,
			TYPEAttr:       "",
			CONTENTIDSAttr: nil,
			XlinkLabelAttr: nil,
			Mptr:           nil,
			Fptr:           make([]*mets.Fptr, 0),
			Div:            nil,
		}
		for _, u := range uuids {
			div.Fptr = append(div.Fptr, &mets.Fptr{
				XMLName:        xml.Name{},
				IDAttr:         "",
				FILEIDAttr:     u,
				CONTENTIDSAttr: nil,
				Par:            nil,
				Seq:            nil,
				Area:           nil,
			})
		}
		structMapSemantical.Div.Div = append(structMapSemantical.Div.Div, div)
	}
	structMaps = append(structMaps, structMapSemantical)

	premisStruct := &premis.PremisComplexType{
		XMLName:           xml.Name{},
		XMLNS:             "http://www.loc.gov/premis/v3",
		XMLXLinkNS:        "http://www.w3.org/1999/xlink",
		XMLNSXSI:          "http://www.w3.org/2001/XMLSchema-instance",
		XSISchemaLocation: "http://www.loc.gov/premis/v3\nschemas/premis.xsd\nhttp://www.w3.org/1999/xlink\nschemas/xlink.xsd",
		VersionAttr:       "3.0",
		Object:            premisFiles,
		Event:             premisEvents,
		Agent:             []*premis.AgentComplexType{},
		Rights:            []*premis.RightsComplexType{},
	}

	premisBytes, err := xml.MarshalIndent(premisStruct, "", "  ")
	if err != nil {
		return errors.Wrap(err, "cannot marshal PREMIS")
	}

	premisChecksum := fmt.Sprintf("%x", sha512.Sum512(premisBytes))

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
		var mimetype string
		switch strings.ToUpper(metaType) {
		case "MARC":
			mimetype = "application/marc"
		case "MARCXML":
			mimetype = "application/marcxml+xml"
		case "JSON":
			mimetype = "text/json"
		case "XML":
			mimetype = "text/xml"
		default:
			mimetype = "application/octet-stream"
		}

		// remove any existing mdSecs with the same checksum
		// todo: do it for internal and external name separately
		mdSecs2 := dmdSecs
		dmdSecs = make([]*mets.MdSecType, 0, len(mdSecs2))
		for i := 0; i < len(mdSecs2); i++ {
			if mdSecs2[i].MdRef.CHECKSUMAttr != foundChecksum {
				dmdSecs = append(dmdSecs, mdSecs2[i])
			}
		}
		mdSecs2 = nil

		if len(found.InternalName) > 0 {
			dmdSecs = append(dmdSecs, newMDSec(
				fmt.Sprintf("dmdSec-int-%s-%s", slug.Make(object.GetID()), head),
				"primary-metadata",
				found.InternalName[0],
				"OTHER",
				"URL:internal",
				mimetype,
				0,
				"",
				foundChecksum,
				string(inventory.GetDigestAlgorithm()),
			))
		}
		if len(found.VersionName[head]) > 0 {
			dmdSecs = append(dmdSecs, newMDSec(
				fmt.Sprintf("dmdSec-ext-%s-%s", slug.Make(object.GetID()), head),
				"primary-metadata",
				found.VersionName[head][0],
				"URL",
				"",
				mimetype,
				0,
				"",
				foundChecksum,
				string(inventory.GetDigestAlgorithm()),
			))
		}
	}

	metsFileGrps := []*mets.FileGrp{}
	for a, files := range metsFiles {
		if len(files) == 0 {
			continue
		}
		metsFileGrps = append(metsFileGrps, &mets.FileGrp{
			XMLName: xml.Name{},
			FileGrpType: &mets.FileGrpType{
				XMLName:      xml.Name{},
				IDAttr:       "uuid-" + fileGrpUUID[a],
				VERSDATEAttr: "",
				ADMIDAttr:    nil,
				USEAttr:      a,
				FileGrp:      nil,
				File:         files,
			},
		})
	}

	var amdSecs = []*mets.AmdSecType{}
	if premisNames != nil {
		sec := &mets.AmdSecType{
			XMLName:  xml.Name{},
			IDAttr:   "uuid-" + uuid.NewString(),
			TechMD:   nil,
			RightsMD: nil,
			SourceMD: nil,
			DigiprovMD: []*mets.MdSecType{
				newMDSec("uuid-"+uuid.NewString(), "", premisNames.InternalPath, "OTHER", "URL:internal", "application/xml", int64(len(premisBytes)), "PREMIS", premisChecksum, "SHA-512"),
			},
		}
		for _, ext := range premisNames.ExternalPaths {
			sec.DigiprovMD = append(sec.DigiprovMD, newMDSec("uuid-"+uuid.NewString(), "", ext, "URL", "", "application/xml", int64(len(premisBytes)), "PREMIS", premisChecksum, "SHA-512"))
		}
		amdSecs = append(amdSecs, sec)
	}

	m := &mets.Mets{
		XMLNS:             "http://www.loc.gov/METS/",
		XMLXLinkNS:        "http://www.w3.org/1999/xlink",
		XMLNSXSI:          "http://www.w3.org/2001/XMLSchema-instance",
		XSISchemaLocation: "http://www.loc.gov/METS/\nschemas/mets.xsd\nhttp://www.w3.org/1999/xlink\nschemas/xlink.xsd",
		MetsType: &mets.MetsType{
			XMLName:     xml.Name{},
			IDAttr:      "",
			OBJIDAttr:   metadata.ID,
			LABELAttr:   fmt.Sprintf("METS Container for Object %s version %s - %s", metadata.ID, head, v.Message),
			TYPEAttr:    "AIP",
			PROFILEAttr: "http://www.ra.ee/METS/v01/IP.xml",
			MetsHdr: &mets.MetsHdr{
				XMLName:          xml.Name{},
				IDAttr:           "",
				ADMIDAttr:        nil,
				CREATEDATEAttr:   v.Created.Format("2006-01-02T15:04:05"),
				LASTMODDATEAttr:  "",
				RECORDSTATUSAttr: "NEW",
				Agent: []*mets.Agent{
					&mets.Agent{
						XMLName:       xml.Name{},
						IDAttr:        "",
						ROLEAttr:      "CREATOR",
						OTHERROLEAttr: "",
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
						XMLName:       xml.Name{},
						IDAttr:        "",
						ROLEAttr:      "ARCHIVIST",
						OTHERROLEAttr: "",
						TYPEAttr:      "",
						OTHERTYPEAttr: "",
						Name:          v.User.Name.String(),
						Note: []*mets.Note{
							&mets.Note{
								XMLName: xml.Name{},
								Value:   v.User.Address.String(),
							},
						},
					},
				},
				AltRecordID: nil,
				MetsDocumentID: &mets.MetsDocumentID{
					XMLName:  xml.Name{},
					IDAttr:   "",
					TYPEAttr: "",
					Value:    "mets.xml",
				},
			},
			DmdSec: dmdSecs,
			AmdSec: amdSecs,
			FileSec: &mets.FileSec{
				XMLName: xml.Name{},
				IDAttr:  "uuid-" + uuid.NewString(),
				FileGrp: metsFileGrps,
			},
			StructMap:   structMaps,
			StructLink:  nil,
			BehaviorSec: nil,
		}}

	metsBytes, err := xml.MarshalIndent(m, "", "  ")
	if err != nil {
		return errors.Wrap(err, "cannot marshal METS")
	}

	switch strings.ToLower(me.StorageType) {
	case "area", "path":

		//		if err := object.AddReader(io.NopCloser(bytes.NewBuffer(metsBytes)), []string{metsName}, area, true, false); err != nil {
		if err := object.AddData(metsBytes, metsName, false, area, true, false); err != nil {
			return errors.Wrapf(err, "cannot write '%s'", metsName)
		}
		if err := object.AddReader(io.NopCloser(bytes.NewBuffer(premisBytes)), []string{premisName}, area, true, false); err != nil {
			return errors.Wrapf(err, "cannot write '%s'", premisName)
		}
		//		if err := object.AddReader(io.NopCloser(bytes.NewBuffer(specs.METSXSD)), []string{"schemas/mets.xsd"}, area, true, false); err != nil {
		if err := object.AddData(specs.METSXSD, "schemas/mets.xsd", true, area, true, false); err != nil {
			return errors.Wrapf(err, "cannot write '%s'", "schemas/mets.xsd")
		}
		//		if err := object.AddReader(io.NopCloser(bytes.NewBuffer(specs.PremisXSD)), []string{"schemas/premis.xsd"}, area, true, false); err != nil {
		if err := object.AddData(specs.PremisXSD, "schemas/premis.xsd", true, area, true, false); err != nil {
			return errors.Wrapf(err, "cannot write '%s'", "schemas/premis.xsd")
		}
		//if err := object.AddReader(io.NopCloser(bytes.NewBuffer(specs.XLinkXSD)), []string{"schemas/xlink.xsd"}, area, true, false); err != nil {
		if err := object.AddData(specs.XLinkXSD, "schemas/xlink.xsd", true, area, true, false); err != nil {
			return errors.Wrapf(err, "cannot write '%s'", "schemas/xlink.xsd")
		}
	case "extension":
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

func newMDSec(id, groupid, href, loctype, otherloctype, mimetype string, size int64, mdType, checksum, checksumType string) *mets.MdSecType {
	if mdType == "" {
		mdType = "OTHER"
	}
	return &mets.MdSecType{
		IDAttr:      id,
		GROUPIDAttr: groupid,
		ADMIDAttr:   nil,
		CREATEDAttr: "",
		STATUSAttr:  "",
		MdRef: &mets.MdRef{
			XMLName:           xml.Name{},
			IDAttr:            "",
			LABELAttr:         "",
			XPTRAttr:          "",
			TypeAttr:          "simple",
			XlinkHrefAttr:     href,
			XlinkRoleAttr:     "",
			XlinkArcroleAttr:  "",
			XlinkTitleAttr:    "",
			XlinkShowAttr:     "",
			XlinkActuateAttr:  "",
			LOCTYPEAttr:       loctype,
			OTHERLOCTYPEAttr:  otherloctype,
			MDTYPEAttr:        mdType,
			OTHERMDTYPEAttr:   "",
			MDTYPEVERSIONAttr: "",
			MIMETYPEAttr:      mimetype,
			SIZEAttr:          size,
			CREATEDAttr:       "",
			CHECKSUMAttr:      checksum,
			CHECKSUMTYPEAttr:  checksumTypeToMets(checksumType),
		},
		MdWrap: nil,
	}
}

// check interface satisfaction
var (
	_ ocfl.ExtensionObjectChange = &Mets{}
)

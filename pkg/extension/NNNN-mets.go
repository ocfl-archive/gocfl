package extension

import (
	"bytes"
	"crypto/sha512"
	"emperror.dev/errors"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/google/uuid"
	"github.com/je4/filesystem/v2/pkg/writefs"
	"github.com/je4/gocfl/v2/pkg/dilcis/mets"
	"github.com/je4/gocfl/v2/pkg/dilcis/premis"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	"github.com/je4/gocfl/v2/version"
	"github.com/je4/indexer/v2/pkg/indexer"
	"github.com/op/go-logging"
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
			Param:         "source",
			//File:          "Source",
			Description: "url with metadata file. $ID will be replaced with object ID i.e. file:///c:/temp/$ID.json",
		},
	}
}

func NewMetsFS(fsys fs.FS, logger *logging.Logger) (*Mets, error) {
	data, err := fs.ReadFile(fsys, "config.json")
	if err != nil {
		return nil, errors.Wrap(err, "cannot read config.json")
	}

	var config = &MetsConfig{}
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
	StorageType string `json:"storageType"`
	StorageName string `json:"storageName"`
}
type Mets struct {
	*MetsConfig
	fsys   fs.FS
	logger *logging.Logger
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

	fileGrpId, err := uuid.NewUUID()
	if err != nil {
		return errors.Wrap(err, "cannot create uuid")
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
	var area string
	var names *ocfl.NamesStruct
	var internalRelativePath, externalRelativePath, internalRelativePathCurrentVersion string
	switch strings.ToLower(me.StorageType) {
	case "area":
		metsName = "mets.xml"
		premisName = "premis.xml"
		area = me.StorageName
		names, err = object.BuildNames([]string{metsName}, area)
		if err != nil {
			return errors.Wrapf(err, "cannot build names for %s", metsName)
		}
	case "path":
		path, err := object.GetAreaPath("content")
		if err != nil {
			return errors.Wrapf(err, "cannot get area path for '%s'", "content")
		}
		metsName = strings.TrimLeft(filepath.ToSlash(filepath.Join(path, me.StorageName, "mets.xml")), "/")
		premisName = strings.TrimLeft(filepath.ToSlash(filepath.Join(path, me.StorageName, "premis.xml")), "/")
		area = ""
		names, err = object.BuildNames([]string{metsName}, area)
		if err != nil {
			return errors.Wrapf(err, "cannot build names for %s", metsName)
		}
	case "extension":
		metsName = strings.TrimLeft(filepath.ToSlash(filepath.Join(me.StorageName, fmt.Sprintf("mets-%s.xml", object.GetVersion()))), "/")
		premisName = strings.TrimLeft(filepath.ToSlash(filepath.Join(me.StorageName, fmt.Sprintf("premis-%s.xml", object.GetVersion()))), "/")
		names = &ocfl.NamesStruct{
			ExternalPaths: []string{},
			InternalPath:  metsName,
			ManifestPath:  "",
		}
	default:
		return errors.Errorf("unsupported storage type '%s'", me.StorageType)
	}
	if len(names.ExternalPaths) > 1 {
		return errors.Errorf("multiple external paths for mets file not supported - %v", names.ExternalPaths)
	}
	if len(names.ExternalPaths) == 1 {
		parts := strings.Split(names.ExternalPaths[0], "/")
		for i := 1; i < len(parts); i++ {
			externalRelativePath += "../"
		}
	}
	parts := strings.Split(names.InternalPath, "/")
	for i := 1; i < len(parts)+2; i++ {
		internalRelativePath += "../"
	}
	for i := 1; i < len(parts); i++ {
		internalRelativePathCurrentVersion += "../"
	}

	metsFiles := []*mets.FileType{}
	premisFiles := []*premis.File{}
	structMaps := []*mets.StructMapType{}
	internalPrefix := fmt.Sprintf("%s/content/", head)
	structPhysical := map[string]map[string][]string{}
	structSemantical := map[string][]string{}
	// file section
	if contentSubPath != nil {
		for _, cse := range contentSubPath {
			structSemantical[cse.Description] = []string{}
		}
	} else {
		structSemantical["Payload"] = []string{}
	}
	for _, v := range inventory.GetVersionStrings() {
		structPhysical[v] = map[string][]string{}
		if contentSubPath != nil {
			for area, _ := range contentSubPath {
				structPhysical[v][area] = []string{}
			}
		} else {
			structPhysical[v]["content"] = []string{}
		}
	}

	for cs, metaFile := range metadata.Files {
		uuid, err := uuid.NewUUID()
		if err != nil {
			return errors.Wrap(err, "cannot create uuid")
		}
		uuidString := "uuid-" + uuid.String()
		var size int64
		var creationString string
		//		var fLocat = []*mets.FLocat{}
		if ext, ok := metaFile.Extension[FilesystemName]; ok {
			extFSL, ok := ext.(map[string][]*FileSystemLine)
			if !ok {
				return errors.Wrapf(err, "invalid type: %v", ext)
			}
			if fslHead, ok := extFSL[head]; ok {
				if len(fslHead) > 0 {
					creationString = fslHead[0].Meta.CTime.Format("2006-01-02T15:04:05")
					size = int64(fslHead[0].Meta.Size)
				}
			}

		}
		metsFile := &mets.FileType{
			XMLName: xml.Name{},
			FILECORE: &mets.FILECORE{
				MIMETYPEAttr:     "",
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
			XSIType:     "premis:file",
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
			intVer := parts[0]
			if parts[1] != "content" {
				return errors.Wrapf(err, "no content in %s", intPath)
			}
			var intArea = "content"
			var intSemantic = "Other Payload"
			if len(parts) > 3 {
				if contentSubPath != nil {
					intArea = parts[2]
					intSemantic = ""
					for area, cse := range contentSubPath {
						if cse.Path == intArea {
							intArea = area
							intSemantic = cse.Description
							break
						}
					}
				}
			}
			structPhysical[intVer][intArea] = append(structPhysical[intVer][intArea], "uuid-"+uuid.String())
			if intSemantic != "" {
				structSemantical[intSemantic] = append(structSemantical[intSemantic], "uuid-"+uuid.String())
			}
			href := internalRelativePath + intPath
			if strings.HasPrefix(intPath, internalPrefix) {
				href = internalRelativePathCurrentVersion + intPath[len(internalPrefix):]
			}
			metsFile.FLocat = append(metsFile.FLocat, &mets.FLocat{
				LOCATION: &mets.LOCATION{
					LOCTYPEAttr:      "URL",
					OTHERLOCTYPEAttr: "internal",
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
					ContentLocationType:  premis.NewStringPlusAuthority("OCFL internal", "", "", ""),
					ContentLocationValue: href,
				},
				StorageMedium: premis.NewStringPlusAuthority("OCFL Object Root", "", "", ""),
			})
		}
		if extNames, ok := metaFile.VersionName[head]; ok {
			for _, extPath := range extNames {
				metsFile.FLocat = append(metsFile.FLocat, &mets.FLocat{
					LOCATION: &mets.LOCATION{
						LOCTYPEAttr:      "URL",
						OTHERLOCTYPEAttr: "external",
					},
					SimpleLink: &mets.SimpleLink{
						//XMLName:          xml.Name{},
						TypeAttr:         "simple",
						XlinkHrefAttr:    externalRelativePath + extPath,
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
						ContentLocationType:  premis.NewStringPlusAuthority("origin", "", "", ""),
						ContentLocationValue: externalRelativePath + extPath,
					},
					StorageMedium: premis.NewStringPlusAuthority("extracted OCFL", "", "", ""),
				})

			}
		}
		metsFiles = append(metsFiles, metsFile)
		premisFiles = append(premisFiles, premisFile)
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
			IDAttr:         "",
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
	for ver, areaList := range structPhysical {
		structMapPhysicalDivVer := &mets.DivType{
			XMLName: xml.Name{},
			ORDERLABELS: &mets.ORDERLABELS{
				ORDERAttr:      0,
				ORDERLABELAttr: "",
				LABELAttr:      "Version " + ver,
			},
			Div: make([]*mets.DivType, 0),
		}

		for area, uuids := range areaList {
			div := &mets.DivType{
				XMLName: xml.Name{},
				ORDERLABELS: &mets.ORDERLABELS{
					ORDERAttr:      0,
					ORDERLABELAttr: "",
					LABELAttr:      area,
				},
				IDAttr:         "",
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
			structMapPhysicalDivVer.Div = append(structMapPhysicalDivVer.Div, div)
		}

		structMapPhysical.Div.Div = append(structMapPhysical.Div.Div, structMapPhysicalDivVer)
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
			IDAttr:         "",
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
		div := &mets.DivType{
			XMLName: xml.Name{},
			ORDERLABELS: &mets.ORDERLABELS{
				ORDERAttr:      0,
				ORDERLABELAttr: "",
				LABELAttr:      area,
			},
			IDAttr:         "",
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
		XMLName:     xml.Name{},
		VersionAttr: "",
		Object:      premisFiles,
		Event:       []*premis.EventComplexType{},
		Agent:       []*premis.AgentComplexType{},
		Rights:      []*premis.RightsComplexType{},
	}

	premisBytes, err := xml.MarshalIndent(premisStruct, "", "  ")
	if err != nil {
		return errors.Wrap(err, "cannot marshal PREMIS")
	}

	premisChecksum := fmt.Sprintf("%x", sha512.Sum512(premisBytes))

	m := &mets.Mets{
		XMLNS:             "http://www.loc.gov/METS/",
		XMLXLinkNS:        "http://www.w3.org/1999/xlink",
		XMLNSXSI:          "http://www.w3.org/2001/XMLSchema-instance",
		XSISchemaLocation: "http://www.loc.gov/METS/\nhttps://www.loc.gov/standards/mets/mets.xsd\nhttp://www.w3.org/1999/xlink\nhttps://www.w3.org/XML/2008/06/xlink.xsd",
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
			DmdSec: nil,
			AmdSec: []*mets.AmdSecType{
				&mets.AmdSecType{
					XMLName:  xml.Name{},
					IDAttr:   "uuid-" + uuid.NewString(),
					TechMD:   nil,
					RightsMD: nil,
					SourceMD: nil,
					DigiprovMD: []*mets.MdSecType{
						&mets.MdSecType{
							//XMLName:     xml.Name{},
							IDAttr:      "uuid-" + uuid.NewString(),
							GROUPIDAttr: "",
							ADMIDAttr:   nil,
							CREATEDAttr: "",
							STATUSAttr:  "",
							MdRef: &mets.MdRef{
								XMLName:          xml.Name{},
								LOCTYPEAttr:      "URL",
								OTHERLOCTYPEAttr: "internal",
								TypeAttr:         "",
								XlinkHrefAttr:    "premis.xml",
								XlinkRoleAttr:    "",
								XlinkArcroleAttr: "",
								XlinkTitleAttr:   "",
								XlinkShowAttr:    "",
								XlinkActuateAttr: "",
								MDTYPEAttr:       "PREMIS",
								IDAttr:           "",
								LABELAttr:        "",
								XPTRAttr:         "",
								MIMETYPEAttr:     "application/xml",
								SIZEAttr:         int64(len(premisBytes)),
								CREATEDAttr:      time.Now().Format("2006-01-02T15:04:05"),
								CHECKSUMAttr:     premisChecksum,
								CHECKSUMTYPEAttr: "SHA-512",
							},
							MdWrap: nil,
						},
					},
				},
			},
			FileSec: &mets.FileSec{
				XMLName: xml.Name{},
				IDAttr:  "",
				FileGrp: []*mets.FileGrp{
					&mets.FileGrp{
						XMLName: xml.Name{},
						FileGrpType: &mets.FileGrpType{
							XMLName:      xml.Name{},
							IDAttr:       "uuid-" + fileGrpId.String(),
							VERSDATEAttr: "",
							ADMIDAttr:    nil,
							USEAttr:      fmt.Sprintf("Version %s", head),
							FileGrp:      nil,
							File:         metsFiles,
						},
					},
				},
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
		if err := object.AddReader(io.NopCloser(bytes.NewBuffer(metsBytes)), []string{metsName}, area, true, false); err != nil {
			return errors.Wrapf(err, "cannot write '%s'", metsName)
		}
		if err := object.AddReader(io.NopCloser(bytes.NewBuffer(premisBytes)), []string{premisName}, area, true, false); err != nil {
			return errors.Wrapf(err, "cannot write '%s'", premisName)
		}
	case "extension":
		if err := writefs.WriteFile(me.fsys, metsName, metsBytes); err != nil {
			return errors.Wrapf(err, "cannot write file '%v/%s'", me.fsys, metsName)
		}
		if err := writefs.WriteFile(me.fsys, premisName, metsBytes); err != nil {
			return errors.Wrapf(err, "cannot write file '%v/%s'", me.fsys, premisName)
		}
	default: // cannot happen here
		return errors.Errorf("unsupported storage type '%s'", me.StorageType)
	}

	return nil
}

// check interface satisfaction
var (
	_ ocfl.ExtensionObjectChange = &Mets{}
)

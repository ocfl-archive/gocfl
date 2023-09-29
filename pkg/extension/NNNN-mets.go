package extension

import (
	"bytes"
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

	files := []*mets.FileType{}
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

	premisStruct := &premis.PremisComplexType{
		XMLName:     xml.Name{},
		VersionAttr: "",
		Object:      []*premis.ObjectComplexType{},
		Event:       []*premis.EventComplexType{},
		Agent:       []*premis.AgentComplexType{},
		Rights:      []*premis.RightsComplexType{},
	}
	_ = premisStruct

	for cs, metaFile := range metadata.Files {
		uuid, err := uuid.NewUUID()
		if err != nil {
			return errors.Wrap(err, "cannot create uuid")
		}
		var size int64
		var creation time.Time
		var mime string
		var fLocat = []*mets.FLocat{}
		if ext, ok := metaFile.Extension[FilesystemName]; ok {
			extFSL, ok := ext.(map[string][]*FileSystemLine)
			if !ok {
				return errors.Wrapf(err, "invalid type: %v", ext)
			}
			if fslHead, ok := extFSL[head]; ok {
				if len(fslHead) > 0 {
					creation = fslHead[0].Meta.CTime
					size = int64(fslHead[0].Meta.Size)
				}
			}

		}
		if ext, ok := metaFile.Extension[IndexerName]; ok {
			extIndexer, ok := ext.(*indexer.ResultV2)
			if !ok {
				return errors.Wrapf(err, "invalid type: %v", ext)
			}
			mime = extIndexer.Mimetype
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
			var intSemantic = "Payload"
			if len(parts) >= 3 {
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
			href := "../../../" + intPath
			if strings.HasPrefix(intPath, internalPrefix) {
				href = "../" + intPath[len(internalPrefix):]
			}
			fLocat = append(fLocat, &mets.FLocat{
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

		}
		if extNames, ok := metaFile.VersionName[head]; ok {
			for _, extPath := range extNames {
				fLocat = append(fLocat, &mets.FLocat{
					LOCATION: &mets.LOCATION{
						LOCTYPEAttr:      "URL",
						OTHERLOCTYPEAttr: "external",
					},
					SimpleLink: &mets.SimpleLink{
						//XMLName:          xml.Name{},
						TypeAttr:         "simple",
						XlinkHrefAttr:    "../" + extPath,
						XlinkRoleAttr:    "",
						XlinkArcroleAttr: "",
						XlinkTitleAttr:   "",
						XlinkShowAttr:    "",
						XlinkActuateAttr: "",
					},
					IDAttr:  "",
					USEAttr: "",
				})
			}
		}

		file := &mets.FileType{
			XMLName: xml.Name{},
			FILECORE: &mets.FILECORE{
				MIMETYPEAttr:     mime,
				SIZEAttr:         size,
				CREATEDAttr:      creation.Format("2006-01-02T15:04:05"),
				CHECKSUMAttr:     cs,
				CHECKSUMTYPEAttr: "SHA-512",
			},
			IDAttr:        "uuid-" + uuid.String(),
			SEQAttr:       0,
			OWNERIDAttr:   "",
			ADMIDAttr:     nil,
			DMDIDAttr:     nil,
			GROUPIDAttr:   "",
			USEAttr:       "Datafile",
			BEGINAttr:     "",
			ENDAttr:       "",
			BETYPEAttr:    "",
			FLocat:        fLocat,
			FContent:      nil,
			Stream:        nil,
			TransformFile: nil,
			File:          nil,
		}
		files = append(files, file)
	}

	for ver, areaList := range structPhysical {
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
					LABELAttr:      "Version " + ver,
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
			structMapPhysical.Div.Div = append(structMapPhysical.Div.Div, div)
		}
		structMaps = append(structMaps, structMapPhysical)
	}

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
			AmdSec: nil,
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
							File:         files,
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
	case "area":
		targetname := "mets.xml"
		if err := object.AddReader(io.NopCloser(bytes.NewBuffer(metsBytes)), []string{targetname}, me.StorageName, true, false); err != nil {
			return errors.Wrapf(err, "cannot write '%s'", targetname)
		}
	case "path":
		path, err := object.GetAreaPath("content")
		if err != nil {
			return errors.Wrapf(err, "cannot get area path for '%s'", "content")
		}
		targetname := strings.TrimLeft(filepath.ToSlash(filepath.Join(path, me.StorageName, "mets.xml")), "/")

		//targetname := fmt.Sprintf("%s/%s_%s.jsonl%s", name, storageName, head, ext)
		if err := object.AddReader(io.NopCloser(bytes.NewBuffer(metsBytes)), []string{targetname}, "", true, false); err != nil {
			return errors.Wrapf(err, "cannot write '%s'", targetname)
		}
	case "extension":
		targetname := strings.TrimLeft(filepath.ToSlash(filepath.Join(me.StorageName, "mets.xml")), "/")
		if err := writefs.WriteFile(me.fsys, targetname, metsBytes); err != nil {
			return errors.Wrapf(err, "cannot write file '%v/%s'", me.fsys, targetname)
		}
	default:
		return errors.Errorf("unsupported storage type '%s'", me.StorageType)
	}

	return nil
}

// check interface satisfaction
var (
	_ ocfl.ExtensionObjectChange = &Mets{}
)

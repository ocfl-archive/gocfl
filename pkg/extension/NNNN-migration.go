package extension

import (
	"bytes"
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"github.com/andybalholm/brotli"
	"github.com/je4/gocfl/v2/pkg/migration"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	"io"
	"regexp"
)

const MigrationName = "NNNN-migration"
const MigrationDescription = "preservation management - file migration"

type MigrationConfig struct {
	*ocfl.ExtensionConfig
	StorageType string
	StorageName string
	Compress    string
}

type MigrationTarget struct {
	Name            string
	Strategy        string        // add or replace
	FilenameRegexp  regexp.Regexp // target expression
	FilenameReplace string        // replacement string
	Command         string        // command to execute (stdin --> stdout)
}

// map pronom to migration
type MigrationMap map[string]*MigrationTarget

// map checksum to migration
type MigrationFiles map[string]*MigrationTarget

type Migration struct {
	*MigrationConfig
	fs             ocfl.OCFLFSRead
	lastHead       string
	migration      *migration.Migration
	buffer         *bytes.Buffer
	writer         *brotli.Writer
	migrationFiles map[string]*migration.Function
}

func NewMigrationFS(fs ocfl.OCFLFSRead, migration *migration.Migration) (*Migration, error) {
	fp, err := fs.Open("config.json")
	if err != nil {
		return nil, errors.Wrap(err, "cannot open config.json")
	}
	defer fp.Close()
	data, err := io.ReadAll(fp)
	if err != nil {
		return nil, errors.Wrap(err, "cannot read config.json")
	}

	var config = &MigrationConfig{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal DirectCleanConfig '%s'", string(data))
	}
	ext, err := NewMigration(config, migration)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create new indexer")
	}
	return ext, nil
}
func NewMigration(config *MigrationConfig, migration *migration.Migration) (*Migration, error) {
	sl := &Migration{
		MigrationConfig: config,
		migration:       migration,
		buffer:          bytes.NewBuffer(nil),
	}
	sl.writer = brotli.NewWriter(sl.buffer)
	if config.ExtensionName != sl.GetName() {
		return nil, errors.New(fmt.Sprintf("invalid extension name'%s'for extension %s", config.ExtensionName, sl.GetName()))
	}
	return sl, nil
}

func (mi *Migration) GetConfigString() string {
	str, _ := json.MarshalIndent(mi.MigrationConfig, "", "  ")
	return string(str)
}

func (mi *Migration) IsRegistered() bool { return false }

func (mi *Migration) GetName() string { return MigrationName }

func (mi *Migration) SetFS(fs ocfl.OCFLFSRead) { mi.fs = fs }

func (mi *Migration) SetParams(params map[string]string) error {
	return nil
}

func (mi *Migration) WriteConfig() error {
	if mi.fs == nil {
		return errors.New("no filesystem set")
	}
	fsRW, ok := mi.fs.(ocfl.OCFLFS)
	if !ok {
		return errors.Errorf("filesystem is read only - '%s'", mi.fs.String())
	}

	configWriter, err := fsRW.Create("config.json")
	if err != nil {
		return errors.Wrap(err, "cannot open config.json")
	}
	defer configWriter.Close()
	jenc := json.NewEncoder(configWriter)
	jenc.SetIndent("", "   ")
	if err := jenc.Encode(mi.MigrationConfig); err != nil {
		return errors.Wrapf(err, "cannot encode config to file")
	}
	return nil
}

func (mi *Migration) UpdateObjectBefore(object ocfl.Object) error {
	return nil
}

func (mi *Migration) UpdateObjectAfter(object ocfl.Object) error {
	// first get the metadata from the object
	meta, err := object.GetMetadata()
	if err != nil {
		return errors.Wrapf(err, "cannot get metadata from object %s", object.GetID())
	}
	for cs, m := range meta.Files {
		indexerMetaAny, ok := m.Extension["indexer"]
		if !ok {
			continue
		}
		indexerMeta, ok := indexerMetaAny.(map[string]interface{})
		if !ok {
			continue
		}
		pronomAny, ok := indexerMeta["pronom"]
		if !ok {
			continue
		}
		pronom, ok := pronomAny.(string)
		if !ok {
			continue
		}
		migration, err := mi.migration.GetFunctionByPronom(pronom)
		if err != nil {
			continue
			//return errors.Wrapf(err, "cannot get migration function for pronom %s", pronom)
		}
		mi.migrationFiles[cs] = migration
	}
	inventory := object.GetInventory()
	if inventory == nil {
		return errors.Errorf("inventory is nil")
	}
	mi.lastHead = inventory.GetHead()
	return nil
}

var (
	_ ocfl.Extension             = &Migration{}
	_ ocfl.ExtensionObjectChange = &Migration{}
)

package extension

import (
	"emperror.dev/errors"
	"encoding/json"
	"github.com/je4/gocfl/v2/pkg/ocfl"
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
	Name         string
	CreationType string        // add or replace
	TargetName   regexp.Regexp // target expression
	Command      string        // command to execute (stdin --> stdout)
}

// map pronom to migration
type MigrationMap map[string]*MigrationTarget

// map checksum to migration
type MigrationFiles map[string]*MigrationTarget

type Migration struct {
	*MigrationConfig
	fs             ocfl.OCFLFSRead
	migrationMap   MigrationMap
	migrationFiles MigrationFiles
	lastHead       string
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
		migration, ok := mi.migrationMap[pronom]
		if !ok {
			continue
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

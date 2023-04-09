package extension

import (
	"bytes"
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"github.com/andybalholm/brotli"
	"github.com/je4/gocfl/v2/pkg/migration"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	"github.com/je4/indexer/v2/pkg/indexer"
	"io"
	"os"
	"path/filepath"
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
	migratedFiles  map[string]map[string]string
	sourceFS       ocfl.OCFLFSRead
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
func NewMigration(config *MigrationConfig, mig *migration.Migration) (*Migration, error) {
	sl := &Migration{
		MigrationConfig: config,
		migration:       mig,
		buffer:          bytes.NewBuffer(nil),
		migrationFiles:  map[string]*migration.Function{},
		migratedFiles:   map[string]map[string]string{},
	}
	sl.writer = brotli.NewWriter(sl.buffer)
	if config.ExtensionName != sl.GetName() {
		return nil, errors.New(fmt.Sprintf("invalid extension name'%s'for extension %s", config.ExtensionName, sl.GetName()))
	}
	if mig != nil {
		sl.sourceFS = mig.SourceFS
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

func (mi *Migration) alreadyMigrated(cs string) bool {
	for _, mf := range mi.migratedFiles {
		if _, ok := mf[cs]; ok {
			return true
		}
	}
	return false
}

func (mi *Migration) UpdateObjectAfter(object ocfl.Object) error {
	// first get the metadata from the object
	meta, err := object.GetMetadata()
	if err != nil {
		return errors.Wrapf(err, "cannot get metadata from object %s", object.GetID())
	}
	for cs, m := range meta.Files {
		indexerMetaAny, ok := m.Extension[IndexerName]
		if !ok {
			continue
		}
		indexerMeta, ok := indexerMetaAny.(*indexer.ResultV2)
		if !ok {
			continue
		}
		migration, err := mi.migration.GetFunctionByPronom(indexerMeta.Pronom)
		if err != nil {
			continue
			//return errors.Wrapf(err, "cannot get migration function for pronom %s", pronom)
		}
		if mi.alreadyMigrated(cs) {
			continue
		}
		if migrationMetaAny, ok := m.Extension[MigrationName]; ok {
			if migrationMetaMap, ok := migrationMetaAny.(map[string]any); ok {
				if _, ok := migrationMetaMap[cs]; ok {
					continue
				}
			}
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

func (mi *Migration) NeedNewVersion(object ocfl.Object) (bool, error) {
	return len(mi.migrationFiles) > 0, nil
}

func (mi *Migration) DoNewVersion(object ocfl.Object) error {
	var err error
	inventory := object.GetInventory()
	//files := inventory.GetFiles()
	head := inventory.GetHead()
	manifest := inventory.GetManifest()
	if _, ok := mi.migratedFiles[head]; !ok {
		mi.migratedFiles[head] = map[string]string{}
	}
	for cs, mig := range mi.migrationFiles {
		// todo: do it more efficient
		var found = false
		for _, mf := range mi.migratedFiles {
			if _, ok := mf[cs]; ok {
				found = true
			}
		}
		if found {
			continue
		}

		var targetNames = []string{}

		manifestFiles, ok := manifest[cs]
		if !ok {
			return errors.Errorf("cannot find file with checksum '%s' in object '%s'", cs, object.GetID())
		}
		for _, f := range manifestFiles {
			targetNames = append(targetNames, mig.GetDestinationName(f))
		}

		var file io.ReadCloser
		fs := object.GetFS()
		if fs != nil {
			file, err = fs.Open(manifestFiles[0])
			if err != nil {
				file = nil
			}
		}
		if file == nil {
			if mi.sourceFS != nil {
				stateFiles, err := inventory.GetStateFiles("", cs)
				if err != nil {
					return errors.Wrapf(err, "cannot get state files for checksum '%s' in object '%s'", cs, object.GetID())
				}
				if len(stateFiles) == 0 {
					return errors.Errorf("zero state file for checksum '%s' in object '%s'", cs, object.GetID())
				}
				external, err := object.GetExtensionManager().BuildObjectExternalPath(object, stateFiles[len(stateFiles)-1])
				if err != nil {
					return errors.Wrapf(err, "cannot build external path for file '%s' in object '%s'", stateFiles[len(stateFiles)-1], object.GetID())
				}
				file, err = mi.sourceFS.Open(external)
				if err != nil {
					return errors.Wrapf(err, "cannot open file '%s/%s' in source filesystem", mi.sourceFS.String(), targetNames[len(targetNames)-1])
				}
			}
		}
		tmpFile, err := os.CreateTemp(os.TempDir(), "gocfl*"+filepath.Ext(targetNames[len(targetNames)-1]))
		if err != nil {
			return errors.Wrap(err, "cannot create temp file")
		}
		if _, err := io.Copy(tmpFile, file); err != nil {
			tmpFile.Close()
			return errors.Wrap(err, "cannot copy file")
		}
		if err := file.Close(); err != nil {
			return errors.Wrap(err, "cannot close file")
		}
		tmpFilename := filepath.ToSlash(tmpFile.Name())
		targetFilename := filepath.ToSlash(filepath.Join(filepath.Dir(tmpFilename), "target."+filepath.Base(tmpFilename)))

		if err := tmpFile.Close(); err != nil {
			return errors.Wrap(err, "cannot close temp file")
		}
		if err := mig.Migrate(tmpFilename, targetFilename); err != nil {
			os.Remove(tmpFilename)
			return errors.Wrapf(err, "cannot migrate file '%v' to object '%s'", targetNames, object.GetID())
		}
		if err := os.Remove(tmpFilename); err != nil {
			return errors.Wrapf(err, "cannot remove temp file '%s'", tmpFilename)
		}
		mFile, err := os.Open(targetFilename)
		if err != nil {
			return errors.Wrapf(err, "cannot open file '%s'", targetFilename)
		}
		if err := object.AddReader(mFile, targetNames, "", false); err != nil {
			return errors.Wrapf(err, "cannot migrate file '%v' to object '%s'", targetNames, object.GetID())
		}
		/*
			if err := mFile.Close(); err != nil {
				return errors.Wrapf(err, "cannot close file '%s'", targetFilename)
			}
		*/
		os.Remove(targetFilename)
	}
	return nil
}

func (mi *Migration) GetMetadata(object ocfl.Object) (map[string]any, error) {
	return map[string]any{MigrationName: nil}, nil
}

var (
	_ ocfl.Extension             = &Migration{}
	_ ocfl.ExtensionObjectChange = &Migration{}
	_ ocfl.ExtensionMetadata     = &Migration{}
	_ ocfl.ExtensionNewVersion   = &Migration{}
)

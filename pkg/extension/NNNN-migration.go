package extension

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"io/fs"
	"path/filepath"
	"regexp"

	"emperror.dev/errors"
	"github.com/andybalholm/brotli"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/ocfl-archive/gocfl/v2/pkg/appendfs"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl"
	extensiontypes "github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"
	inventorytypes "github.com/ocfl-archive/gocfl/v2/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfllogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/subsystem/migration"
	"github.com/ocfl-archive/indexer/v3/pkg/indexer"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

const MigrationName = "NNNN-migration"
const MigrationDescription = "preservation management - file migration"

func NewMigration(mig *migration.Migration) *Migration {
	config := &MigrationConfig{
		ExtensionConfig: &extensiontypes.ExtensionConfig{ExtensionName: MigrationName},
	}
	sl := &Migration{
		MigrationConfig: config,
		migration:       mig,
		buffer:          map[string]*bytes.Buffer{},
		migrationFiles:  map[string]*migration.Function{},
		migratedFiles:   map[string]map[string]string{},
	}
	//	sl.writer = brotli.NewWriter(sl.buffer)
	if mig != nil {
		sl.sourceFS = mig.SourceFS
	}
	return sl
}

type MigrationConfig struct {
	*extensiontypes.ExtensionConfig
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

type MigrationResult struct {
	Source string `json:"source"`
	Error  string `json:"error,omitempty"`
	ID     string `json:"id"`
}

type migrationLine struct {
	Path      string           `json:"path"`
	Migration *MigrationResult `json:"migration"`
}

// map pronom to migration
type MigrationMap map[string]*MigrationTarget

// map checksum to migration
type MigrationFiles map[string]*MigrationTarget

type Migration struct {
	*MigrationConfig
	//targetFS  appendfs.FS
	lastHead  *inventorytypes.VersionNumber
	migration *migration.Migration
	//buffer *bytes.Buffer
	buffer         map[string]*bytes.Buffer
	writer         *brotli.Writer
	migrationFiles map[string]*migration.Function
	migratedFiles  map[string]map[string]string
	sourceFS       fs.FS
	currentHead    string
	done           bool
	logger         ocfllogger.OCFLLogger
}

func (mi *Migration) WithLogger(logger ocfllogger.OCFLLogger) extensiontypes.Extension {
	mi.logger = logger.With("extension", MigrationName)
	return mi
}

func (mi *Migration) Load(fsys fs.FS) error {
	data, err := fs.ReadFile(fsys, "config.json")
	if err != nil {
		return errors.Wrap(err, "cannot read config.json")
	}
	if err := json.Unmarshal(data, mi.MigrationConfig); err != nil {
		return errors.Wrapf(err, "cannot unmarshal MigrationConfig '%s'", string(data))
	}
	return nil
}

func (mi *Migration) Terminate() error {
	return nil
}

/*
func (mi *Migration) GetFS() fs.FS {
	return mi.targetFS
}

func (mi *Migration) SetFS(targetFS fs.FS, create bool) {
	if sfs, ok := targetFS.(appendfs.FS); ok {
		mi.targetFS = sfs
	}
}

*/

func (mi *Migration) GetConfig() any {
	return mi.MigrationConfig
}

func (mi *Migration) IsRegistered() bool { return false }

func (mi *Migration) GetName() string { return MigrationName }

func (mi *Migration) SetParams(map[string]string) error {
	return nil
}

func (mi *Migration) WriteConfig(fsys appendfs.FS) error {
	jsonData, _ := json.MarshalIndent(mi.MigrationConfig, "", "  ")
	if _, err := writefs.WriteFile(fsys, "config.json", jsonData); err != nil {
		return errors.Wrap(err, "cannot write config.json")
	}
	return nil
}

func (mi *Migration) UpdateObjectBefore(object.VersionWriter) error {
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

func (mi *Migration) UpdateObjectAfter(obj object.VersionWriter) error {
	inventory := obj.GetInventory()
	if inventory == nil {
		return errors.Errorf("inventory is nil")
	}

	// first get the metadata from the object
	extractor := obj.GetExtractor(obj.GetFS())
	meta, err := extractor.GetMetadata()
	if err != nil {
		return errors.Wrapf(err, "cannot get metadata from object %s", obj.GetID())
	}
	for _, m := range meta.Files {
		// check whether already migrated
		migrationMetaAny, ok := m.Extension[MigrationName]
		if !ok {
			continue
		}
		migrationMeta, ok := migrationMetaAny.(*MigrationResult)
		if !ok {
			continue
		}
		if _, ok := mi.migratedFiles[inventory.GetHead().String()]; !ok {
			mi.migratedFiles[inventory.GetHead().String()] = map[string]string{}
		}
		mi.migratedFiles[inventory.GetHead().String()][migrationMeta.Source] = migrationMeta.ID

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
		migrationFunctions, err := mi.migration.GetFunctionByPronom(indexerMeta.Pronom)
		if err != nil {
			continue
			//return errors.Wrapf(err, "cannot get migration function for pronom %s", pronom)
		}
		if mi.alreadyMigrated(cs) {
			continue
		}
		mi.migrationFiles[cs] = migrationFunctions
	}

	mi.lastHead = inventory.GetHead()
	return nil
}

func (mi *Migration) NeedNewVersion(object.VersionWriter) (bool, error) {
	return len(mi.migrationFiles) > 0 && !mi.done, nil
}

// DoNewVersion todo: check for second migration step and do different naming
func (mi *Migration) DoNewVersion(obj object.VersionWriter) error {
	defer func() {
		mi.migrationFiles = map[string]*migration.Function{}
		mi.done = true
	}()

	migrationMetadata, err := mi.GetMetadata(nil, obj)
	if err != nil {
		return errors.Wrapf(err, "cannot get migration metadata for object '%s'", obj.GetID())
	}
	migratedChecksums := maps.Keys(migrationMetadata)
	for _, metaAny := range migrationMetadata {
		if meta, ok := metaAny.(*migrationLine); ok {
			_ = meta
		}
	}
	inv := obj.GetInventory()
	head := inv.GetHead()
	/*
		extensionManager := object.GetExtensionManager()
		if extensionManager == nil {
			return errors.Errorf("extension manager is nil")
		}
	*/
	mi.buffer[head.String()] = &bytes.Buffer{}
	mi.writer = brotli.NewWriter(mi.buffer[head.String()])
	//files := inv.GetFiles()

	versions := ocfl.SeqToSlice(inv.GetVersions().GetVersionNumbers())
	if len(versions) < 2 {
		return errors.Errorf("cannot migrate files in object '%s' - no previous version", obj.GetID())
	}
	manifest := inv.GetManifest()
	if _, ok := mi.migratedFiles[head.String()]; !ok {
		mi.migratedFiles[head.String()] = map[string]string{}
	}
	for cs, mig := range mi.migrationFiles {
		var isMigrated bool
		if slices.Contains(migratedChecksums, cs) {
			isMigrated = true
		}
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
		manifestFiles, err := manifest.GetFiles(cs)
		if err != nil {
			return errors.Errorf("cannot find file with checksum '%s' in object '%s'", cs, obj.GetID())
		}
		// get the files from last version
		ver := inv.GetVersions().GetVersion(versions[len(versions)-2])
		if ver == nil {
			return errors.Errorf("cannot get state files for checksum '%s' in object '%s' version %s", cs, obj.GetID(), versions[len(versions)-2])
		}
		externalFiles, err := ver.GetState().GetFiles(cs)
		if err != nil {
			return errors.Errorf("cannot get state files for checksum '%s' in object '%s'", cs, obj.GetID())
		}
		for _, externalFile := range externalFiles {
			t := mig.GetDestinationName(externalFile, head.String(), isMigrated)
			if t == "" {
				return errors.Errorf("cannot get destination name for file '%s' in object '%s'", externalFile, obj.GetID())
			}
			targetNames = append(targetNames, t)
		}

		mi.migratedFiles[head.String()][cs] = manifestFiles[0]

		var file io.ReadCloser
		var ext string
		// no direct object filesystem available here; fall back to sourceFS branch
		file = nil
		ext = ""
		if file == nil {
			if mi.sourceFS != nil {
				stateFiles, err := inv.GetVersions().GetVersion(inv.GetHead()).GetState().GetFiles(cs)
				if err != nil {
					return errors.Wrapf(err, "cannot get state files for checksum '%s' in object '%s'", cs, obj.GetID())
				}
				if len(stateFiles) == 0 {
					return errors.Errorf("zero state file for checksum '%s' in object '%s'", cs, obj.GetID())
				}
				external, err := obj.GetExtensionManager().BuildObjectExtractPath(stateFiles[len(stateFiles)-1], "")
				if err != nil {
					return errors.Wrapf(err, "cannot build external path for file '%s' in object '%s'", stateFiles[len(stateFiles)-1], obj.GetID())
				}
				file, err = mi.sourceFS.Open(external)
				if err != nil {
					return errors.Wrapf(err, "cannot open file '%v/%s' in source filesystem", mi.sourceFS, targetNames[len(targetNames)-1])
				}
				ext = filepath.Ext(external)
			} else {
				// todo: this is not correct
				continue
			}
		}
		var ml *migrationLine
		/*
			path, err := extensionManager.BuildObjectManifestPath(object, targetNames[0], "content")
			if err != nil {
				return errors.Wrapf(err, "cannot build state path for file '%s' in object '%s'", targetNames[0], object.GetID())
			}
		*/
		extractTargetNames := []string{}
		for _, targetName := range targetNames {
			extractTargetName, err := obj.GetExtensionManager().BuildObjectExtractPath(targetName, "")
			if err != nil {
				return errors.Wrapf(err, "cannot build extract path for file '%s' in object '%s'", targetName, obj.GetID())
			}
			extractTargetNames = append(extractTargetNames, extractTargetName)
		}
		manifestName, err := obj.GetExtensionManager().BuildObjectManifestPath(extractTargetNames[0], "content")
		if err != nil {
			return errors.Wrapf(err, "cannot build manifest path for file '%s' in object '%s'", extractTargetNames[0], obj.GetID())
		}
		path := inv.BuildManifestName(manifestName)
		if err := migration.DoMigrate(obj, mig, ext, extractTargetNames, file); err != nil {
			ml = &migrationLine{
				Path: path,
				Migration: &MigrationResult{
					Source: manifestFiles[0],
					Error:  err.Error(),
					ID:     mig.GetID(),
				},
			}
			//			return err
		} else {
			ml = &migrationLine{
				Path: path,
				Migration: &MigrationResult{
					Source: manifestFiles[0],
					ID:     mig.GetID(),
				},
			}
			switch mig.Strategy {
			case migration.StrategyReplace:
				for _, externalFile := range externalFiles {
					if slices.Contains(targetNames, externalFile) {
						continue
					}
					if err := obj.DeleteFile(externalFile, cs); err != nil {
						return errors.Wrapf(err, "cannot delete file '%s' in object '%s'", externalFile, obj.GetID())
					}
				}
			case migration.StrategyFolder:
				for _, src := range externalFiles {
					if slices.Contains(targetNames, src) {
						continue
					}
					var dest string
					if !isMigrated {
						dest = filepath.ToSlash(filepath.Join(filepath.Dir(src), filepath.Base(src), filepath.Base(src)))
					}
					if err := obj.RenameFile(src, dest, cs); err != nil {
						return errors.Wrapf(err, "cannot delete file '%s' in object '%s'", src, obj.GetID())
					}
				}
			}
		}
		data, err := json.Marshal(ml)
		if err != nil {
			return errors.Wrapf(err, "cannot marshal migration line for file '%s' in object '%s'", targetNames[0], obj.GetID())
		}
		if _, err := mi.writer.Write(append(data, []byte("\n")...)); err != nil {
			return errors.Wrapf(err, "cannot write migration line for file '%s' in object '%s'", targetNames[0], obj.GetID())
		}
	}
	if err := mi.writer.Flush(); err != nil {
		return errors.Wrapf(err, "cannot flush migration line writer for object '%s'", obj.GetID())
	}
	if err := mi.writer.Close(); err != nil {
		return errors.Wrapf(err, "cannot close migration line writer for object '%s'", obj.GetID())
	}
	buffer, ok := mi.buffer[head.String()]
	if !ok {
		return nil
	}
	if err := WriteJsonL(
		obj.GetFS(),
		obj,
		"migration",
		buffer.Bytes(),
		mi.MigrationConfig.Compress,
		mi.StorageType,
		mi.StorageName,
	); err != nil {
		return errors.Wrap(err, "cannot write jsonl")
	}
	return nil
}

func (mi *Migration) GetMetadata(sourceFS fs.FS, obj object.Object) (map[string]any, error) {
	var err error
	var result = map[string]any{}

	inventory := obj.GetInventory()
	manifest := inventory.GetManifest()
	path2digest := map[string]string{}
	for checksum, names := range manifest.Iterate() {
		for _, name := range names {
			path2digest[name] = checksum
		}
	}
	for v := range inventory.GetVersions().GetVersionNumbers() {
		var data []byte
		if buf, ok := mi.buffer[v.String()]; ok && buf.Len() > 0 {
			//		if v == inventory.GetHead() && sl.buffer.Len() > 0 {
			// need a new reader on the buffer
			reader := brotli.NewReader(bytes.NewBuffer(buf.Bytes()))
			data, err = io.ReadAll(reader)
			if err != nil {
				return nil, errors.Wrapf(err, "cannot read buffer for '%s' '%s'", obj.GetID(), v)
			}
		} else {
			data, err = ReadJsonL(sourceFS, obj, v, "migration", mi.MigrationConfig.Compress, mi.StorageType, mi.StorageName)
			if err != nil {
				continue
				// return nil, errors.Wrapf(err, "cannot read jsonl for '%s' version '%s'", object.GetID(), v)
			}
		}

		reader := bytes.NewReader(data)
		r := bufio.NewScanner(reader)
		r.Buffer(make([]byte, 128*1024), 16*1024*1024)
		r.Split(bufio.ScanLines)
		for r.Scan() {
			line := r.Text()
			var meta = migrationLine{}
			if err := json.Unmarshal([]byte(line), &meta); err != nil {
				return nil, errors.Wrapf(err, "cannot unmarshal line from for '%s' %s - [%s]", obj.GetID(), v, line)
			}
			var digest string
			for cs, names := range manifest.Iterate() {
				for _, name := range names {
					if name == meta.Migration.Source {
						digest = cs
						break
					}
				}
			}
			if digest == "" {
				return nil, errors.Errorf("cannot find checksum for file '%s' in object '%s'", meta.Migration.Source, obj.GetID())
			}
			cs, ok := path2digest[meta.Path]
			if !ok && meta.Migration.Error != "" {
				cs, ok = path2digest[meta.Migration.Source]
			}
			if !ok {
				return nil, errors.Errorf("cannot find checksum for file '%s' in object '%s'", meta.Path, obj.GetID())
			}
			meta.Migration.Source = digest
			result[cs] = meta.Migration
		}
		if err := r.Err(); err != nil {
			return nil, errors.Wrapf(err, "cannot scan lines for '%s' %s", obj.GetID(), v)
		}
	}
	return result, nil
}

var (
	_ extensiontypes.Extension     = &Migration{}
	_ object.ExtensionObjectChange = &Migration{}
	_ object.ExtensionMetadata     = &Migration{}
	_ object.ExtensionNewVersion   = &Migration{}
)

package migration

import (
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"emperror.dev/errors"
	"github.com/google/shlex"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object"
)

func GetMigrations(conf *ConfigMigration) (*Migration, error) {
	m := &Migration{
		Functions: map[string]*Function{},
	}

	for name, fn := range conf.Function {
		parts, err := shlex.Split(fn.Command)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot parse Migration.Function.%s", name)
		}
		if len(parts) < 1 {
			return nil, errors.Errorf("Migration.Function.%s is empty", name)
		}
		re, err := regexp.Compile(fn.FilenameRegexp)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot parse Migration.Function.%s", name)
		}
		timeout := fn.Timeout
		var pronoms []string
		for _, pro := range fn.Pronoms {
			pronoms = append(pronoms, strings.TrimSpace(pro))
		}
		strategy, ok := Strategies[fn.Strategy]
		if !ok {
			return nil, errors.Errorf("unknown strategy '%s' in Migration.Function.%s", fn.Strategy, name)
		}
		m.Functions[name] = &Function{
			title:    fn.Title,
			id:       fn.ID,
			command:  parts[0],
			args:     parts[1:],
			Strategy: strategy,
			regexp:   re,
			replace:  fn.FilenameReplacement,
			timeout:  time.Duration(timeout),
			pronoms:  pronoms,
		}
	}
	return m, nil
}

func DoMigrate(obj object.VersionWriter, mig *Function, ext string, targetNames []string, file io.ReadCloser) error {
	tmpFile, err := os.CreateTemp(os.TempDir(), "gocfl_*"+ext)
	if err != nil {
		return errors.Wrap(err, "cannot create temp file")
	}
	if _, err := io.Copy(tmpFile, file); err != nil {
		_ = tmpFile.Close()
		return errors.Wrap(err, "cannot copy file")
	}
	if err := file.Close(); err != nil {
		return errors.Wrap(err, "cannot close file")
	}
	tmpFilename := filepath.ToSlash(tmpFile.Name())
	if len(targetNames) == 0 {
		return errors.New("targetNames is empty")
	}
	targetFilename := filepath.ToSlash(filepath.Join(filepath.Dir(tmpFilename), "target."+filepath.Base(tmpFilename)+filepath.Ext(targetNames[0])))

	if err := tmpFile.Close(); err != nil {
		return errors.Wrap(err, "cannot close temp file")
	}
	defer func() {
		_ = os.Remove(tmpFilename)
		_ = os.Remove(targetFilename)
	}()
	if err := mig.Migrate(tmpFilename, targetFilename); err != nil {
		return errors.Wrapf(err, "cannot migrate file '%v' to object '%s'", targetNames, obj.GetID())
	}

	mFile, err := os.Open(targetFilename)
	if err != nil {
		return errors.Wrapf(err, "cannot open file '%s'", targetFilename)
	}
	if _, err := obj.AddReader(mFile, targetNames, "content", false, false); err != nil {
		return errors.Wrapf(err, "cannot migrate file '%v' to object '%s'", targetNames, obj.GetID())
	}
	if err := mFile.Close(); err != nil {
		return errors.Wrapf(err, "cannot close file '%s'", targetFilename)
	}
	return nil
}

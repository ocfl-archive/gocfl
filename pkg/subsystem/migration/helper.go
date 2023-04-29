package migration

import (
	"emperror.dev/errors"
	"github.com/google/shlex"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	"github.com/spf13/viper"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func anyToStringMapString(dataAny any) (map[string]string, error) {
	result := map[string]string{}
	data, ok := dataAny.(map[string]interface{})
	if !ok {
		return nil, errors.Errorf("cannot convert to map[string]interface{}")
	}
	for k, v := range data {
		str, ok := v.(string)
		if !ok {
			return nil, errors.Errorf("cannot convert '%s' to string", k)
		}
		result[strings.ToLower(k)] = str
	}
	return result, nil
}

func GetMigrations() (*Migration, error) {
	m := &Migration{
		Functions: map[string]*Function{},
		//Sources:   viper.GetStringMapString("Migration.Source"),
	}
	cmdstrings := viper.GetStringMap("Migration.Function")

	for name, cmdAny := range cmdstrings {
		cmdMap, err := anyToStringMapString(cmdAny)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot parse Migration.Function.%s", name)
		}
		parts, err := shlex.Split(cmdMap["command"])
		if err != nil {
			return nil, errors.Wrapf(err, "cannot parse Migration.Function.%s", name)
		}
		if len(parts) < 1 {
			return nil, errors.Errorf("Migration.Function.%s is empty", name)
		}
		re, err := regexp.Compile(cmdMap["filenameregexp"])
		if err != nil {
			return nil, errors.Wrapf(err, "cannot parse Migration.Function.%s", name)
		}
		timeoutString := cmdMap["timeout"]
		timeout, err := time.ParseDuration(timeoutString)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot parse timeout of Migration.Function.%s", name)
		}
		var pronoms []string
		pros := strings.Split(cmdMap["pronoms"], ",")
		for _, pro := range pros {
			pronoms = append(pronoms, strings.TrimSpace(pro))
		}
		strategy, ok := Strategies[cmdMap["strategy"]]
		if !ok {
			return nil, errors.Errorf("unknown strategy '%s' in Migration.Function.%s", cmdMap["strategy"], name)
		}
		m.Functions[name] = &Function{
			title:    cmdMap["title"],
			id:       cmdMap["id"],
			command:  parts[0],
			args:     parts[1:],
			Strategy: strategy,
			regexp:   re,
			replace:  cmdMap["filenamereplacement"],
			timeout:  timeout,
			pronoms:  pronoms,
		}
	}
	return m, nil
}

func DoMigrate(object ocfl.Object, mig *Function, targetNames []string, file io.ReadCloser) error {
	tmpFile, err := os.CreateTemp(os.TempDir(), "gocfl_*"+filepath.Ext(targetNames[len(targetNames)-1]))
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
	targetFilename := filepath.ToSlash(filepath.Join(filepath.Dir(tmpFilename), "target."+filepath.Base(tmpFilename)))

	if err := tmpFile.Close(); err != nil {
		return errors.Wrap(err, "cannot close temp file")
	}
	if err := mig.Migrate(tmpFilename, targetFilename); err != nil {
		_ = os.Remove(tmpFilename)
		return errors.Wrapf(err, "cannot migrate file '%v' to object '%s'", targetNames, object.GetID())
	}
	if err := os.Remove(tmpFilename); err != nil {
		return errors.Wrapf(err, "cannot remove temp file '%s'", tmpFilename)
	}

	mFile, err := os.Open(targetFilename)
	if err != nil {
		return errors.Wrapf(err, "cannot open file '%s'", targetFilename)
	}
	if err := object.AddReader(mFile, targetNames, "content", false); err != nil {
		return errors.Wrapf(err, "cannot migrate file '%v' to object '%s'", targetNames, object.GetID())
	}
	if err := mFile.Close(); err != nil {
		return errors.Wrapf(err, "cannot close file '%s'", targetFilename)
	}
	if err := os.Remove(targetFilename); err != nil {
		return errors.Wrapf(err, "cannot remove temp file '%s'", targetFilename)
	}
	return nil
}

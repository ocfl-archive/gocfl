package migration

import (
	"context"
	"emperror.dev/errors"
	"github.com/google/shlex"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	"github.com/spf13/viper"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type Function struct {
	command  string
	args     []string
	Strategy string
	regexp   *regexp.Regexp
	replace  string
	timeout  time.Duration
}

func (f *Function) GetDestinationName(src string) string {
	return f.regexp.ReplaceAllString(src, f.replace)
}

func (f *Function) Migrate(source string, dest string) error {
	ctx, cancel := context.WithTimeout(context.Background(), f.timeout)
	defer cancel()
	args := []string{}
	for _, arg := range f.args {
		//		arg = strings.ReplaceAll(arg, "{source}", filepath.Base(source))
		//		arg = strings.ReplaceAll(arg, "{destination}", filepath.Base(dest))
		arg = strings.ReplaceAll(arg, "{source}", filepath.ToSlash(source))
		arg = strings.ReplaceAll(arg, "{destination}", filepath.ToSlash(dest))

		args = append(args, arg)
	}
	cmd := exec.CommandContext(ctx, f.command, args...)
	cmd.Dir = filepath.Dir(source)
	return errors.WithStack(cmd.Run())
}

type Migration struct {
	Functions map[string]*Function
	Sources   map[string]string
	SourceFS  ocfl.OCFLFSRead
}

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
		Sources:   viper.GetStringMapString("Migration.Source"),
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
		m.Functions[name] = &Function{
			command:  parts[0],
			args:     parts[1:],
			Strategy: cmdMap["strategy"],
			regexp:   re,
			replace:  cmdMap["filenamereplacement"],
			timeout:  timeout,
		}
	}
	return m, nil
}

func (m *Migration) GetFunctionByName(name string) (*Function, error) {
	if f, ok := m.Functions[strings.ToLower(name)]; ok {
		return f, nil
	}
	return nil, errors.Errorf("Migration.Function.%s does not exist", name)
}

func (m *Migration) GetFunctionByPronom(pronom string) (*Function, error) {
	if f, ok := m.Sources[pronom]; ok {
		return m.GetFunctionByName(f)
	}
	return nil, errors.Errorf("Migration.Source.%s does not exist", pronom)
}

func (m *Migration) SetSourceFS(fs ocfl.OCFLFSRead) {
	m.SourceFS = fs
}

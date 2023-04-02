package migration

import (
	"emperror.dev/errors"
	"github.com/google/shlex"
	"github.com/spf13/viper"
	"io"
	"os/exec"
	"regexp"
	"strings"
)

type Function struct {
	command  *exec.Cmd
	Strategy string
	regexp   *regexp.Regexp
	replace  string
}

func (f *Function) GetDestinationName(src string) string {
	return f.regexp.ReplaceAllString(src, f.replace)
}

func (f *Function) Migrate(r io.Reader, w io.Writer) error {
	f.command.Stdin = r
	f.command.Stdout = w
	return errors.WithStack(f.command.Run())
}

type Migration struct {
	Functions map[string]*Function
	Sources   map[string]string
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
		result[k] = strings.ToLower(str)
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
		cmd := exec.Command(parts[0], parts[1:]...)
		re, err := regexp.Compile(cmdMap["regexp"])
		if err != nil {
			return nil, errors.Wrapf(err, "cannot parse Migration.Function.%s", name)
		}
		m.Functions[name] = &Function{
			command:  cmd,
			Strategy: cmdMap["strategy"],
			regexp:   re,
			replace:  cmdMap["replace"],
		}
	}
	return m, nil
}

func (m *Migration) GetFunctionByName(name string) (*Function, error) {
	if f, ok := m.Functions[name]; ok {
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

package migration

import (
	"context"
	"emperror.dev/errors"
	"fmt"
	"io/fs"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type Strategy string

const (
	StrategyReplace Strategy = "replace"
	StrategyAdd     Strategy = "add"
	StrategyFolder  Strategy = "folder"
)

var Strategies = map[string]Strategy{
	"replace": StrategyReplace,
	"add":     StrategyAdd,
	"folder":  StrategyFolder,
}

type Function struct {
	command  string
	args     []string
	Strategy Strategy
	regexp   *regexp.Regexp
	replace  string
	timeout  time.Duration
	title    string
	id       string
	pronoms  []string
}

var migrationVersionRegexp = regexp.MustCompile(`^([^.]+)\.(.+)$`)

func (f *Function) GetDestinationName(src string, head string, isMigrated bool) string {
	dest := f.regexp.ReplaceAllString(src, f.replace)
	if f.Strategy == StrategyFolder {
		if isMigrated {
			parts := migrationVersionRegexp.FindStringSubmatch(filepath.Base(src))
			if parts == nil {
				return ""
			}
			dest = filepath.ToSlash(filepath.Join(filepath.Dir(src), fmt.Sprintf("%s.%s", head, parts[2])))
		} else {
			dest = filepath.ToSlash(filepath.Join(src, fmt.Sprintf("%s.%s", head, dest)))
		}
	}
	return dest
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
	return errors.Wrapf(cmd.Run(), "cannot run command '%s %s'", f.command, strings.Join(args, " "))
}

func (f *Function) GetID() string {
	return f.id
}

type Migration struct {
	Functions map[string]*Function
	//Sources   map[string]string
	SourceFS fs.FS
}

func (m *Migration) GetFunctionByName(name string) (*Function, error) {
	if f, ok := m.Functions[strings.ToLower(name)]; ok {
		return f, nil
	}
	return nil, errors.Errorf("Migration.Function.%s does not exist", name)
}

func (m *Migration) GetFunctionByPronom(pronom string) (*Function, error) {
	for _, f := range m.Functions {
		for _, pro := range f.pronoms {
			if pro == pronom {
				return f, nil
			}
		}
	}
	return nil, errors.Errorf("Migration.Source.%s does not exist", pronom)
}

func (m *Migration) SetSourceFS(fs fs.FS) {
	m.SourceFS = fs
}

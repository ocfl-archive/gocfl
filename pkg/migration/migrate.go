package migration

import (
	"context"
	"emperror.dev/errors"
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
)

var Strategies = map[string]Strategy{
	"replace": StrategyReplace,
	"add":     StrategyAdd,
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

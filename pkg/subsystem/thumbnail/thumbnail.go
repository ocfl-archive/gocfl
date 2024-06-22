package thumbnail

import (
	"context"
	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/zLogger"
	"io/fs"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type ThumbnailMeta struct {
	Ext    string
	Width  uint64
	Height uint64
	Mime   string
}

type Function struct {
	thumb   *Thumbnail
	command string
	args    []string
	timeout time.Duration
	title   string
	id      string
	pronoms []string
	mime    []*regexp.Regexp
}

func (f *Function) Thumbnail(source string, dest string, width uint64, height uint64, logger zLogger.ZLogger) error {
	ctx, cancel := context.WithTimeout(context.Background(), f.timeout)
	defer cancel()
	args := []string{}
	for _, arg := range f.args {
		arg = strings.ReplaceAll(arg, "{source}", filepath.ToSlash(source))
		arg = strings.ReplaceAll(arg, "{destination}", filepath.ToSlash(dest))
		arg = strings.ReplaceAll(arg, "{background}", f.thumb.Background)
		arg = strings.ReplaceAll(arg, "{width}", strconv.FormatUint(width, 10))
		arg = strings.ReplaceAll(arg, "{height}", strconv.FormatUint(height, 10))
		args = append(args, arg)
	}
	logger.Debug().Msgf("%s %v", f.command, args)
	cmd := exec.CommandContext(ctx, f.command, args...)
	cmd.Dir = filepath.Dir(source)
	return errors.Wrapf(cmd.Run(), "cannot run command '%s %s'", f.command, strings.Join(args, " "))
}

func (f *Function) GetID() string {
	return f.id
}

type Thumbnail struct {
	Functions  map[string]*Function
	SourceFS   fs.FS
	Background string
}

func (m *Thumbnail) GetFunctionByName(name string) (*Function, error) {
	if f, ok := m.Functions[strings.ToLower(name)]; ok {
		return f, nil
	}
	return nil, errors.Errorf("Thumbnail.Function.%s does not exist", name)
}

func (m *Thumbnail) GetFunctionByPronom(pronom string) (*Function, error) {
	for _, f := range m.Functions {
		for _, pro := range f.pronoms {
			if pro == pronom {
				return f, nil
			}
		}
	}
	return nil, errors.Errorf("Thumbnail.Source.%s does not exist", pronom)
}

func (m *Thumbnail) GetFunctionByMimetype(mime string) (*Function, error) {
	for _, f := range m.Functions {
		for _, re := range f.mime {
			if re.MatchString(mime) {
				return f, nil
			}
		}
	}
	return nil, errors.Errorf("Thumbnail.Source.%s does not exist", mime)
}

func (m *Thumbnail) SetSourceFS(fs fs.FS) {
	m.SourceFS = fs
}

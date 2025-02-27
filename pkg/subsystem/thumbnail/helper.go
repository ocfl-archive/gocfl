package thumbnail

import (
	"emperror.dev/errors"
	"github.com/google/shlex"
	"github.com/ocfl-archive/gocfl/v2/config"
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

func GetThumbnails(conf *config.GOCFLConfig) (*Thumbnail, error) {
	m := &Thumbnail{
		Functions:  map[string]*Function{},
		Background: conf.Thumbnail.Background,
	}

	for name, fn := range conf.Thumbnail.Function {
		parts, err := shlex.Split(fn.Command)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot parse Thumbnail.Function.%s", name)
		}
		if len(parts) < 1 {
			return nil, errors.Errorf("Thumbnail.Function.%s is empty", name)
		}
		timeout := fn.Timeout
		if err != nil {
			return nil, errors.Wrapf(err, "cannot parse timeout of Thumbnail.Function.%s", name)
		}
		var mimeRes = []*regexp.Regexp{}
		var typeRes = []*regexp.Regexp{}
		for _, mime := range fn.Mime {
			re, err := regexp.Compile(mime)
			if err != nil {
				return nil, errors.Wrapf(err, "cannot parse Migration.Function.%s", name)
			}
			mimeRes = append(mimeRes, re)
		}
		for _, t := range fn.Types {
			re, err := regexp.Compile(t)
			if err != nil {
				return nil, errors.Wrapf(err, "cannot parse Migration.Function.%s", name)
			}
			typeRes = append(typeRes, re)
		}
		var pronoms []string
		for _, pro := range fn.Pronoms {
			pronoms = append(pronoms, strings.TrimSpace(pro))
		}
		m.Functions[name] = &Function{
			thumb:   m,
			title:   fn.Title,
			id:      fn.ID,
			command: parts[0],
			args:    parts[1:],
			timeout: time.Duration(timeout),
			pronoms: pronoms,
			mime:    mimeRes,
			types:   typeRes,
		}
	}
	return m, nil
}

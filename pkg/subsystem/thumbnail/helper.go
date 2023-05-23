package thumbnail

import (
	"emperror.dev/errors"
	"github.com/google/shlex"
	"github.com/spf13/viper"
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

func GetThumbnails() (*Thumbnail, error) {
	m := &Thumbnail{
		Functions:  map[string]*Function{},
		Background: viper.GetString("Thumbnail.Background"),
		//Sources:   viper.GetStringMapString("Thumbnail.Source"),
	}
	cmdstrings := viper.GetStringMap("Thumbnail.Function")

	for name, cmdAny := range cmdstrings {
		cmdMap, err := anyToStringMapString(cmdAny)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot parse Thumbnail.Function.%s", name)
		}
		parts, err := shlex.Split(cmdMap["command"])
		if err != nil {
			return nil, errors.Wrapf(err, "cannot parse Thumbnail.Function.%s", name)
		}
		if len(parts) < 1 {
			return nil, errors.Errorf("Thumbnail.Function.%s is empty", name)
		}
		timeoutString := cmdMap["timeout"]
		timeout, err := time.ParseDuration(timeoutString)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot parse timeout of Thumbnail.Function.%s", name)
		}
		var re *regexp.Regexp
		if cmdMap["mime"] != "" {
			re, err = regexp.Compile(cmdMap["mime"])
			if err != nil {
				return nil, errors.Wrapf(err, "cannot parse Migration.Function.%s", name)
			}
		}
		var pronoms []string
		pros := strings.Split(cmdMap["pronoms"], ",")
		for _, pro := range pros {
			pronoms = append(pronoms, strings.TrimSpace(pro))
		}
		m.Functions[name] = &Function{
			thumb:   m,
			title:   cmdMap["title"],
			id:      cmdMap["id"],
			command: parts[0],
			args:    parts[1:],
			timeout: timeout,
			pronoms: pronoms,
			mime:    re,
		}
	}
	return m, nil
}

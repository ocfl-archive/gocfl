package uri

import (
	"emperror.dev/errors"
	"regexp"
)

var uriRegexp = regexp.MustCompile(`^(?P<scheme>[a-z]+[a-z0-9+-.]+):(//(?P<authority>((?P<userinfo>[^@]+)@)?(?P<host>[^:/?#]+)(:(?P<port>[0-9]+)))/?)?(?P<path>[^#?]*)?(\?(?P<query>[^#]*))?(#(?P<anchor>.*))?$`)

type URI struct {
	Scheme    string
	Authority string
	Userinfo  string
	Host      string
	Port      string
	Path      string
	Query     string
	Fragment  string
}

func Parse(str string) (*URI, error) {
	u := &URI{}
	groupNames := uriRegexp.SubexpNames()
	matches := uriRegexp.FindAllStringSubmatch(str, -1)
	if matches == nil {
		return nil, errors.Errorf("'%s' does not match regexp '%s'", str, uriRegexp.String())
	}
	for _, match := range matches {
		for groupIdx, group := range match {
			switch groupNames[groupIdx] {
			case "scheme":
				u.Scheme = group
			case "authority":
				u.Authority = group
			case "userinfo":
				u.Userinfo = group
			case "host":
				u.Host = group
			case "port":
				u.Port = group
			case "path":
				u.Path = group
			case "query":
				u.Query = group
			case "fragment":
				u.Fragment = group
			}
		}
	}
	return u, nil
}

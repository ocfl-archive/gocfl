package info

import (
	"fmt"
	"runtime/debug"
	"time"
)

var VCSRevision string
var VCSTime time.Time
var MainVersion string
var GoVersion string
var BuildInfo string
var Version string

func init() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}
	BuildInfo = info.String()
	MainVersion = info.Main.Version
	GoVersion = info.GoVersion
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			VCSRevision = setting.Value
		case "vcs.time":
			VCSTime, _ = time.Parse(time.RFC3339, setting.Value)
		}
	}
	Version = fmt.Sprintf("%s %s (%s) %s", MainVersion, VCSRevision, VCSTime.Format(time.RFC3339), GoVersion)
}

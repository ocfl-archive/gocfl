// Custom view/summary objects for consumers of RO-CRATE metadata.
package rocrate

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const stringSeparator = "; "

// GocflSummary provides a summary compatible with gocfl user's
// expectations for the info.json object.
//
/*
	-- via RO-CRATE.
	signature       = id
	title           = name
	description     = description
	created         = datePublished
	sets            = @type
	keywords        = keywords
	licenses        = license

	-- output at runtime.
	last_changed    = now()

	-- via GOCFL config.
	organisation_id = user.config
	organisation    = user.config
	user            = user.config
	address         = user.config

*/
//
type GocflSummary struct {
	// provided by ro-crate.
	Signature   string `json:"signature"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Created     string `json:"created"`
	Sets        string `json:"sets"`
	Keywords    string `json:"keywords"`
	Licenses    string `json:"licenses"`
	// generated at runtime, e.g. time.Now().
	LastChanged string `json:"last_changed"`
	// provided by caller.
	OrganisationID string `json:"organisation_id"`
	Organisation   string `json:"organisation"`
	User           string `json:"user"`
	Address        string `json:"address"`
}

// newGocflSummary returns an initialized gocflSummary object for
// maximum safety.
func newGocflSummary() GocflSummary {
	return GocflSummary{
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		// provided by caller.
		"",
		"",
		"",
		"",
		"",
	}
}

// String provides stringer functions for the gocfl summary object.
func (gocflSummary GocflSummary) String() string {
	ret, err := json.MarshalIndent(gocflSummary, " ", " ")
	if err != nil {
		return fmt.Sprintf("%s: %s", StringerError, err)
	}
	return string(ret)
}

// joinStrings simplifies string join functions for us so that we can
// exchange string separator values easily.
func joinStrings(inputString []string) string {
	return strings.Join(inputString, stringSeparator)
}

// currentTime provides a module level approach to replacing the time
// functions, e.g. for testing.
var currentTime = getTime

// getTime provides a helpter to return the current time formatted as
// a string.
func getTime(utc bool) string {
	if utc {
		t := time.Now().UTC()
		return t.Format("2006-01-02T15:04:05Z")
	}
	t := time.Now()
	return t.Format("2006-01-02T15:04:05")
}

func (rcMeta rocrateMeta) GOCFLSummary() (GocflSummary, error) {
	if len(rcMeta.Graph) == 0 {
		return GocflSummary{}, fmt.Errorf("ro-crate-metadata.json is empty")
	}
	if len(rcMeta.Graph) == 1 {
		return GocflSummary{}, fmt.Errorf("ro-crate-metadata.json is non-conformant")
	}
	summary := newGocflSummary()
	summary.Signature = rcMeta.Graph[1].ID
	if rcMeta.Graph[1].Name != nil {
		name := rcMeta.Graph[1].Name.Value()
		if len(name) > 0 {
			summary.Title = rcMeta.Graph[1].Name.Value()[0]
		}
	}
	if rcMeta.Graph[1].Type != nil {
		summary.Sets = joinStrings(rcMeta.Graph[1].Type.Value())
	}
	if rcMeta.Graph[1].Description != nil {
		summary.Description = joinStrings(rcMeta.Graph[1].Description.Value())
	}
	summary.Created = rcMeta.Graph[1].DatePublished
	if rcMeta.Graph[1].License != nil {
		summary.Licenses = joinStrings(rcMeta.Graph[1].License.StringSlice())
	}
	if rcMeta.Graph[1].Keywords != nil {
		summary.Keywords = joinStrings(rcMeta.Graph[1].Keywords.Value())
	}
	summary.LastChanged = currentTime(true)
	return summary, nil
}

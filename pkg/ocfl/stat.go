package ocfl

type StatInfo int64

const (
	StatObjectFolders StatInfo = iota
	StatExtensionConfigs
	StatObjects
	StatObjectVersions
	StatObjectVersionState
	StatObjectManifest
	StatObjectExtensionConfigs
)

var StatInfoString = map[string]StatInfo{
	"ObjectFolders":          StatObjectFolders,
	"ExtensionConfigs":       StatExtensionConfigs,
	"Objects":                StatObjects,
	"ObjectVersions":         StatObjectVersions,
	"ObjectVersionState":     StatObjectVersionState,
	"ObjectManifest":         StatObjectManifest,
	"ObjectExtensionConfigs": StatObjectExtensionConfigs,
}

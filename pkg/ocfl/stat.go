package ocfl

type StatInfo int64

const (
	StatObjectFolders StatInfo = iota
	StatExtension
	StatExtensionConfigs
	StatObjects
	StatObjectVersions
	StatObjectVersionState
	StatObjectManifest
	StatObjectExtension
	StatObjectExtensionConfigs
)

var StatInfoString = map[string]StatInfo{
	"ObjectFolders":          StatObjectFolders,
	"Extension":              StatExtension,
	"ExtensionConfigs":       StatExtensionConfigs,
	"Objects":                StatObjects,
	"ObjectVersions":         StatObjectVersions,
	"ObjectVersionState":     StatObjectVersionState,
	"ObjectManifest":         StatObjectManifest,
	"ObjectExtension":        StatObjectExtension,
	"ObjectExtensionConfigs": StatObjectExtensionConfigs,
}

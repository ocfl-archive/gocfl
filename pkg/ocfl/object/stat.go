package object

// StatInfo is a type for statistical information about an OCFL object.
type StatInfo int64

const (
	StatObjectFolders          StatInfo = iota // Count of folders in the object.
	StatExtension                              // Extension information.
	StatExtensionConfigs                       // Extension configuration information.
	StatObjects                                // Count of objects.
	StatObjectVersions                         // Count of object versions.
	StatObjectVersionState                     // State of object versions.
	StatObjectManifest                         // Object manifest information.
	StatObjectExtension                        // Object extension information.
	StatObjectExtensionConfigs                 // Object extension configuration information.
)

// StatInfoString maps string representations to StatInfo values.
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

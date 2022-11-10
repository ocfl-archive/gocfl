package ocfl

type ExtensionConfig struct {
	ExtensionName string `json:"extensionName"`
}

type Extension interface {
	GetName() string
	WriteConfig(fs OCFLFS) error
}

type StoragerootPath interface {
	WriteLayout(fs OCFLFS) error
	BuildStorageRootPath(storageRoot StorageRoot, id string) (string, error)
}

type ObjectContentPath interface {
	BuildObjectContentPath(object Object, originalPath string) (string, error)
}

type ContentChange interface {
	AddFile(object Object, source, dest string)
	UpdateFile(object Object, source, dest string)
	DeleteFile(object Object, dest string)
}

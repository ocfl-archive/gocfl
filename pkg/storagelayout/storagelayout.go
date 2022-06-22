package storagelayout

const MAX_DIR_LEN = 255

type StorageLayout interface {
	ID2Path(id string) (string, error)
	Name() string
}

type Config struct {
	ExtensionName string `json:"extensionName"`
}

package config

import (
	"bytes"
	"encoding"
	"strings"

	"github.com/BurntSushi/toml"
)

type MiniConfig map[string]any

func (m MiniConfig) MarshalText() (text []byte, err error) {
	// Konvertiere flache Map mit Punkt-Trennern in verschachtelte Map
	nested := make(map[string]any)
	for key, value := range m {
		parts := strings.Split(strings.ToLower(key), ".")
		curr := nested
		for i, part := range parts {
			if i == len(parts)-1 {
				curr[part] = value
			} else {
				if _, ok := curr[part]; !ok {
					curr[part] = make(map[string]any)
				}
				if next, ok := curr[part].(map[string]any); ok {
					curr = next
				} else {
					// Fall, dass ein Zwischenknoten bereits ein Wert ist (z.B. "a" und "a.b")
					// Wir überschreiben den Wert mit einer Map, um den Pfad fortsetzen zu können.
					next := make(map[string]any)
					curr[part] = next
					curr = next
				}
			}
		}
	}

	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	if err := enc.Encode(nested); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

var _ encoding.TextMarshaler = MiniConfig{}

package thumbnail

import configutil "github.com/je4/utils/v2/pkg/config"

type ConfigThumbnailFunction struct {
	ID      string              `toml:"id"`
	Title   string              `toml:"title"`
	Command string              `toml:"command"`
	Timeout configutil.Duration `toml:"timeout"`
	Pronoms []string            `toml:"pronoms"`
	Mime    []string            `toml:"mime"`
}

type ConfigThumbnail struct {
	Enabled    bool                                `toml:"enabled"`
	Background string                              `toml:"background"`
	Function   map[string]*ConfigThumbnailFunction `toml:"function"`
}

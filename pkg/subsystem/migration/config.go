package migration

import configutil "github.com/je4/utils/v2/pkg/config"

type ConfigMigrationFunction struct {
	ID                  string              `toml:"id"`
	Title               string              `toml:"title"`
	Command             string              `toml:"command"`
	Strategy            string              `toml:"strategy"`
	FilenameRegexp      string              `toml:"filenameregexp"`
	FilenameReplacement string              `toml:"filenamereplacement"`
	Timeout             configutil.Duration `toml:"timeout"`
	Pronoms             []string            `toml:"pronoms"`
}

type ConfigMigration struct {
	Enabled  bool                                `toml:"enabled"`
	Function map[string]*ConfigMigrationFunction `toml:"function"`
}

package run

type config struct {
	SqlitePath string `env:"SQLITE_PATH" envDefault:"db.sqlite3"`
	ArchiveDir string `env:"ARCHIVE_DIR" envDefault:"./archive"`
}

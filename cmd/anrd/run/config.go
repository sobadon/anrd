package run

type config struct {
	SqlitePath string `env:"SQLITE_PATH" envDefault:"db.sqlite3"`
}

package orm

type ORMOption struct {
	DSN     string            `yaml:"dsn"`
	ShowSQL bool              `yaml:"showsql"`
	Extra   map[string]string `yaml:"extra"`
}

package napoleon

import "database/sql"

type initPaths struct {
	rootPath    string
	folderNames []string
}

type cookieConfig struct {
	name     string
	lifetime string
	parsist  string
	srcure   string
	domain   string
}

type DatabaseConfig struct {
	dsn      string
	database string
}

type Database struct {
	DataType string
	Pool     *sql.DB
}

type RedisConfig struct {
	host     string
	password string
	prefix   string
}

package migrations

import "time"

type DB struct {
	URL                 string        `json:"url"`
	DatabaseName        string        `json:"dbname"`
	SchemaName          string        `json:"schema_name"`
	MaxOpenConns        int           `json:"max_open_conns"`
	MaxIdleConns        int           `json:"max_idle_conns"`
	MigrationsPath      string        `json:"migrations_path"`
	ConnMaxLifetime     time.Duration `json:"conn_max_lifetime"`
	ConnMaxIdleLifetime time.Duration `json:"conn_max_idle_lifetime"`
	ContextTimeout      time.Duration `json:"context_timeout"`
	MigrationVersion    int
}

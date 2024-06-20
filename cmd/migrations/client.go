package migrations

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Client interface {
	Open() (*sqlx.DB, error)
	Close() error
}

type DBClientConfig struct {
	MaxOpenConns        int
	ConnMaxLifetime     time.Duration
	ConnMaxIdleLifetime time.Duration
	urlDBConnection     string
	dbSchemaName        string
}

type client struct {
	DB *sqlx.DB

	logger       *zap.Logger
	clientConfig DBClientConfig
}

// NewClient возвращает новый объект клиента, готового к подключению к БД.
func NewClient(conf *DB) Client {
	clientConfig := DBClientConfig{
		MaxOpenConns:        conf.MaxOpenConns,
		ConnMaxLifetime:     conf.ConnMaxLifetime,
		ConnMaxIdleLifetime: conf.ConnMaxIdleLifetime,
		urlDBConnection:     conf.URL,
		dbSchemaName:        conf.SchemaName,
	}

	return &client{
		clientConfig: clientConfig,
		logger:       zap.L().With(zap.String("pkg", "migration_db")),
	}
}

// Open возвращает открытое, настроенное соединение с БД.
func (c *client) Open() (*sqlx.DB, error) {
	c.logger.Debug("client.Open try connect to pg")
	db, err := sqlx.Open("postgres", c.clientConfig.urlDBConnection)
	if err != nil {
		c.logger.Error("client.Open sql.Open error", zap.Error(err))
		return nil, errors.Wrap(err, fmt.Sprintf("client.Open sql.Open error %v", err))
	}

	err = db.Ping()
	if err != nil {
		c.logger.Error("client.Open db.Open error", zap.Error(err))
		return nil, errors.Wrap(err, fmt.Sprintf("client.Open db.Ping error %v", err))
	}

	c.logger.Debug("db.Ping() pg success")

	db.SetMaxOpenConns(c.clientConfig.MaxOpenConns)
	db.SetMaxIdleConns(c.clientConfig.MaxOpenConns)
	db.SetConnMaxLifetime(c.clientConfig.ConnMaxLifetime)
	db.SetConnMaxIdleTime(c.clientConfig.ConnMaxIdleLifetime)

	c.logger.Debug("client.Open() connect to pg success")

	c.DB = db
	return c.DB, nil
}

// Close закрыть соединение с БД.
func (c *client) Close() error {
	return c.DB.Close()
}

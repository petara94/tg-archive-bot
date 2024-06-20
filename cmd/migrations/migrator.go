package migrations

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"go.uber.org/zap"
)

const (
	migrationsTableName  = "golang-migrate-migrations"
	migrationsSource     = "migration_embedded_sql_files"
	delayForTryMigration = 5 * time.Second
)

//go:embed migrations/*.sql
var FS embed.FS

// MigratorPostgres обертка над golang-migrate.
type MigratorPostgres struct {
	logger       *zap.Logger
	dbClient     Client
	migDriver    source.Driver
	driverConfig postgres.Config

	pathToMigrations string
	migrationVersion uint
}

// NewMigratorPostgres возвращает новый объект мигратора.
func NewMigratorPostgres(conf *DB) (*MigratorPostgres, error) {
	logger := zap.L().With(zap.String("pkg", "migrator_postgres"))
	logger.Debug("MigratorPostgres.NewMigratorPostgres Try connect to create migration driver")

	m := &MigratorPostgres{
		logger:           logger,
		pathToMigrations: conf.MigrationsPath,
		migrationVersion: uint(conf.MigrationVersion),
	}

	// клиентское соединение, готовое к работе
	m.dbClient = NewClient(conf)

	// конфиги драйвера БД
	m.driverConfig = postgres.Config{
		DatabaseName:    conf.DatabaseName,
		SchemaName:      conf.SchemaName,
		MigrationsTable: migrationsTableName,
	}

	// Открываем драйвер файлов миграции
	migDriver, errMig := iofs.New(FS, m.pathToMigrations)
	if errMig != nil {
		m.logger.Error("MigratorPostgres.NewMigratorPostgres iofs.New error", zap.Error(errMig))
		return nil, fmt.Errorf("create migration driver error: %w", errMig)
	}

	m.migDriver = migDriver

	return m, nil
}

// ApplyMigrationsToLast поднять миграции до последней версии.
func (m *MigratorPostgres) ApplyMigrationsToLast() error {
	m.logger.Debug("MigratorPostgres.ApplyMigrationsToLast begin")

	// устанавливаем соединение с БД
	db, errClient := m.dbClient.Open()
	if errClient != nil {
		m.logger.Error("MigratorPostgres.NewMigratorPostgres Open error", zap.Error(errClient))
		return fmt.Errorf("MigratorPostgres.ApplyMigrationsToLast Open error: %w", errClient)
	}

	// закрываем клиентское соединение с БД
	defer func() {
		if errClose := m.dbClient.Close(); errClose != nil {
			m.logger.Error("MigratorPostgres.DropMigrations close db error", zap.Error(errClose))
		}
	}()

	// проверяем схему в БД, и если ее нет, то создаем
	if errPrepare := prepareForMigration(m.driverConfig.SchemaName, db.DB); errPrepare != nil {
		m.logger.Error("MigratorPostgres.ApplyMigrationsToLast prepareForMigration error", zap.Error(errPrepare))
		return errPrepare
	}

	// Настраиваем драйвер миграции (схема и название БД)
	m.logger.Debug("MigratorPostgres.ApplyMigrationsToLast postgres.WithInstance db config", zap.Any("config", m.driverConfig))

	dbDriver, errDb := postgres.WithInstance(db.DB, &m.driverConfig)
	if errDb != nil {
		m.logger.Error("MigratorPostgres.ApplyMigrationsToLast postgres.WithInstance error config",
			zap.Error(errDb),
			zap.Any("config", m.driverConfig))
		return fmt.Errorf("MigratorPostgres.NewMigratorPostgres create db driver db config %+v error : %w", m.driverConfig, errDb)
	}

	// Настраиваем миграцию golang-migrate
	migratorUnit, err := migrate.NewWithInstance(
		migrationsSource,
		m.migDriver,
		m.driverConfig.DatabaseName,
		dbDriver)
	if err != nil {
		m.logger.Warn("MigratorPostgres.ApplyMigrationsToLast NewWithInstance warning", zap.Error(err))
		return err
	}

	// применяем миграции
	if errUp := migratorUnit.Up(); errUp != nil {
		if errors.Is(errUp, migrate.ErrNoChange) {
			m.logger.Warn("MigratorPostgres.ApplyMigrationsToLast Up() warning", zap.Error(errUp))
		} else {
			m.logger.Error("MigratorPostgres.ApplyMigrationsToLast Up() error", zap.Error(errUp))
			return errUp
		}
	}

	return nil
}

// ApplySpecificVersion поднять миграции до конкретной версии.
func (m *MigratorPostgres) ApplySpecificVersion() error {
	m.logger.Debug("MigratorPostgres.ApplySpecificVersion begin")

	// устанавливаем соединение с БД
	db, errClient := m.dbClient.Open()
	if errClient != nil {
		m.logger.Error("MigratorPostgres.NewMigratorPostgres NewClient error", zap.Error(errClient))
		return fmt.Errorf("MigratorPostgres.ApplySpecificVersion Open error: %w", errClient)
	}

	// закрываем клиентское соединение с БД
	defer func() {
		if errClose := m.dbClient.Close(); errClose != nil {
			m.logger.Error("MigratorPostgres.ApplySpecificVersion close db error", zap.Error(errClose))
		}
	}()

	// проверяем схему в БД, и если ее нет, то создаем
	if errPrepare := prepareForMigration(m.driverConfig.SchemaName, db.DB); errPrepare != nil {
		m.logger.Error("MigratorPostgres.NewMigratorPostgres prepareForMigration error", zap.Error(errPrepare))
		return errPrepare
	}

	// Настраиваем драйвер миграции (схема и название БД)
	m.logger.Debug("MigratorPostgres.ApplySpecificVersion postgres.WithInstance db config", zap.Any("config", m.driverConfig))

	dbDriver, errDb := postgres.WithInstance(db.DB, &m.driverConfig)
	if errDb != nil {
		m.logger.Error("MigratorPostgres.ApplySpecificVersion postgres.WithInstance error config",
			zap.Error(errDb),
			zap.Any("config", m.driverConfig))
		return fmt.Errorf("MigratorPostgres.ApplySpecificVersion create db driver db config %+v error : %w", m.driverConfig, errDb)
	}

	// Настраиваем миграцию golang-migrate
	migratorUnit, err := migrate.NewWithInstance(
		migrationsSource,
		m.migDriver,
		m.driverConfig.DatabaseName,
		dbDriver)
	if err != nil {
		m.logger.Warn("MigratorPostgres.ApplySpecificVersion NewWithInstance warning", zap.Error(err))
		return err
	}

	// Получаем текущую версию миграции
	version, dirty, errVersion := migratorUnit.Version()
	if errVersion != nil && !errors.Is(errVersion, migrate.ErrNilVersion) {
		m.logger.Error("MigratorPostgres.ApplySpecificVersion Version() error", zap.Error(errVersion))
		return errVersion
	}

	// если dirty - true https://stackoverflow.com/questions/59616263/dirty-database-version-error-when-using-golang-migrate
	if dirty {
		errDirty := migrate.ErrDirty{Version: int(version)}
		m.logger.Error(
			fmt.Sprintf("MigratorPostgres.ApplySpecificVersion Version() check migrations version: %d and %d", version, version+1),
			zap.Error(errDirty))
		return err
	}

	// Проверяем ready to run миграции, последнюю версию миграции.
	_, migName, errRead := m.migDriver.ReadUp(m.migrationVersion)
	if errRead != nil {
		m.logger.Error(
			fmt.Sprintf("MigratorPostgres.ApplySpecificVersion ReadUp() mig name : %s current version: %d your version : %d", migName, version, m.migrationVersion),
			zap.Error(errRead))
		return err
	}

	m.logger.Warn(fmt.Sprintf("MigratorPostgres.ApplySpecificVersion current version %d next migration %s", version, migName))

	// накатываем миграцию специфической версии
	if errMigrate := migratorUnit.Migrate(m.migrationVersion); errMigrate != nil {
		if errors.Is(errMigrate, migrate.ErrNoChange) {
			m.logger.Warn("MigratorPostgres.ApplySpecificVersion Up() warning", zap.Error(errMigrate))
		} else {
			m.logger.Warn("MigratorPostgres.ApplySpecificVersion Up() error", zap.Error(errMigrate))
			return err
		}
	}

	return nil
}

// DropMigrations откатывает миграции до НУЛЕВОЙ версии (для тестов). Оставит только схему и таблицу с миграциями.
func (m *MigratorPostgres) DropMigrations() error {
	m.logger.Debug("MigratorPostgres.DropMigrations begin")

	// устанавливаем соединение с БД
	db, errClient := m.dbClient.Open()
	if errClient != nil {
		m.logger.Error("MigratorPostgres.NewMigratorPostgres NewClient error", zap.Error(errClient))
		return fmt.Errorf("MigratorPostgres.DropMigrations Open error: %w", errClient)
	}

	// закрываем клиентское соединение с БД
	defer func() {
		if errClose := m.dbClient.Close(); errClose != nil {
			m.logger.Error("MigratorPostgres.DropMigrations close db error", zap.Error(errClose))
		}
	}()

	// проверяем схему в БД, и если ее нет, то создаем
	if errPrepare := prepareForMigration(m.driverConfig.SchemaName, db.DB); errPrepare != nil {
		m.logger.Error("MigratorPostgres.NewMigratorPostgres prepareForMigration error", zap.Error(errPrepare))
		return errPrepare
	}

	// Настраиваем драйвер миграции (схема и название БД)
	m.logger.Debug("MigratorPostgres.DropMigrations postgres.WithInstance db config", zap.Any("config", m.driverConfig))

	dbDriver, errDb := postgres.WithInstance(db.DB, &m.driverConfig)
	if errDb != nil {
		m.logger.Error("MigratorPostgres.DropMigrations postgres.WithInstance error config",
			zap.Error(errDb),
			zap.Any("config", m.driverConfig))
		return fmt.Errorf("MigratorPostgres.DropMigrations create db driver db config %+v error : %w", m.driverConfig, errDb)
	}

	// Настраиваем миграцию golang-migrate
	migratorUnit, err := migrate.NewWithInstance(
		migrationsSource,
		m.migDriver,
		m.driverConfig.DatabaseName,
		dbDriver)
	if err != nil {
		m.logger.Warn("MigratorPostgres.DropMigrations NewWithInstance warning", zap.Error(err))
		return err
	}

	// проверяем схему в БД, и если ее нет, то создаем
	if errPrepare := prepareForMigration(m.driverConfig.SchemaName, db.DB); errPrepare != nil {
		m.logger.Error("MigratorPostgres.DropMigrations prepareForMigration error", zap.Error(errPrepare))
		return err
	}

	// применяем миграции
	if errDown := migratorUnit.Down(); errDown != nil {
		if errors.Is(errDown, migrate.ErrNoChange) {
			m.logger.Warn("Migrator.DropMigrations Down() warning", zap.Error(errDown))
		} else {
			m.logger.Error("Migrator.DropMigrations Down() error", zap.Error(errDown))
			return errDown
		}
	}

	return nil
}

// prepareForMigration подготавливаем БД к миграции (создаем схему).
func prepareForMigration(schema string, db *sql.DB) error {
	_, err := db.Exec(`CREATE SCHEMA IF NOT EXISTS ` + schema)
	if err != nil {
		return err
	}

	_, err = db.Exec(`SET search_path TO ` + schema)
	if err != nil {
		return err
	}

	return nil
}

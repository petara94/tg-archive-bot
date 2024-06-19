package log

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var rootLogger *zap.Logger
var logLevel zap.AtomicLevel
var mu sync.Mutex

// RootLogger Создает корневой логгер, использовать только для создания первого корневого логгера
// для создания дочерних логгеров используйте zap.L().With( .... )
func RootLogger(env string, level string) (*zap.Logger, error) {
	mu.Lock()
	defer mu.Unlock()

	if rootLogger != nil {
		return rootLogger, nil
	}

	var config zap.Config
	if env == "development" {
		config = zap.NewDevelopmentConfig()
	} else {
		config = zap.NewProductionConfig()
	}

	var err error
	logLevel = zap.NewAtomicLevel()
	if err = logLevel.UnmarshalText([]byte(level)); err != nil {
		return nil, err
	}
	config.Level = logLevel

	rootLogger, err = config.Build()
	if err != nil {
		return nil, err
	}

	zap.ReplaceGlobals(rootLogger)

	return rootLogger, nil
}

// SetLogLevel изменить уровень логирования корневого и всех дочерних логеров
func SetLogLevel(levelText string) error {
	level, err := zapcore.ParseLevel(levelText)
	if err != nil {
		return err
	}
	logLevel.SetLevel(level)
	return nil
}

// GetLogLevel получить уровень логирования
func GetLogLevel() string {
	return logLevel.String()
}

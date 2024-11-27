package main

import (
	"fmt"
	"museff/internal/app"
	"strings"

	"github.com/hagarashi8/go-envbinder"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

type DBCreds struct {
	Host     string
	Port     int
	Sslmode  string
	Username string
	Password string
	TimeZone string
	Database string
}

func (c *DBCreds) GenerateDSN() string {
	return fmt.Sprintf("host=%s port=%d sslmode=%s user=%s database=%s password=%s TimeZone=%s", c.Host, c.Port, c.Sslmode, c.Username, c.Database, c.Password, c.TimeZone)
}

// @title           Музыкальная библиотека
// @version         1.0
func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	lg := e.Logger
	lg.SetLevel(log.INFO)
	var port int
	b := envbinder.EnvBinder{}
	dbc := DBCreds{}
	cfg := app.Config{
		Echo: e,
	}
	var loglvl string

	lg.Info("Подготовка к запуску")
	lg.Debug("Парсим переменные среды")
	// Насртойки подключения к базе данных
	err := b.StringOrDef(&dbc.Sslmode, "POSTGRES_SSLMODE", "disable").
		IntOrDef(&dbc.Port, "POSTGRES_PORT", 5432).
		StringOrDef(&dbc.Host, "POSTGRES_HOST", "localhost").
		StringOrDef(&dbc.Username, "POSTGRES_USER", "postgres").
		StringOrDef(&dbc.Password, "POSTGRES_PASSWORD", "").
		StringOrDef(&dbc.TimeZone, "POSTGRES_TIMEZONE", "Etc/GMT").
		StringOrDef(&dbc.Database, "POSTGRES_DB", "postgres").
		// Настройки приложения
		IntOrDef(&port, "LIBRARY_SERVICE_PORT", 3001).
		String(&cfg.MusicInfoServiceUrl, "MUSIC_INFO_SERVICE_ADDRESS").
		StringOrDef(&loglvl, "LIBRARY_SERVICE_LOG_LEVEL", "INFO").
		BindError()
	if err != nil {
		lg.Fatalf("Ошибка парсинга переменных среды: %s\n\tCFG: %v\n\tDSN:%s", err.Error(), cfg, dbc.GenerateDSN())
	}
	var ok bool

	cfg.LogLevel, ok = ParseLogLevel(loglvl)

	if !ok {
		lg.Warnf("Неправильный уровень логирования: %s\nВозможные варианты: DEBUG, INFO, WARN, ERROR, OFF. Продолжаем с уровнем логирования по умолчанию(INFO)", loglvl)
		cfg.LogLevel = log.INFO
	}

	lg.Debug("Создаём объект приложения")
	cfg.DSN = dbc.GenerateDSN()
	a, err := app.NewApp(cfg)

	if err != nil {
		lg.Fatalf("Не удалось подготовить сервис с текущим конфигом\n\tОшибка: %s\n\tCFG: %v\n\tDSN: %s", err.Error(), &cfg, dbc.GenerateDSN())
	}

	lg.Info("Запускаем сервис...")
	lg.Fatal(a.Start(fmt.Sprintf(":%d", port)))
}

func ParseLogLevel(lvl string) (l log.Lvl, ok bool) {
	switch strings.ToLower(lvl) {
	case "debug":
		l = log.DEBUG
		ok = true
		return
	case "info":
		l = log.INFO
		ok = true
		return
	case "warn":
		l = log.WARN
		ok = true
		return
	case "error":
		l = log.ERROR
		ok = true
		return
	case "off":
		l = log.OFF
		ok = true
		return
	default:
		ok = false
		return
	}
}

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	// Библиотека для миграций
	"github.com/golang-migrate/migrate/v4"
	"github.com/ilyakaznacheev/cleanenv"

	// Драйвер для выполнения миграций SQLite 3
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	// Драйвер для получения миграций из файлов
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Config struct {
	Database struct {
		URL             string `yaml:"url" env:"DATABASE_URL" env-required:"true"`
		MigrationsPath  string `yaml:"migrations_path" env:"MIGRATIONS_PATH" env-required:"true"`
		MigrationsTable string `yaml:"migrations_table" env:"MIGRATIONS_TABLE" env-default:"migrations"`
	} `yaml:"database"`
}

func main() {
	cfg := MustLoad()

	if cfg.Database.URL == "" {
		panic("storage-path is required")
	}
	if cfg.Database.MigrationsPath == "" {
		panic("migrations-path is required")
	}

	println(cfg.Database.MigrationsPath)

	m, err := migrate.New(
		"file://"+cfg.Database.MigrationsPath,
		cfg.Database.URL,
	)
	if err != nil {
		panic(err)
	}
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply")

			return
		}

		panic(err)
	}

	fmt.Println("migrations applied")
	os.Exit(0)
}

func MustLoad() *Config {
	path := fetchConfigPath()
	if path == "" {
		// Если путь не указан, пробуем загрузить из .env или переменных окружения
		var cfg Config
		err := cleanenv.ReadEnv(&cfg)
		if err != nil {
			panic("failed to read config from env: " + err.Error())
		}
		return &cfg
	}

	return MustLoadByPath(path)
}

func MustLoadByPath(configPath string) *Config {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}

	return &cfg
}

func fetchConfigPath() string {
	var res string

	// Пробуем получить путь из флага
	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	// Если не задано флагом, пробуем из переменной окружения
	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}

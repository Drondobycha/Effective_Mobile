package config

import (
	"flag"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	DatabaseURL string `yaml:"database_url" env-required:"true"`
	Port        string `yaml:"port" env-default:"8080"`
	LogLevel    string `yaml:"log_level" env-default:"info"`
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

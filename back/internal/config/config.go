package config

import (
	"errors"
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

var (
	errEmptyConfigPath = errors.New("config path cannot be empty")
)

type Config struct {
	App    AppCfg    `yaml:"app"`
	Server ServerCfg `yaml:"server"`
	DB     DBCfg     `yaml:"db"`
	TG     TGCfg     `yaml:"tg"`
}

type AppCfg struct {
	Name       string `yaml:"name"`
	Env        string `yaml:"enviroment"`
	PubURL     string `yaml:"public_url"`
	MiniAppURL string `yaml:"mini_app_url"`
}

type ServerCfg struct {
	Host         string        `yaml:"host"`
	Port         string        `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout"`
}

type DBCfg struct {
	User     string `env:"DB_USER"`
	Password string `env:"DB_PASS"`
	Host     string `env:"DB_HOST"`
	Port     string `env:"DB_PORT"`
	Name     string `env:"DB_NAME"`
}

type TGCfg struct {
	Token   string        `env:"BOT_TOKEN"`
	ExpTime time.Duration `yaml:"expiration_time"`
}

func MustLoad() Config {
	path := fetchConfigPath()

	if path == "" {
		panic(errEmptyConfigPath)
	}

	if _, err := os.Stat(path); err != nil {
		panic(err)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic(err)
	}

	if err := cleanenv.ReadConfig(".env", &cfg.DB); err != nil {
		panic(err)
	}

	if err := cleanenv.ReadConfig(".env", &cfg.TG); err != nil {
		panic(err)
	}

	return cfg
}

func fetchConfigPath() string {
	var cfgPath string

	flag.StringVar(&cfgPath, "config", "", "path to config file")
	flag.Parse()

	if cfgPath == "" {
		cfgPath = os.Getenv("CONFIG_PATH")
	}

	return cfgPath
}

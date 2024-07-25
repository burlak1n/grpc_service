package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env string `yaml:"env" env-default:"local"`
	StoragePath string `yaml:"storage_path" env-required:"true"`
	TokenTTL time.Duration `yaml:"token_ttl" env-required:"true"`
	GRPC GRPCConfig `yaml:"grpc"`
}

type GRPCConfig struct {
	Port int `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

func MustLoad() (cfg *Config) {
	path := fetchConfigPath()
	if path == "" {
		panic("config path is empty")
	}
	
	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file does not exist: " + path)
	}

	cfg = new(Config)

	if err := cleanenv.ReadConfig(path, cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}

	return cfg
}

// flag > env > default
// default value is empty string
func fetchConfigPath() (res string) {

	// --config="path/to/config.yaml"
	flag.StringVar(&res, "config", "../../config/local.yaml", "path to config file")
	flag.Parse()

	if res == "" {
		os.Getenv("CONFIG_PATH")
	}
	
	return res
}
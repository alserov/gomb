package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Addr string `yaml:"addr"`
}

const (
	CONFIG_PATH = "./config/default.conf.yaml"

	ENV_ADDR = "GOMB_ADDR"
)

func MustLoad() *Config {
	b, err := os.ReadFile(CONFIG_PATH)
	if err != nil {
		panic("failed to read config file: " + err.Error())
	}

	var cfg Config
	if err = yaml.Unmarshal(b, &cfg); err != nil {
		panic("failed to unmarshal config file: " + err.Error())
	}

	return &cfg
}

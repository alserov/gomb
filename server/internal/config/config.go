package config

import (
	"gopkg.in/yaml.v3"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Addr             string `yaml:"addr"`
	PredefinedTopics map[string]uint
}

const (
	DEFAULT_BUFFER_SIZE = 32
	CONFIG_PATH         = "./config/default.conf.yaml"

	ENV_ADDR   = "GOMB_ADDR"
	ENV_TOPICS = "GOMB_TOPICS"
)

func MustLoad() *Config {
	b, err := os.ReadFile(CONFIG_PATH)
	if err != nil {
		panic("failed to read config file: " + err.Error())
	}

	cfg := Config{
		PredefinedTopics: map[string]uint{},
	}
	if err = yaml.Unmarshal(b, &cfg); err != nil {
		panic("failed to unmarshal config file: " + err.Error())
	}

	envAddr := os.Getenv(ENV_ADDR)
	if envAddr != "" {
		cfg.Addr = envAddr
	}

	topics := os.Getenv(ENV_TOPICS)
	if t := strings.Split(topics, ","); topics != "" {
		for _, topic := range t {
			topicParams := strings.Split(topic, ":")
			if len(topicParams) > 1 {
				buffSize, err := strconv.Atoi(topicParams[1])
				if err != nil {
					panic("invalid buffer size: " + topicParams[1])
				}

				cfg.PredefinedTopics[topicParams[0]] = uint(buffSize)
				continue
			}

			cfg.PredefinedTopics[topicParams[0]] = DEFAULT_BUFFER_SIZE
		}
	}

	return &cfg
}

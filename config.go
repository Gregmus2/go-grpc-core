package core

import "github.com/caarlos0/env"

type Config struct {
	LogLevel      string   `env:"LOG_LEVEL" envDefault:"debug"`
	ListenAddress []string `env:"HOSTS" envSeparator:","`
}

func NewConfig() (*Config, error) {
	c := new(Config)
	if err := env.Parse(c); err != nil {
		return nil, err
	}

	return c, nil
}

package core

import (
	"github.com/caarlos0/env"
	"os"
)

type Config struct {
	LogLevel      string   `env:"LOG_LEVEL" envDefault:"debug"`
	ListenAddress []string `env:"HOSTS" envSeparator:","`
}

func NewConfig() (*Config, error) {
	c := new(Config)
	if err := env.Parse(c); err != nil {
		return nil, err
	}

	var hosts []string
	if len(os.Args) > 1 {
		hosts = os.Args[1:]
	}
	if len(hosts) > 0 {
		c.ListenAddress = hosts
	}

	return c, nil
}

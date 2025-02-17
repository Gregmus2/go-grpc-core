package core

import (
	"github.com/caarlos0/env"
	"os"
)

type TLS struct {
	Certificate    string `env:"CERTIFICATE"`
	CertificateKey string `env:"CERTIFICATE_KEY"`
}

type Config struct {
	LogLevel      string   `env:"LOG_LEVEL" envDefault:"debug"`
	ListenAddress []string `env:"HOSTS" envSeparator:","`
	TLS           TLS
}

func NewConfig() (*Config, error) {
	tls := TLS{}
	if err := env.Parse(&tls); err != nil {
		return nil, err
	}

	c := new(Config)
	if err := env.Parse(c); err != nil {
		return nil, err
	}

	c.TLS = tls

	var hosts []string
	if len(os.Args) > 1 {
		hosts = os.Args[1:]
	}
	if len(hosts) > 0 {
		c.ListenAddress = hosts
	}

	return c, nil
}

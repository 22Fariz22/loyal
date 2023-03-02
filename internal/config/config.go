package config

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v7"
)

type Config struct {
	RunAddress           string `env:"RUN_ADDRESS"`
	DatabaseURI          string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

func NewConfig() *Config {
	cfg := Config{}

	flag.StringVar(&cfg.RunAddress, "a", "", "server address") //localhost:8080
	flag.StringVar(&cfg.DatabaseURI, "d", "", "database address")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "", "accural system")

	flag.Parse()

	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
	}

	return &cfg
}

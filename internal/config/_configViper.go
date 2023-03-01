package config

import (
	"flag"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	RunAddress           string `env:"RUN_ADDRESS"` //envDefault:":8080"
	DatabaseURI          string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

func NewConfig() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.RunAddress, "a", "localhost:8080", "server address")
	flag.StringVar(&cfg.DatabaseURI, "d", "", "database address")        //postgres://postgres:55555@127.0.0.1:5432/gophermart
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "", "accrual system") // http://127.0.0.1:8080

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	return cfg
}

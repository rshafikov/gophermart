package app

import (
	"github.com/caarlos0/env/v6"
	"log"
)

type envParams struct {
	RunAddress     string `env:"RUN_ADDRESS"`
	LogLevel       string `env:"LOG_LEVEL"`
	DatabaseURI    string `env:"DATABASE_DSN"`
	AccrualAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

var Env envParams

func ParseEnv() error {
	err := env.Parse(&Env)

	if err != nil {
		log.Println("Unable to parse ENV:", err)
		return err
	}

	return nil
}

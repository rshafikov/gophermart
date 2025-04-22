package app

import (
	"log"
	"net"
)

func InitConfig() {
	InitFlags()

	err := ParseEnv()

	if err != nil {
		log.Fatal("unable parse ENV:", err)
	}

	if Env.RunAddress != "" {
		host, port, err := net.SplitHostPort(Env.RunAddress)
		if err != nil {
			log.Fatal("invalid RUN_ADDRESS environment variable: ", Env.RunAddress)
		}
		Config.RunAddress.Host = host
		Config.RunAddress.Port = port
	}

	if Env.AccrualAddress != "" {
		host, port, err := net.SplitHostPort(Env.AccrualAddress)
		if err != nil {
			log.Fatal("invalid RUN_ADDRESS environment variable: ", Env.AccrualAddress)
		}
		Config.AccrualAddress.Host = host
		Config.AccrualAddress.Port = port
	}

	if Env.LogLevel != "" {
		Config.LogLevel = Env.LogLevel
	}

	if Env.DatabaseURI != "" {
		err := Config.DB.Set(Env.DatabaseURI)
		if err != nil {
			log.Fatal("Unable to parse DATABASE_URI environment variable: ", Env.DatabaseURI)
		}
	}
}

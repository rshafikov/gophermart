package app

import (
	"fmt"
	"github.com/rshafikov/gophermart/internal/core/logger"
	"log"
	"net"
	"net/url"
	"strings"
)

func InitConfig() {
	InitFlags()

	err := ParseEnv()

	if err != nil {
		log.Fatal("unable parse ENV:", err)
	}

	if Env.RunAddress != "" {
		host, port, err := ParseHostPort(Env.RunAddress)
		if err != nil {
			log.Fatal("invalid RUN_ADDRESS environment variable: ", Env.RunAddress, err)
		}
		Config.RunAddress.Host = host
		Config.RunAddress.Port = port
	}

	if Env.AccrualAddress != "" {
		host, port, err := ParseHostPort(Env.AccrualAddress)
		if err != nil {
			log.Fatal("invalid ACCRUAL_SYSTEM_ADDRESS environment variable: ", Env.AccrualAddress, err)
		}
		Config.AccrualAddress.Host = host
		Config.AccrualAddress.Port = port
	}

	if Env.LogLevel != "" {
		Config.LogLevel = Env.LogLevel
	}

	if Env.DatabaseURI != "" {
		err = Config.DB.Set(Env.DatabaseURI)
		if err != nil {
			log.Fatalf("unable to parse DATABASE_URI '%s': %s", Env.DatabaseURI, err)
		}
	}

	dbURI := Config.DB.String()
	Config.DB.URI = dbURI

	err = logger.Initialize(Config.LogLevel)
	if err != nil {
		log.Fatal("unable to initialize logger:", err)
	}

	initMessage := "\033[1;36m╭────────────────────────────────────────\033[0m\n" +
		"\033[1;36m│ \033[1;34m🚀 Server Initialized Successfully \033[0m\n" +
		"\033[1;36m├────────────────────────────────────────\033[0m\n" +
		"\033[1;36m│ \033[1;33m📡 Server Address:   \033[0;37m%-39s\033[0m\n" +
		"\033[1;36m│ \033[1;33m🐘 Database DSN:     \033[0;37m%-39s\033[0m\n" +
		"\033[1;36m│ \033[1;33m◉  Accrual Address:  \033[0;37m%-39s\033[0m\n" +
		"\033[1;36m│ \033[1;33m📝 Logging Level:    \033[0;37m%-39s\033[0m\n" +
		"\033[1;36m╰────────────────────────────────────────\033[0m\n"

	fmt.Printf(
		initMessage,
		Config.RunAddress.String(),
		Config.DB.URI,
		Config.AccrualAddress.String(),
		Config.LogLevel,
	)

	if dbURI == "" {
		log.Fatal("DATABASE_URI is empty, please set it using ENV or CLI flag '-d'")
	}
}

func ParseHostPort(rawURL string) (host, port string, err error) {
	rawURL = strings.TrimSuffix(rawURL, "/")

	parsedURL, err := url.Parse(rawURL)
	if err == nil && parsedURL.Host != "" && parsedURL.Scheme != "" {
		host, port, err = net.SplitHostPort(parsedURL.Host)
		if err != nil {
			return "", "", fmt.Errorf("failed to split host and port: %w", err)
		}
		return host, port, nil
	}

	parsedURL, err = url.Parse("http://" + rawURL)
	if err == nil && parsedURL.Host != "" {
		host, port, err = net.SplitHostPort(parsedURL.Host)
		if err != nil {
			return "", "", fmt.Errorf("failed to split host and port with scheme: %w", err)
		}
		return host, port, nil
	}

	host, port, err = net.SplitHostPort(rawURL)
	if err != nil {
		if !strings.Contains(rawURL, ":") {
			return rawURL, "", nil
		}
		return "", "", fmt.Errorf("failed to parse host and port: %w", err)
	}

	return host, port, nil
}

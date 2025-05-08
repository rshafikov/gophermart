package app

import (
	"errors"
	"flag"
	"fmt"
	"strconv"
	"strings"
)

const (
	defaultServerHost = "localhost"
	defaultServerPort = "8080"
	defaultLogLevel   = "info"
)

type dbSettings struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	URI      string
}

// postgres://postgres:password@localhost:5432/postgres
func (d *dbSettings) String() string {
	url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", d.User, d.Password, d.Host, d.Port, d.Name)

	if url == "postgres://:@:/" {
		url = ""
	}

	return url
}

func (d *dbSettings) Set(s string) error {
	parsed := strings.Split(s, "://")
	if len(parsed) != 2 || parsed[0] == "" {
		return errors.New("supported format: postgres://user:password@host:port/dbname")
	}

	credentialsHostDB := parsed[1]
	atIndex := strings.LastIndex(credentialsHostDB, "@")
	if atIndex == -1 {
		return errors.New("invalid url: missing '@'")
	}

	credentials := credentialsHostDB[:atIndex]
	hostDB := credentialsHostDB[atIndex+1:]

	creds := strings.SplitN(credentials, ":", 2)
	if len(creds) != 2 {
		return errors.New("invalid url: missing or invalid credentials")
	}
	d.User = creds[0]
	d.Password = creds[1]

	slashIndex := strings.LastIndex(hostDB, "/")
	if slashIndex == -1 {
		return errors.New("invalid url: missing '/' before dbname")
	}
	hostPort := hostDB[:slashIndex]
	d.Name = hostDB[slashIndex+1:]

	colonIndex := strings.LastIndex(hostPort, ":")
	if colonIndex == -1 {
		return errors.New("invalid url: missing ':' in host:port")
	}
	d.Host = hostPort[:colonIndex]
	d.Port = hostPort[colonIndex+1:]

	return nil
}

type netAddr struct {
	Host string
	Port string
}

func (na *netAddr) String() string {
	return fmt.Sprintf("%s:%s", na.Host, na.Port)
}

func (na *netAddr) Set(s string) error {
	hp := strings.Split(s, ":")
	if len(hp) != 2 {
		return errors.New("supported format: host:port")
	}
	_, err := strconv.Atoi(hp[1])
	if err != nil {
		return err
	}
	na.Host = hp[0]
	na.Port = hp[1]
	return nil
}

type defaultConfig struct {
	DB             dbSettings
	RunAddress     netAddr
	AccrualAddress netAddr
	LogLevel       string
}

var Config = defaultConfig{
	DB:             dbSettings{},
	RunAddress:     netAddr{Host: defaultServerHost, Port: defaultServerPort},
	AccrualAddress: netAddr{},
	LogLevel:       defaultLogLevel,
}

func InitFlags() {
	_ = flag.Value(&Config.RunAddress)
	flag.Var(&Config.RunAddress, "a", "run address")

	_ = flag.Value(&Config.AccrualAddress)
	flag.Var(&Config.AccrualAddress, "r", "address of accrual system")

	_ = flag.Value(&Config.DB)
	flag.Var(&Config.DB, "d", "database URI")

	flag.StringVar(&Config.LogLevel, "l", defaultLogLevel, "log level")

	flag.Parse()
}

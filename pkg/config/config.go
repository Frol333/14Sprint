package config

import (
	"flag"
	"os"
	"strconv"

	"github.com/rs/zerolog/log"
)

var Password string

type Config struct {
	LogLevel string
	Port     int64
	DBFile   string
}

func New() *Config {
	c := &Config{}

	flag.StringVar(&c.LogLevel, "logLevel", LookupEnvString("LOG_LEVEL", "debug"), "Set log level")
	flag.StringVar(&c.DBFile, "dabFile", LookupEnvString("TODO_DBFILE", "scheduler.db"), "Set log DB file")
	flag.Int64Var(&c.Port, "port", LookupEnvInt("TODO_PORT", 7540), "Set port")
	flag.StringVar(&Password, "password", LookupEnvString("TODO_PASSWORD", "pass"), "Set default password")

	flag.Parse()

	return c
}

func LookupEnvString(envName, def string) string {
	if v := os.Getenv(envName); v != "" {
		return v
	}
	return def
}

func LookupEnvInt(envName string, def int64) int64 {
	if v := os.Getenv(envName); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			return int64(p)
		}
		log.Fatal().Str("envName", envName).Int64("default", def)
	}
	return def
}

package main

import (
	"fmt"

	"github.com/Frol333/14Sprint/pkg/api"
	"github.com/Frol333/14Sprint/pkg/config"
	"github.com/Frol333/14Sprint/pkg/db"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func initLogger(logLevel string) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	ll, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse log level with err")
		panic(err)
	}
	zerolog.SetGlobalLevel(ll)
}

var (
	shutdown = make(chan struct{})
)

func main() {
	cfg := config.New()
	initLogger(cfg.LogLevel)
	log.Info().Msg("the configuration was initialized successfully")

	// Инициализация БД
	if err := db.Init(cfg.DBFile); err != nil {
		log.Panic().Err(err).Msg("Failed to init db")
	}
	defer db.DB.Close()

	log.Info().Msg("the DB was initialized successfully")
	listenAddr := fmt.Sprintf("127.0.0.1:%v", cfg.Port)
	srv := api.New(listenAddr)

	log.Info().Msgf("Start running server on port %v", cfg.Port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal().Err(err).Msg("Server wasn't starting")
	}

	log.Info().Msg("Server was stopping")
}

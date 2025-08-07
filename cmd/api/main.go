package main

import (
	"log"

	"github.com/Neroframe/crypto-tracker/config"
	"github.com/Neroframe/crypto-tracker/pkg/logger"
)

func main() {
	cfg, err := config.Load("config/dev.yaml")
	if err != nil {
		log.Fatal(err)
	}

	log := logger.New(logger.Config(cfg.Log))
	log.Info("config loaded", "version", cfg.Version)
}

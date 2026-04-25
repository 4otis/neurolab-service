package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/4otis/neurolab-service/config"
	"github.com/4otis/neurolab-service/internal/app"
)

// @title nuerolab-service
// @version 1.0
// @description сервис для автоматизации выдачи/проверки/приема лабораторных работ
// @host localhost:8080
// @BasePath /

func main() {
	cfg := config.Load()

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("failed to create application: %v", err)
	}

	err = application.Start()
	if err != nil {
		log.Fatalf("error while starting app: %v", err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	application.Stop()
}

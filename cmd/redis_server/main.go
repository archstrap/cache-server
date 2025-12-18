package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/archstrap/cache-server/internal/config"
	"github.com/archstrap/cache-server/internal/tcpserver"
)

func main() {

	appConfig, err := config.NewAppConfig()
	if err != nil {
		log.Fatalln("Error reading config file: ", err.Error())
	}

	rootContext, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	server := tcpserver.NewServerFromConfig(appConfig)

	server.Start(rootContext)

}

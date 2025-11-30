package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/codecrafters-io/redis-starter-go/internal/config"
	"github.com/codecrafters-io/redis-starter-go/internal/tcpserver"
)

func main() {

	appConfig, err := config.NewAppConfig()
	if err != nil {
		log.Fatalln("Error reading config file: ", err.Error())
	}

	shutdownSignal := make(chan struct{})

	go graceFullyShutDown(shutdownSignal)

	serverAddress := fmt.Sprintf("%s:%s", appConfig.GetHost(), appConfig.GetPort())

	server := tcpserver.NewServer(serverAddress, 10)

	server.Start(shutdownSignal)

}

func graceFullyShutDown(shutDownSignal chan struct{}) {

	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	log.Println("App is running press CTRL+C to exit")

	receivedSignal := <-sigChan
	log.Println("Received signal:", receivedSignal)
	log.Println("App is shutting down gracefully..................")

	close(shutDownSignal)
}

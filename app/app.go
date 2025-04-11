package main

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/config"
	"github.com/codecrafters-io/redis-starter-go/app/tcpserver"
	"log"
)

func main() {

	appConfig, err := config.NewAppConfig()
	if err != nil {
		log.Fatalln("Error reading config file: ", err.Error())
	}
	serverAddress := fmt.Sprintf("%s:%s", appConfig.GetHost(), appConfig.GetPort())

	server := tcpserver.NewServer(serverAddress, 10)

	server.Start()

}

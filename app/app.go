package main

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/config"
	"github.com/codecrafters-io/redis-starter-go/app/tcpserver"
	"os"
)

func main() {

	appConfig, err := config.NewAppConfig()
	if err != nil {
		fmt.Println("Error reading config file: ", err.Error())
		os.Exit(1)
	}
	serverAddress := fmt.Sprintf("%s:%s", appConfig.GetHost(), appConfig.GetPort())

	server := tcpserver.NewServer(serverAddress, 10)

	server.Start()

}

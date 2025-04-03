package main

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/config"
	"github.com/codecrafters-io/redis-starter-go/app/tcpserver"
	"io"
	"net"
	"os"
)

func main() {

	appConfig, err := config.NewAppConfig()
	if err != nil {
		fmt.Println("Error reading config file: ", err.Error())
		os.Exit(1)
	}
	serverAddress := fmt.Sprintf("%s:%s", appConfig.GetHost(), appConfig.GetPort())

	server, err := tcpserver.NewServer(serverAddress)
	if err != nil {
		fmt.Println("Error creating server: ", err.Error())
		os.Exit(1)
	}

	server.Start()
}

func startServerAndReturnConnection(serverAddress string) net.Listener {
	listener, err := net.Listen("tcp", serverAddress)
	if err != nil {
		fmt.Println("failed to bind to server address", serverAddress, " : ", err)
		os.Exit(1)
	}
	fmt.Printf("server started at %s\n", serverAddress)
	return listener
}

func handleServerRequest(l net.Listener) {
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {

	fmt.Printf("Handling client: %v\n", conn.RemoteAddr())

	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)

		if err != nil {
			fmt.Println("Error", err)
			conn.Close()
			fmt.Printf("Connection from: %s is closed\n", conn.RemoteAddr())
			if err == io.EOF {
				break
			}
		}

		receivedMessage := string(buf[:n])
		fmt.Printf("client address:%s , sent message:%s \n", conn.RemoteAddr(), receivedMessage)

		message := []byte("+PONG\r\n")
		conn.Write(message)
	}

}

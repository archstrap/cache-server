package main

import (
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	host := "0.0.0.0"
	port := "6379"

	serverAddress := fmt.Sprintf("%s:%s", host, port)
	l := startServerAndReturnConnection(serverAddress)
	handleServerRequest(l)
}

func startServerAndReturnConnection(serverAddress string) net.Listener {
	listener, err := net.Listen("tcp", serverAddress)
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
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

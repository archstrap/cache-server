package tcpserver

import (
	"fmt"
	"net"
)

type Server struct {
	listener net.Listener
	address  string
}

func (server *Server) Start() {

	fmt.Printf("server started at %s\n", server.address)
	connection, err := server.listener.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
	}
	fmt.Printf("Handling client: %v\n", connection.RemoteAddr())
	for {
		buf := make([]byte, 1024)
		n, err := connection.Read(buf)
		if err != nil {
			fmt.Println("Error", err)
		}

		receivedMessage := string(buf[:n])
		fmt.Println(receivedMessage)
	}
}

func NewServer(address string) (*Server, error) {
	conn, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Printf("Unable to bind %v to tcp server", address)
	}
	return &Server{listener: conn, address: address}, nil
}

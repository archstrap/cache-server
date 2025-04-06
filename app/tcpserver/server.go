package tcpserver

import (
	"github.com/codecrafters-io/redis-starter-go/app/eventloop"
	"log"
	"net"
)

type Server struct {
	address   string
	eventLoop *eventloop.EventLoop
}

func NewServer(address string, maxParallelization int) *Server {
	return &Server{
		address: address,
		eventLoop: &eventloop.EventLoop{
			Tasks: make(chan eventloop.RedisTask, maxParallelization)}}
}

func (server *Server) Start() {
	// 1. address -> start the server on the preferred location
	listener, err := net.Listen("tcp", server.address)
	if err != nil {
		log.Fatalf("Failed to start server at address: %s. Error Details: %v\n", server.address, err)
	}
	log.Println("Server started at:", server.address)

	// 2. run the event loop in a separate go-routine
	eventLoop := server.eventLoop
	go eventLoop.Start()

	// 3. current go-routine will monitor the incoming tasks

	for {
		conn, err := listener.Accept()
		if err != nil {
		}
		task := eventloop.RedisTask{Connection: conn}
		eventLoop.AddEvent(task)
	}
}

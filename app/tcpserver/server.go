package tcpserver

import (
	"github.com/codecrafters-io/redis-starter-go/app/eventloop"
	"log"
	"net"
	"strings"
	"sync"
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

func (server *Server) Start(shutDownSignal <-chan struct{}) {
	// 1. address -> start the server on the preferred location
	listener, err := net.Listen("tcp", server.address)
	if err != nil {
		log.Fatalf("Failed to start server at address: %s. Error Details: %v\n", server.address, err)
	}
	log.Println("Server started at:", server.address)

	// 2. run the event loop in a separate go-routine
	eventLoop := server.eventLoop
	go eventLoop.Start(shutDownSignal)

	// 3. current go-routine will monitor the incoming tasks
	var wg sync.WaitGroup
	wg.Add(1)

	go handleIncomingRequests(listener, eventLoop, &wg)

	go func() {
		<-shutDownSignal
		log.Println("Closing the listener")
		listener.Close()
	}()

	wg.Wait()
}

func handleIncomingRequests(
	listener net.Listener,
	eventLoop *eventloop.EventLoop,
	wg *sync.WaitGroup) {

	defer wg.Done()

	for {

		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error occurred while accepting a connection..........%v", err)
			if strings.Contains(err.Error(), "use of closed network connection") {
				log.Println("closing socket connections..........")
				return
			}
			continue
		} else {
			task := eventloop.RedisTask{Connection: conn}
			eventLoop.AddEvent(task)
		}

	}
}

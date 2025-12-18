package tcpserver

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"

	"github.com/archstrap/cache-server/internal/config"
	"github.com/archstrap/cache-server/internal/eventloop"
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

func NewServerFromConfig(config *config.AppConfig) *Server {
	return NewServer(config.GetServerAddress(), config.GetMaxParallelization())
}

func printBanner() {
	fmt.Print(`
	██╗  ██╗ █████╗ ██╗     ██╗███╗   ██╗██████╗ ██╗    ██████╗ ██████╗ 
	██║ ██╔╝██╔══██╗██║     ██║████╗  ██║██╔══██╗██║    ██╔══██╗██╔══██╗
	█████╔╝ ███████║██║     ██║██╔██╗ ██║██║  ██║██║    ██║  ██║██████╔╝
	██╔═██╗ ██╔══██║██║     ██║██║╚██╗██║██║  ██║██║    ██║  ██║██╔══██╗
	██║  ██╗██║  ██║███████╗██║██║ ╚████║██████╔╝██║    ██████╔╝██████╔╝
	╚═╝  ╚═╝╚═╝  ╚═╝╚══════╝╚═╝╚═╝  ╚═══╝╚═════╝ ╚═╝    ╚═════╝ ╚═════╝

`)
}

func (server *Server) Start(ctx context.Context) {
	// 1. address -> start the server on the preferred location
	listener, err := net.Listen("tcp", server.address)
	if err != nil {
		log.Fatalf("Failed to start server at address: %s. Error Details: %v\n", server.address, err)
	}
	// print banner
	printBanner()
	log.Println("Server started at:", server.address)

	// 2. run the event loop in a separate go-routine
	eventLoop := server.eventLoop
	go eventLoop.Start(ctx)

	// 3. The current go-routine will monitor the incoming tasks
	var wg sync.WaitGroup
	wg.Add(1)

	go closeClientConnection(ctx, listener)
	go handleIncomingRequests(listener, eventLoop, &wg)

	wg.Wait()
}

func closeClientConnection(ctx context.Context, listener net.Listener) {
	<-ctx.Done()
	log.Println("Closing the listener")
	if listener == nil {
		return
	}
	err := listener.Close()
	if err != nil {
		return
	}
}

func handleIncomingRequests(
	listener net.Listener,
	eventLoop *eventloop.EventLoop,
	wg *sync.WaitGroup) {

	defer wg.Done()

	for {

		conn, err := listener.Accept()

		if err != nil {

			if errors.Is(err, net.ErrClosed) {
				log.Println("Socket Connection closed!!!")
				return
			}

			log.Printf("Error occurred while accepting a connection..........%v", err)
			if strings.Contains(err.Error(), "use of closed network connection") {
				log.Println("closing socket connections..........")
				return
			}
			continue
		}

		task := eventloop.RedisTask{Connection: conn}
		eventLoop.AddEvent(task)

	}
}

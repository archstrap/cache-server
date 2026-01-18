package tcpserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
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
		slog.Error("Failed to start server", "address", server.address, "error", err)
		os.Exit(1)
	}
	// print banner
	printBanner()
	slog.Info("Cache server started at", slog.String("port", server.address))

	// 2. run the event loop in a separate go-routine
	eventLoop := server.eventLoop
	go eventLoop.Start(ctx)

	// 3. The current go-routine will monitor the incoming tasks
	var wg sync.WaitGroup
	wg.Add(1)

	go closeClientConnection(ctx, listener)
	go handleIncomingRequests(listener, eventLoop, &wg, ctx)

	wg.Wait()
}

func closeClientConnection(ctx context.Context, listener net.Listener) {
	<-ctx.Done()
	slog.Info("Closing the listener")
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
	wg *sync.WaitGroup, ctx context.Context) {

	defer wg.Done()

	for {

		conn, err := listener.Accept()

		if err != nil {

			if errors.Is(err, net.ErrClosed) {
				slog.Info("Socket Connection closed")
				return
			}

			slog.Error("Error occurred while accepting a connection", "error", err)
			if strings.Contains(err.Error(), "use of closed network connection") {
				slog.Info("closing socket connections")
				return
			}
			continue
		}

		task := eventloop.RedisTask{Connection: conn, Context: ctx}
		eventLoop.AddEvent(task)

	}
}

package eventloop

import (
	"context"
	"io"
	"log/slog"
	"net"
	"os"

	"github.com/archstrap/cache-server/internal/command"
	parserLib "github.com/archstrap/cache-server/pkg/parser"
)

type RedisTask struct {
	Connection net.Conn
	Context    context.Context
}

func (conn *RedisTask) exec() {

	connection := conn.Connection

	for {
		select {
		case <-conn.Context.Done():
			slog.Info("Shutdown server")
			return

		default:
			data, err := parserLib.Parse(connection)
			if err != nil {
				if err == io.EOF {
					slog.Info("Client disconnected", "address", connection.RemoteAddr())
					return
				}
				slog.Error("Error occurred", "error", err)
				os.Exit(1)
			}

			factory := command.GetCommandHandlerFactory()
			output := factory.ProcessCommand(data)
			_, err = connection.Write([]byte(output))

			if err == nil {
				slog.Info("Output flushed successfully")
			}
		}

	}

}

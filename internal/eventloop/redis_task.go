package eventloop

import (
	"context"
	"io"
	"log/slog"
	"net"
	"os"
	"strings"

	"github.com/archstrap/cache-server/internal/command"
	"github.com/archstrap/cache-server/internal/rdb"
	"github.com/archstrap/cache-server/internal/replication"
	parserLib "github.com/archstrap/cache-server/pkg/parser"
)

type RedisTask struct {
	Connection net.Conn
	Context    context.Context
}

func (conn *RedisTask) exec() {

	connection := conn.Connection
	parser := parserLib.NewRespParser(connection)

	for {
		select {
		case <-conn.Context.Done():
			slog.Info("Shutdown server")
			return

		default:
			data, err := parser.Parse()
			if err != nil {
				if err == io.EOF {
					slog.Info("Client disconnected", "address", connection.RemoteAddr())
					return
				}
				slog.Error("Error occurred", "error", err)
				os.Exit(1)
			}

			factory := command.GetCommandHandlerFactory()
			output := factory.ProcessCommand(connection, data)
			_, err = connection.Write([]byte(output))

			if err != nil {
				slog.Error("Error while sending outputs .", slog.Any("details", err))
				break
			}
			go sendExtraPayloadIfPossible(connection, output)
		}

	}

}

func sendExtraPayloadIfPossible(conn net.Conn, output string) {

	if strings.Contains(output, "FULLRESYNC") {
		slog.Info("Sending RDB snapshots", slog.Any("connection details", conn.RemoteAddr().String()))
		data := rdb.GetRDBSnapshot()
		_, err := conn.Write(data)
		if err != nil {
			slog.Error("Error while sending RDB snapshots. ", slog.Any("details", err))
		} else {
			replication.GetReplicationStore().Add(conn)
			slog.Info("rdb snapshots sent and registered to master's replication store",
				slog.Any("details", conn.RemoteAddr().String()))
		}
	}

}

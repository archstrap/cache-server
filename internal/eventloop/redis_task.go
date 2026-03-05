package eventloop

import (
	"context"
	"io"
	"log/slog"
	"net"
	"os"
	"slices"
	"strings"

	"github.com/archstrap/cache-server/internal/command"
	"github.com/archstrap/cache-server/internal/rdb"
	"github.com/archstrap/cache-server/internal/replication"
	"github.com/archstrap/cache-server/internal/shared"
	"github.com/archstrap/cache-server/pkg/model"
	parserLib "github.com/archstrap/cache-server/pkg/parser"
	"github.com/archstrap/cache-server/util"
)

type RedisTask struct {
	Connection net.Conn
	Context    context.Context
}

func (task *RedisTask) exec() {

	conn := task.Connection
	parser := parserLib.NewRespParser(conn)

	for {
		select {
		case <-task.Context.Done():
			slog.Info("Shutdown server")
			return

		default:
			input, err := parser.Parse()
			if err != nil {
				if err == io.EOF {
					slog.Info("Client disconnected", "address", conn.RemoteAddr())
					return
				}
				slog.Error("Error occurred", "error", err)
				os.Exit(1)
			}

			var output string
			if !IsCommand(input, "EXEC", "DISCARD") && shared.GetMultiTransactionStore().IsTransactionInitialized(conn) {
				shared.GetMultiTransactionStore().AddCommand(conn, input)
				output = parserLib.ParseOutput(model.NewRespOutput(model.TypeSimpleString, "QUEUED"))
			} else if sr := command.GetSpecialRegistry(); sr.Contains(input.Command) {
				output = sr.Process(conn, input)
				slog.Info("SPECIAL", slog.Any("OUTPUT", output))
			} else {
				output = command.GetCommandHandlerFactory().ProcessCommand(conn, input)
				slog.Info("NORMAL", slog.Any("OUTPUT", output))
			}

			if IsCommand(input, "MULTI") {
				shared.GetMultiTransactionStore().Add(conn)
			}

			// for ACK subcommand we don't have to respond
			if util.IsInputAck(input) {
				continue
			}
			_, err = conn.Write([]byte(output))

			if err != nil {
				slog.Error("Error while sending outputs .", slog.Any("details", err))
				break
			}
			sendExtraPayloadIfPossible(conn, output)
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

func IsCommand(input *model.RespValue, expectedNames ...string) bool {
	actualCommandName := strings.ToUpper(strings.TrimSpace(input.Command))
	return slices.Contains(expectedNames, actualCommandName)
}

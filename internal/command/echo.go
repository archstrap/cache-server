package command

import (
	"log/slog"

	"github.com/archstrap/cache-server/pkg/model"
)

type EchoCommand struct {
	CommandName string
}

func (echo *EchoCommand) Name() string {
	return echo.CommandName
}

func (echo *EchoCommand) Process(value *model.RespValue) *model.RespOutput {

	data := value.Value.([]string)

	if len(data) != 2 {
		return model.NewRespOutput(model.TypeError, "ERR wrong number of arguments for 'echo' command")
	}

	args := data[1]
	slog.Info("ECHO", "args", args)

	return model.NewRespOutput(model.TypeBulkString, args)
}

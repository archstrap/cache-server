package command

import (
	"log/slog"

	"github.com/archstrap/cache-server/internal/config"
	"github.com/archstrap/cache-server/pkg/model"
)

type ConfigCommand struct {
}

func (c *ConfigCommand) Process(value *model.RespValue) *model.RespOutput {

	data := value.Value.([]string)
	if len(data) < 3 {
		return model.NewRespOutput(model.TypeError, "ERR wrong number of arguments for 'config' command")
	}

	slog.Info("", slog.Any("sub", data[0]), slog.Any("arg", data[1]))

	switch data[1] {
	case "GET":
		arg := data[2]
		result, ok := config.Store[arg]
		slog.Info("From Config Command: ", slog.Any("Arg", arg), slog.Any("Result", result))
		if !ok {
			return model.NewRespOutput(model.TypeError, "ERR unknown command")
		}
		return model.NewRespOutput(model.TypeArray, []string{arg, result})
	}

	return model.NewRespOutput(model.TypeError, "ERR unknown details provided")
}

func (c *ConfigCommand) Name() string {
	return "CONFIG"
}

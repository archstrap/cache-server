package command

import (
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
	switch data[1] {
	case "GET":
		arg := data[2]
		result, ok := config.Store[arg]
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

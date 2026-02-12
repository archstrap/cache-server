package command

import (
	"github.com/archstrap/cache-server/internal/replication"
	"github.com/archstrap/cache-server/pkg/model"
)

type InfoCommand struct{}

var InfoCommandInstance = &InfoCommand{}

func (i *InfoCommand) Process(value *model.RespValue) *model.RespOutput {
	args := value.Value.([]string)

	if len(args) < 2 {
		return model.NewUnknownCommandOutput("ERR wrong number of arguments for 'info' command")
	}

	switch args[1] {
	case "replication":
		return model.NewRespOutput(model.TypeBulkString, replication.FormatDetails())
	}

	return model.NewUnknownCommandOutput("ERR wrong number of arguments for 'info' command")
}

func (i *InfoCommand) Name() string {
	return "INFO"
}

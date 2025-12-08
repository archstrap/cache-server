package command

import (
	"log"

	"github.com/archstrap/cache-server/pkg/model"
)

type EchoCommand struct {
}

func (echo *EchoCommand) Name() string {
	return "ECHO"
}

func (echo *EchoCommand) Process(value *model.RespValue) *model.RespOutput {

	data := value.Value.([]string)

	if len(data) != 2 {
		return model.NewRespOutput(model.TypeError, "ERR wrong number of arguments for 'echo' command")
	}

	args := data[1]
	log.Println("ECHO: ", args)

	return model.NewRespOutput(model.TypeBulkString, args)
}

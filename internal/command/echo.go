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
	}

	args := data[1]
	log.Println("ECHO: ", args)

	return model.NewRespOutput(model.TypeSimpleString, args)
}

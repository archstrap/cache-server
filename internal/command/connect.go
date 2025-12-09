package command

import "github.com/archstrap/cache-server/pkg/model"

type ConnectCommand struct {
	CommandName string
}

func (command *ConnectCommand) Name() string {
	return command.CommandName
}

func (command *ConnectCommand) Process(value *model.RespValue) *model.RespOutput {
	return model.NewRespOutput(model.TypeSimpleString, "OK")
}

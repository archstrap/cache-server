package command

import "github.com/archstrap/cache-server/pkg/model"

type PingCommand struct {
	CommandName string
}

func (pc *PingCommand) Name() string {
	return pc.CommandName
}

func (pc *PingCommand) Process(value *model.RespValue) *model.RespOutput {
	return model.NewRespOutput(model.TypeSimpleString, "PONG")
}

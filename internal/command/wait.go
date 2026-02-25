package command

import "github.com/archstrap/cache-server/pkg/model"

type WaitCommand struct{}

var WaitCommandInstance *WaitCommand = &WaitCommand{}

func (command *WaitCommand) Process(input *model.RespValue) *model.RespOutput {
	return model.NewRespOutput(model.TypeInteger, 0)
}

func (command *WaitCommand) Name() string {
	return "WAIT"
}

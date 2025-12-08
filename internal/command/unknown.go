package command

import (
	"fmt"

	"github.com/archstrap/cache-server/pkg/model"
)

type UnknownCommand struct {
	CommandName string
}

func (uc *UnknownCommand) Name() string {
	return uc.CommandName
}

func (uc *UnknownCommand) Process(value *model.RespValue) *model.RespOutput {
	errorMessage := fmt.Sprintf("ERR unknown command '%s', with args beginning with:", value.Command)
	return model.NewUnknownCommandOutput(errorMessage)
}

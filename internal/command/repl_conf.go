package command

import (
	"github.com/archstrap/cache-server/pkg/model"
)

type ReplConf struct{}

var ReplConfCommandInstance *ReplConf = &ReplConf{}

func (r *ReplConf) Name() string {
	return "REPLCONF"
}

func (r *ReplConf) Process(input *model.RespValue) *model.RespOutput {

	args := input.ArgsToStringSlice()
	if len(args) < 3 {
		return model.NewWrongNumberOfOutput("REPLCONF")
	}

	subCommand := args[1]
	switch subCommand {
	case "GETACK":
		return model.NewRespOutput(model.TypeArray, []string{"REPLCONF", "ACK", "0"})
	default:
	}
	return model.NewRespOutput(model.TypeSimpleString, "OK")
}

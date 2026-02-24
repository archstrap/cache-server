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

	args := input.Value.([]string)
	if len(args) < 3 {
		return model.NewWrongNumberOfOutput("REPLCONF")
	}
	return model.NewRespOutput(model.TypeSimpleString, "OK")
}

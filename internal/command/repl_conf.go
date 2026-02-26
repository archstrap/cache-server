package command

import (
	"strconv"

	"github.com/archstrap/cache-server/internal/shared"
	"github.com/archstrap/cache-server/pkg/model"
	"github.com/archstrap/cache-server/util"
)

type ReplConf struct {
	offset int
}

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
		return model.NewRespOutput(model.TypeArray, []string{"REPLCONF", "ACK", strconv.Itoa(r.offset)})
	case "ACK":
		replicationOffset, err := strconv.Atoi(args[2])
		if err != nil {
			return model.NewRespOutput(model.TypeError, err.Error())
		}
		communication := shared.GeAckState().GetCommunication()
		communication <- replicationOffset

	default:
	}
	return model.NewRespOutput(model.TypeSimpleString, "OK")
}

func (r *ReplConf) UpdateOffSet(input *model.RespValue) {
	r.offset += util.GetBytes(input)
}

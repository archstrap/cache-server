package command

import (
	"strconv"

	"github.com/archstrap/cache-server/internal/replication"
	"github.com/archstrap/cache-server/pkg/model"
)

type WaitCommand struct{}

var WaitCommandInstance *WaitCommand = &WaitCommand{}

func (command *WaitCommand) Process(input *model.RespValue) *model.RespOutput {

	args := input.ArgsToStringSlice()

	if len(args) != 3 {
		return model.NewWrongNumberOfOutput(command.Name())
	}

	noOfReplicas, err := strconv.Atoi(args[1])
	if err != nil {
		return model.NewRespOutput(model.TypeError, err)
	}

	if noOfReplicas == 0 {
		return model.NewRespOutput(model.TypeInteger, 0)
	}

	return model.NewRespOutput(model.TypeInteger, replication.GetReplicationStore().ActiveReplicationCount())
}

func (command *WaitCommand) Name() string {
	return "WAIT"
}

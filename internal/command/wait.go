package command

import (
	"strconv"
	"time"

	"github.com/archstrap/cache-server/internal/replication"
	"github.com/archstrap/cache-server/internal/shared"
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

	timeOut, err := strconv.Atoi(args[2])
	if err != nil {
		return model.NewRespOutput(model.TypeError, err)
	}

	return model.NewRespOutput(model.TypeInteger, command.findActiveAcknowledgements(time.Duration(timeOut)*time.Millisecond, noOfReplicas, replication.GetReplicationStore()))
}

func (command *WaitCommand) Name() string {
	return "WAIT"
}

func (command *WaitCommand) findActiveAcknowledgements(timeout time.Duration, noOfReplicas int, store *replication.ReplicationStore) int {
	masterServerState := replication.GetServerState()
	if masterServerState.GetMasterState() == 0 {
		return store.ActiveReplicationCount()
	}

	// ask for an acknowledgement from replicas to know about their offset
	replication.GetReplicationStore().Propagate(&model.RespValue{
		DataType: model.TypeArray,
		Command:  "REPLCONF",
		Value:    []string{"REPLCONF", "GETACK", "*"},
	})

	ackCount := 0
	timer := time.After(timeout)
	ackState := shared.GeAckState()
	serverState := replication.GetServerState()

	for {
		select {
		case <-timer:
			return ackCount
		case offset := <-ackState.GetCommunication():
			if offset >= serverState.GetMasterState() {
				ackCount++
			}
			if ackCount >= noOfReplicas {
				return ackCount
			}
		}
	}
}

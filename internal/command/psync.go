package command

import (
	"fmt"
	"strconv"

	"github.com/archstrap/cache-server/internal/replication"
	"github.com/archstrap/cache-server/pkg/model"
)

type Psync struct{}

var PsyncInstance *Psync = &Psync{}

func (p *Psync) Name() string {
	return "PSYNC"
}

func (p *Psync) Process(input *model.RespValue) *model.RespOutput {

	args := input.ArgsToStringSlice()
	if len(args) < 3 {
		return model.NewWrongNumberOfOutput("PSYNC")
	}

	offset, err := strconv.Atoi(args[2])
	if err != nil {
		return model.NewRespOutput(model.TypeError, err)
	}

	var result string

	if args[1] == "?" {
		result = fmt.Sprintf("FULLRESYNC %s %d", replication.GetServerReplicationID(), (offset + 1))
	}

	return model.NewRespOutput(model.TypeSimpleString, result)

}

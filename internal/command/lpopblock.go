package command

import (
	"strconv"

	"github.com/archstrap/cache-server/internal/store"
	"github.com/archstrap/cache-server/pkg/model"
)

type Blpop struct{}

var cBlpop = &Blpop{}

func init() {
	Registry.RegisterCommand(cBlpop)
}

func (c *Blpop) Name() string {
	return "BLPOP"
}

func (c *Blpop) Process(input *model.RespValue) *model.RespOutput {
	args := input.ArgsToStringSlice()
	if len(args) < 3 {
		return model.NewWrongNumberOfOutput(c.Name())
	}

	key := args[1]
	timeOut, _ := strconv.Atoi(args[2])
	result := store.GetContainer().BlockDelete(key, timeOut)

	return model.NewRespOutput(model.TypeArray, result)
}

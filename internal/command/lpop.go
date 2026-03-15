package command

import (
	"strconv"

	"github.com/archstrap/cache-server/internal/store"
	"github.com/archstrap/cache-server/pkg/model"
)

type Lpop struct{}

var cLpop = &Lpop{}

func init() {
	Registry.RegisterCommand(cLpop)
}

func (c *Lpop) Name() string {
	return "LPOP"
}

func (c *Lpop) Process(input *model.RespValue) *model.RespOutput {
	args := input.ArgsToStringSlice()
	length := len(args)
	if length < 2 {
		return model.NewWrongNumberOfOutput(c.Name())
	}

	deleteCount := 1
	if length > 2 {
		deleteCount, _ = strconv.Atoi(args[2])
	}

	key := args[1]
	container := store.GetContainer()
	deletedItems := container.Delete(key, deleteCount)

	if len(deletedItems) == 0 {
		return model.NewRespOutput(model.TypeBulkString, "-1")
	}

	if len(deletedItems) == 1 {
		return model.NewRespOutput(model.TypeBulkString, deletedItems[0])
	}
	return model.NewRespOutput(model.TypeArray, deletedItems)
}

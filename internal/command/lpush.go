// Package command provides all the commands we can execute in redis
package command

import (
	"github.com/archstrap/cache-server/internal/store"
	"github.com/archstrap/cache-server/pkg/model"
)

type Lpush struct{}

var cLpush = &Lpush{}

func init() {
	Registry.RegisterCommand(cLpush)
}

func (c *Lpush) Name() string {
	return "LPUSH"
}

func (c *Lpush) Process(input *model.RespValue) *model.RespOutput {
	args := input.ArgsToStringSlice()
	if len(args) < 3 {
		return model.NewWrongNumberOfOutput(input.Command)
	}

	key := args[1]
	elements := args[2:]

	container := store.GetContainer()
	insertedItems := container.Prepend(key, elements...)
	return model.NewRespOutput(model.TypeInteger, insertedItems)
}

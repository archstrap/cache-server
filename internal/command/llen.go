// Package command provides all the commands we can execute in redis
package command

import (
	"github.com/archstrap/cache-server/internal/store"
	"github.com/archstrap/cache-server/pkg/model"
)

type Llen struct{}

var cLlen = &Llen{}

func init() {
	Registry.RegisterCommand(cLlen)
}

func (c *Llen) Name() string {
	return "LLEN"
}

func (c *Llen) Process(input *model.RespValue) *model.RespOutput {
	args := input.ArgsToStringSlice()
	if len(args) < 2 {
		return model.NewWrongNumberOfOutput(input.Command)
	}

	key := args[1]

	container := store.GetContainer()
	insertedItems := container.GetLen(key)
	return model.NewRespOutput(model.TypeInteger, insertedItems)
}

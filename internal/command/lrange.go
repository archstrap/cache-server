// Package command provides all the commands we can execute in redis
package command

import (
	"strconv"

	"github.com/archstrap/cache-server/internal/store"
	"github.com/archstrap/cache-server/pkg/model"
)

type Lrange struct{}

var cLrange = &Lrange{}

func init() {
	Registry.RegisterCommand(cLrange)
}

func (c *Lrange) Name() string {
	return "LRANGE"
}

func (c *Lrange) Process(input *model.RespValue) *model.RespOutput {
	args := input.ArgsToStringSlice()
	if len(args) < 4 {
		return model.NewWrongNumberOfOutput(input.Command)
	}

	key := args[1]
	start, err := strconv.Atoi(args[2])
	if err != nil {
		return model.NewRespOutput(model.TypeError, "Provide integer value of start")
	}
	end, err := strconv.Atoi(args[3])
	if err != nil {
		return model.NewRespOutput(model.TypeError, "Provide integer value of end")
	}

	container := store.GetContainer()
	data := container.Get(key, start, end)
	return model.NewRespOutput(model.TypeArray, data)
}

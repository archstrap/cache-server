package command

import (
	"github.com/archstrap/cache-server/internal/store"
	"github.com/archstrap/cache-server/pkg/model"
)

type XRANGE struct{}

var cXRANGE = &XRANGE{}

func init() {
	Registry.RegisterCommand(cXRANGE)
}

func (x *XRANGE) Name() string {
	return "XRANGE"
}

func (x *XRANGE) Process(input *model.RespValue) *model.RespOutput {

	args := input.ArgsToStringSlice()
	if len(args)%2 != 0 {
		return model.NewWrongNumberOfOutput("XRANGE")
	}

	key := args[1]
	start := args[2]
	end := args[3]

	result := store.StreamStoreInstance.SearchInRange(key, start, end)

	return model.NewRespOutput(model.TypeArray, result)
}

package command

import (
	"github.com/archstrap/cache-server/internal/store"
	"github.com/archstrap/cache-server/pkg/model"
)

type XADD struct{}

var cXadd = &XADD{}

func init() {
	Registry.RegisterCommand(cXadd)
}

func (x *XADD) Name() string {
	return "XADD"
}

func (x *XADD) Process(input *model.RespValue) *model.RespOutput {

	args := input.ArgsToStringSlice()
	if len(args)%2 == 0 {
		return model.NewWrongNumberOfOutput(x.Name())
	}

	key := args[1]
	data := map[string]string{
		"id": args[2],
	}

	for i := 3; i < len(args); i += 2 {
		k, v := args[i], args[i+1]
		data[k] = v
	}

	insertedId := store.StreamStoreInstance.AddItem(key, data)

	return model.NewRespOutput(model.TypeBulkString, insertedId)
}

package command

import (
	"github.com/archstrap/cache-server/internal/store"
	"github.com/archstrap/cache-server/pkg/model"
)

type ZREM struct{}

func init() {
	Registry.RegisterCommand(&ZREM{})
}

func (c *ZREM) Name() string {
	return "ZREM"
}

func (c *ZREM) Process(input *model.RespValue) *model.RespOutput {

	args := input.ArgsToStringSlice()
	if len(args) != 3 {
		return model.NewWrongNumberOfOutput(c.Name())
	}

	key := args[1]
	member := args[2]
	bucket := store.GetSkipListBucket()

	return model.NewRespOutput(model.TypeInteger, bucket.Remove(key, member))
}

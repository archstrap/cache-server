package command

import (
	"github.com/archstrap/cache-server/internal/store"
	"github.com/archstrap/cache-server/pkg/model"
)

type ZCARD struct{}

func init() {
	Registry.RegisterCommand(&ZCARD{})
}

func (c *ZCARD) Name() string {
	return "ZCARD"
}

func (c *ZCARD) Process(input *model.RespValue) *model.RespOutput {
	args := input.ArgsToStringSlice()
	if len(args) != 2 {
		return model.NewWrongNumberOfOutput(c.Name())
	}

	key := args[1]
	bucket := store.GetSkipListBucket()
	return model.NewRespOutput(model.TypeInteger, bucket.Count(key))
}

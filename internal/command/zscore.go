package command

import (
	"github.com/archstrap/cache-server/internal/store"
	"github.com/archstrap/cache-server/pkg/model"
)

type ZSCORE struct{}

func init() {
	Registry.RegisterCommand(&ZSCORE{})
}

func (c *ZSCORE) Name() string {
	return "ZSCORE"
}

func (c *ZSCORE) Process(input *model.RespValue) *model.RespOutput {
	args := input.ArgsToStringSlice()
	if len(args) != 3 {
		return model.NewWrongNumberOfOutput(c.Name())
	}

	key := args[1]
	member := args[2]
	bucket := store.GetSkipListBucket()

	return model.NewRespOutput(model.TypeBulkString, bucket.Score(key, member))
}

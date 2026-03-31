package command

import (
	"github.com/archstrap/cache-server/internal/store"
	"github.com/archstrap/cache-server/pkg/model"
)

type ZRANK struct{}

func init() {
	Registry.RegisterCommand(&ZRANK{})
}

func (c *ZRANK) Name() string {
	return "ZRANK"
}

func (c *ZRANK) Process(input *model.RespValue) *model.RespOutput {
	args := input.ArgsToStringSlice()
	if len(args) < 3 {
		return model.NewWrongNumberOfOutput(c.Name())
	}

	key := args[1]
	member := args[2]
	bucket := store.GetSkipListBucket()

	result := bucket.Rank(key, member)
	if result == -1 {
		return model.NewRespOutput(model.TypeBulkString, "-1")
	}

	return model.NewRespOutput(model.TypeInteger, result)
}

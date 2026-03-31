package command

import (
	"github.com/archstrap/cache-server/internal/store"
	"github.com/archstrap/cache-server/pkg/model"
)

type ZADD struct{}

func init() {
	Registry.RegisterCommand(&ZADD{})
}

func (c *ZADD) Name() string {
	return "ZADD"
}

func (c *ZADD) Process(input *model.RespValue) *model.RespOutput {
	args := input.ArgsToStringSlice()
	if len(args) < 4 {
		return model.NewWrongNumberOfOutput(c.Name())
	}

	setItem := store.NewSetItem(args[1], args[2], args[3])
	bucket := store.GetSkipListBucket()

	return model.NewRespOutput(model.TypeInteger, bucket.Insert(setItem))

}

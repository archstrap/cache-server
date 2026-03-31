package command

import (
	"strconv"

	"github.com/archstrap/cache-server/internal/store"
	"github.com/archstrap/cache-server/pkg/model"
)

type ZRANGE struct{}

func init() {
	Registry.RegisterCommand(&ZRANGE{})
}

func (c *ZRANGE) Name() string {
	return "ZRANGE"
}

func (c *ZRANGE) Process(input *model.RespValue) *model.RespOutput {

	args := input.ArgsToStringSlice()
	if len(args) < 4 {
		return model.NewWrongNumberOfOutput(c.Name())
	}

	key := args[1]
	start, err := strconv.Atoi(args[2])
	if err != nil {
		return model.NewRespOutput(model.TypeError, err.Error())
	}
	end, err := strconv.Atoi(args[3])
	if err != nil {
		return model.NewRespOutput(model.TypeError, err.Error())
	}

	bucket := store.GetSkipListBucket()
	return model.NewRespOutput(model.TypeArray, bucket.Range(key, start, end))
}

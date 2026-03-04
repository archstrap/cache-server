package command

import (
	"time"

	"github.com/archstrap/cache-server/internal/store"
	"github.com/archstrap/cache-server/pkg/model"
)

type Incr struct{}

var cIncr = &Incr{}

func init() {
	Registry.RegisterCommand(cIncr)
}

func (c *Incr) Name() string {
	return "INCR"
}

func (c *Incr) Process(input *model.RespValue) *model.RespOutput {
	args := input.ArgsToStringSlice()
	if len(args) != 2 {
		return model.NewWrongNumberOfOutput(c.Name())
	}

	key := args[1]
	cacheStore := store.CacheStore

	cacheStore.Lock()
	defer cacheStore.Unlock()

	item, ok := cacheStore.Get(key)

	if !ok {
		item = cacheStore.Add(key, store.NewCacheItem("0", time.Time{}, model.ValueTypeString))
	}

	updated, ok := item.IncrementBy(1)
	if !ok {
		return model.NewRespOutput(model.TypeError, "ERR value is not an integer or out of range")
	}

	return model.NewRespOutput(model.TypeInteger, updated)
}

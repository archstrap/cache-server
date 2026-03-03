package command

import (
	"github.com/archstrap/cache-server/internal/store"
	"github.com/archstrap/cache-server/pkg/model"
)

type TypeCommand struct{}

var TypeCommandInstance *TypeCommand = &TypeCommand{}

func (t *TypeCommand) Name() string {
	return "TYPE"
}

func (t *TypeCommand) Process(input *model.RespValue) *model.RespOutput {

	args := input.ArgsToStringSlice()
	if len(args) != 2 {
		return model.NewWrongNumberOfOutput(t.Name())
	}

	cacheStore := store.GetCacheStore()
	cacheStore.Lock()
	defer cacheStore.Unlock()

	key := args[1]
	val, ok := cacheStore.Get(key)
	resultType := "none"
	if ok {
		if val.IsExpired() {
			cacheStore.Delete(key)
		} else {
			resultType = string(val.ValueType())
		}
	}

	if store.StreamStoreInstance.ContainsKey(key) {
		resultType = "stream"
	}

	return model.NewRespOutput(model.TypeSimpleString, resultType)
}

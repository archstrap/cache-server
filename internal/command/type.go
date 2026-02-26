package command

import (
	"time"

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

	CacheStore.mu.Lock()
	defer CacheStore.mu.Unlock()

	key := args[1]
	val, ok := CacheStore.data[key]
	resultType := "none"
	if ok {
		if !val.expiresAt.IsZero() && time.Now().After(val.expiresAt) {
			delete(CacheStore.data, key)
		} else {
			resultType = string(val.valueType)
		}
	}

	if store.StreamStoreInstance.ContainsKey(key) {
		resultType = "stream"
	}

	return model.NewRespOutput(model.TypeSimpleString, resultType)
}

package command

import (
	"log/slog"
	"time"

	"github.com/archstrap/cache-server/internal/store"
	"github.com/archstrap/cache-server/pkg/model"
	"github.com/archstrap/cache-server/util"
)

type KeyCommand struct {
}

var KeyCommandInstance = &KeyCommand{}

func (k *KeyCommand) Process(input *model.RespValue) *model.RespOutput {

	args := input.Value.([]string)
	if len(args) != 2 {
		return model.NewRespOutput(model.TypeError, "ERR wrong number of arguments for 'keys' command")
	}

	pattern := args[1]
	cacheStore := store.GetCacheStore()
	keys := make([]string, 0)

	cleanableKeys := make([]string, 0)

	for k := range cacheStore.GetData() {
		matched, err := util.MatchString(pattern, k)
		if err != nil {
			continue
		}

		if matched {
			v, _ := cacheStore.Get(k)
			now := time.Now()
			// delete the key if expires
			if v.IsExpiredNow(now) {
				cleanableKeys = append(cleanableKeys, k)
				continue
			}

			keys = append(keys, k)

		}
	}

	go cleanup(&cleanableKeys)

	return model.NewRespOutput(model.TypeArray, keys)
}

func (k *KeyCommand) Name() string {
	return "KEYS"
}

func cleanup(cleanableKeys *[]string) {
	store := store.GetCacheStore()
	for i := range *cleanableKeys {
		key := (*cleanableKeys)[i]
		store.Delete(key)
	}
	slog.Info("Delete expired keys ", "count", len(*cleanableKeys))
}

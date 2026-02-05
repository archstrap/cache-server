package command

import (
	"log/slog"
	"time"

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
	store := GetCacheStore()
	keys := make([]string, 0)

	cleanableKeys := make([]string, 0)

	for k := range store.data {
		matched, err := util.MatchString(pattern, k)
		if err != nil {
			continue
		}

		if matched {
			v := store.data[k]
			now := time.Now()
			// delete the key if expires
			if !v.expiresAt.IsZero() && now.After(v.expiresAt) {
				cleanableKeys = append(cleanableKeys, k)
				continue
			}

			keys = append(keys, k)

		}
	}

	go cleanup(cleanableKeys)

	return model.NewRespOutput(model.TypeArray, keys)
}

func (k *KeyCommand) Name() string {
	return "KEYS"
}

func cleanup(cleanableKeys []string) {
	store := GetCacheStore()
	for i := range cleanableKeys {
		delete(store.data, cleanableKeys[i])
	}
	slog.Info("Delete expired keys ", "count", len(cleanableKeys))
}

package command

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/archstrap/cache-server/pkg/model"
)

type GetCommand struct {
	CommandName string
}

type SetCommand struct {
	CommandName string
}

type CacheItem struct {
	item      any
	expiresAt time.Time
}

type Cache struct {
	mu   sync.RWMutex
	data map[string]CacheItem
}

var cache = &Cache{
	data: make(map[string]CacheItem),
}

var subCommands = map[string]bool{
	"EX": true,
	"PX": true,
}

func (command *GetCommand) Name() string {
	return command.CommandName
}

func (command *SetCommand) Name() string {
	return command.CommandName
}

func (command *GetCommand) Process(value *model.RespValue) *model.RespOutput {

	data := value.Value.([]string)

	// GET K
	if len(data) != 2 {
		return model.NewRespOutput(model.TypeError, "ERR wrong number of arguments for 'get' command")
	}

	cache.mu.RLock()
	defer cache.mu.RUnlock()

	key := data[1]
	cacheItem, ok := cache.data[key]
	val := cacheItem.item
	if !ok || !cacheItem.expiresAt.IsZero() && time.Now().After(cacheItem.expiresAt) {
		delete(cache.data, key)
		val = "-1"
	}

	return model.NewRespOutput(model.TypeBulkString, val)
}

func (command *SetCommand) Process(value *model.RespValue) *model.RespOutput {

	data := value.Value.([]string)

	// SET k v
	if len(data) != 3 && len(data) != 5 {
		return model.NewRespOutput(model.TypeError, "ERR wrong number of arguments for 'set' command")
	}

	var expiresAt time.Time
	if len(data) == 5 {

		optionalCommandName := strings.TrimSpace(data[3])
		optionalCommand := strings.ToUpper(optionalCommandName)

		if !subCommands[optionalCommand] {
			return model.NewRespOutput(model.TypeError, fmt.Sprintf("ERR invalid key element for '%s' command", optionalCommand))
		}

		timeArgument, err := strconv.Atoi(data[4])

		if err != nil || timeArgument <= 0 {
			return model.NewRespOutput(model.TypeError, fmt.Sprintf("ERR invalid value for '%s' command", optionalCommandName))
		}

		switch optionalCommand {
		// SET K V EX 10
		case "EX":
			expiresAt = time.Now().Add(time.Duration(timeArgument) * time.Second)
		// SET K V PX 1000
		case "PX":
			expiresAt = time.Now().Add(time.Duration(timeArgument) * time.Millisecond)
		}

	}

	cache.mu.Lock()
	defer cache.mu.Unlock()

	key := data[1]
	val := data[2]
	cache.data[key] = CacheItem{item: val, expiresAt: expiresAt}

	return model.NewRespOutput(model.TypeSimpleString, "OK")
}

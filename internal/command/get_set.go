package command

import (
	"sync"

	"github.com/archstrap/cache-server/pkg/model"
)

type GetCommand struct {
	CommandName string
}

type SetCommand struct {
	CommandName string
}

type Cache struct {
	mu   sync.RWMutex
	data map[string]any
}

var cache = &Cache{
	data: make(map[string]any),
}

func (command *GetCommand) Name() string {
	return command.CommandName
}

func (command *SetCommand) Name() string {
	return command.CommandName
}

func (command *GetCommand) Process(value *model.RespValue) *model.RespOutput {

	data := value.Value.([]string)

	if len(data) != 3 {
		return model.NewRespOutput(model.TypeError, "(error) ERR wrong number of arguments for 'get' command")
	}

	cache.mu.RLock()
	defer cache.mu.Unlock()

	key := data[1]
	val, ok := cache.data[key]
	if !ok {
		val = "-1"
	}

	return model.NewRespOutput(model.TypeBulkString, val)
}

func (command *SetCommand) Process(value *model.RespValue) *model.RespOutput {

	data := value.Value.([]string)

	if len(data) < 3 {
		return model.NewRespOutput(model.TypeError, "(error) ERR wrong number of arguments for 'set' command")
	}

	// TODO for options like NX

	cache.mu.Lock()
	defer cache.mu.Unlock()

	key := data[1]
	val := data[2]
	cache.data[key] = val

	return model.NewRespOutput(model.TypeSimpleString, "OK")
}

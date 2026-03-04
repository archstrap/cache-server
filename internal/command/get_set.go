package command

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/archstrap/cache-server/internal/store"
	"github.com/archstrap/cache-server/pkg/model"
)

type GetCommand struct {
	CommandName string
}

type SetCommand struct {
	CommandName string
}

var SetCommandInstance = &SetCommand{
	CommandName: "SET",
}

var GetCommandInstance = &GetCommand{
	CommandName: "GET",
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

	data := value.ArgsToStringSlice()

	// GET K
	if len(data) != 2 {
		return model.NewRespOutput(model.TypeError, "ERR wrong number of arguments for 'get' command")
	}

	cacheStore := store.GetCacheStore()
	cacheStore.RLock()
	defer cacheStore.RUnlock()

	key := data[1]
	cacheItem, ok := cacheStore.Get(key)

	// case-1: key is not present
	if !ok {
		return model.NewRespOutput(model.TypeBulkString, "-1")
	}

	// case-2: Key is present
	val := cacheItem.Item()
	// case-2(b): but key is expired
	if cacheItem.IsExpired() {
		val = "-1"
		cacheStore.Delete(key)
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

	cacheStore := store.GetCacheStore()
	cacheStore.Lock()
	defer cacheStore.Unlock()

	key := data[1]
	val := data[2]
	valueType := model.ValueTypeString

	cacheStore.Add(key, store.NewCacheItem(val, expiresAt, valueType))

	return model.NewRespOutput(model.TypeSimpleString, "OK")
}

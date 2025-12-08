package command

import (
	"log"
	"strings"

	"github.com/archstrap/cache-server/pkg/model"
	"github.com/archstrap/cache-server/pkg/parser"
)

type HandlerFactory struct {
	handlers map[string]ICommand
}

func NewCommandHandlerFactory() *HandlerFactory {
	handlerFactory := &HandlerFactory{
		handlers: make(map[string]ICommand),
	}
	handlerFactory.registerAllCommands()
	return handlerFactory
}

func (hcf *HandlerFactory) registerCommandHandler(command ICommand) {
	hcf.handlers[command.Name()] = command
}

func (hcf *HandlerFactory) registerAllCommands() {
	hcf.registerCommandHandler(&EchoCommand{})
	hcf.registerCommandHandler(&UnknownCommand{CommandName: "UNKNOWN"})
}

func getOrDefault(mp map[string]ICommand, key string, defaultValue ICommand) ICommand {
	value, ok := mp[key]
	if !ok {
		value = defaultValue
	}
	return value
}

func (hcf *HandlerFactory) ProcessCommand(input *model.RespValue) string {
	command := strings.ToUpper(strings.TrimSpace(input.Command))

	log.Println("Received command ", command)

	switch command {
	case "COMMAND":
		return "+OK\r\n"
	case "PING":
		return "+PONG\r\n"
	default:
		iCommand := getOrDefault(hcf.handlers, command, &UnknownCommand{CommandName: "UNKNOWN"})
		respOutput := iCommand.Process(input)
		return parser.ParseOutputV2(respOutput)
	}

}

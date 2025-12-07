package command

import (
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
}

func (hcf *HandlerFactory) ProcessCommand(input *model.RespValue) string {
	command := strings.ToUpper(strings.TrimSpace(input.Command))

	switch command {
	case "COMMAND":
		return "+OK\r\n"
	case "PING":
		return "+PONG\r\n"
	default:
		iCommand := hcf.handlers[command]
		respOutput := iCommand.Process(input)
		return parser.ParseOutputV2(respOutput)
	}

}

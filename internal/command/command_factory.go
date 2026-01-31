package command

import (
	"log/slog"
	"strings"
	"sync"

	"github.com/archstrap/cache-server/pkg/model"
	"github.com/archstrap/cache-server/pkg/parser"
)

type HandlerFactory struct {
	handlers map[string]ICommand
}

func NewCommandHandlerFactory() *HandlerFactory {

	slog.Info("Command Handlers Initialized")

	handlerFactory := &HandlerFactory{
		handlers: make(map[string]ICommand),
	}
	handlerFactory.registerAllCommands()
	return handlerFactory
}

var (
	handlerFactoryInstance *HandlerFactory
	mu                     sync.Mutex
)

func GetCommandHandlerFactory() *HandlerFactory {

	if handlerFactoryInstance == nil {
		mu.Lock()
		defer mu.Unlock()
		if handlerFactoryInstance == nil {
			handlerFactoryInstance = NewCommandHandlerFactory()
		} else {
			slog.Info("Giving Existing Initialized Handlers - 1")
		}
	} else {
		slog.Info("Giving Existing Initialized Handlers - 2")
	}

	return handlerFactoryInstance
}

func (hcf *HandlerFactory) registerCommandHandler(command ICommand) {
	hcf.handlers[command.Name()] = command
}

func (hcf *HandlerFactory) registerAllCommands() {

	var commandHandlers []ICommand

	commandHandlers = append(commandHandlers, &EchoCommand{CommandName: "ECHO"})
	commandHandlers = append(commandHandlers, &PingCommand{CommandName: "PING"})
	commandHandlers = append(commandHandlers, &ConnectCommand{CommandName: "COMMAND"})
	commandHandlers = append(commandHandlers, &GetCommand{CommandName: "GET"})
	commandHandlers = append(commandHandlers, &SetCommand{CommandName: "SET"})
	commandHandlers = append(commandHandlers, &UnknownCommand{CommandName: "UNKNOWN"})

	for _, handler := range commandHandlers {
		hcf.registerCommandHandler(handler)
	}

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

	slog.Info("Received command", "command", command)

	iCommand := getOrDefault(hcf.handlers, command, &UnknownCommand{CommandName: "UNKNOWN"})
	respOutput := iCommand.Process(input)
	return parser.ParseOutput(respOutput)

}

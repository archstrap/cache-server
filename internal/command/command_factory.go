package command

import (
	"log/slog"
	"net"
	"strings"
	"sync"

	"github.com/archstrap/cache-server/internal/replication"
	"github.com/archstrap/cache-server/internal/shared"
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
	shared.SetCommandProcessor(handlerFactory)
	return handlerFactory
}

var (
	handlerFactoryInstance *HandlerFactory
	mu                     sync.Mutex
	modifiableCommands     map[string]bool = map[string]bool{
		"SET": true,
	}
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
	commandHandlers = append(commandHandlers, GetCommandInstance)
	commandHandlers = append(commandHandlers, SetCommandInstance)
	commandHandlers = append(commandHandlers, &ConfigCommand{})
	commandHandlers = append(commandHandlers, KeyCommandInstance)
	commandHandlers = append(commandHandlers, InfoCommandInstance)
	commandHandlers = append(commandHandlers, ReplConfCommandInstance)
	commandHandlers = append(commandHandlers, PsyncInstance)
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

func (hcf *HandlerFactory) ProcessCommand(conn net.Conn, input *model.RespValue) string {
	command := strings.ToUpper(strings.TrimSpace(input.Command))

	slog.Info("Start Processing ", slog.Any("command", input.Command))
	MonitorReplicaConnectionIfPossible(conn, input)

	iCommand := getOrDefault(hcf.handlers, command, &UnknownCommand{CommandName: "UNKNOWN"})
	// process the output
	respOutput := iCommand.Process(input)

	AddPropagationIfPossible(input, respOutput)

	return parser.ParseOutput(respOutput)

}

func (hcf *HandlerFactory) ProcessCommandsSilently(input *model.RespValue) string {
	command := strings.ToUpper(strings.TrimSpace(input.Command))

	slog.Info("Start Processing Silently", slog.Any("command", input.Command))

	iCommand := getOrDefault(hcf.handlers, command, &UnknownCommand{CommandName: "UNKNOWN"})
	// process the output
	respOutput := iCommand.Process(input)

	return parser.ParseOutput(respOutput)

}

func MonitorReplicaConnectionIfPossible(conn net.Conn, input *model.RespValue) {
	if input.Command == "REPLCONF" {
		args := input.ArgsToStringSlice()
		if args[1] == "listening-port" {
			port := args[2]
			replication.GetReplicationStore().Add(conn, port)
		}
	}
}

func AddPropagationIfPossible(input *model.RespValue, output *model.RespOutput) {
	// if we are getting non error message we are going to propagate the WRITE commands to replica
	command := strings.ToUpper(input.Command)

	if !modifiableCommands[command] || output.RespType == model.TypeError {
		slog.Info("Not able to propagate commands")
		return
	}

	go replication.GetReplicationStore().Propagate(input)
}

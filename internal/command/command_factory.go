package command

import (
	"log/slog"
	"net"
	"strings"
	"sync"

	"github.com/archstrap/cache-server/internal/config"
	"github.com/archstrap/cache-server/internal/replication"
	"github.com/archstrap/cache-server/internal/shared"
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
		}
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
	commandHandlers = append(commandHandlers, WaitCommandInstance)
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

	iCommand := getOrDefault(hcf.handlers, command, &UnknownCommand{CommandName: "UNKNOWN"})
	// process the output
	respOutput := iCommand.Process(input)

	MonitorReplicaConnectionIfPossible(conn, input)
	AddPropagationIfPossible(input, respOutput)

	return parser.ParseOutput(respOutput)

}

func (hcf *HandlerFactory) Process(conn net.Conn, input *model.RespValue) {
	command := strings.ToUpper(strings.TrimSpace(input.Command))

	slog.Info("Start Processing Silently", slog.Any("command", input.Command))

	iCommand := getOrDefault(hcf.handlers, command, &UnknownCommand{CommandName: "UNKNOWN"})
	// process the output
	result := iCommand.Process(input)

	if command == "REPLCONF" {
		output := parser.ParseOutput(result)
		if _, err := conn.Write([]byte(output)); err != nil {
			slog.Error("Error Occurred while responding back to REPLCONF command.", slog.Any("Details", err))
		}
	}

	ReplConfCommandInstance.UpdateOffSet(input)

}

func MonitorReplicaConnectionIfPossible(conn net.Conn, input *model.RespValue) {
	if input.Command == "REPLCONF" {
		args := input.ArgsToStringSlice()
		if args[1] == "listening-port" {
			port := args[2]
			replication.GetReplicationStore().AddInPending(conn, port)
		}
	}
}

func AddPropagationIfPossible(input *model.RespValue, output *model.RespOutput) {
	// if we are getting non error message we are going to propagate the WRITE commands to replica
	command := strings.ToUpper(input.Command)

	if config.Store["replicaof"] != "" || !modifiableCommands[command] || output.RespType == model.TypeError {
		return
	}

	go replication.GetReplicationStore().Propagate(input)
}

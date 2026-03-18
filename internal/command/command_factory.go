package command

import (
	"fmt"
	"log/slog"
	"net"
	"strings"

	"github.com/archstrap/cache-server/internal/config"
	"github.com/archstrap/cache-server/internal/replication"
	"github.com/archstrap/cache-server/internal/shared"
	"github.com/archstrap/cache-server/pkg/model"
	"github.com/archstrap/cache-server/pkg/parser"
	"github.com/archstrap/cache-server/util"
)

type HandlerFactory struct {
	handlers map[string]ICommand
}

type SpecialRegistry struct {
	commands map[string]SCommand
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
	Registry           *HandlerFactory
	modifiableCommands map[string]bool = map[string]bool{
		"SET":  true,
		"INCR": true,
	}
	SRegistry *SpecialRegistry = &SpecialRegistry{
		commands: make(map[string]SCommand),
	}
)

func init() {
	Registry = NewCommandHandlerFactory()
}

func GetCommandHandlerFactory() *HandlerFactory {
	return Registry
}

func GetSpecialRegistry() *SpecialRegistry {
	return SRegistry
}

func (cR *HandlerFactory) RegisterCommand(command ICommand) {
	cR.handlers[command.Name()] = command
}

func (cR *HandlerFactory) registerAllCommands() {

	var commandHandlers []ICommand

	commandHandlers = append(commandHandlers, &EchoCommand{CommandName: "ECHO"})
	commandHandlers = append(commandHandlers, &ConnectCommand{CommandName: "COMMAND"})
	commandHandlers = append(commandHandlers, GetCommandInstance)
	commandHandlers = append(commandHandlers, SetCommandInstance)
	commandHandlers = append(commandHandlers, &ConfigCommand{})
	commandHandlers = append(commandHandlers, KeyCommandInstance)
	commandHandlers = append(commandHandlers, InfoCommandInstance)
	commandHandlers = append(commandHandlers, ReplConfCommandInstance)
	commandHandlers = append(commandHandlers, PsyncInstance)
	commandHandlers = append(commandHandlers, WaitCommandInstance)
	commandHandlers = append(commandHandlers, TypeCommandInstance)
	commandHandlers = append(commandHandlers, &UnknownCommand{CommandName: "UNKNOWN"})

	for _, handler := range commandHandlers {
		cR.RegisterCommand(handler)
	}

}

func getOrDefault(mp map[string]ICommand, key string, defaultValue ICommand) ICommand {
	value, ok := mp[key]
	if !ok {
		value = defaultValue
	}
	return value
}

func (cR *HandlerFactory) ProcessCommand(conn net.Conn, input *model.RespValue) string {
	command := strings.ToUpper(strings.TrimSpace(input.Command))

	slog.Info("Start Processing ", slog.Any("command", command))

	iCommand := getOrDefault(cR.handlers, command, &UnknownCommand{CommandName: "UNKNOWN"})

	if command != "QUIT" && command != "RESET" && shared.GetChannelStore().IsSubscribed(conn) {
		return parser.ParseOutput(model.NewRespOutput(model.TypeError, fmt.Sprintf("ERR Can't execute '%s': only (P|S)SUBSCRIBE / (P|S)UNSUBSCRIBE / PING / QUIT / RESET are allowed in this context ", input.Command)))
	}

	// process the output
	respOutput := iCommand.Process(input)

	MonitorReplicaConnectionIfPossible(conn, input)
	AddPropagationIfPossible(input, respOutput)

	return parser.ParseOutput(respOutput)

}

func (cR *HandlerFactory) Process(conn net.Conn, input *model.RespValue) {
	command := strings.ToUpper(strings.TrimSpace(input.Command))

	slog.Info("Start Processing Silently", slog.Any("command", input.Command))

	iCommand := getOrDefault(cR.handlers, command, &UnknownCommand{CommandName: "UNKNOWN"})
	// process the output
	result := iCommand.Process(input)

	if command == "REPLCONF" {
		output := parser.ParseOutput(result)
		slog.Info("Sending back from replica ", slog.Any("details", output))
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
	// it is a write command so we can increment the master offset
	replication.GetServerState().IncrementMasterState(util.GetBytes(input))
}

func (s *SpecialRegistry) Contains(command string) bool {
	command = strings.ToUpper(strings.TrimSpace(command))
	_, ok := s.commands[command]
	return ok
}

func (s *SpecialRegistry) Process(conn net.Conn, input *model.RespValue) string {

	command := strings.ToUpper(strings.TrimSpace(input.Command))
	slog.Any("Start Processing special", slog.Any("command", command))
	executableCommand := s.commands[command]
	result := executableCommand.Execute(conn, input)
	return parser.ParseOutput(result)
}

package shared

import (
	"net"

	"github.com/archstrap/cache-server/pkg/model"
)

type CommandProcessor interface {
	Process(conn net.Conn, input *model.RespValue)
}

var CommandProcessorInstance CommandProcessor

func SetCommandProcessor(processor CommandProcessor) {
	CommandProcessorInstance = processor
}

func GetCommandProcessor() CommandProcessor {
	return CommandProcessorInstance
}

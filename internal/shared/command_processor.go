package shared

import "github.com/archstrap/cache-server/pkg/model"

type CommandProcessor interface {
	ProcessCommandsSilently(input *model.RespValue) string
}

var CommandProcessorInstance CommandProcessor

func SetCommandProcessor(processor CommandProcessor) {
	CommandProcessorInstance = processor
}

func GetCommandProcessor() CommandProcessor {
	return CommandProcessorInstance
}

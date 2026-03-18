package command

import (
	"net"

	"github.com/archstrap/cache-server/internal/shared"
	"github.com/archstrap/cache-server/pkg/model"
)

type PingCommand struct{}

func init() {
	SRegistry.commands["PING"] = &PingCommand{}
}

func (pc *PingCommand) Name() string {
	return "PING"
}

func (pc *PingCommand) Execute(conn net.Conn, value *model.RespValue) *model.RespOutput {

	if shared.GetChannelStore().IsSubscribed(conn) {
		return model.NewRespOutput(model.TypeArray, []any{"pong", ""})
	}

	return model.NewRespOutput(model.TypeSimpleString, "PONG")
}

package command

import (
	"net"

	"github.com/archstrap/cache-server/internal/shared"
	"github.com/archstrap/cache-server/pkg/model"
)

type Subscribe struct{}

func init() {
	SRegistry.commands["SUBSCRIBE"] = &Subscribe{}
}

func (c *Subscribe) Name() string {
	return "SUBSCRIBE"
}

func (c *Subscribe) Execute(conn net.Conn, input *model.RespValue) *model.RespOutput {

	args := input.ArgsToStringSlice()
	if len(args) < 2 {
		return model.NewWrongNumberOfOutput(c.Name())
	}

	channelName := args[1]
	channelStore := shared.GetChannelStore()
	result := channelStore.Subscribe(conn, channelName)

	return model.NewRespOutput(model.TypeArray, result)
}

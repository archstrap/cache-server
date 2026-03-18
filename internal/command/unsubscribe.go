package command

import (
	"net"

	"github.com/archstrap/cache-server/internal/shared"
	"github.com/archstrap/cache-server/pkg/model"
)

type Unsubscribe struct{}

var cUnsubscribe = &Unsubscribe{}

func init() {
	SRegistry.commands[cUnsubscribe.Name()] = cUnsubscribe
}

func (c *Unsubscribe) Name() string {
	return "UNSUBSCRIBE"
}

func (c *Unsubscribe) Execute(conn net.Conn, input *model.RespValue) *model.RespOutput {

	args := input.ArgsToStringSlice()
	if len(args) < 2 {
		return model.NewWrongNumberOfOutput(c.Name())
	}

	store := shared.GetChannelStore()
	channelName := args[1]
	result := store.Unsubscribe(conn, channelName)
	return model.NewRespOutput(model.TypeArray, result)

}

package command

import (
	"net"

	"github.com/archstrap/cache-server/internal/shared"
	"github.com/archstrap/cache-server/pkg/model"
)

type Publish struct{}

func init() {
	SRegistry.commands["PUBLISH"] = &Publish{}
}

func (c *Publish) Name() string {
	return "PUBLISH"
}

func (c *Publish) Execute(conn net.Conn, input *model.RespValue) *model.RespOutput {
	args := input.ArgsToStringSlice()
	if len(args) < 3 {
		return model.NewWrongNumberOfOutput(c.Name())
	}

	store := shared.GetChannelStore()
	channelName := args[1]
	messages := args[2]

	result := store.Publish(channelName, messages)
	return model.NewRespOutput(model.TypeInteger, result)

}

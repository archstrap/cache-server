package command

import "github.com/archstrap/cache-server/pkg/model"

type Multi struct{}

var cMulti = &Multi{}

func init() {
	Registry.RegisterCommand(cMulti)
}

func (c *Multi) Name() string {
	return "MULTI"
}

func (c *Multi) Process(input *model.RespValue) *model.RespOutput {
	return model.NewRespOutput(model.TypeSimpleString, "OK")
}

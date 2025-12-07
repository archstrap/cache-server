package command

import (
	"github.com/archstrap/cache-server/pkg/model"
)

type ICommand interface {
	Process(value *model.RespValue) *model.RespOutput
	Name() string
}

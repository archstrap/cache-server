package command

import (
	"net"

	"github.com/archstrap/cache-server/pkg/model"
)

type SCommand interface {
	Execute(conn net.Conn, input *model.RespValue) *model.RespOutput
}

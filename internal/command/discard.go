package command

import (
	"net"

	"github.com/archstrap/cache-server/internal/shared"
	"github.com/archstrap/cache-server/pkg/model"
)

type Discard struct{}

var cDiscard = &Discard{}

func init() {
	GetSpecialRegistry().commands["DISCARD"] = cDiscard
}

func (d *Discard) Execute(conn net.Conn, input *model.RespValue) *model.RespOutput {
	mts := shared.GetMultiTransactionStore()
	if !mts.IsTransactionInitialized(conn) {
		return model.NewRespOutput(model.TypeError, "ERR DISCARD without MULTI")
	}
	defer mts.Remove(conn)
	mts.Discard(conn)
	return model.NewRespOutput(model.TypeSimpleString, "OK")
}

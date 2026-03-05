package command

import (
	"net"

	"github.com/archstrap/cache-server/internal/shared"
	"github.com/archstrap/cache-server/pkg/model"
)

type Exec struct{}

var cExec = &Exec{}

func init() {
	SRegistry.commands["EXEC"] = cExec
}

func (c *Exec) Execute(conn net.Conn, input *model.RespValue) *model.RespOutput {

	mts := shared.GetMultiTransactionStore()

	if !mts.IsTransactionInitialized(conn) {
		return model.NewRespOutput(model.TypeError, "ERR EXEC without MULTI")
	}

	defer mts.Remove(conn)

	if !mts.AreQueuedCommandsAvailable(conn) {
		return model.NewRespOutput(model.TypeArray, []string{})
	}

	return nil
}

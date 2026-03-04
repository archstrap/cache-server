package shared

import (
	"log/slog"
	"net"

	"github.com/archstrap/cache-server/pkg/model"
)

type MultiTransaction struct {
	details map[net.Conn]*Transaction
}

type Transaction struct {
	commands      chan *model.RespValue
	isInitialized bool
	isClosed      bool
}

func NewTransaction() *Transaction {
	return &Transaction{
		commands:      make(chan *model.RespValue, 25),
		isInitialized: true,
		isClosed:      false,
	}
}

var SMultiTransaction = &MultiTransaction{
	details: make(map[net.Conn]*Transaction),
}

func (c *MultiTransaction) Add(conn net.Conn) {
	c.details[conn] = NewTransaction()
	slog.Info("added connection ", slog.Any("details", conn.RemoteAddr()))
}

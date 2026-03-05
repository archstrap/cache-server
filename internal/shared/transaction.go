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

func GetMultiTransactionStore() *MultiTransaction {
	return SMultiTransaction
}

func (c *MultiTransaction) Add(conn net.Conn) {
	if c.IsTransactionInitialized(conn) {
		return
	}
	c.details[conn] = NewTransaction()
	slog.Info("Connection gets added in Multi session transactions", slog.Any("details", conn.RemoteAddr()))
}

func (c *MultiTransaction) IsTransactionInitialized(conn net.Conn) bool {
	_, ok := c.details[conn]
	return ok
}

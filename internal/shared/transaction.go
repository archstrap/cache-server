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
	commands      []*model.RespValue
	isInitialized bool
	isClosed      bool
}

func (t *Transaction) IsEmpty() bool {
	return len(t.commands) == 0
}

func (t *Transaction) AddCommand(input *model.RespValue) {
	t.commands = append(t.commands, input)
}

func NewTransaction() *Transaction {
	return &Transaction{
		commands:      make([]*model.RespValue, 0),
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

func (c *MultiTransaction) AreQueuedCommandsAvailable(conn net.Conn) bool {
	transactionDetails, ok := c.details[conn]
	if !ok {
		slog.Info("Unable to get")
		return false
	}
	return !transactionDetails.IsEmpty()
}

func (c *MultiTransaction) Remove(conn net.Conn) {
	delete(c.details, conn)
}

func (c *MultiTransaction) AddCommand(conn net.Conn, input *model.RespValue) {
	c.details[conn].AddCommand(input)
}

func (c *MultiTransaction) GetCommands(conn net.Conn) []*model.RespValue {
	return c.details[conn].commands
}

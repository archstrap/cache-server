package shared

import (
	"fmt"
	"log/slog"
	"net"
)

type ChannelInfo struct {
	channels []string
	count    int
}

func NewChannelInfo() *ChannelInfo {
	return &ChannelInfo{
		channels: make([]string, 0),
		count:    0,
	}
}

func (c *ChannelInfo) Subscribe(channelName string) {
	c.channels = append(c.channels, channelName)
	c.count = len(c.channels)
}

func (c *ChannelInfo) ChannelCount() int {
	return c.count
}

/* ---------------- [ Channel Store ] -------------------------*/

type ChannelStore struct {
	details map[net.Conn]*ChannelInfo
}

var cChannelStore *ChannelStore = &ChannelStore{
	details: make(map[net.Conn]*ChannelInfo),
}

func GetChannelStore() *ChannelStore {
	return cChannelStore
}

func (c *ChannelStore) Subscribe(conn net.Conn, channelName string) []any {
	if c.details[conn] == nil {
		c.details[conn] = NewChannelInfo()
	}
	c.details[conn].Subscribe(channelName)
	return []any{"subscribe", channelName, c.SubcriptionCount(conn)}
}

func (c *ChannelStore) SubcriptionCount(conn net.Conn) int {
	return c.details[conn].ChannelCount()
}

func (c *ChannelStore) Remove(conn net.Conn) {
	if !c.IsSubscribed(conn) {
		return
	}
	channelDetails := c.details[conn].channels
	delete(c.details, conn)
	slog.Info(fmt.Sprintf("Removed connection: %s with channels: %v from subscription", conn.RemoteAddr(), channelDetails))
}

func (c *ChannelStore) IsSubscribed(conn net.Conn) bool {
	_, ok := c.details[conn]
	return ok
}

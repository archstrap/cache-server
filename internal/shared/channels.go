package shared

import (
	"fmt"
	"log/slog"
	"net"
	"slices"

	"github.com/archstrap/cache-server/pkg/model"
	"github.com/archstrap/cache-server/pkg/parser"
)

type ChannelInfo struct {
	channels []string
	count    int
}

func NewChannelInfo() *ChannelInfo {
	return &ChannelInfo{
		channels: make([]string, 0),
	}
}

func (c *ChannelInfo) Subscribe(channelName string) {
	c.channels = append(c.channels, channelName)
	c.count = len(c.channels)
}

func (c *ChannelInfo) Unsubscribe(channelName string) {
	index := slices.Index(c.channels, channelName)
	if index == -1 {
		return
	}
	c.channels = append(c.channels[:index], c.channels[index+1:]...)

}

func (c *ChannelInfo) ChannelCount() int {
	return len(c.channels)
}

/* ---------------- [ Channel Store ] -------------------------*/

type ChannelStore struct {
	details  map[net.Conn]*ChannelInfo
	channels map[string][]net.Conn
}

var cChannelStore *ChannelStore = &ChannelStore{
	details:  make(map[net.Conn]*ChannelInfo),
	channels: make(map[string][]net.Conn),
}

func GetChannelStore() *ChannelStore {
	return cChannelStore
}

func (c *ChannelStore) Subscribe(conn net.Conn, channelName string) []any {
	if c.details[conn] == nil {
		c.details[conn] = NewChannelInfo()
	}
	c.details[conn].Subscribe(channelName)
	c.channels[channelName] = append(c.channels[channelName], conn)
	return []any{"subscribe", channelName, c.SubcriptionCount(conn)}
}

func (c *ChannelStore) Unsubscribe(conn net.Conn, channelName string) []any {

	channelDetails := c.details[conn]
	channelDetails.Unsubscribe(channelName)

	for i, localConn := range c.channels[channelName] {
		if localConn == conn {
			c.channels[channelName] = append(c.channels[channelName][:i], c.channels[channelName][i+1:]...)
			break
		}
	}

	if len(c.channels[channelName]) == 0 {
		delete(c.channels, channelName)
	}

	return []any{"unsubscribe", channelName, c.SubcriptionCount(conn)}
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

	for i := range channelDetails {
		channelName := channelDetails[i]
		for j, localConn := range c.channels[channelName] {
			if localConn == conn {
				c.channels[channelName] = append(c.channels[channelName][:j], c.channels[channelName][j+1:]...)
				break
			}
		}

		if len(c.channels[channelName]) == 0 {
			delete(c.channels, channelName)
		}
	}

	slog.Info(fmt.Sprintf("Removed connection: %s with channels: %v from subscription", conn.RemoteAddr(), channelDetails))
}

func (c *ChannelStore) IsSubscribed(conn net.Conn) bool {
	_, ok := c.details[conn]
	return ok
}

func (c *ChannelStore) Publish(channelName, messages string) int {

	connections := c.channels[channelName]

	for _, conn := range connections {

		finalMessage := []any{"message", channelName, messages}
		result := parser.ParseOutput(model.NewRespOutput(model.TypeArray, finalMessage))
		conn.Write([]byte(result))
	}

	return len(connections)
}

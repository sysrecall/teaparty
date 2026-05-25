package room

import (
	"context"

	"github.com/coder/websocket"
	"github.com/google/uuid"
)

type Client struct {
	Id      string
	Room    *Room
	Conn    *websocket.Conn
	Matched chan struct{}
	Send    chan []byte
}

func (c *Client) WritePump(ctx context.Context) {
	for msg := range c.Send {
		c.Conn.Write(ctx, websocket.MessageText, msg)
	}

	c.Conn.Close(websocket.StatusGoingAway, "peer disconnected")
}

type ClientState int

const (
	Waiting ClientState = iota
	Matched
	Disconnected
)

type Room struct {
	Id          string
	Client1     *Client
	Client2     *Client
	TurnServers string
}

func NewRoom(turnServers string) *Room {
	id := uuid.NewString()
	return &Room{
		Id:          id,
		TurnServers: turnServers,
	}
}

func (room *Room) Peer(client *Client) *Client {
	if room.Client1 == client {
		return room.Client2
	}
	return room.Client1
}

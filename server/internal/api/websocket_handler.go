package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"
	"golang.org/x/time/rate"
)


type Client struct {
	Id string
	Room *Room
	Conn *websocket.Conn
	Matched chan struct{}
	// State ClientState
	// Send chan []byte
	// IceCandidates any
	// Sdf           any
}

type ClientState int

const (
	Waiting ClientState = iota
	Matched
	Disconnected
)

type Room struct {
	Id string
	Client1 *Client
	Client2 *Client
}

func NewRoom() *Room {
	id := uuid.NewString()
	return &Room{
		Id: id,
	}
}

func (room *Room) Peer(client *Client) *Client {
	if room.Client1 == client {
		return room.Client2
	}
	return room.Client1
}


type Queue interface {
	Enqueue(*websocket.Conn) *Client
}

type WebsocketHandler struct {
	logger *log.Logger
	Queue Queue
}

func NewWebsocketHandler(logger *log.Logger, queue Queue) *WebsocketHandler {
	return &WebsocketHandler{
		logger: logger,
		Queue: queue,
	}
}

func (wh *WebsocketHandler) HandleWebsocket(w http.ResponseWriter, r *http.Request) {
	connection, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		// Subprotocols: []string{"echo"},
	})
	if err != nil {
		wh.logger.Printf("%v", err)
		return
	}

	client := wh.Queue.Enqueue(connection)

	// block until there is a match or client disconnect
	select {
	case <-client.Matched:
		// paired
	case <-r.Context().Done():
		// disconnected
		connection.Close(websocket.StatusGoingAway, "disconnected while waiting")
		return
	}
 
	// notify client about match
	ctx := r.Context()
	err = wh.WriteToConnection(ctx, connection, Message{
		MessageType:    "matched",
		MessageContent: client.Room.Id,
	})
	if err != nil {
		wh.logger.Printf("failed to send matched message: %v", err)
		return
	}
 
	// relay messages
	peer := client.Room.Peer(client)
	for {
		_, data, err := connection.Read(ctx)
		if err != nil {
			wh.logger.Printf("client %v disconnected: %v", client.Id, err)
			peer.Conn.Close(websocket.StatusGoingAway, "peer disconnected")
			return
		}
		if err := peer.Conn.Write(ctx, websocket.MessageText, data); err != nil {
			wh.logger.Printf("failed to relay to peer: %v", err)
			return
		}
	}
}

// echo reads from the WebSocket connection and then writes
// the received message back to it.
// The entire function has 60m to complete.
func echo(connection *websocket.Conn, limiter *rate.Limiter) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute * 60)
	defer cancel()

	err := limiter.Wait(ctx)
	if err != nil {
		return err
	}

	messageType, r, err := connection.Reader(ctx)
	if err != nil {
		return err
	}

	writer, err := connection.Writer(ctx, messageType)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, r)
	if err != nil {
		return fmt.Errorf("failed to io.Copy: %w", err)
	}

	err = writer.Close()
	return err
}

type Message struct {
	MessageType string `json:"type"`
	MessageContent string `json:"content"`
}

func (wh *WebsocketHandler) WriteToConnection(ctx context.Context, connection *websocket.Conn, message Message) error {
	data, err := json.Marshal(message)

	if err != nil {
		return fmt.Errorf("unable to serialize message: %w", err)
	}

	return connection.Write(ctx, websocket.MessageText, data)
}
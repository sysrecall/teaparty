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
	"github.com/joho/godotenv"
	"golang.org/x/time/rate"

	"teaparty/internal/room"
)

type Queue interface {
	Enqueue(*websocket.Conn) *room.Client
	Skip(*room.Client)
}

type WebsocketHandler struct {
	logger *log.Logger
	Queue  Queue
}

func NewWebsocketHandler(logger *log.Logger, queue Queue) *WebsocketHandler {
	// load .env file to os
	godotenv.Load()

	return &WebsocketHandler{
		logger: logger,
		Queue:  queue,
	}
}

const READ_LIMIT_BYTES = 1024 * 10

func (wh *WebsocketHandler) HandleWebsocket(w http.ResponseWriter, r *http.Request) {
	// create connection
	connection, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		// Subprotocols: []string{"echo"},
		OriginPatterns: []string{"*"},
	})

	if err != nil {
		wh.logger.Printf("%v", err)
		return
	}

	// config
	connection.SetReadLimit(READ_LIMIT_BYTES)

	// send to waiting queue
	client := wh.Queue.Enqueue(connection)

	// start write pump
	go client.WritePump(r.Context())

	// block until there is a match or client disconnect
	select {
	case <-client.Matched:
		// paired
	case <-r.Context().Done():
		// disconnected
		client.Conn.Close(websocket.StatusGoingAway, "disconnected while waiting")
		return
	}

	// notify client about match
	wh.sendInitOffer(client)

	// relay messages between clinets, blocking
	wh.relayMessages(r, client)
}

func (wh *WebsocketHandler) sendInitOffer(client *room.Client) {

	var messageContent string
	if client == client.Room.Client1 {
		messageContent = "offerer"
	} else {
		messageContent = "answerer"
	}

	message := Message{
		MessageType:    "match",
		MessageContent: messageContent,
	}

	messageJson, err := json.Marshal(message)
	if err != nil {
		fmt.Printf("unable to serialize message: %v", err)
	}

	client.Send <- messageJson
}

func (wh *WebsocketHandler) relayMessages(r *http.Request, client *room.Client) {
	peer := client.Room.Peer(client)

	// send turn servers
	if client.Room.TurnServers != "" {
		msg, err := json.Marshal(Message{
			MessageType:    "servers",
			MessageContent: client.Room.TurnServers,
		})

		if err == nil {
			client.Send <- msg
		}
	}

	// relay messages
	for {
		// read from client
		// if there are any error we skip ?
		// otherwise send the message to the
		_, data, err := client.Conn.Read(r.Context())
		if err != nil {
			wh.logger.Printf("client %v disconnected: %v", client.Id, err)
			skipMsg, _ := json.Marshal(Message{MessageType: "skip"})
			// peer.Conn.Write(relayContext, websocket.MessageText, skipMsg)
			peer.Send <- skipMsg
			// peer.Conn.Close(websocket.StatusGoingAway, "peer disconnected")
			close(client.Send)
			close(peer.Send)
			return
		}

		peer.Send <- data
	}

}

// echo reads from the WebSocket connection and then writes
// the received message back to it.
// The entire function has 60m to complete.
func echo(connection *websocket.Conn, limiter *rate.Limiter) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*60)
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
	MessageType    string `json:"type"`
	MessageContent string `json:"message"`
}

func (wh *WebsocketHandler) WriteToConnection(ctx context.Context, connection *websocket.Conn, message Message) error {
	data, err := json.Marshal(message)

	if err != nil {
		return fmt.Errorf("unable to serialize message: %w", err)
	}

	return connection.Write(ctx, websocket.MessageText, data)
}

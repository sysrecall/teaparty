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
	"golang.org/x/time/rate"
)


type Queue interface {
	Enqueue(*websocket.Conn)
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

	wh.Queue.Enqueue(connection)

	<- r.Context().Done()

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
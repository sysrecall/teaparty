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

type WebsocketHandler struct {
	logger *log.Logger
	ConnectionsChannel chan *websocket.Conn
}

func NewWebsocketHandler(logger *log.Logger) *WebsocketHandler {
	return &WebsocketHandler{
		logger: logger,
		ConnectionsChannel: make(chan *websocket.Conn, 100),
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

	wh.ConnectionsChannel <- connection

	defer connection.CloseNow()

	// if connection.Subprotocol() != "echo" {
	// 	connection.Close(websocket.StatusPolicyViolation, "client must speak the echo subprotocol")
	// 	return
	// }

	limiter := rate.NewLimiter(rate.Every(time.Millisecond*100), 10)
	for {
		err = echo(connection, limiter)
		if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
			return
		}
		if err != nil {
			wh.logger.Printf("failed to echo with %v: %v", r.RemoteAddr, err)
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

func (wh *WebsocketHandler) WriteToConnection(toConnection *websocket.Conn, message Message, limiter *rate.Limiter) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute * 60)
	defer cancel()

	err := limiter.Wait(ctx)
	if err != nil {
		return err
	}

	messageType, _, err := toConnection.Reader(ctx)
	if err != nil {
		return err
	}

	writer, err := toConnection.Writer(ctx, messageType)
	if err != nil {
		return err
	}

	messageSerialized, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("unable to serialize message: %w", err)
	}

	_, err = writer.Write(messageSerialized)
	if err != nil {
		return fmt.Errorf("failed to write: %w", err)
	}

	err = writer.Close()
	return err
}
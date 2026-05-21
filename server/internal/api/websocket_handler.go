package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"golang.org/x/time/rate"
)

type Client struct {
	Id      string
	Room    *Room
	Conn    *websocket.Conn
	Matched chan struct{}
	// IceCandidates any
	// Offer         any
	// State ClientState
	// Send chan []byte
}

type ClientState int

const (
	Waiting ClientState = iota
	Matched
	Disconnected
)

type Room struct {
	Id      string
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

type turnCredentials struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	ExpiryInSeconds int    `json:"expiryInSeconds"`
	Label           string `json:"label,omitempty"`
	ApiKey          string `json:"apiKey"`
}

type turnServer struct {
	Urls        string `json:"urls"`
	Username    string `json:"username"`
	Credentials string `json:"credentials"`
}

func (wh *WebsocketHandler) getTurnCredentials() (turnCredentials, error) {
	// load env variables
	domain, foundDomain := os.LookupEnv("METERED_DOMAIN")
	if !foundDomain {
		wh.logger.Fatal("TURN server domain does not exist in env")
	}

	turnSecret, foundTurnSecret := os.LookupEnv("METERED_SECRET_KEY")
	if !foundTurnSecret {
		wh.logger.Fatal("TURN server secret does not exist in env")

	}

	turnExipry, foundTurnExpiry := os.LookupEnv("TURN_CREDENTIALS_EXPIRY_SECONDS")
	if !foundTurnExpiry {
		wh.logger.Fatal("TURN server expiry does not exist in env")
	}

	// request credentials
	url := fmt.Sprintf("https://%v/api/v1/turn/credential", domain)
	requestUrl := fmt.Sprintf("%v?secretKey=%v", url, turnSecret)

	data := map[string]string{"expiryInSeconds": turnExipry}
	jsonData, _ := json.Marshal(data)

	res, err := http.Post(requestUrl, "application/json", bytes.NewBuffer(jsonData))

	if err != nil {
		wh.logger.Printf("Error requesting turn credentials: %v", err)
	}

	defer res.Body.Close()

	var turnCredentials turnCredentials

	err = json.NewDecoder(res.Body).Decode(&turnCredentials)

	if err != nil {
		wh.logger.Printf("Unable to parse turn credetial response: %v", err)
		return turnCredentials, err
	}

	if res.StatusCode != http.StatusOK {
		return turnCredentials, fmt.Errorf("unexpected status %d", res.StatusCode)
	}

	return turnCredentials, nil
}

func (wh *WebsocketHandler) getTurnServers(apiKey string) (string, error) {
	// load env variables
	domain, foundDomain := os.LookupEnv("METERED_DOMAIN")
	if !foundDomain {
		wh.logger.Fatal("TURN server domain does not exist in env")
	}

	// request ice servers
	url := fmt.Sprintf("https://%v/api/v1/turn/credentials", domain)
	requestUrl := fmt.Sprintf("%v?apiKey=%v", url, apiKey)

	res, err := http.Get(requestUrl)

	if err != nil {
		wh.logger.Printf("Error requesting turn servers: %v", err)
		return "", err
	}

	defer res.Body.Close()

	// map servers
	turnServers, err := io.ReadAll(res.Body)
	if err != nil {
		wh.logger.Printf("Error parsing body: %v", err)
		return "", err
	}

	if res.StatusCode != 200 {
		wh.logger.Printf("%s", string(turnServers))
		return "", fmt.Errorf("Invalid Request")
	}

	return string(turnServers), nil
}

func (wh *WebsocketHandler) HandleWebsocket(w http.ResponseWriter, r *http.Request) {
	connection, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		// Subprotocols: []string{"echo"},
		OriginPatterns: []string{"localhost*"},
	})

	if err != nil {
		wh.logger.Printf("%v", err)
		return
	}

	connection.SetReadLimit(READ_LIMIT_BYTES)

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

	err = wh.WriteToConnection(ctx, connection, message)
	if err != nil {
		wh.logger.Printf("failed to send matched message: %v", err)
		return
	}

	// relay messages
	peer := client.Room.Peer(client)
	relayContext, cancelRelay := context.WithCancel(context.Background())
	defer cancelRelay()

	// get turn credentials
	turnCredentials, turnCredentialsError := wh.getTurnCredentials()

	if turnCredentialsError == nil {
		// get turn servers
		turnServers, turnServersError := wh.getTurnServers(turnCredentials.ApiKey)

		// send turn server list
		if turnServersError == nil {
			msg, err := json.Marshal(Message{
				MessageType:    "servers",
				MessageContent: turnServers,
			})

			if err == nil {
				client.Conn.Write(relayContext, websocket.MessageText, msg)
			}
		}
	}

	for {
		_, data, err := connection.Read(ctx)
		if err != nil {
			wh.logger.Printf("client %v disconnected: %v", client.Id, err)
			peer.Conn.Close(websocket.StatusGoingAway, "peer disconnected")
			return
		}
		if err := peer.Conn.Write(relayContext, websocket.MessageText, data); err != nil {
			wh.logger.Printf("failed to relay to peer: %v", err)
			return
		}
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

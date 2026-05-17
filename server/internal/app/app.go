package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"teaparty/internal/api"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"
)

type Client struct {
	Id string
	Room *Room
	Conn *websocket.Conn
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

type Application struct {
	Logger *log.Logger
	WebsocketHandler *api.WebsocketHandler
	WaitingQueue chan *websocket.Conn
	Rooms map[string]*Room
	mu sync.Mutex
}

func (app *Application) Enqueue(connection *websocket.Conn) {
	app.WaitingQueue <- connection
}

func NewApplication() *Application {
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	app := &Application{
		Logger: logger,
		WaitingQueue: make(chan *websocket.Conn, 100),
	}

	websocketHandler := api.NewWebsocketHandler(logger, app)
	app.WebsocketHandler = websocketHandler

	return app
}

func (app *Application) HealthCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Status is available\n")
}

func (app *Application) MakePairs() {
	for {
		connection1 := <- app.WaitingQueue
		connection2 := <- app.WaitingQueue

		// limiter := rate.NewLimiter(rate.Every(time.Second * 2), 3)



		// create a room and populate the room

		room := NewRoom()

		client1 := &Client{
			Room: room,
			Conn: connection1,
		}

		client2 := &Client{
			Room: room,
			Conn: connection2,
		}

		room.Client1 = client1
		room.Client2 = client2

		// update rooms
		app.mu.Lock()
		app.Rooms[room.Id] = room
		app.mu.Unlock()


		// send both clients info about the other
		message1 := api.Message {
			MessageType: "matched",
			MessageContent: room.Id,
		}

		message2 := api.Message {
			MessageType: "matched",
			MessageContent: room.Id,
		}

		ctx, err := context.WithTimeout(context.Background(), time.Second * 10)
		if err != nil {
			fmt.Printf("could not create context: %w", err)
			continue
		}

		app.WebsocketHandler.WriteToConnection(ctx, connection1, message1)
		app.WebsocketHandler.WriteToConnection(ctx, connection2, message2)
	}
}

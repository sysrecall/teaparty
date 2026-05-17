package app

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"teaparty/internal/api"
	"time"

	"github.com/coder/websocket"
	"golang.org/x/time/rate"
)

type Client struct {
	Conn *websocket.Conn
	Peer *Client
	State ClientState
	Send chan []byte
	// IceCandidates any
	// Sdf           any
}

type ClientState int

const (
	Waiting ClientState = iota
	Matched
	Disconnected
)

type Application struct {
	Logger *log.Logger
	WebsocketHandler *api.WebsocketHandler
	WaitingQueue chan *websocket.Conn
	Clients []*Client
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
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		<- ticker.C
		
		connection1 := <- app.WaitingQueue
		connection2 := <- app.WaitingQueue

		limiter := rate.NewLimiter(rate.Every(time.Second * 2), 3)

		message1 := api.Message {
			MessageType: "client",
			MessageContent: fmt.Sprint(connection2),
		}

		message2 := api.Message {
			MessageType: "client",
			MessageContent: fmt.Sprint(connection1),
		}

		app.WebsocketHandler.WriteToConnection(connection1, message1, limiter)
		app.WebsocketHandler.WriteToConnection(connection2, message2, limiter)

	}
}

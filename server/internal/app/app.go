package app

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"teaparty/internal/api"

	"github.com/coder/websocket"
	"github.com/google/uuid"
)

type Application struct {
	Logger           *log.Logger
	WebsocketHandler *api.WebsocketHandler
	WaitingQueue     chan *api.Client
	Rooms            map[string]*api.Room
	mu               sync.Mutex
}

func (app *Application) Enqueue(connection *websocket.Conn) *api.Client {
	client := &api.Client{
		Id:      uuid.NewString(),
		Conn:    connection,
		Matched: make(chan struct{}),
	}

	app.WaitingQueue <- client
	return client
}

func NewApplication() *Application {
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	app := &Application{
		Logger:       logger,
		WaitingQueue: make(chan *api.Client, 100),
		mu:           sync.Mutex{},
		Rooms:        make(map[string]*api.Room),
	}

	app.WebsocketHandler = api.NewWebsocketHandler(logger, app)

	return app
}

func (app *Application) HealthCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Status is available\n")
}

func (app *Application) MakePairs() {
	for {
		client1 := <-app.WaitingQueue
		client2 := <-app.WaitingQueue

		// limiter := rate.NewLimiter(rate.Every(time.Second * 2), 3)

		// create a room and populate the room

		room := &api.Room{
			Id:      uuid.NewString(),
			Client1: client1,
			Client2: client2,
		}

		client1.Room = room
		client2.Room = room

		// update rooms
		app.mu.Lock()
		app.Rooms[room.Id] = room
		app.mu.Unlock()

		// close matching channel
		close(client1.Matched)
		close(client2.Matched)

	}
}

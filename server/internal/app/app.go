package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

func (app *Application) DeleteRoom(roomId string) {
	delete(app.Rooms, roomId)
}

func (app *Application) Skip(client *api.Client) {
	client1 := client.Room.Client1
	client2 := client.Room.Client2

	app.DeleteRoom(client.Room.Id)

	// queue for new pair
	app.WaitingQueue <- client1
	app.WaitingQueue <- client2
}

func (app *Application) MakePairs() {
	for {
		client1 := <-app.WaitingQueue
		client2 := <-app.WaitingQueue

		// assign turn servers on pair
		turnServers := app.fetchTurnServers()

		// create a room and populate the room
		room := &api.Room{
			Id:          uuid.NewString(),
			Client1:     client1,
			Client2:     client2,
			TurnServers: turnServers,
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

func (app *Application) fetchTurnServers() string {
	creds, err := app.getTurnCredentials()
	if err != nil {
		app.Logger.Printf("failed to get turn credentials: %v", err)
		return ""
	}

	servers, err := app.getTurnServers(creds.ApiKey)
	if err != nil {
		app.Logger.Printf("failed to get turn servers: %v", err)
		return ""
	}

	return servers
}

func (app *Application) getTurnCredentials() (turnCredentials, error) {
	// load env variables
	domain, foundDomain := os.LookupEnv("METERED_DOMAIN")
	if !foundDomain {
		app.Logger.Fatal("TURN server domain does not exist in env")
	}

	turnSecret, foundTurnSecret := os.LookupEnv("METERED_SECRET_KEY")
	if !foundTurnSecret {
		app.Logger.Fatal("TURN server secret does not exist in env")

	}

	turnExipry, foundTurnExpiry := os.LookupEnv("TURN_CREDENTIALS_EXPIRY_SECONDS")
	if !foundTurnExpiry {
		app.Logger.Fatal("TURN server expiry does not exist in env")
	}

	// request credentials
	url := fmt.Sprintf("https://%v/api/v1/turn/credential", domain)
	requestUrl := fmt.Sprintf("%v?secretKey=%v", url, turnSecret)

	data := map[string]string{"expiryInSeconds": turnExipry}
	jsonData, _ := json.Marshal(data)

	res, err := http.Post(requestUrl, "application/json", bytes.NewBuffer(jsonData))

	if err != nil {
		app.Logger.Printf("Error requesting turn credentials: %v", err)
	}

	defer res.Body.Close()

	var turnCredentials turnCredentials

	err = json.NewDecoder(res.Body).Decode(&turnCredentials)

	if err != nil {
		app.Logger.Printf("Unable to parse turn credetial response: %v", err)
		return turnCredentials, err
	}

	if res.StatusCode != http.StatusOK {
		return turnCredentials, fmt.Errorf("unexpected status %d", res.StatusCode)
	}

	return turnCredentials, nil
}

func (app *Application) getTurnServers(apiKey string) (string, error) {
	// load env variables
	domain, foundDomain := os.LookupEnv("METERED_DOMAIN")
	if !foundDomain {
		app.Logger.Fatal("TURN server domain does not exist in env")
	}

	// request ice servers
	url := fmt.Sprintf("https://%v/api/v1/turn/credentials", domain)
	requestUrl := fmt.Sprintf("%v?apiKey=%v", url, apiKey)

	res, err := http.Get(requestUrl)

	if err != nil {
		app.Logger.Printf("Error requesting turn servers: %v", err)
		return "", err
	}

	defer res.Body.Close()

	// map servers
	turnServers, err := io.ReadAll(res.Body)
	if err != nil {
		app.Logger.Printf("Error parsing body: %v", err)
		return "", err
	}

	if res.StatusCode != 200 {
		app.Logger.Printf("%s", string(turnServers))
		return "", fmt.Errorf("Invalid Request")
	}

	return string(turnServers), nil
}

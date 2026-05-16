package app

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"teaparty/internal/api"
	"time"

	"golang.org/x/time/rate"
)

type Application struct {
	Logger *log.Logger
	WebsocketHandler *api.WebsocketHandler
}

func NewApplication() *Application {
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	websocketHandler := api.NewWebsocketHandler(logger)

	return &Application{
		Logger: logger,
		WebsocketHandler: websocketHandler,
	}

}

func (app *Application) HealthCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Status is available\n")
}

func (app *Application) MakePair() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		<- ticker.C
		
		connection1 := <- app.WebsocketHandler.ConnectionsChannel
		connection2 := <- app.WebsocketHandler.ConnectionsChannel

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
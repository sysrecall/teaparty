package app

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"teaparty/internal/api"
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
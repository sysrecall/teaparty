package routes

import (
	"teaparty/internal/app"

	"github.com/go-chi/chi/v5"
)

func SetupRoutes(app *app.Application) *chi.Mux {
	r := chi.NewRouter()
	
	r.Get("/health", app.HealthCheck)
	r.Get("/ws", app.WebsocketHandler.HandleWebsocket)

	return r
}
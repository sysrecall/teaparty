package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"teaparty/internal/app"
	"teaparty/internal/routes"
	"time"
)

func main() {
	log.SetFlags(0)

	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return errors.New("please provide an address to listen on as the first argument")
	}

	listener, err := net.Listen("tcp", os.Args[1])
	if err != nil {
		return err
	}
	log.Printf("listening on http://%v", listener.Addr())

	app := app.NewApplication()
	handler := routes.SetupRoutes(app)

	server := &http.Server{
		Handler: handler,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}

	errorChannel := make(chan error, 1)
	go func() {
		errorChannel <- server.Serve(listener)
	}()

	signalsChannel := make(chan os.Signal, 1)
	signal.Notify(signalsChannel, os.Interrupt)

	select {
	case err := <-errorChannel:
		log.Printf("failed to serve: %v", err)
	case sig := <-signalsChannel:
		log.Printf("terminating: %v", sig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	return server.Shutdown(ctx)
}
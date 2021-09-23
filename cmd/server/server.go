package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"tcp-serv-test/internal/server"
	"time"
)

func main() {
	ctx := context.Background()
	arguments := os.Args
	if len(arguments) == 1 {
		log.Fatal("address must be provided")
	}
	log.Println("starting server")
	srv := server.New(arguments[1])
	go srv.Serve()
	waitStopSignal(ctx, srv)
}

func waitStopSignal(ctx context.Context, srv *server.Server) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	<-c
	log.Println("stopping server")
	ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	go func() {
		<-c
		log.Println("force stopping server")
		cancel()
	}()
	srv.Stop(ctx)

	log.Println("stopped")
}

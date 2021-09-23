package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tcp-serv-test/internal/client"
)

func main() {
	ctx := context.Background()
	arguments := os.Args
	if len(arguments) == 1 {
		log.Fatal("address must be provided")
	}
	log.Println("starting client")
	c := client.New(arguments[1])
	go c.Start()
	waitStopSignal(ctx, c)
}

func waitStopSignal(ctx context.Context, c *client.Client) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	<-ch
	log.Println("stopping client")
	ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	go func() {
		<-ch
		log.Println("force stopping client")
		cancel()
	}()
	c.Stop()

	log.Println("stopped")
}
